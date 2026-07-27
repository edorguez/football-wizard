package scraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

type Parser struct {
	client *Client
	log    *zap.Logger
}

func NewParser(client *Client, log *zap.Logger) *Parser {
	return &Parser{client: client, log: log}
}

func (p *Parser) ParseMatchResults(season int) ([]ScrapedMatch, error) {
	url := fmt.Sprintf("https://fbref.com/en/comps/24/%d/schedule/%d-Serie-A-Scores-and-Fixtures", season, season)
	p.log.Info("parsing match results", zap.String("url", url), zap.Int("season", season))

	p.client.Progress <- ProgressMsg{Text: fmt.Sprintf("fetching FBref season %d...", season)}
	html, err := p.client.FetchHTML(url)
	if err != nil {
		return nil, fmt.Errorf("scraping matches for season %d: %w", season, err)
	}

	p.client.Progress <- ProgressMsg{Text: fmt.Sprintf("parsing HTML with goquery (%d bytes)...", len(html))}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML for season %d: %w", season, err)
	}

	var matches []ScrapedMatch
	table := doc.Find("table#results tbody tr, table.stats_table tbody tr")
	rowCount := table.Length()
	p.client.Progress <- ProgressMsg{Text: fmt.Sprintf("found %d rows in match table", rowCount)}

	table.Each(func(i int, row *goquery.Selection) {
		if row.HasClass("thead") || row.Children().Length() < 10 {
			return
		}

		match := ScrapedMatch{}

		dateStr := strings.TrimSpace(row.Find("td[data-stat='date']").Text())
		if dateStr != "" {
			formats := []string{"2006-01-02", "Jan 2, 2006", "2006-01-02 15:04:05"}
			parsed := false
			for _, f := range formats {
				t, err := time.Parse(f, dateStr)
				if err == nil {
					match.Date = t
					parsed = true
					break
				}
			}
			if !parsed {
				p.log.Warn("unrecognized date format", zap.String("date", dateStr))
				return
			}
		}

		match.Season = season
		match.HomeTeam = strings.TrimSpace(row.Find("td[data-stat='home_team']").Text())
		match.AwayTeam = strings.TrimSpace(row.Find("td[data-stat='away_team']").Text())

		scoreStr := strings.TrimSpace(row.Find("td[data-stat='score']").Text())
		if scoreStr != "" && strings.Contains(scoreStr, "–") {
			parts := strings.Split(scoreStr, "–")
			if len(parts) == 2 {
				match.HomeGoals, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
				match.AwayGoals, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
			}
		}

		match.Referee = strings.TrimSpace(row.Find("td[data-stat='referee']").Text())

		attStr := strings.ReplaceAll(row.Find("td[data-stat='attendance']").Text(), ",", "")
		match.Attendance, _ = strconv.Atoi(strings.TrimSpace(attStr))

		matches = append(matches, match)

		if i%10 == 0 && i > 0 {
			p.client.Progress <- ProgressMsg{Text: fmt.Sprintf("parsed %d/%d rows...", i, rowCount)}
		}
	})

	p.client.Progress <- ProgressMsg{Text: fmt.Sprintf("parsed %d matches total", len(matches))}
	p.log.Info("parsed match results", zap.Int("count", len(matches)))
	return matches, nil
}

func (p *Parser) ParseMatchStats(url string) (*ScrapedMatch, error) {
	p.log.Info("parsing match stats", zap.String("url", url))

	p.client.Progress <- ProgressMsg{Text: fmt.Sprintf("fetching match stats page...")}
	html, err := p.client.FetchHTML(url)
	if err != nil {
		return nil, fmt.Errorf("fetching match stats: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing match stats HTML: %w", err)
	}

	match := &ScrapedMatch{}

	doc.Find("div#team_stats table tr").Each(func(i int, row *goquery.Selection) {
		stat := strings.TrimSpace(row.Find("th").Text())
		cells := make([]string, 0)
		row.Find("td").Each(func(j int, cell *goquery.Selection) {
			cells = append(cells, strings.TrimSpace(cell.Text()))
		})

		if len(cells) != 2 {
			return
		}

		homeVal, _ := strconv.ParseFloat(cells[0], 64)
		awayVal, _ := strconv.ParseFloat(cells[1], 64)

		switch strings.ToLower(stat) {
		case "possession":
			match.Possession = [2]float64{homeVal, awayVal}
		case "shots on goal":
			match.ShotsOnTarget = [2]int{int(homeVal), int(awayVal)}
		case "shot attempts":
			match.Shots = [2]int{int(homeVal), int(awayVal)}
		case "fouls":
			match.Fouls = [2]int{int(homeVal), int(awayVal)}
		case "yellow cards":
			match.YellowCards = [2]int{int(homeVal), int(awayVal)}
		case "red cards":
			match.RedCards = [2]int{int(homeVal), int(awayVal)}
		case "offsides":
			match.Offsides = [2]int{int(homeVal), int(awayVal)}
		case "corners":
			match.Corners = [2]int{int(homeVal), int(awayVal)}
		}
	})

	return match, nil
}

func (p *Parser) ParseExpectedGoals(url string) ([2]float64, error) {
	p.client.Progress <- ProgressMsg{Text: fmt.Sprintf("fetching expected goals data...")}
	html, err := p.client.FetchHTML(url)
	if err != nil {
		return [2]float64{}, fmt.Errorf("fetching expected goals: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return [2]float64{}, fmt.Errorf("parsing expected goals HTML: %w", err)
	}

	xg := [2]float64{}

	doc.Find("div#scoring table tbody tr").Each(func(i int, row *goquery.Selection) {
		player := strings.TrimSpace(row.Find("th a").Text())
		xgStr := strings.TrimSpace(row.Find("td[data-stat='xg']").Text())

		if player == "" || xgStr == "" {
			return
		}

		xgVal, err := strconv.ParseFloat(xgStr, 64)
		if err != nil {
			return
		}

		team := strings.TrimSpace(row.Find("td[data-stat='team']").Text())
		if strings.Contains(team, "Home") {
			xg[0] += xgVal
		} else {
			xg[1] += xgVal
		}
	})

	return xg, nil
}

func (p *Parser) ParseFixtures(season int) ([]ScrapedFixture, error) {
	url := fmt.Sprintf("https://fbref.com/en/comps/24/%d/schedule/%d-Serie-A-Scores-and-Fixtures", season, season)
	p.log.Info("parsing fixtures", zap.String("url", url), zap.Int("season", season))

	p.client.Progress <- ProgressMsg{Text: fmt.Sprintf("fetching fixtures for season %d...", season)}
	html, err := p.client.FetchHTML(url)
	if err != nil {
		return nil, fmt.Errorf("scraping fixtures for season %d: %w", season, err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing fixtures HTML: %w", err)
	}

	var fixtures []ScrapedFixture
	now := time.Now()

	p.client.Progress <- ProgressMsg{Text: fmt.Sprintf("parsing fixture rows...")}
	doc.Find("table#results tbody tr").Each(func(i int, row *goquery.Selection) {
		if row.HasClass("thead") {
			return
		}

		scoreStr := strings.TrimSpace(row.Find("td[data-stat='score']").Text())
		if scoreStr != "" {
			return
		}

		dateStr := strings.TrimSpace(row.Find("td[data-stat='date']").Text())

		fixture := ScrapedFixture{}
		formats := []string{"2006-01-02", "Jan 2, 2006"}
		parsed := false
		for _, f := range formats {
			t, err := time.Parse(f, dateStr)
			if err == nil {
				fixture.Date = t
				parsed = true
				break
			}
		}
		if !parsed {
			return
		}

		if fixture.Date.Before(now) {
			return
		}

		fixture.Season = season
		fixture.HomeTeam = strings.TrimSpace(row.Find("td[data-stat='home_team']").Text())
		fixture.AwayTeam = strings.TrimSpace(row.Find("td[data-stat='away_team']").Text())

		fixtures = append(fixtures, fixture)
	})

	p.client.Progress <- ProgressMsg{Text: fmt.Sprintf("found %d upcoming fixtures", len(fixtures))}
	p.log.Info("parsed fixtures", zap.Int("count", len(fixtures)))
	return fixtures, nil
}

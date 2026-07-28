package scraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func ParseMatchResults(season int, html string) ([]ScrapedMatch, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	var matches []ScrapedMatch

	doc.Find("table#matchlogs tbody tr").Each(func(_ int, row *goquery.Selection) {
		m, ok := parseMatchRow(season, row)
		if ok {
			matches = append(matches, m)
		}
	})

	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found in HTML")
	}

	return matches, nil
}

func parseMatchRow(season int, row *goquery.Selection) (ScrapedMatch, bool) {
	roundClass, exists := row.Attr("data-row")
	if !exists {
		return ScrapedMatch{}, false
	}

	round, err := strconv.Atoi(roundClass)
	if err != nil {
		return ScrapedMatch{}, false
	}

	dateStr := strings.TrimSpace(row.Find("td[data-stat='date']").Text())
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return ScrapedMatch{}, false
	}

	homeTeam := strings.TrimSpace(row.Find("td[data-stat='home_team'] a").Text())
	if homeTeam == "" {
		homeTeam = strings.TrimSpace(row.Find("td[data-stat='home_team']").Text())
	}

	awayTeam := strings.TrimSpace(row.Find("td[data-stat='away_team'] a").Text())
	if awayTeam == "" {
		awayTeam = strings.TrimSpace(row.Find("td[data-stat='away_team']").Text())
	}

	if homeTeam == "" || awayTeam == "" {
		return ScrapedMatch{}, false
	}

	scoreStr := strings.TrimSpace(row.Find("td[data-stat='score']").Text())
	var homeGoals, awayGoals int

	if strings.Contains(scoreStr, "–") {
		parts := strings.SplitN(scoreStr, "–", 2)
		homeGoals, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
		awayGoals, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
	}

	refereeName := strings.TrimSpace(row.Find("td[data-stat='referee'] a").Text())
	if refereeName == "" {
		refereeName = strings.TrimSpace(row.Find("td[data-stat='referee']").Text())
	}

	m := ScrapedMatch{
		Season:      season,
		Round:       round,
		Date:        date,
		HomeTeam:    homeTeam,
		AwayTeam:    awayTeam,
		HomeGoals:   homeGoals,
		AwayGoals:   awayGoals,
		RefereeName: refereeName,
		HomeShots:           parseIntAttr(row, "shots", "home_"),
		AwayShots:           parseIntAttr(row, "shots", "away_"),
		HomeShotsOnTarget:   parseIntAttr(row, "shots_on_target", "home_"),
		AwayShotsOnTarget:   parseIntAttr(row, "shots_on_target", "away_"),
		HomeCorners:         parseIntAttr(row, "corner_kicks", "home_"),
		AwayCorners:         parseIntAttr(row, "corner_kicks", "away_"),
		HomeYellowCards:     parseIntAttr(row, "yellow_cards", "home_"),
		AwayYellowCards:     parseIntAttr(row, "yellow_cards", "away_"),
		HomeRedCards:        parseIntAttr(row, "red_cards", "home_"),
		AwayRedCards:        parseIntAttr(row, "red_cards", "away_"),
	}

	return m, true
}

func parseIntAttr(row *goquery.Selection, stat, prefix string) *int {
	val := strings.TrimSpace(row.Find(fmt.Sprintf("td[data-stat='%s%s']", prefix, stat)).Text())
	if val == "" {
		return nil
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return nil
	}
	return &n
}

func ParseFixtures(season int, html string) ([]ScrapedFixture, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing fixtures HTML: %w", err)
	}

	var fixtures []ScrapedFixture

	doc.Find("table#fixtures tbody tr").Each(func(_ int, row *goquery.Selection) {
		f, ok := parseFixtureRow(season, row)
		if ok {
			fixtures = append(fixtures, f)
		}
	})

	return fixtures, nil
}

func parseFixtureRow(season int, row *goquery.Selection) (ScrapedFixture, bool) {
	roundClass, exists := row.Attr("data-row")
	if !exists {
		return ScrapedFixture{}, false
	}

	round, err := strconv.Atoi(roundClass)
	if err != nil {
		return ScrapedFixture{}, false
	}

	dateStr := strings.TrimSpace(row.Find("td[data-stat='date']").Text())
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return ScrapedFixture{}, false
	}

	homeTeam := strings.TrimSpace(row.Find("td[data-stat='home_team'] a").Text())
	if homeTeam == "" {
		homeTeam = strings.TrimSpace(row.Find("td[data-stat='home_team']").Text())
	}

	awayTeam := strings.TrimSpace(row.Find("td[data-stat='away_team'] a").Text())
	if awayTeam == "" {
		awayTeam = strings.TrimSpace(row.Find("td[data-stat='away_team']").Text())
	}

	if homeTeam == "" || awayTeam == "" {
		return ScrapedFixture{}, false
	}

	return ScrapedFixture{
		Season:   season,
		Round:    round,
		Date:     date,
		HomeTeam: homeTeam,
		AwayTeam: awayTeam,
	}, true
}

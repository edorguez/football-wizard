package scraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var errNoMatches = fmt.Errorf("no matches found in HTML")

func tableID(season int) string {
	return fmt.Sprintf("sched_%d_24_1", season)
}

func ParseMatchResults(season int, html string) ([]ScrapedMatch, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	table := doc.Find(fmt.Sprintf("table#%s", tableID(season)))
	if table.Length() == 0 {
		return nil, fmt.Errorf("table %q not found", tableID(season))
	}

	var matches []ScrapedMatch

	table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		score := strings.TrimSpace(row.Find("td[data-stat='score']").Text())
		if score == "vs" || score == "" {
			return
		}

		m, ok := parseMatchRow(season, row)
		if ok {
			matches = append(matches, m)
		}
	})

	if len(matches) == 0 {
		return nil, fmt.Errorf("no completed matches found: %w", errNoMatches)
	}

	return matches, nil
}

func parseMatchRow(season int, row *goquery.Selection) (ScrapedMatch, bool) {
	round, ok := parseGameweek(row)
	if !ok {
		return ScrapedMatch{}, false
	}

	date, ok := parseDate(row)
	if !ok {
		return ScrapedMatch{}, false
	}

	homeTeam := extractTeam(row, "home_team")
	if homeTeam == "" {
		return ScrapedMatch{}, false
	}

	awayTeam := extractTeam(row, "away_team")
	if awayTeam == "" {
		return ScrapedMatch{}, false
	}

	scoreStr := strings.TrimSpace(row.Find("td[data-stat='score']").Text())
	homeGoals, awayGoals, ok := parseScore(scoreStr)
	if !ok {
		return ScrapedMatch{}, false
	}

	refereeName := strings.TrimSpace(row.Find("td[data-stat='referee'] a").Text())
	if refereeName == "" {
		refereeName = strings.TrimSpace(row.Find("td[data-stat='referee']").Text())
	}

	return ScrapedMatch{
		Season:      season,
		Round:       round,
		Date:        date,
		HomeTeam:    homeTeam,
		AwayTeam:    awayTeam,
		HomeGoals:   homeGoals,
		AwayGoals:   awayGoals,
		RefereeName: refereeName,
	}, true
}

func parseGameweek(row *goquery.Selection) (int, bool) {
	sel := row.Find("th[data-stat='gameweek'], td[data-stat='gameweek']")
	text := strings.TrimSpace(sel.Text())
	if text == "" {
		return 0, false
	}
	n, err := strconv.Atoi(text)
	if err != nil {
		return 0, false
	}
	return n, true
}

func parseDate(row *goquery.Selection) (time.Time, bool) {
	dateStr := strings.TrimSpace(row.Find("td[data-stat='date']").Text())
	if dateStr == "" {
		return time.Time{}, false
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, false
	}
	return date, true
}

func extractTeam(row *goquery.Selection, stat string) string {
	name := strings.TrimSpace(row.Find(fmt.Sprintf("td[data-stat='%s'] a", stat)).Text())
	if name == "" {
		name = strings.TrimSpace(row.Find(fmt.Sprintf("td[data-stat='%s']", stat)).Text())
	}
	return name
}

func parseScore(scoreStr string) (int, int, bool) {
	scoreStr = strings.ReplaceAll(scoreStr, "\u2013", "-")
	scoreStr = strings.ReplaceAll(scoreStr, "\u2014", "-")

	parts := strings.SplitN(scoreStr, "-", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}

	home, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, false
	}

	away, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, false
	}

	return home, away, true
}

func ParseFixtures(season int, html string) ([]ScrapedFixture, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing fixtures HTML: %w", err)
	}

	table := doc.Find(fmt.Sprintf("table#%s", tableID(season)))
	if table.Length() == 0 {
		return nil, fmt.Errorf("table %q not found", tableID(season))
	}

	var fixtures []ScrapedFixture

	table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		score := strings.TrimSpace(row.Find("td[data-stat='score']").Text())
		if score != "vs" {
			return
		}

		f, ok := parseFixtureRow(season, row)
		if ok {
			fixtures = append(fixtures, f)
		}
	})

	return fixtures, nil
}

func parseFixtureRow(season int, row *goquery.Selection) (ScrapedFixture, bool) {
	round, ok := parseGameweek(row)
	if !ok {
		return ScrapedFixture{}, false
	}

	date, ok := parseDate(row)
	if !ok {
		return ScrapedFixture{}, false
	}

	homeTeam := extractTeam(row, "home_team")
	if homeTeam == "" {
		return ScrapedFixture{}, false
	}

	awayTeam := extractTeam(row, "away_team")
	if awayTeam == "" {
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

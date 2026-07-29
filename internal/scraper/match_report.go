package scraper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ParseMatchReport(html string) (ScrapedMatchReport, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return ScrapedMatchReport{}, fmt.Errorf("parsing match report HTML: %w", err)
	}

	report := ScrapedMatchReport{
		HomePlayerStats: map[string]ScrapedPlayerMatchStat{},
		AwayPlayerStats: map[string]ScrapedPlayerMatchStat{},
	}

	scoreBox := doc.Find(".scorebox")
	report.HomeTeam = strings.TrimSpace(scoreBox.Find("div[itemprop='performer']").First().Find("a, strong").First().Text())
	report.AwayTeam = strings.TrimSpace(scoreBox.Find("div[itemprop='performer']").Last().Find("a, strong").First().Text())

	parseTeamStats(doc, &report)
	parsePlayerStats(doc, &report)

	return report, nil
}

func parseTeamStats(doc *goquery.Document, report *ScrapedMatchReport) {
	doc.Find("div#team_stats_extra div").Each(func(_ int, div *goquery.Selection) {
		text := strings.TrimSpace(div.Text())
		if strings.Contains(text, "Fouls") {
			vals := extractTwoValues(text)
			if len(vals) == 2 {
				report.HomeFouls = parseInt(vals[0])
				report.AwayFouls = parseInt(vals[1])
			}
		}
		if strings.Contains(text, "Corners") {
			vals := extractTwoValues(text)
			if len(vals) == 2 {
				report.HomeCorners = parseInt(vals[0])
				report.AwayCorners = parseInt(vals[1])
			}
		}
		if strings.Contains(text, "Crosses") {
			vals := extractTwoValues(text)
			if len(vals) == 2 {
				report.HomeCrosses = parseInt(vals[0])
				report.AwayCrosses = parseInt(vals[1])
			}
		}
		if strings.Contains(text, "Offsides") {
			vals := extractTwoValues(text)
			if len(vals) == 2 {
				report.HomeOffsides = parseInt(vals[0])
				report.AwayOffsides = parseInt(vals[1])
			}
		}
	})

	doc.Find("div#team_stats div").Each(func(_ int, div *goquery.Selection) {
		strong := strings.TrimSpace(div.Find("strong").Text())
		if strong == "Possession" {
			vals := extractTwoValues(div.Text())
			if len(vals) == 2 {
				report.HomePossession = parseInt(vals[0])
				report.AwayPossession = parseInt(vals[1])
			}
		}
	})
}

func extractTwoValues(text string) []string {
	text = strings.ReplaceAll(text, "%", "")
	parts := strings.Fields(text)
	var vals []string
	for _, p := range parts {
		if isNumeric(p) {
			vals = append(vals, p)
		}
	}
	if len(vals) >= 2 {
		return vals[:2]
	}
	return nil
}

func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func parsePlayerStats(doc *goquery.Document, report *ScrapedMatchReport) {
	doc.Find("table[id^='stats_']").Each(func(_ int, table *goquery.Selection) {
		tableID, hasID := table.Attr("id")
		if !hasID {
			return
		}

		isSummary := strings.HasSuffix(tableID, "_summary")
		isKeeper := strings.HasSuffix(tableID, "_keeper")
		if !isSummary && !isKeeper {
			return
		}

		table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
			if row.HasClass("thead") || row.HasClass("spacer") {
				return
			}

			stat := parsePlayerStatRow(row, table)
			if stat.Name == "" {
				return
			}

			if report.HomeTeam != "" && strings.Contains(tableID, "_") {
				if len(report.HomeLineup) == 0 {
					report.HomePlayerStats[stat.Name] = stat
				} else {
					isHome := false
					for _, lp := range report.HomeLineup {
						if lp.Name == stat.Name {
							isHome = true
							break
						}
					}
					if isHome {
						report.HomePlayerStats[stat.Name] = stat
					} else {
						report.AwayPlayerStats[stat.Name] = stat
					}
				}
			}
		})
	})
}

func parsePlayerStatRow(row *goquery.Selection, table *goquery.Selection) ScrapedPlayerMatchStat {
	name := strings.TrimSpace(row.Find("th[data-stat='player'] a").First().Text())
	if name == "" {
		return ScrapedPlayerMatchStat{}
	}

	return ScrapedPlayerMatchStat{
		Name:     name,
		Position: strings.TrimSpace(row.Find("td[data-stat='position']").Text()),
		Minutes:  mustParseInt(row.Find("td[data-stat='minutes']").Text()),
		Goals:    parseIntFromRow(row, "goals"),
		Assists:  parseIntFromRow(row, "assists"),
		Shots:    parseIntFromRow(row, "shots"),
		ShotsOnTarget: parseIntFromRow(row, "shots_on_target"),
		Fouls:    parseIntFromRow(row, "fouls"),
		Fouled:   parseIntFromRow(row, "fouled"),
		Offsides: parseIntFromRow(row, "offsides"),
		Crosses:  parseIntFromRow(row, "crosses"),
		YellowCards: parseIntFromRow(row, "cards_yellow"),
		RedCards:  parseIntFromRow(row, "cards_red"),
		Saves:    parseIntFromRow(row, "saves"),
		Tackles:  parseIntFromRow(row, "tackles"),
		Interceptions: parseIntFromRow(row, "interceptions"),
		Passes:   parseIntFromRow(row, "passes"),
	}
}

func parseIntFromRow(row *goquery.Selection, stat string) *int {
	val := strings.TrimSpace(row.Find(fmt.Sprintf("td[data-stat='%s']", stat)).Text())
	if val == "" {
		return nil
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return nil
	}
	return &n
}

func mustParseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

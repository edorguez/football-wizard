package scraper

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var goalMinuteRe = regexp.MustCompile(`(\d+)(?:\+(\d+))?\s*’`)

func ParseMatchReport(html string) (ScrapedMatchReport, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return ScrapedMatchReport{}, fmt.Errorf("parsing match report HTML: %w", err)
	}

	report := ScrapedMatchReport{
		HomePlayerStats: map[string]ScrapedPlayerMatchStat{},
		AwayPlayerStats: map[string]ScrapedPlayerMatchStat{},
	}

	teamEls := doc.Find(".scorebox_team strong a")
	if teamEls.Length() >= 2 {
		report.HomeTeam = strings.TrimSpace(teamEls.First().Text())
		report.AwayTeam = strings.TrimSpace(teamEls.Last().Text())
	}

	parseTeamStats(doc, &report)
	parseLineups(doc, &report)
	parsePlayerStats(doc, &report)
	parseScoringSummary(doc, &report)

	return report, nil
}

// parseScoringSummary extracts first/second-half goal counts and first-goal
// minutes from the FBref scoring summary (div.event[id="a"] = home,
// [id="b"] = away).
func parseScoringSummary(doc *goquery.Document, report *ScrapedMatchReport) {
	report.HomeFirstGoalMinute, report.HomeGoalsFirstHalf,
		report.HomeGoalsSecondHalf, report.HomeSecondGoalMinute = scoringSummary(doc, "a")

	report.AwayFirstGoalMinute, report.AwayGoalsFirstHalf,
		report.AwayGoalsSecondHalf, report.AwaySecondGoalMinute = scoringSummary(doc, "b")
}

// scoringSummary returns the first-goal minute (any half), the first- and
// second-half goal counts, and the first-goal minute in the second half.
//
// When the summary exists but the side never scored, counts are 0 and goal
// minutes are nil. When the summary itself is missing, everything is nil.
func scoringSummary(doc *goquery.Document, side string) (*int, *int, *int, *int) {
	firstMinute, secondHalfMinute, firstHalfCount, secondHalfCount := -1, -1, 0, 0

	summary := doc.Find(fmt.Sprintf("div.event[id='%s']", side))
	if summary.Length() == 0 {
		return nil, nil, nil, nil
	}

	summary.Find("div.event_icon.goal").Each(func(_ int, icon *goquery.Selection) {
		m := goalMinuteRe.FindStringSubmatch(icon.Parent().Text())
		if m == nil {
			return
		}

		base := mustParseInt(m[1])
		added := 0
		if m[2] != "" {
			added = mustParseInt(m[2])
		}

		total := base + added
		if firstMinute == -1 || total < firstMinute {
			firstMinute = total
		}
		if base <= 45 {
			firstHalfCount++
		} else {
			secondHalfCount++
			if secondHalfMinute == -1 || total < secondHalfMinute {
				secondHalfMinute = total
			}
		}
	})

	if firstMinute == -1 {
		return nil, &firstHalfCount, &secondHalfCount, nil
	}

	firstMinutePtr := &firstMinute
	var secondHalfMinutePtr *int
	if secondHalfMinute != -1 {
		secondHalfMinutePtr = &secondHalfMinute
	}

	return firstMinutePtr, &firstHalfCount, &secondHalfCount, secondHalfMinutePtr
}

func parseTeamStats(doc *goquery.Document, report *ScrapedMatchReport) {
	parseExtraStats(doc, report)
	parseMainStats(doc, report)
}

func parseMainStats(doc *goquery.Document, report *ScrapedMatchReport) {
	doc.Find("div#team_stats table tr").Each(func(_ int, row *goquery.Selection) {
		th := strings.TrimSpace(row.Find("th").Text())

		tds := row.Next().Find("td")
		if tds.Length() < 2 {
			return
		}

		homeText := tds.First().Text()
		awayText := tds.Last().Text()

		switch th {
		case "Possession":
			report.HomePossession = parseStrongPct(tds.First())
			report.AwayPossession = parseStrongPct(tds.Last())

		case "Shots on Target":
			hs, hst := parseShotsOnTarget(homeText)
			as, ast := parseShotsOnTarget(awayText)
			report.HomeShots = hs
			report.AwayShots = as
			report.HomeShotsOnTarget = hst
			report.AwayShotsOnTarget = ast
			if hs != nil && hst != nil {
				off := *hs - *hst
				report.HomeShotsOffTarget = &off
			}
			if as != nil && ast != nil {
				off := *as - *ast
				report.AwayShotsOffTarget = &off
			}

		case "Saves":
			report.HomeSaves = parseSaves(homeText)
			report.AwaySaves = parseSaves(awayText)
		}
	})
}

func parseStrongPct(td *goquery.Selection) *int {
	return extractPct(strings.TrimSpace(td.Find("strong").First().Text()))
}

func parseShotsOnTarget(text string) (totalShots, shotsOnTarget *int) {
	text = strings.TrimSpace(text)
	parts := strings.SplitN(text, "—", 2)
	var shotPart string
	for _, p := range parts {
		if strings.Contains(p, " of ") {
			shotPart = p
			break
		}
	}
	if shotPart == "" {
		return nil, nil
	}
	idx := strings.Index(shotPart, " of ")
	onTarget := parseInt(strings.TrimSpace(shotPart[:idx]))
	rest := strings.TrimSpace(shotPart[idx+4:])
	space := strings.Index(rest, " ")
	if space > 0 {
		rest = rest[:space]
	}
	total := parseInt(rest)
	if onTarget == nil || total == nil {
		return nil, nil
	}
	return total, onTarget
}

func parseSaves(text string) *int {
	text = strings.TrimSpace(text)
	parts := strings.SplitN(text, "—", 2)
	for _, p := range parts {
		idx := strings.Index(p, " of ")
		if idx >= 0 {
			return parseInt(strings.TrimSpace(p[:idx]))
		}
	}
	return nil
}

func parseExtraStats(doc *goquery.Document, report *ScrapedMatchReport) {
	doc.Find("div#team_stats_extra > div").Each(func(_ int, container *goquery.Selection) {
		var vals []string
		container.ChildrenFiltered("div").Each(func(_ int, d *goquery.Selection) {
			if d.HasClass("th") {
				return
			}
			vals = append(vals, strings.TrimSpace(d.Text()))
		})

		for i := 0; i+2 < len(vals); i += 3 {
			stat := vals[i+1]
			switch stat {
			case "Fouls":
				report.HomeFouls = parseInt(vals[i])
				report.AwayFouls = parseInt(vals[i+2])
			case "Corners":
				report.HomeCorners = parseInt(vals[i])
				report.AwayCorners = parseInt(vals[i+2])
			case "Crosses":
				report.HomeCrosses = parseInt(vals[i])
				report.AwayCrosses = parseInt(vals[i+2])
			case "Offsides":
				report.HomeOffsides = parseInt(vals[i])
				report.AwayOffsides = parseInt(vals[i+2])
			}
		}
	})
}

func parseLineups(doc *goquery.Document, report *ScrapedMatchReport) {
	lineupTables := doc.Find("div.lineup table")
	if lineupTables.Length() < 2 {
		return
	}

	report.HomeLineup = parseLineupTable(lineupTables.First())
	report.AwayLineup = parseLineupTable(lineupTables.Last())
}

func parseLineupTable(table *goquery.Selection) []ScrapedLineupPlayer {
	var lineups []ScrapedLineupPlayer
	isStarter := true

	table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		th := row.Find("th")
		if th.Length() > 0 {
			if strings.Contains(th.Text(), "Bench") {
				isStarter = false
			}
			return
		}

		name := strings.TrimSpace(row.Find("td a").Text())
		if name == "" {
			return
		}

		var shirtNum *int
		if n := parseInt(strings.TrimSpace(row.Find("td").First().Text())); n != nil {
			shirtNum = n
		}

		hasSub := row.Find("div.event_icon.substitute_in").Length() > 0

		lineups = append(lineups, ScrapedLineupPlayer{
			Name:       name,
			ShirtNum:   shirtNum,
			IsStarter:  isStarter,
			HasSubIcon: hasSub,
		})
	})

	return lineups
}

func extractPct(s string) *int {
	s = strings.TrimSpace(strings.ReplaceAll(s, "%", ""))
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &n
}

func parsePlayerStats(doc *goquery.Document, report *ScrapedMatchReport) {
	doc.Find("table[id^='stats_'], table[id^='keeper_stats_']").Each(func(_ int, table *goquery.Selection) {
		tableID, hasID := table.Attr("id")
		if !hasID {
			return
		}

		isSummary := strings.HasSuffix(tableID, "_summary")
		isKeeper := strings.HasPrefix(tableID, "keeper_stats_")
		if !isSummary && !isKeeper {
			return
		}

		caption := strings.TrimSpace(table.Find("caption").Text())

		isHome := report.HomeTeam != "" && caption != "" &&
			(strings.Contains(caption, report.HomeTeam) || strings.Contains(report.HomeTeam, caption))
		isAway := !isHome && report.AwayTeam != "" && caption != "" &&
			(strings.Contains(caption, report.AwayTeam) || strings.Contains(report.AwayTeam, caption))

		if !isHome && !isAway {
			return
		}

		table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
			if row.HasClass("thead") || row.HasClass("spacer") {
				return
			}

			stat := parsePlayerStatRow(row, tableID)
			if stat.Name == "" {
				return
			}

			if isHome {
				report.HomePlayerStats[stat.Name] = stat
			} else {
				report.AwayPlayerStats[stat.Name] = stat
			}
		})
	})
}

func parsePlayerStatRow(row *goquery.Selection, tableID string) ScrapedPlayerMatchStat {
	name := strings.TrimSpace(row.Find("th[data-stat='player'] a").First().Text())
	if name == "" {
		return ScrapedPlayerMatchStat{}
	}

	isKeeper := strings.HasPrefix(tableID, "keeper_stats_")

	stat := ScrapedPlayerMatchStat{
		Name:     name,
		Position: strings.TrimSpace(row.Find("td[data-stat='position']").Text()),
		Minutes:  mustParseInt(row.Find("td[data-stat='minutes']").Text()),
	}

	if isKeeper {
		stat.Saves = parseIntFromRow(row, "gk_saves")
		stat.Goals = parseIntFromRow(row, "gk_goals_against")
		stat.ShotsOnTarget = parseIntFromRow(row, "gk_shots_on_target_against")
	} else {
		stat.Goals = parseIntFromRow(row, "goals")
		stat.Assists = parseIntFromRow(row, "assists")
		stat.Shots = parseIntFromRow(row, "shots")
		stat.ShotsOnTarget = parseIntFromRow(row, "shots_on_target")
		stat.Fouls = parseIntFromRow(row, "fouls")
		stat.Fouled = parseIntFromRow(row, "fouled")
		stat.Offsides = parseIntFromRow(row, "offsides")
		stat.Crosses = parseIntFromRow(row, "crosses")
		stat.YellowCards = parseIntFromRow(row, "cards_yellow")
		stat.RedCards = parseIntFromRow(row, "cards_red")
		stat.Tackles = parseIntFromRow(row, "tackles")
		stat.Interceptions = parseIntFromRow(row, "interceptions")
		stat.Passes = parseIntFromRow(row, "passes")
	}

	return stat
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

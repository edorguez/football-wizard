package scraper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func ParseSquad(season int, html string) (ScrapedSquad, error) {
	players, err := ParseSquadPlayers(season, html)
	if err != nil {
		return ScrapedSquad{}, err
	}

	teamFBRefID := extractTeamFBRefID(html)

	return ScrapedSquad{
		TeamFBRefID: teamFBRefID,
		Players:     players,
	}, nil
}

func ParseSquadPlayers(season int, html string) ([]ScrapedPlayer, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing squad HTML: %w", err)
	}

	var players []ScrapedPlayer

	doc.Find("table[id^='stats_standard'] tbody tr").Each(func(_ int, row *goquery.Selection) {
		if row.HasClass("thead") || row.HasClass("spacer") {
			return
		}

		name := strings.TrimSpace(row.Find("th[data-stat='player'] a, td[data-stat='player'] a").Text())
		if name == "" || name == "Player" {
			return
		}

		nationality := strings.TrimSpace(row.Find("td[data-stat='nationality'] a").Text())
		if nationality == "" {
			nationality = strings.TrimSpace(row.Find("td[data-stat='nationality']").Text())
		}
		position := strings.TrimSpace(row.Find("td[data-stat='position']").Text())

		shirtNum := parseNullableInt(row, "number")

		players = append(players, ScrapedPlayer{
			Name:        name,
			Nationality: nationality,
			Position:    position,
			ShirtNum:    shirtNum,
			Season:      season,
		})
	})

	if len(players) == 0 {
		return nil, fmt.Errorf("no players found")
	}

	return players, nil
}

func extractTeamFBRefID(html string) string {
	idx := strings.Index(html, "data-team=\"")
	if idx == -1 {
		return ""
	}
	start := idx + len("data-team=\"")
	end := strings.Index(html[start:], "\"")
	if end == -1 {
		return ""
	}
	return html[start : start+end]
}

func ParseSquadURLs(season int, html string) (map[string]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parsing schedule HTML: %w", err)
	}

	table := doc.Find(fmt.Sprintf("table#%s", tableID(season)))
	if table.Length() == 0 {
		return nil, fmt.Errorf("table %q not found", tableID(season))
	}

	teamURLs := map[string]string{}

	table.Find("td[data-stat='home_team'] a, td[data-stat='away_team'] a").Each(func(_ int, link *goquery.Selection) {
		teamName := strings.TrimSpace(link.Text())
		href, exists := link.Attr("href")
		if !exists || teamName == "" {
			return
		}
		teamURLs[teamName] = fbrefBase + href
	})

	return teamURLs, nil
}

func teamIDFromURL(url string) string {
	parts := strings.Split(url, "/")
	for i, p := range parts {
		if p == "squads" && i+1 < len(parts) {
			return parts[i+1]
		}
		if p == "en" && i+2 < len(parts) && parts[i+1] == "squads" {
			return parts[i+2]
		}
	}

	return ""
}

func parseInt(s string) *int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &n
}

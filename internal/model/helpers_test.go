package model

import "time"

func intPtr(v int) *int { return &v }

func makeMatch(season, round int, home, away uint, hg, ag int) MatchRow {
	return MatchRow{
		Season:     season,
		Round:      round,
		Date:       time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, round),
		HomeTeamID: home,
		AwayTeamID: away,
		HomeGoals:  intPtr(hg),
		AwayGoals:  intPtr(ag),
	}
}

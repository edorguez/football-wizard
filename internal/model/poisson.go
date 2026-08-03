package model

import (
	"math"

	"gonum.org/v1/gonum/stat/distuv"
)

// PoissonModel is a classic attack/defense Poisson model: each team has a
// league-relative attack and defense strength, and a fixture's expected goals
// are derived by combining them with the league's home/away scoring averages.
type PoissonModel struct {
	HomeAvg   float64
	AwayAvg   float64
	LeagueAvg float64
	Attack    map[uint]float64
	Defense   map[uint]float64
}

// FitPoisson estimates league scoring averages and per-team attack/defense
// strengths from completed matches.
func FitPoisson(matches []MatchRow) *PoissonModel {
	homeGoals, awayGoals, homeGames, awayGames := 0, 0, 0, 0

	goalsFor := map[uint]int{}
	goalsAgainst := map[uint]int{}
	games := map[uint]int{}

	for _, m := range matches {
		hg, ag := 0, 0
		if m.HomeGoals != nil {
			hg = *m.HomeGoals
		}
		if m.AwayGoals != nil {
			ag = *m.AwayGoals
		}

		homeGoals += hg
		awayGoals += ag
		homeGames++
		awayGames++

		goalsFor[m.HomeTeamID] += hg
		goalsAgainst[m.HomeTeamID] += ag
		goalsFor[m.AwayTeamID] += ag
		goalsAgainst[m.AwayTeamID] += hg
		games[m.HomeTeamID]++
		games[m.AwayTeamID]++
	}

	if homeGames == 0 || awayGames == 0 {
		return &PoissonModel{
			Attack:  map[uint]float64{},
			Defense: map[uint]float64{},
		}
	}

	model := &PoissonModel{
		HomeAvg:   float64(homeGoals) / float64(homeGames),
		AwayAvg:   float64(awayGoals) / float64(awayGames),
		LeagueAvg: float64(homeGoals+awayGoals) / float64(homeGames+awayGames),
		Attack:    map[uint]float64{},
		Defense:   map[uint]float64{},
	}

	if model.LeagueAvg == 0 {
		return model
	}

	for teamID, n := range games {
		if n == 0 {
			continue
		}
		model.Attack[teamID] = float64(goalsFor[teamID]) / float64(n) / model.LeagueAvg
		model.Defense[teamID] = float64(goalsAgainst[teamID]) / float64(n) / model.LeagueAvg
	}

	return model
}

// ExpectedGoals returns the expected goals for home and away in a fixture.
func (m *PoissonModel) ExpectedGoals(homeID, awayID uint) (float64, float64) {
	homeG := m.HomeAvg * m.attack(homeID) * m.defense(awayID)
	awayG := m.AwayAvg * m.attack(awayID) * m.defense(homeID)
	return homeG, awayG
}

// ProbHomeWin, ProbDraw and ProbAwayWin sum the scoreline distribution.
func (m *PoissonModel) ProbHomeWin(homeID, awayID uint) float64 {
	homeG, awayG := m.ExpectedGoals(homeID, awayID)
	return m.sumScorelines(homeG, awayG, func(h, a int) bool { return h > a })
}

func (m *PoissonModel) ProbDraw(homeID, awayID uint) float64 {
	homeG, awayG := m.ExpectedGoals(homeID, awayID)
	return m.sumScorelines(homeG, awayG, func(h, a int) bool { return h == a })
}

func (m *PoissonModel) ProbAwayWin(homeID, awayID uint) float64 {
	homeG, awayG := m.ExpectedGoals(homeID, awayID)
	return m.sumScorelines(homeG, awayG, func(h, a int) bool { return h < a })
}

// ProbOver returns P(total goals > line), e.g. line=2.5 means 3+ goals.
func (m *PoissonModel) ProbOver(homeID, awayID uint, line float64) float64 {
	homeG, awayG := m.ExpectedGoals(homeID, awayID)
	return m.sumScorelines(homeG, awayG, func(h, a int) bool { return float64(h+a) > line })
}

// ProbExactScore returns the probability of a given scoreline.
func (m *PoissonModel) ProbExactScore(homeID, awayID uint, home, away int) float64 {
	homeG, awayG := m.ExpectedGoals(homeID, awayID)
	return m.pmf(homeG, float64(home)) * m.pmf(awayG, float64(away))
}

const maxGoals = 12

func (m *PoissonModel) sumScorelines(homeG, awayG float64, keep func(h, a int) bool) float64 {
	home := distuv.Poisson{Lambda: homeG}
	away := distuv.Poisson{Lambda: awayG}

	var sum float64
	for h := 0; h <= maxGoals; h++ {
		for a := 0; a <= maxGoals; a++ {
			if keep(h, a) {
				sum += home.Prob(float64(h)) * away.Prob(float64(a))
			}
		}
	}
	return math.Min(1, sum)
}

func (m *PoissonModel) pmf(lambda, k float64) float64 {
	return distuv.Poisson{Lambda: lambda}.Prob(k)
}

func (m *PoissonModel) attack(teamID uint) float64 {
	if v, ok := m.Attack[teamID]; ok && v > 0 {
		return v
	}
	return 1
}

func (m *PoissonModel) defense(teamID uint) float64 {
	if v, ok := m.Defense[teamID]; ok && v > 0 {
		return v
	}
	return 1
}

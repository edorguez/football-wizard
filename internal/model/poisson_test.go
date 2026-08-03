package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFitPoisson(t *testing.T) {
	t.Parallel()

	// Balanced round-robin of 4 teams: home always wins 2-1, so every team
	// has identical strength and expected goals equal the league averages.
	teams := []uint{0, 1, 2, 3}
	matches := make([]MatchRow, 0, 12)
	for _, home := range teams {
		for _, away := range teams {
			if home == away {
				continue
			}
			matches = append(matches, makeMatch(2025, len(matches), home, away, 2, 1))
		}
	}

	model := FitPoisson(matches)
	require.NotNil(t, model)

	is := assert.New(t)
	is.InDelta(2.0, model.HomeAvg, 1e-9)
	is.InDelta(1.0, model.AwayAvg, 1e-9)

	hg, ag := model.ExpectedGoals(1, 2)
	is.InDelta(2.0, hg, 1e-9)
	is.InDelta(1.0, ag, 1e-9)
}

func TestPoissonProbabilitiesSumToOne(t *testing.T) {
	t.Parallel()

	matches := make([]MatchRow, 0, 20)
	for i := 0; i < 20; i++ {
		matches = append(matches, makeMatch(2025, i, uint(i), uint(i+1), 2, 1))
	}
	model := FitPoisson(matches)

	homeWin := model.ProbHomeWin(1, 2)
	draw := model.ProbDraw(1, 2)
	awayWin := model.ProbAwayWin(1, 2)

	is := assert.New(t)
	is.InDelta(1.0, homeWin+draw+awayWin, 1e-6)
	is.True(homeWin > awayWin, "home advantage should favor the home side")

	over := model.ProbOver(1, 2, 2.5)
	is.True(over > 0 && over < 1)
}

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrainerEndToEnd(t *testing.T) {
	t.Parallel()

	// Build a fake season: every home team wins 2-1, so each of the 6 teams
	// is balanced (attack == defense == 1) and the Poisson model should favor
	// the home side.
	matches := make([]MatchRow, 0, 60)
	for i := 0; i < 60; i++ {
		m := makeMatch(2025, i%30, uint(i%6), uint((i+1)%6), 2, 1)
		m.HomeCorners = intPtr(6)
		m.AwayCorners = intPtr(4)
		m.HomeYellowCards = intPtr(2)
		m.AwayYellowCards = intPtr(2)
		m.HomeOffsides = intPtr(1)
		m.AwayOffsides = intPtr(1)
		m.HomeGoalsFirstHalf = intPtr(1)
		m.AwayGoalsFirstHalf = intPtr(0)
		m.HomeFirstGoalMinute = intPtr(20)
		m.AwayFirstGoalMinute = intPtr(0)
		matches = append(matches, m)
	}

	trainer := NewTrainer()
	predictor, err := trainer.Train(matches)
	require.NoError(t, err)

	pred := predictor.PredictMatch(1, 2, "Flamengo", "Palmeiras")
	is := assert.New(t)

	is.InDelta(1.0, pred.HomeWin+pred.Draw+pred.AwayWin, 1e-4)
	is.True(pred.HomeWin > pred.Draw, "home win %.3f should beat draw %.3f", pred.HomeWin, pred.Draw)
	is.True(pred.HomeWin > pred.AwayWin, "home win %.3f should beat away win %.3f", pred.HomeWin, pred.AwayWin)

	require.NotEmpty(t, pred.Markets)
	for _, m := range pred.Markets {
		is.True(m.Probability >= 0 && m.Probability <= 1, "probability out of range for %s/%s", m.Market, m.Outcome)
	}
}

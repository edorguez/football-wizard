package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateHeldOut(t *testing.T) {
	t.Parallel()

	matches := make([]MatchRow, 0, 60)
	for i := 0; i < 60; i++ {
		m := makeMatch(2025, i%30, uint(i%6), uint((i+1)%6), 2, 1)
		m.HomeCorners = intPtr(6)
		m.AwayCorners = intPtr(4)
		m.HomeYellowCards = intPtr(2)
		m.AwayYellowCards = intPtr(2)
		m.HomeOffsides = intPtr(1)
		m.AwayOffsides = intPtr(1)
		matches = append(matches, m)
	}

	summary, err := Evaluate(matches, 15, NewTrainer())
	require.NoError(t, err)

	is := assert.New(t)
	is.Equal(30, summary.TrainCount)
	is.Equal(30, summary.TestCount)
	is.NotEmpty(summary.Reports)

	byMarket := map[Market]EvalReport{}
	for _, r := range summary.Reports {
		byMarket[r.Market] = r
		is.True(r.Samples > 0)
		is.True(r.Accuracy >= 0 && r.Accuracy <= 1)
	}
	is.Contains(byMarket, Market("1X2"))
}

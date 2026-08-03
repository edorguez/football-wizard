package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatPrediction(t *testing.T) {
	t.Parallel()

	pred := &MatchPrediction{
		HomeTeam:           "Flamengo",
		AwayTeam:           "Palmeiras",
		HomeWin:            0.6,
		Draw:               0.25,
		AwayWin:            0.15,
		ExpectedHomeGoals:  1.8,
		ExpectedAwayGoals:  0.9,
		ExpectedTotalGoals: 2.7,
		Markets: []MarketPrediction{
			{Market: MarketBTTS, Outcome: "Yes", Probability: 0.5, Confidence: ConfidenceLow},
			{Market: MarketTotalGoals, Outcome: "Over", Line: 2.5, Probability: 0.6, Confidence: ConfidenceMedium},
		},
	}

	out := FormatPrediction(pred)

	is := assert.New(t)
	is.Contains(out, "Flamengo vs Palmeiras")
	is.Contains(out, "expected goals: 1.80 - 0.90")
	is.Contains(out, "60.0%")
	is.Contains(out, "Medium")
	is.Contains(out, "btts")
}

func TestEvalSummaryTable(t *testing.T) {
	t.Parallel()

	summary := &EvalSummary{
		SplitRound: 19,
		TrainCount: 180,
		TestCount:  200,
		Reports: []EvalReport{
			{Market: MarketBTTS, Samples: 200, Correct: 106, Accuracy: 0.53},
			{Market: "1X2", Samples: 200, Correct: 109, Accuracy: 0.545},
			{Market: MarketCardsOU, Line: 3.5, HasLine: true, Samples: 197, Correct: 135, Accuracy: 0.685},
		},
	}

	out := summary.Table()
	require.NotEmpty(t, out)

	is := assert.New(t)
	is.True(strings.Contains(out, "MARKET"))
	is.Contains(out, "btts")
	is.Contains(out, "3.5")
	is.Contains(out, "68.5%")
}

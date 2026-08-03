package model

import (
	"fmt"
	"sort"
	"strings"
)

// DefaultSplitRound is the round at which training data is split from held-out
// data during evaluation (the second half of a Brasileirão season).
const DefaultSplitRound = 19

// EvalReport is accuracy for a single market on held-out matches.
type EvalReport struct {
	Market   Market
	Line     float64
	HasLine  bool
	Samples  int
	Correct  int
	Accuracy float64
}

// EvalSummary aggregates the held-out evaluation.
type EvalSummary struct {
	SplitRound int
	TrainCount int
	TestCount  int
	Reports    []EvalReport
}

// Evaluate trains on matches before splitRound and measures accuracy on the
// rest. This mirrors the weekly retrain flow: the trainer only ever sees past
// matches, then forecasts upcoming ones.
func Evaluate(rows []MatchRow, splitRound int, trainer *Trainer) (*EvalSummary, error) {
	train, test := splitByRound(rows, splitRound)
	if len(train) == 0 || len(test) == 0 {
		return nil, fmt.Errorf("evaluation needs both train and test samples (got %d/%d)", len(train), len(test))
	}

	predictor, err := trainer.Train(train)
	if err != nil {
		return nil, err
	}

	reports := map[Market]*EvalReport{}
	score := func(market Market, line float64, hasLine bool, predicted, actual bool) {
		r, ok := reports[market]
		if !ok {
			r = &EvalReport{Market: market, Line: line, HasLine: hasLine}
			reports[market] = r
		}
		r.Samples++
		if predicted == actual {
			r.Correct++
		}
	}

	for _, m := range test {
		pred := predictor.PredictMatch(m.HomeTeamID, m.AwayTeamID, "", "")

		probs := map[Market]float64{}
		goalsOver := map[float64]float64{}
		for _, mp := range pred.Markets {
			if mp.Market == MarketTotalGoals {
				goalsOver[mp.Line] = mp.Probability
				continue
			}
			probs[mp.Market] = mp.Probability
		}

		// 1X2: the market with the highest Poisson probability.
		score1X2(pred, m, score)

		for _, spec := range binaryMarketSpecs {
			if predictor.logistic[spec.Market] == nil {
				continue
			}
			label, ok := spec.Label(m, predictor.line(spec.Market))
			if !ok {
				continue
			}
			p := probs[spec.Market]
			score(spec.Market, predictor.line(spec.Market), spec.HasThreshold, p > 0.5, label > 0.5)
		}

		// Total goals lines from the Poisson model.
		for _, line := range predictor.goalLines {
			hg, ag := 0, 0
			if m.HomeGoals != nil {
				hg = *m.HomeGoals
			}
			if m.AwayGoals != nil {
				ag = *m.AwayGoals
			}
			actual := float64(hg+ag) > line
			score(MarketTotalGoals, line, true, goalsOver[line] > 0.5, actual)
		}
	}

	summary := &EvalSummary{
		SplitRound: splitRound,
		TrainCount: len(train),
		TestCount:  len(test),
		Reports:    make([]EvalReport, 0, len(reports)),
	}
	for _, r := range reports {
		if r.Samples > 0 {
			r.Accuracy = float64(r.Correct) / float64(r.Samples)
		}
		summary.Reports = append(summary.Reports, *r)
	}
	return summary, nil
}

func splitByRound(rows []MatchRow, splitRound int) ([]MatchRow, []MatchRow) {
	var train, test []MatchRow
	for _, m := range rows {
		if m.Round < splitRound {
			train = append(train, m)
		} else {
			test = append(test, m)
		}
	}
	return train, test
}

func score1X2(pred *MatchPrediction, m MatchRow, score func(Market, float64, bool, bool, bool)) {
	if m.HomeGoals == nil || m.AwayGoals == nil {
		return
	}
	predicted := drawOutcome(pred)
	switch {
	case *m.HomeGoals > *m.AwayGoals:
		score(Market("1X2"), 0, false, predicted == "home", true)
	case *m.HomeGoals < *m.AwayGoals:
		score(Market("1X2"), 0, false, predicted == "away", true)
	default:
		score(Market("1X2"), 0, false, predicted == "draw", true)
	}
}

func drawOutcome(pred *MatchPrediction) string {
	mostLikely := "home"
	best := pred.HomeWin
	if pred.Draw > best {
		mostLikely, best = "draw", pred.Draw
	}
	if pred.AwayWin > best {
		mostLikely = "away"
	}
	return mostLikely
}

// Table renders the evaluation report as a fixed-width table.
func (s *EvalSummary) Table() string {
	reports := make([]EvalReport, len(s.Reports))
	copy(reports, s.Reports)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Market < reports[j].Market
	})

	var b strings.Builder
	fmt.Fprintf(&b, "%-24s %-8s %8s %8s\n", "MARKET", "LINE", "SAMPLES", "ACCURACY")
	b.WriteString(strings.Repeat("-", 56))
	b.WriteString("\n")
	for _, r := range reports {
		line := "-"
		if r.HasLine {
			line = fmt.Sprintf("%.1f", r.Line)
		}
		fmt.Fprintf(&b, "%-24s %-8s %8d %7.1f%%\n", r.Market, line, r.Samples, r.Accuracy*100)
	}
	return b.String()
}

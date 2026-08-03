package model

import (
	"fmt"
	"sort"
)

// TrainingRow is one match paired with its causal feature vector.
type TrainingRow struct {
	Features *Features
	Match    MatchRow
}

func sortedByDate(matches []MatchRow) []MatchRow {
	ordered := make([]MatchRow, len(matches))
	copy(ordered, matches)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Date.Before(ordered[j].Date)
	})
	return ordered
}

// BuildDataset computes causal features for every match in chronological
// order. Matches are sorted by date before processing.
func BuildDataset(matches []MatchRow) []TrainingRow {
	ordered := sortedByDate(matches)

	engine := NewEngine()
	rows := make([]TrainingRow, 0, len(ordered))
	for _, m := range ordered {
		rows = append(rows, TrainingRow{
			Features: engine.Features(m.HomeTeamID, m.AwayTeamID),
			Match:    m,
		})
		engine.Apply(m)
	}
	return rows
}

// Trainer fits the Poisson model and every market classifier from matches.
type Trainer struct {
	Thresholds map[Market]float64
	GoalLines  []float64
	Options    LogisticOptions
}

func NewTrainer() *Trainer {
	thresholds := make(map[Market]float64, len(DefaultThresholds))
	for market, line := range DefaultThresholds {
		thresholds[market] = line
	}

	goalLines := make([]float64, len(DefaultGoalLines))
	copy(goalLines, DefaultGoalLines)

	return &Trainer{
		Thresholds: thresholds,
		GoalLines:  goalLines,
	}
}

func (t *Trainer) threshold(m Market) float64 {
	if v, ok := t.Thresholds[m]; ok {
		return v
	}
	return DefaultThresholds[m]
}

// Train fits all models and returns a Predictor ready for new fixtures. The
// predictor keeps the engine state at the end of the dataset so that live
// features reflect the latest form.
func (t *Trainer) Train(matches []MatchRow) (*Predictor, error) {
	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches to train on")
	}

	rows := BuildDataset(matches)

	engine := NewEngine()
	for _, m := range sortedByDate(matches) {
		engine.Apply(m)
	}

	predictor := &Predictor{
		poisson:    FitPoisson(matches),
		logistic:   map[Market]*LogisticModel{},
		thresholds: t.Thresholds,
		goalLines:  t.GoalLines,
		engine:     engine,
	}

	for _, spec := range binaryMarketSpecs {
		model, err := t.fitSpec(rows, spec)
		if err != nil {
			return nil, fmt.Errorf("training %s: %w", spec.Market, err)
		}
		if model != nil {
			predictor.logistic[spec.Market] = model
		}
	}

	if err := t.fitFirstScorer(predictor, rows); err != nil {
		return nil, err
	}

	return predictor, nil
}

// fitSpec returns nil when the market has no labeled samples (e.g. the DB has
// not yet been re-scraped for that stat), so the market is simply skipped.
func (t *Trainer) fitSpec(rows []TrainingRow, spec MarketSpec) (*LogisticModel, error) {
	X := make([][]float64, 0, len(rows))
	y := make([]float64, 0, len(rows))

	for _, r := range rows {
		label, ok := spec.Label(r.Match, t.threshold(spec.Market))
		if !ok {
			continue
		}
		X = append(X, r.Features.Vector())
		y = append(y, label)
	}

	if len(X) == 0 {
		return nil, nil
	}

	return FitLogistic(X, y, t.Options)
}

func (t *Trainer) fitFirstScorer(predictor *Predictor, rows []TrainingRow) error {
	homeX, homeY := make([][]float64, 0, len(rows)), []float64{}
	awayX, awayY := make([][]float64, 0, len(rows)), []float64{}
	for _, r := range rows {
		if label, ok := r.Match.homeScoredFirst(); ok {
			homeX = append(homeX, r.Features.Vector())
			homeY = append(homeY, label)
		}
		if label, ok := r.Match.awayScoredFirst(); ok {
			awayX = append(awayX, r.Features.Vector())
			awayY = append(awayY, label)
		}
	}

	predictor.firstScorer = &FirstScorerModel{}
	var err error
	if len(homeX) > 0 {
		if predictor.firstScorer.Home, err = FitLogistic(homeX, homeY, t.Options); err != nil {
			return fmt.Errorf("training %s (home first): %w", MarketFirstScorer, err)
		}
	}
	if len(awayX) > 0 {
		if predictor.firstScorer.Away, err = FitLogistic(awayX, awayY, t.Options); err != nil {
			return fmt.Errorf("training %s (away first): %w", MarketFirstScorer, err)
		}
	}
	return nil
}

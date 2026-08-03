package main

import (
	"fmt"

	"github.com/edorguez/football-wizard/internal/config"
	"github.com/edorguez/football-wizard/internal/model"
	"github.com/edorguez/football-wizard/internal/repository"
)

func buildTrainer(cfg *config.Config) *model.Trainer {
	t := model.NewTrainer()
	if len(cfg.Model.GoalLines) > 0 {
		t.GoalLines = cfg.Model.GoalLines
	}
	markets := map[string]model.Market{
		"cards":           model.MarketCardsOU,
		"corners":         model.MarketCornersOU,
		"offsides":        model.MarketOffsidesOU,
		"shots":           model.MarketShotsOU,
		"shots_on_target": model.MarketShotsOnTargetOU,
		"saves":           model.MarketSavesOU,
	}
	for key, market := range markets {
		if v, ok := cfg.Model.OverUnder[key]; ok {
			t.Thresholds[market] = v
		}
	}
	t.Options = model.LogisticOptions{
		Epochs:    cfg.Model.Epochs,
		LearnRate: cfg.Model.LearnRate,
		L2:        cfg.Model.L2,
	}
	return t
}

func runTrain(matchesRepo *repository.MatchRepository, cfg *config.Config) error {
	rows, err := matchesRepo.ListRows()
	if err != nil {
		return fmt.Errorf("loading matches: %w", err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("no completed matches in the database")
	}
	fmt.Printf("loaded %d completed matches\n", len(rows))

	summary, err := model.Evaluate(rows, model.DefaultSplitRound, buildTrainer(cfg))
	if err != nil {
		return fmt.Errorf("evaluating models: %w", err)
	}

	fmt.Printf("train/test split at round %d (%d train, %d test)\n\n", summary.SplitRound, summary.TrainCount, summary.TestCount)
	fmt.Print(summary.Table())
	return nil
}

func runPredict(matchesRepo *repository.MatchRepository, teamsRepo *repository.TeamRepository, cfg *config.Config, home, away string) error {
	rows, err := matchesRepo.ListRows()
	if err != nil {
		return fmt.Errorf("loading matches: %w", err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("no completed matches in the database")
	}

	homeTeam, err := teamsRepo.FindByName(home)
	if err != nil {
		return fmt.Errorf("finding home team %q: %w", home, err)
	}
	awayTeam, err := teamsRepo.FindByName(away)
	if err != nil {
		return fmt.Errorf("finding away team %q: %w", away, err)
	}

	predictor, err := buildTrainer(cfg).Train(rows)
	if err != nil {
		return fmt.Errorf("training models: %w", err)
	}

	pred := predictor.PredictMatch(homeTeam.ID, awayTeam.ID, homeTeam.Name, awayTeam.Name)
	fmt.Print(model.FormatPrediction(pred))
	return nil
}

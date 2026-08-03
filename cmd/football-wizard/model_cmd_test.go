package main

import (
	"testing"

	"github.com/edorguez/football-wizard/internal/config"
	"github.com/edorguez/football-wizard/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestBuildTrainerMapsConfig(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			GoalLines: []float64{0.5, 2.5, 4.5},
			OverUnder: map[string]float64{
				"cards": 4.5,
				"shots": 20.0,
				"saves": 3.5,
			},
			Epochs:    100,
			LearnRate: 0.1,
			L2:        0.5,
		},
	}

	trainer := buildTrainer(cfg)

	is := assert.New(t)
	is.Equal([]float64{0.5, 2.5, 4.5}, trainer.GoalLines)
	is.Equal(4.5, trainer.Thresholds[model.MarketCardsOU])
	is.Equal(20.0, trainer.Thresholds[model.MarketShotsOU])
	is.Equal(3.5, trainer.Thresholds[model.MarketSavesOU])
	is.Equal(100, trainer.Options.Epochs)
	is.Equal(0.1, trainer.Options.LearnRate)
	is.Equal(0.5, trainer.Options.L2)

	// Markets absent from config fall back to their defaults.
	is.Equal(9.5, trainer.Thresholds[model.MarketCornersOU])
	is.Equal(3.5, trainer.Thresholds[model.MarketOffsidesOU])
}

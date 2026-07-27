package predictor

import (
	"fmt"
	"time"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/model"
	"github.com/edorguez/football-wizard/internal/repository"
	"github.com/edorguez/football-wizard/pkg/utils"
	"go.uber.org/zap"
)

type Engine struct {
	trainer      *model.Trainer
	featureEng   *model.FeatureEngine
	fixtureRepo  *repository.FixtureRepo
	predictionRepo *repository.PredictionRepo
	teamRepo     *repository.TeamRepo
	log          *zap.Logger
}

func NewEngine(
	trainer *model.Trainer,
	featureEng *model.FeatureEngine,
	fixtureRepo *repository.FixtureRepo,
	predictionRepo *repository.PredictionRepo,
	teamRepo *repository.TeamRepo,
	log *zap.Logger,
) *Engine {
	return &Engine{
		trainer:        trainer,
		featureEng:     featureEng,
		fixtureRepo:    fixtureRepo,
		predictionRepo: predictionRepo,
		teamRepo:       teamRepo,
		log:            log,
	}
}

func (e *Engine) Predict(homeTeamID, awayTeamID int64, matchDate time.Time) (*database.Prediction, error) {
	e.log.Info("predicting match",
		zap.Int64("home", homeTeamID),
		zap.Int64("away", awayTeamID),
		zap.Time("date", matchDate),
	)

	homeTeam, err := e.teamRepo.FindByID(homeTeamID)
	if err != nil {
		return nil, fmt.Errorf("home team not found: %w", err)
	}
	awayTeam, err := e.teamRepo.FindByID(awayTeamID)
	if err != nil {
		return nil, fmt.Errorf("away team not found: %w", err)
	}

	features, err := e.featureEng.Calculate(homeTeamID, awayTeamID, 5)
	if err != nil {
		return nil, fmt.Errorf("calculating features: %w", err)
	}

	poisson := e.trainer.GetPoissonModel()
	logistic := e.trainer.GetLogisticModel()

	homeWinProb := poisson.HomeWinProb()
	drawProb := poisson.DrawProb()
	awayWinProb := poisson.AwayWinProb()
	bttsProb := poisson.BttsProb()

	over05 := poisson.Over05Prob()
	over15 := poisson.Over15Prob()
	over25 := poisson.Over25Prob()
	over35 := poisson.Over35Prob()

	confLevel, confScore := e.trainer.CalculateConfidence(features, 5)

	yellowTotal := utils.Round(features.HomeFormGoalsScored+features.AwayFormGoalsScored+2, 2)
	cornerTotal := utils.Round(features.HomeCornersAvg+features.AwayCornersAvg+1, 2)

	prediction := &database.Prediction{
		HomeWinProb:      utils.Round(homeWinProb, 4),
		DrawProb:         utils.Round(drawProb, 4),
		AwayWinProb:      utils.Round(awayWinProb, 4),
		Over05Prob:       utils.Round(over05, 4),
		Over15Prob:       utils.Round(over15, 4),
		Over25Prob:       utils.Round(over25, 4),
		Over35Prob:       utils.Round(over35, 4),
		BttsYesProb:      utils.Round(bttsProb, 4),
		Over35YellowProb: utils.Round(yellowTotal*0.4, 4),
		Over45YellowProb: utils.Round(yellowTotal*0.3, 4),
		HomeRedProb:      utils.Round(features.HomeFormGoalsScored*0.05, 4),
		AwayRedProb:      utils.Round(features.AwayFormGoalsScored*0.05, 4),
		Over85CornersProb:  utils.Round(cornerTotal*0.3, 4),
		Over95CornersProb:  utils.Round(cornerTotal*0.2, 4),
		Over105CornersProb: utils.Round(cornerTotal*0.1, 4),
		HomeFirstHalfProb:  utils.Round(homeWinProb*0.45, 4),
		AwayFirstHalfProb:  utils.Round(awayWinProb*0.45, 4),
		ConfidenceLevel:    confLevel,
		ConfidenceScore:    utils.Round(confScore, 2),
		ModelVersion:       "v1",
	}

	_ = homeTeam
	_ = awayTeam
	_ = logistic

	e.log.Info("prediction complete",
		zap.Float64("home_win", prediction.HomeWinProb),
		zap.Float64("draw", prediction.DrawProb),
		zap.Float64("away_win", prediction.AwayWinProb),
		zap.String("confidence", prediction.ConfidenceLevel),
	)

	return prediction, nil
}

func (e *Engine) PredictFixtures(limit int) ([]database.Prediction, error) {
	fixtures, err := e.fixtureRepo.FindUpcoming(limit)
	if err != nil {
		return nil, fmt.Errorf("finding upcoming fixtures: %w", err)
	}

	var predictions []database.Prediction
	for _, f := range fixtures {
		pred, err := e.Predict(f.HomeTeamID, f.AwayTeamID, f.Date)
		if err != nil {
			e.log.Warn("error predicting fixture", zap.Int64("fixture_id", f.ID), zap.Error(err))
			continue
		}
		pred.FixtureID = f.ID
		if err := e.predictionRepo.Create(pred); err != nil {
			e.log.Warn("error saving prediction", zap.Error(err))
			continue
		}
		predictions = append(predictions, *pred)
	}

	return predictions, nil
}

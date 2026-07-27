package model

import (
	"fmt"
	"time"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/repository"
	"go.uber.org/zap"
)

type Trainer struct {
	featureEngine *FeatureEngine
	matchRepo     *repository.MatchRepo
	teamRepo      *repository.TeamRepo
	poissonModel  *PoissonModel
	logisticModel *LogisticModel
	log           *zap.Logger
}

func NewTrainer(
	featureEngine *FeatureEngine,
	matchRepo *repository.MatchRepo,
	teamRepo *repository.TeamRepo,
	log *zap.Logger,
) *Trainer {
	return &Trainer{
		featureEngine: featureEngine,
		matchRepo:     matchRepo,
		teamRepo:      teamRepo,
		logisticModel: NewLogisticModel(),
		log:           log,
	}
}

func (t *Trainer) Train(seasons []int) (string, error) {
	version := fmt.Sprintf("v%d", time.Now().Unix())
	t.log.Info("starting model training", zap.String("version", version), zap.Ints("seasons", seasons))

	var allMatches []database.Match
	for _, season := range seasons {
		matches, err := t.matchRepo.FindBySeason(season)
		if err != nil {
			return "", fmt.Errorf("loading season %d: %w", season, err)
		}
		allMatches = append(allMatches, matches...)
	}

	if len(allMatches) == 0 {
		return "", fmt.Errorf("no matches found for seasons %v", seasons)
	}

	t.log.Info("loaded matches", zap.Int("count", len(allMatches)))

	teams, err := t.teamRepo.FindAll()
	if err != nil {
		return "", fmt.Errorf("loading teams: %w", err)
	}

	t.log.Info("loaded teams", zap.Int("count", len(teams)))

	teamStats := make(map[int64]struct {
		GoalsFor  int
		GoalsAgainst int
		Matches   int
	})

	for _, m := range allMatches {
		home := teamStats[m.HomeTeamID]
		home.GoalsFor += m.HomeGoals
		home.GoalsAgainst += m.AwayGoals
		home.Matches++
		teamStats[m.HomeTeamID] = home

		away := teamStats[m.AwayTeamID]
		away.GoalsFor += m.AwayGoals
		away.GoalsAgainst += m.HomeGoals
		away.Matches++
		teamStats[m.AwayTeamID] = away
	}

	var totalGF, totalGA int
	var totalMatches int
	for _, s := range teamStats {
		totalGF += s.GoalsFor
		totalGA += s.GoalsAgainst
		totalMatches += s.Matches
	}

	avgGF := float64(totalGF) / float64(totalMatches) * float64(len(teamStats))
	avgGA := float64(totalGA) / float64(totalMatches) * float64(len(teamStats))
	leagueAvg := (float64(totalGF) + float64(totalGA)) / float64(totalMatches)

	t.poissonModel = NewPoissonModel(avgGF, avgGA, avgGF, avgGA, leagueAvg, 1)
	t.logisticModel.Train(nil, nil, 100, 0.01)

	t.log.Info("model training complete", zap.String("version", version))
	return version, nil
}

func (t *Trainer) GetPoissonModel() *PoissonModel {
	return t.poissonModel
}

func (t *Trainer) GetLogisticModel() *LogisticModel {
	return t.logisticModel
}

func (t *Trainer) TrainDecisionTree(features [][]float64, labels []float64) {
	if len(features) == 0 {
		return
	}
}

func (t *Trainer) CalculateConfidence(features *Features, matchCount int) (string, float64) {
	if matchCount < 3 {
		return "low", 0.2
	}
	if matchCount < 5 {
		return "low", 0.4
	}
	if matchCount < 10 {
		return "medium", 0.6
	}

	score := 0.7

	if features.HomeFormGoalsScored > 1.5 {
		score += 0.05
	}
	if features.AwayFormGoalsConceded > 1.5 {
		score += 0.05
	}
	if features.HomeFormGoalsConceded < 0.8 {
		score += 0.05
	}
	if features.AwayFormGoalsScored < 0.8 {
		score += 0.05
	}
	if features.HomeXgLast5 > 1.5 {
		score += 0.05
	}
	if features.AwayXgLast5 < 1.0 {
		score += 0.05
	}

	if score > 1.0 {
		score = 1.0
	}

	level := "medium"
	if score >= 0.8 {
		level = "high"
	} else if score <= 0.4 {
		level = "low"
	}

	return level, score
}

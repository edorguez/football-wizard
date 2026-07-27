package model

import (
	"fmt"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/repository"
	"gorm.io/gorm"
)

type Features struct {
	HomeFormGoalsScored   float64
	HomeFormGoalsConceded float64
	AwayFormGoalsScored   float64
	AwayFormGoalsConceded float64
	HomeXgLast5           float64
	AwayXgLast5           float64
	RefereeYellowAvg      float64
	RefereeRedAvg         float64
	HomeCornersAvg        float64
	AwayCornersAvg        float64
	IsDerby               bool
	DistanceKm            float64
}

type FeatureEngine struct {
	db        *gorm.DB
	matchRepo *repository.MatchRepo
	teamRepo  *repository.TeamRepo
}

func NewFeatureEngine(db *gorm.DB, matchRepo *repository.MatchRepo, teamRepo *repository.TeamRepo) *FeatureEngine {
	return &FeatureEngine{db: db, matchRepo: matchRepo, teamRepo: teamRepo}
}

func (e *FeatureEngine) Calculate(homeTeamID, awayTeamID int64, formCount int) (*Features, error) {
	homeMatches, err := e.matchRepo.FindByTeam(homeTeamID, formCount)
	if err != nil {
		return nil, fmt.Errorf("loading home team matches: %w", err)
	}

	awayMatches, err := e.matchRepo.FindByTeam(awayTeamID, formCount)
	if err != nil {
		return nil, fmt.Errorf("loading away team matches: %w", err)
	}

	f := &Features{
		HomeFormGoalsScored:   avgGoalsScored(homeMatches, homeTeamID, true),
		HomeFormGoalsConceded: avgGoalsConceded(homeMatches, homeTeamID, true),
		AwayFormGoalsScored:   avgGoalsScored(awayMatches, awayTeamID, false),
		AwayFormGoalsConceded: avgGoalsConceded(awayMatches, awayTeamID, false),
		HomeXgLast5:           e.avgXg(homeMatches, homeTeamID, true),
		AwayXgLast5:           e.avgXg(awayMatches, awayTeamID, false),
		HomeCornersAvg:        e.avgCorners(homeMatches, homeTeamID, true),
		AwayCornersAvg:        e.avgCorners(awayMatches, awayTeamID, false),
	}

	return f, nil
}

func avgGoalsScored(matches []database.Match, teamID int64, isHome bool) float64 {
	if len(matches) == 0 {
		return 1.0
	}
	total := 0
	count := 0
	for _, m := range matches {
		if isHome && m.HomeTeamID == teamID {
			total += m.HomeGoals
			count++
		} else if !isHome && m.AwayTeamID == teamID {
			total += m.AwayGoals
			count++
		}
	}
	if count == 0 {
		return 1.0
	}
	return float64(total) / float64(count)
}

func avgGoalsConceded(matches []database.Match, teamID int64, isHome bool) float64 {
	if len(matches) == 0 {
		return 1.0
	}
	total := 0
	count := 0
	for _, m := range matches {
		if isHome && m.HomeTeamID == teamID {
			total += m.AwayGoals
			count++
		} else if !isHome && m.AwayTeamID == teamID {
			total += m.HomeGoals
			count++
		}
	}
	if count == 0 {
		return 1.0
	}
	return float64(total) / float64(count)
}

func (e *FeatureEngine) avgXg(matches []database.Match, teamID int64, isHome bool) float64 {
	if len(matches) == 0 {
		return 0.5
	}
	total := 0.0
	count := 0
	for _, m := range matches {
		var stat database.MatchStat
		err := e.db.Where("match_id = ?", m.ID).First(&stat).Error
		if err != nil {
			continue
		}
		if isHome && m.HomeTeamID == teamID {
			total += stat.HomeXg
			count++
		} else if !isHome && m.AwayTeamID == teamID {
			total += stat.AwayXg
			count++
		}
	}
	if count == 0 {
		return 0.5
	}
	return total / float64(count)
}

func (e *FeatureEngine) avgCorners(matches []database.Match, teamID int64, isHome bool) float64 {
	if len(matches) == 0 {
		return 4.0
	}
	total := 0
	count := 0
	for _, m := range matches {
		var stat database.MatchStat
		err := e.db.Where("match_id = ?", m.ID).First(&stat).Error
		if err != nil {
			continue
		}
		if isHome && m.HomeTeamID == teamID {
			total += stat.HomeCorners
			count++
		} else if !isHome && m.AwayTeamID == teamID {
			total += stat.AwayCorners
			count++
		}
	}
	if count == 0 {
		return 4.0
	}
	return float64(total) / float64(count)
}

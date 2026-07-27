package repository

import (
	"fmt"
	"time"

	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
)

type MatchRepo struct {
	db *gorm.DB
}

func NewMatchRepo(db *gorm.DB) *MatchRepo {
	return &MatchRepo{db: db}
}

func (r *MatchRepo) FindByTeam(teamID int64, limit int) ([]database.Match, error) {
	var matches []database.Match
	result := r.db.Where("home_team_id = ? OR away_team_id = ?", teamID, teamID).
		Order("date DESC").
		Limit(limit).
		Find(&matches)
	if result.Error != nil {
		return nil, fmt.Errorf("finding matches for team %d: %w", teamID, result.Error)
	}
	return matches, nil
}

func (r *MatchRepo) FindBySeason(season int) ([]database.Match, error) {
	var matches []database.Match
	result := r.db.Where("season = ?", season).
		Order("date ASC").
		Find(&matches)
	if result.Error != nil {
		return nil, fmt.Errorf("finding matches for season %d: %w", season, result.Error)
	}
	return matches, nil
}

func (r *MatchRepo) FindByDateRange(from, to time.Time) ([]database.Match, error) {
	var matches []database.Match
	result := r.db.Where("date BETWEEN ? AND ?", from, to).
		Order("date ASC").
		Find(&matches)
	if result.Error != nil {
		return nil, fmt.Errorf("finding matches in date range: %w", result.Error)
	}
	return matches, nil
}

func (r *MatchRepo) BulkCreate(matches []database.Match) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, m := range matches {
			err := tx.Where("date = ? AND home_team_id = ? AND away_team_id = ?",
				m.Date, m.HomeTeamID, m.AwayTeamID).
				Assign(m).
				FirstOrCreate(&m).Error
			if err != nil {
				return fmt.Errorf("creating match: %w", err)
			}
		}
		return nil
	})
}

func (r *MatchRepo) Count() (int64, error) {
	var count int64
	result := r.db.Model(&database.Match{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("counting matches: %w", result.Error)
	}
	return count, nil
}

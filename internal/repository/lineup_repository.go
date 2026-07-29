package repository

import (
	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
)

type LineupRepository struct {
	db *gorm.DB
}

func NewLineupRepository(db *gorm.DB) *LineupRepository {
	return &LineupRepository{db: db}
}

func (r *LineupRepository) BulkCreate(lineups []database.MatchLineup) error {
	return r.db.CreateInBatches(lineups, 100).Error
}

func (r *LineupRepository) BulkCreatePlayerStats(stats []database.MatchPlayerStat) error {
	return r.db.CreateInBatches(stats, 100).Error
}

func (r *LineupRepository) BulkCreateSubstitutions(subs []database.MatchSubstitution) error {
	return r.db.CreateInBatches(subs, 100).Error
}

func (r *LineupRepository) ListByMatch(matchID uint) ([]database.MatchLineup, error) {
	var lineups []database.MatchLineup
	err := r.db.Where("match_id = ?", matchID).
		Preload("Player").
		Order("is_starter DESC, position ASC").
		Find(&lineups).Error
	return lineups, err
}

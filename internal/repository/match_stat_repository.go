package repository

import (
	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
)

type MatchStatRepository struct {
	db *gorm.DB
}

func NewMatchStatRepository(db *gorm.DB) *MatchStatRepository {
	return &MatchStatRepository{db: db}
}

func (r *MatchStatRepository) Create(stat *database.MatchStat) error {
	return r.db.Create(stat).Error
}

func (r *MatchStatRepository) BulkCreate(stats []database.MatchStat) error {
	return r.db.CreateInBatches(stats, 100).Error
}

func (r *MatchStatRepository) FindByMatchID(matchID uint) (*database.MatchStat, error) {
	var stat database.MatchStat
	err := r.db.Where("match_id = ?", matchID).First(&stat).Error
	if err != nil {
		return nil, err
	}
	return &stat, nil
}

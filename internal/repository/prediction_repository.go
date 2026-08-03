package repository

import (
	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
)

type PredictionRepository struct {
	db *gorm.DB
}

func NewPredictionRepository(db *gorm.DB) *PredictionRepository {
	return &PredictionRepository{db: db}
}

func (r *PredictionRepository) Create(prediction *database.Prediction) error {
	return r.db.Create(prediction).Error
}

func (r *PredictionRepository) ListRecent(limit int) ([]database.Prediction, error) {
	if limit <= 0 {
		limit = 20
	}
	var predictions []database.Prediction
	err := r.db.
		Preload("HomeTeam").
		Preload("AwayTeam").
		Order("created_at DESC").
		Limit(limit).
		Find(&predictions).Error
	return predictions, err
}

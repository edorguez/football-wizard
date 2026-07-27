package repository

import (
	"fmt"

	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
)

type PredictionRepo struct {
	db *gorm.DB
}

func NewPredictionRepo(db *gorm.DB) *PredictionRepo {
	return &PredictionRepo{db: db}
}

func (r *PredictionRepo) Create(prediction *database.Prediction) error {
	err := r.db.Create(prediction).Error
	if err != nil {
		return fmt.Errorf("creating prediction: %w", err)
	}
	return nil
}

func (r *PredictionRepo) FindLast(limit int) ([]database.Prediction, error) {
	var predictions []database.Prediction
	result := r.db.Order("created_at DESC").
		Limit(limit).
		Find(&predictions)
	if result.Error != nil {
		return nil, fmt.Errorf("finding last %d predictions: %w", limit, result.Error)
	}
	return predictions, nil
}

func (r *PredictionRepo) FindByFixtureID(fixtureID int64) (*database.Prediction, error) {
	var prediction database.Prediction
	result := r.db.Where("fixture_id = ?", fixtureID).
		Order("created_at DESC").
		First(&prediction)
	if result.Error != nil {
		return nil, fmt.Errorf("finding prediction for fixture %d: %w", fixtureID, result.Error)
	}
	return &prediction, nil
}

package repository

import (
	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
)

type FixtureRepository struct {
	db *gorm.DB
}

func NewFixtureRepository(db *gorm.DB) *FixtureRepository {
	return &FixtureRepository{db: db}
}

func (r *FixtureRepository) BulkCreate(fixtures []database.Fixture) error {
	return r.db.CreateInBatches(fixtures, 100).Error
}

func (r *FixtureRepository) ListUpcoming() ([]database.Fixture, error) {
	var fixtures []database.Fixture
	err := r.db.Where("status = ?", "scheduled").
		Preload("HomeTeam").
		Preload("AwayTeam").
		Order("date ASC").
		Find(&fixtures).Error
	return fixtures, err
}

func (r *FixtureRepository) ListBySeason(season int) ([]database.Fixture, error) {
	var fixtures []database.Fixture
	err := r.db.Where("season = ?", season).
		Preload("HomeTeam").
		Preload("AwayTeam").
		Order("date ASC").
		Find(&fixtures).Error
	return fixtures, err
}

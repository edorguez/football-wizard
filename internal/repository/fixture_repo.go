package repository

import (
	"fmt"
	"time"

	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
)

type FixtureRepo struct {
	db *gorm.DB
}

func NewFixtureRepo(db *gorm.DB) *FixtureRepo {
	return &FixtureRepo{db: db}
}

func (r *FixtureRepo) FindUpcoming(limit int) ([]database.Fixture, error) {
	var fixtures []database.Fixture
	result := r.db.Where("date >= ? AND status = ?", time.Now(), "scheduled").
		Order("date ASC").
		Limit(limit).
		Find(&fixtures)
	if result.Error != nil {
		return nil, fmt.Errorf("finding upcoming fixtures: %w", result.Error)
	}
	return fixtures, nil
}

func (r *FixtureRepo) FindByDate(date time.Time) ([]database.Fixture, error) {
	var fixtures []database.Fixture
	result := r.db.Where("date(date) = date(?)", date).
		Find(&fixtures)
	if result.Error != nil {
		return nil, fmt.Errorf("finding fixtures by date: %w", result.Error)
	}
	return fixtures, nil
}

func (r *FixtureRepo) FindByID(id int64) (*database.Fixture, error) {
	var fixture database.Fixture
	result := r.db.First(&fixture, id)
	if result.Error != nil {
		return nil, fmt.Errorf("finding fixture %d: %w", id, result.Error)
	}
	return &fixture, nil
}

func (r *FixtureRepo) BulkCreate(fixtures []database.Fixture) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, f := range fixtures {
			err := tx.Where("date = ? AND home_team_id = ? AND away_team_id = ?",
				f.Date, f.HomeTeamID, f.AwayTeamID).
				Assign(f).
				FirstOrCreate(&f).Error
			if err != nil {
				return fmt.Errorf("creating fixture: %w", err)
			}
		}
		return nil
	})
}

func (r *FixtureRepo) UpdateStatus(id int64, status string) error {
	result := r.db.Model(&database.Fixture{}).Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("updating fixture %d status: %w", id, result.Error)
	}
	return nil
}

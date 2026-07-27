package repository

import (
	"fmt"

	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
)

type RefereeRepo struct {
	db *gorm.DB
}

func NewRefereeRepo(db *gorm.DB) *RefereeRepo {
	return &RefereeRepo{db: db}
}

func (r *RefereeRepo) FindByName(name string) (*database.Referee, error) {
	var ref database.Referee
	result := r.db.Where("name = ?", name).First(&ref)
	if result.Error != nil {
		return nil, fmt.Errorf("finding referee %q: %w", name, result.Error)
	}
	return &ref, nil
}

func (r *RefereeRepo) BulkUpsert(referees []database.Referee) error {
	for _, ref := range referees {
		err := r.db.Where("name = ?", ref.Name).Assign(ref).FirstOrCreate(&ref).Error
		if err != nil {
			return fmt.Errorf("upserting referee %q: %w", ref.Name, err)
		}
	}
	return nil
}

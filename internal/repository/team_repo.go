package repository

import (
	"fmt"

	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
)

type TeamRepo struct {
	db *gorm.DB
}

func NewTeamRepo(db *gorm.DB) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) FindAll() ([]database.Team, error) {
	var teams []database.Team
	result := r.db.Order("name ASC").Find(&teams)
	if result.Error != nil {
		return nil, fmt.Errorf("finding all teams: %w", result.Error)
	}
	return teams, nil
}

func (r *TeamRepo) FindByID(id int64) (*database.Team, error) {
	var team database.Team
	result := r.db.First(&team, id)
	if result.Error != nil {
		return nil, fmt.Errorf("finding team %d: %w", id, result.Error)
	}
	return &team, nil
}

func (r *TeamRepo) FindByName(name string) (*database.Team, error) {
	var team database.Team
	result := r.db.Where("name = ?", name).First(&team)
	if result.Error != nil {
		return nil, fmt.Errorf("finding team %q: %w", name, result.Error)
	}
	return &team, nil
}

func (r *TeamRepo) BulkUpsert(teams []database.Team) error {
	for _, team := range teams {
		err := r.db.Where("name = ?", team.Name).Assign(team).FirstOrCreate(&team).Error
		if err != nil {
			return fmt.Errorf("upserting team %q: %w", team.Name, err)
		}
	}
	return nil
}

func (r *TeamRepo) Count() (int64, error) {
	var count int64
	result := r.db.Model(&database.Team{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("counting teams: %w", result.Error)
	}
	return count, nil
}

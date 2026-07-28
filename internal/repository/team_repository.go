package repository

import (
	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TeamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Upsert(team *database.Team) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		UpdateAll: true,
	}).Create(team).Error
}

func (r *TeamRepository) FindByName(name string) (*database.Team, error) {
	var team database.Team
	err := r.db.Where("name = ?", name).First(&team).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) FindByID(id uint) (*database.Team, error) {
	var team database.Team
	err := r.db.First(&team, id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) ListAll() ([]database.Team, error) {
	var teams []database.Team
	err := r.db.Order("name ASC").Find(&teams).Error
	return teams, err
}

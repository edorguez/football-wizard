package repository

import (
	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RefereeRepository struct {
	db *gorm.DB
}

func NewRefereeRepository(db *gorm.DB) *RefereeRepository {
	return &RefereeRepository{db: db}
}

func (r *RefereeRepository) Upsert(referee *database.Referee) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		UpdateAll: true,
	}).Create(referee).Error
}

func (r *RefereeRepository) FindByName(name string) (*database.Referee, error) {
	var referee database.Referee
	err := r.db.Where("name = ?", name).First(&referee).Error
	if err != nil {
		return nil, err
	}
	return &referee, nil
}

func (r *RefereeRepository) FindByID(id uint) (*database.Referee, error) {
	var referee database.Referee
	err := r.db.First(&referee, id).Error
	if err != nil {
		return nil, err
	}
	return &referee, nil
}

func (r *RefereeRepository) Search(name string) ([]database.Referee, error) {
	var referees []database.Referee
	err := r.db.Where("name LIKE ?", "%"+name+"%").Order("name ASC").Limit(20).Find(&referees).Error
	return referees, err
}

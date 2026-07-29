package repository

import (
	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PlayerRepository struct {
	db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) *PlayerRepository {
	return &PlayerRepository{db: db}
}

func (r *PlayerRepository) Upsert(player *database.Player) error {
	if player.DateOfBirth != nil {
		return r.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}, {Name: "date_of_birth"}},
			UpdateAll: true,
		}).Create(player).Error
	}

	return r.db.Where("name = ? AND date_of_birth IS NULL", player.Name).
		Assign(*player).
		FirstOrCreate(player).Error
}

func (r *PlayerRepository) FindByID(id uint) (*database.Player, error) {
	var player database.Player
	err := r.db.First(&player, id).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *PlayerRepository) FindByName(name string) (*database.Player, error) {
	var player database.Player
	err := r.db.Where("name = ?", name).First(&player).Error
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (r *PlayerRepository) UpsertSquadMember(sm *database.TeamSquadMember) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "team_id"}, {Name: "player_id"}, {Name: "season"}},
		UpdateAll: true,
	}).Create(sm).Error
}

func (r *PlayerRepository) ListSquadByTeamAndSeason(teamID uint, season int) ([]database.TeamSquadMember, error) {
	var members []database.TeamSquadMember
	err := r.db.Where("team_id = ? AND season = ?", teamID, season).
		Preload("Player").
		Find(&members).Error
	return members, err
}

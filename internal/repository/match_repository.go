package repository

import (
	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MatchRepository struct {
	db *gorm.DB
}

func NewMatchRepository(db *gorm.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

func (r *MatchRepository) Upsert(match *database.Match) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "season"}, {Name: "round"}, {Name: "home_team_id"}, {Name: "away_team_id"}},
		UpdateAll: true,
	}).Create(match).Error
}

func (r *MatchRepository) BulkCreate(matches []database.Match) error {
	return r.db.CreateInBatches(matches, 100).Error
}

func (r *MatchRepository) FindByID(id uint) (*database.Match, error) {
	var match database.Match
	err := r.db.Preload("HomeTeam").Preload("AwayTeam").First(&match, id).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *MatchRepository) ListBySeason(season int) ([]database.Match, error) {
	var matches []database.Match
	err := r.db.Where("season = ?", season).
		Preload("HomeTeam").
		Preload("AwayTeam").
		Order("date ASC").
		Find(&matches).Error
	return matches, err
}

func (r *MatchRepository) ListByTeam(teamID uint) ([]database.Match, error) {
	var matches []database.Match
	err := r.db.Where("home_team_id = ? OR away_team_id = ?", teamID, teamID).
		Preload("HomeTeam").
		Preload("AwayTeam").
		Order("date DESC").
		Limit(10).
		Find(&matches).Error
	return matches, err
}

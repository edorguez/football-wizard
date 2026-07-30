package repository

import (
	"github.com/edorguez/football-wizard/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LineupRepository struct {
	db *gorm.DB
}

func NewLineupRepository(db *gorm.DB) *LineupRepository {
	return &LineupRepository{db: db}
}

func (r *LineupRepository) UpsertLineups(lineups []database.MatchLineup) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "match_id"}, {Name: "team_id"}, {Name: "player_id"}},
		UpdateAll: true,
	}).CreateInBatches(lineups, 100).Error
}

func (r *LineupRepository) UpsertPlayerStats(stats []database.MatchPlayerStat) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "match_id"}, {Name: "player_id"}},
		UpdateAll: true,
	}).CreateInBatches(stats, 100).Error
}

func (r *LineupRepository) UpsertSubstitutions(subs []database.MatchSubstitution) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "match_id"},
			{Name: "team_id"},
			{Name: "player_on_id"},
			{Name: "minute"},
		},
		DoNothing: true,
	}).CreateInBatches(subs, 100).Error
}

func (r *LineupRepository) ListByMatch(matchID uint) ([]database.MatchLineup, error) {
	var lineups []database.MatchLineup
	err := r.db.Where("match_id = ?", matchID).
		Preload("Player").
		Order("is_starter DESC, position ASC").
		Find(&lineups).Error
	return lineups, err
}

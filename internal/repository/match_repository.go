package repository

import (
	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/model"
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

func (r *MatchRepository) DB() *gorm.DB {
	return r.db
}

func (r *MatchRepository) FindBySeasonRoundTeams(season, round int, homeTeam, awayTeam string) (*database.Match, error) {
	var match database.Match
	err := r.db.
		Joins("JOIN teams AS ht ON ht.id = matches.home_team_id").
		Joins("JOIN teams AS at ON at.id = matches.away_team_id").
		Where("matches.season = ? AND matches.round = ? AND ht.name = ? AND at.name = ?", season, round, homeTeam, awayTeam).
		Preload("HomeTeam").Preload("AwayTeam").
		First(&match).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *MatchRepository) ListByReferee(refereeID uint) ([]database.Match, error) {
	var matches []database.Match
	err := r.db.Where("referee_id = ?", refereeID).
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

// ListRows returns all completed matches paired with their aggregated stats,
// shaped for the prediction models.
func (r *MatchRepository) ListRows() ([]model.MatchRow, error) {
	var matches []database.Match
	if err := r.db.Where("status = ?", "completed").Order("date ASC").Find(&matches).Error; err != nil {
		return nil, err
	}

	var stats []database.MatchStat
	if err := r.db.Find(&stats).Error; err != nil {
		return nil, err
	}

	statsByMatch := map[uint]database.MatchStat{}
	for _, s := range stats {
		statsByMatch[s.MatchID] = s
	}

	rows := make([]model.MatchRow, 0, len(matches))
	for _, m := range matches {
		row := model.MatchRow{
			ID:         m.ID,
			Season:     m.Season,
			Round:      m.Round,
			Date:       m.Date,
			HomeTeamID: m.HomeTeamID,
			AwayTeamID: m.AwayTeamID,
			HomeGoals:  m.HomeGoals,
			AwayGoals:  m.AwayGoals,
			HomeXG:     m.HomeXG,
			AwayXG:     m.AwayXG,
		}

		if s, ok := statsByMatch[m.ID]; ok {
			row.HomeCorners = s.HomeCorners
			row.AwayCorners = s.AwayCorners
			row.HomeOffsides = s.HomeOffsides
			row.AwayOffsides = s.AwayOffsides
			row.HomeYellowCards = s.HomeYellowCards
			row.AwayYellowCards = s.AwayYellowCards
			row.HomeRedCards = s.HomeRedCards
			row.AwayRedCards = s.AwayRedCards
			row.HomeShots = s.HomeShots
			row.AwayShots = s.AwayShots
			row.HomeShotsOnTarget = s.HomeShotsOnTarget
			row.AwayShotsOnTarget = s.AwayShotsOnTarget
			row.HomeSaves = s.HomeSaves
			row.AwaySaves = s.AwaySaves
			row.HomeGoalsFirstHalf = s.HomeGoalsFirstHalf
			row.AwayGoalsFirstHalf = s.AwayGoalsFirstHalf
			row.HomeGoalsSecondHalf = s.HomeGoalsSecondHalf
			row.AwayGoalsSecondHalf = s.AwayGoalsSecondHalf
			row.HomeFirstGoalMinute = s.HomeFirstGoalMinute
			row.AwayFirstGoalMinute = s.AwayFirstGoalMinute
			row.HomeSecondGoalMinute = s.HomeSecondGoalMinute
			row.AwaySecondGoalMinute = s.AwaySecondGoalMinute
		}

		rows = append(rows, row)
	}
	return rows, nil
}

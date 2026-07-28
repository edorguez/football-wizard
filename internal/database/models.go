package database

import "time"

type Team struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	ShortName string    `json:"short_name"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Referee struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Nationality string    `json:"nationality"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Match struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Season      int       `gorm:"uniqueIndex:idx_season_round_teams;not null" json:"season"`
	Round       int       `gorm:"uniqueIndex:idx_season_round_teams;not null" json:"round"`
	Date        time.Time `gorm:"index" json:"date"`
	HomeTeamID  uint      `gorm:"uniqueIndex:idx_season_round_teams;not null" json:"home_team_id"`
	AwayTeamID  uint      `gorm:"uniqueIndex:idx_season_round_teams;not null" json:"away_team_id"`
	HomeGoals   *int      `json:"home_goals"`
	AwayGoals   *int      `json:"away_goals"`
	RefereeID   *uint     `json:"referee_id"`
	Status      string    `gorm:"default:scheduled" json:"status"`
	HomeTeam    Team      `gorm:"foreignKey:HomeTeamID" json:"-"`
	AwayTeam    Team      `gorm:"foreignKey:AwayTeamID" json:"-"`
	Referee     *Referee  `gorm:"foreignKey:RefereeID" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MatchStat struct {
	ID              uint   `gorm:"primaryKey" json:"id"`
	MatchID         uint   `gorm:"uniqueIndex;not null" json:"match_id"`
	HomeShots       *int   `json:"home_shots"`
	AwayShots       *int   `json:"away_shots"`
	HomeShotsOnTarget *int `json:"home_shots_on_target"`
	AwayShotsOnTarget *int `json:"away_shots_on_target"`
	HomeCorners     *int   `json:"home_corners"`
	AwayCorners     *int   `json:"away_corners"`
	HomeYellowCards *int   `json:"home_yellow_cards"`
	AwayYellowCards *int   `json:"away_yellow_cards"`
	HomeRedCards    *int   `json:"home_red_cards"`
	AwayRedCards    *int   `json:"away_red_cards"`
}

type Fixture struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Season     int       `gorm:"index;not null" json:"season"`
	Round      int       `json:"round"`
	Date       time.Time `gorm:"index" json:"date"`
	HomeTeamID uint      `gorm:"not null" json:"home_team_id"`
	AwayTeamID uint      `gorm:"not null" json:"away_team_id"`
	Status     string    `gorm:"default:scheduled" json:"status"`
	HomeTeam   Team      `gorm:"foreignKey:HomeTeamID" json:"-"`
	AwayTeam   Team      `gorm:"foreignKey:AwayTeamID" json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

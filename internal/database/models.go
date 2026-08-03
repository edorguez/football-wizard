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

type Player struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"uniqueIndex:idx_player_name_dob;not null" json:"name"`
	DateOfBirth *time.Time `gorm:"uniqueIndex:idx_player_name_dob" json:"date_of_birth"`
	Nationality string     `json:"nationality"`
	Position    string     `json:"position"`
	Foot        string     `json:"foot"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TeamSquadMember struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	TeamID   uint   `gorm:"uniqueIndex:idx_team_season_player;not null" json:"team_id"`
	PlayerID uint   `gorm:"uniqueIndex:idx_team_season_player;not null" json:"player_id"`
	Season   int    `gorm:"uniqueIndex:idx_team_season_player;not null" json:"season"`
	ShirtNum *int   `json:"shirt_num"`
	Player   Player `gorm:"foreignKey:PlayerID" json:"-"`
	Team     Team   `gorm:"foreignKey:TeamID" json:"-"`
}

type Match struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Season         int       `gorm:"uniqueIndex:idx_season_round_teams;not null" json:"season"`
	Round          int       `gorm:"uniqueIndex:idx_season_round_teams;not null" json:"round"`
	Date           time.Time `gorm:"index" json:"date"`
	HomeTeamID     uint      `gorm:"uniqueIndex:idx_season_round_teams;not null" json:"home_team_id"`
	AwayTeamID     uint      `gorm:"uniqueIndex:idx_season_round_teams;not null" json:"away_team_id"`
	HomeGoals      *int      `json:"home_goals"`
	AwayGoals      *int      `json:"away_goals"`
	HomeXG         *float64  `json:"home_xg"`
	AwayXG         *float64  `json:"away_xg"`
	Venue          string    `json:"venue"`
	Attendance     *int      `json:"attendance"`
	RefereeID      *uint     `json:"referee_id"`
	MatchReportURL string    `json:"match_report_url"`
	Status         string    `gorm:"default:scheduled" json:"status"`
	HomeTeam       Team      `gorm:"foreignKey:HomeTeamID" json:"-"`
	AwayTeam       Team      `gorm:"foreignKey:AwayTeamID" json:"-"`
	Referee        *Referee  `gorm:"foreignKey:RefereeID" json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type MatchStat struct {
	ID                   uint `gorm:"primaryKey" json:"id"`
	MatchID              uint `gorm:"uniqueIndex;not null" json:"match_id"`
	HomeShots            *int `json:"home_shots"`
	AwayShots            *int `json:"away_shots"`
	HomeShotsOnTarget    *int `json:"home_shots_on_target"`
	AwayShotsOnTarget    *int `json:"away_shots_on_target"`
	HomeShotsOffTarget   *int `json:"home_shots_off_target"`
	AwayShotsOffTarget   *int `json:"away_shots_off_target"`
	HomePossession       *int `json:"home_possession"`
	AwayPossession       *int `json:"away_possession"`
	HomeSaves            *int `json:"home_saves"`
	AwaySaves            *int `json:"away_saves"`
	HomeFouls            *int `json:"home_fouls"`
	AwayFouls            *int `json:"away_fouls"`
	HomeCorners          *int `json:"home_corners"`
	AwayCorners          *int `json:"away_corners"`
	HomeCrosses          *int `json:"home_crosses"`
	AwayCrosses          *int `json:"away_crosses"`
	HomeOffsides         *int `json:"home_offsides"`
	AwayOffsides         *int `json:"away_offsides"`
	HomeTackles          *int `json:"home_tackles"`
	AwayTackles          *int `json:"away_tackles"`
	HomeInterceptions    *int `json:"home_interceptions"`
	AwayInterceptions    *int `json:"away_interceptions"`
	HomeYellowCards      *int `json:"home_yellow_cards"`
	AwayYellowCards      *int `json:"away_yellow_cards"`
	HomeRedCards         *int `json:"home_red_cards"`
	AwayRedCards         *int `json:"away_red_cards"`
	HomeGoalsFirstHalf   *int `json:"home_goals_first_half"`
	AwayGoalsFirstHalf   *int `json:"away_goals_first_half"`
	HomeGoalsSecondHalf  *int `json:"home_goals_second_half"`
	AwayGoalsSecondHalf  *int `json:"away_goals_second_half"`
	HomeFirstGoalMinute  *int `json:"home_first_goal_minute"`
	AwayFirstGoalMinute  *int `json:"away_first_goal_minute"`
	HomeSecondGoalMinute *int `json:"home_second_half_goal_minute"`
	AwaySecondGoalMinute *int `json:"away_second_half_goal_minute"`
}

type MatchLineup struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	MatchID   uint   `gorm:"uniqueIndex:idx_lineup;not null" json:"match_id"`
	TeamID    uint   `gorm:"uniqueIndex:idx_lineup;not null" json:"team_id"`
	PlayerID  uint   `gorm:"uniqueIndex:idx_lineup;not null" json:"player_id"`
	IsStarter bool   `json:"is_starter"`
	Position  string `json:"position"`
	ShirtNum  *int   `json:"shirt_num"`
	Player    Player `gorm:"foreignKey:PlayerID" json:"-"`
	Match     Match  `gorm:"foreignKey:MatchID" json:"-"`
	Team      Team   `gorm:"foreignKey:TeamID" json:"-"`
}

type MatchPlayerStat struct {
	ID            uint     `gorm:"primaryKey" json:"id"`
	MatchID       uint     `gorm:"uniqueIndex:idx_match_player;not null" json:"match_id"`
	TeamID        uint     `gorm:"not null" json:"team_id"`
	PlayerID      uint     `gorm:"uniqueIndex:idx_match_player;not null" json:"player_id"`
	MinutesPlayed *int     `json:"minutes_played"`
	Goals         *int     `json:"goals"`
	Assists       *int     `json:"assists"`
	Shots         *int     `json:"shots"`
	ShotsOnTarget *int     `json:"shots_on_target"`
	Passes        *int     `json:"passes"`
	PassAccuracy  *float64 `json:"pass_accuracy"`
	Tackles       *int     `json:"tackles"`
	Interceptions *int     `json:"interceptions"`
	Fouls         *int     `json:"fouls"`
	Fouled        *int     `json:"fouled"`
	Offsides      *int     `json:"offsides"`
	Crosses       *int     `json:"crosses"`
	YellowCards   *int     `json:"yellow_cards"`
	RedCards      *int     `json:"red_cards"`
	Saves         *int     `json:"saves"`
	Player        Player   `gorm:"foreignKey:PlayerID" json:"-"`
	Match         Match    `gorm:"foreignKey:MatchID" json:"-"`
	Team          Team     `gorm:"foreignKey:TeamID" json:"-"`
}

type MatchSubstitution struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	MatchID     uint   `gorm:"uniqueIndex:idx_subs;not null" json:"match_id"`
	TeamID      uint   `gorm:"uniqueIndex:idx_subs;not null" json:"team_id"`
	PlayerOffID uint   `gorm:"not null" json:"player_off_id"`
	PlayerOnID  uint   `gorm:"uniqueIndex:idx_subs;not null" json:"player_on_id"`
	Minute      int    `gorm:"uniqueIndex:idx_subs;not null" json:"minute"`
	PlayerOff   Player `gorm:"foreignKey:PlayerOffID" json:"-"`
	PlayerOn    Player `gorm:"foreignKey:PlayerOnID" json:"-"`
	Match       Match  `gorm:"foreignKey:MatchID" json:"-"`
	Team        Team   `gorm:"foreignKey:TeamID" json:"-"`
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

type Prediction struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Season            int       `json:"season"`
	Round             int       `json:"round"`
	HomeTeamID        uint      `gorm:"not null" json:"home_team_id"`
	AwayTeamID        uint      `gorm:"not null" json:"away_team_id"`
	HomeWin           float64   `json:"home_win"`
	Draw              float64   `json:"draw"`
	AwayWin           float64   `json:"away_win"`
	ExpectedHomeGoals float64   `json:"expected_home_goals"`
	ExpectedAwayGoals float64   `json:"expected_away_goals"`
	Payload           string    `gorm:"type:text" json:"payload"`
	CreatedAt         time.Time `json:"created_at"`
	HomeTeam          Team      `gorm:"foreignKey:HomeTeamID" json:"-"`
	AwayTeam          Team      `gorm:"foreignKey:AwayTeamID" json:"-"`
}

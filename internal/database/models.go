package database

import "time"

type Team struct {
	ID       int64  `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"uniqueIndex;not null" json:"name"`
	ShortName string `json:"short_name"`
	City     string `json:"city"`
	Founded  int    `json:"founded"`
	Stadium  string `json:"stadium"`
}

type Player struct {
	ID          int64  `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Position    string `json:"position"`
	Nationality string `json:"nationality"`
	BirthDate   string `json:"birth_date"`
	TeamID      int64  `gorm:"index" json:"team_id"`
}

type Match struct {
	ID           int64     `gorm:"primaryKey" json:"id"`
	Date         time.Time `gorm:"index;not null" json:"date"`
	Season       int       `gorm:"index;not null" json:"season"`
	Matchday     int       `json:"matchday"`
	HomeTeamID   int64     `gorm:"index;not null" json:"home_team_id"`
	AwayTeamID   int64     `gorm:"index;not null" json:"away_team_id"`
	HomeGoals    int       `json:"home_goals"`
	AwayGoals    int       `json:"away_goals"`
	RefereeID    int64     `gorm:"index" json:"referee_id"`
	Stadium      string    `json:"stadium"`
	Attendance   int       `json:"attendance"`
}

type MatchStat struct {
	MatchID          int64   `gorm:"primaryKey" json:"match_id"`
	HomePossession   float64 `json:"home_possession"`
	AwayPossession   float64 `json:"away_possession"`
	HomeShots        int     `json:"home_shots"`
	AwayShots        int     `json:"away_shots"`
	HomeShotsOnTarget int    `json:"home_shots_on_target"`
	AwayShotsOnTarget int    `json:"away_shots_on_target"`
	HomeCorners      int     `json:"home_corners"`
	AwayCorners      int     `json:"away_corners"`
	HomeYellowCards  int     `json:"home_yellow_cards"`
	AwayYellowCards  int     `json:"away_yellow_cards"`
	HomeRedCards     int     `json:"home_red_cards"`
	AwayRedCards     int     `json:"away_red_cards"`
	HomeOffsides     int     `json:"home_offsides"`
	AwayOffsides     int     `json:"away_offsides"`
	HomeFouls        int     `json:"home_fouls"`
	AwayFouls        int     `json:"away_fouls"`
	HomeXg           float64 `json:"home_xg"`
	AwayXg           float64 `json:"away_xg"`
}

type Lineup struct {
	MatchID      int64  `gorm:"primaryKey" json:"match_id"`
	PlayerID     int64  `gorm:"primaryKey" json:"player_id"`
	TeamID       int64  `gorm:"index" json:"team_id"`
	IsStarter    bool   `json:"is_starter"`
	Position     string `json:"position"`
	MinutesPlayed int   `json:"minutes_played"`
	Goals        int    `json:"goals"`
	Assists      int    `json:"assists"`
	YellowCards  int    `json:"yellow_cards"`
	RedCards     int    `json:"red_cards"`
}

type Referee struct {
	ID           int64  `gorm:"primaryKey" json:"id"`
	Name         string `gorm:"uniqueIndex;not null" json:"name"`
	Nationality  string `json:"nationality"`
	MatchesCount int    `json:"matches_count"`
}

type RefereeStat struct {
	RefereeID        int64 `gorm:"primaryKey" json:"referee_id"`
	MatchID          int64 `gorm:"primaryKey" json:"match_id"`
	YellowCardsShown int   `json:"yellow_cards_shown"`
	RedCardsShown    int   `json:"red_cards_shown"`
	PenaltiesAwarded int   `json:"penalties_awarded"`
	FoulsCalled      int   `json:"fouls_called"`
}

type Fixture struct {
	ID         int64     `gorm:"primaryKey" json:"id"`
	Date       time.Time `gorm:"index;not null" json:"date"`
	Season     int       `gorm:"index;not null" json:"season"`
	Matchday   int       `json:"matchday"`
	HomeTeamID int64     `gorm:"index;not null" json:"home_team_id"`
	AwayTeamID int64     `gorm:"index;not null" json:"away_team_id"`
	RefereeID  int64     `gorm:"index" json:"referee_id"`
	Status     string    `gorm:"default:scheduled" json:"status"`
}

type Prediction struct {
	ID                    int64     `gorm:"primaryKey" json:"id"`
	FixtureID             int64     `gorm:"index;not null" json:"fixture_id"`
	CreatedAt             time.Time `gorm:"autoCreateTime" json:"created_at"`
	ModelVersion          string    `json:"model_version"`
	HomeWinProb           float64   `json:"home_win_prob"`
	DrawProb              float64   `json:"draw_prob"`
	AwayWinProb           float64   `json:"away_win_prob"`
	Over05Prob            float64   `json:"over_0_5_prob"`
	Over15Prob            float64   `json:"over_1_5_prob"`
	Over25Prob            float64   `json:"over_2_5_prob"`
	Over35Prob            float64   `json:"over_3_5_prob"`
	BttsYesProb           float64   `json:"btts_yes_prob"`
	Over35YellowProb      float64   `json:"over_3_5_yellow_prob"`
	Over45YellowProb      float64   `json:"over_4_5_yellow_prob"`
	HomeRedProb           float64   `json:"home_red_prob"`
	AwayRedProb           float64   `json:"away_red_prob"`
	Over85CornersProb     float64   `json:"over_8_5_corners_prob"`
	Over95CornersProb     float64   `json:"over_9_5_corners_prob"`
	Over105CornersProb    float64   `json:"over_10_5_corners_prob"`
	HomeFirstHalfProb     float64   `json:"home_first_half_prob"`
	AwayFirstHalfProb     float64   `json:"away_first_half_prob"`
	ConfidenceLevel       string    `json:"confidence_level"`
	ConfidenceScore       float64   `json:"confidence_score"`
}

type ModelFeature struct {
	PredictionID          int64   `gorm:"primaryKey" json:"prediction_id"`
	HomeFormGoalsScored   float64 `json:"home_form_goals_scored"`
	HomeFormGoalsConceded float64 `json:"home_form_goals_conceded"`
	AwayFormGoalsScored   float64 `json:"away_form_goals_scored"`
	AwayFormGoalsConceded float64 `json:"away_form_goals_conceded"`
	HomeXgLast5           float64 `json:"home_xg_last5"`
	AwayXgLast5           float64 `json:"away_xg_last5"`
	RefereeYellowAvg      float64 `json:"referee_yellow_avg"`
	RefereeRedAvg         float64 `json:"referee_red_avg"`
	HomeCornersAvg        float64 `json:"home_corners_avg"`
	AwayCornersAvg        float64 `json:"away_corners_avg"`
	IsDerby               bool    `json:"is_derby"`
	DistanceKm            float64 `json:"distance_km"`
}

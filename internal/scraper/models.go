package scraper

import "time"

type ScrapedMatch struct {
	Season   int
	Round    int
	Date     time.Time
	HomeTeam string
	AwayTeam string

	HomeGoals  int
	AwayGoals  int
	HomeXG     *float64
	AwayXG     *float64
	Venue      string
	Attendance *int

	RefereeName string

	MatchReportURL string
}

type ScrapedFixture struct {
	Season   int
	Round    int
	Date     time.Time
	HomeTeam string
	AwayTeam string
}

type ScrapedPlayer struct {
	Name        string
	Nationality string
	Position    string
	ShirtNum    *int
	Season      int
}

type ScrapedSquad struct {
	TeamFBRefID string
	TeamName    string
	Players     []ScrapedPlayer
}

type ScrapedLineupPlayer struct {
	Name       string
	Position   string
	ShirtNum   *int
	IsStarter  bool
	HasSubIcon bool
}

type ScrapedMatchReport struct {
	HomeTeam string
	AwayTeam string

	HomePossession     *int
	AwayPossession     *int
	HomeShots          *int
	AwayShots          *int
	HomeShotsOnTarget  *int
	AwayShotsOnTarget  *int
	HomeShotsOffTarget *int
	AwayShotsOffTarget *int
	HomeSaves          *int
	AwaySaves          *int
	HomeFouls          *int
	AwayFouls          *int
	HomeCrosses        *int
	AwayCrosses        *int
	HomeCorners        *int
	AwayCorners        *int
	HomeOffsides       *int
	AwayOffsides       *int
	HomeTackles        *int
	AwayTackles        *int

	HomeGoalsFirstHalf   *int
	AwayGoalsFirstHalf   *int
	HomeGoalsSecondHalf  *int
	AwayGoalsSecondHalf  *int
	HomeFirstGoalMinute  *int
	AwayFirstGoalMinute  *int
	HomeSecondGoalMinute *int
	AwaySecondGoalMinute *int

	HomeLineup []ScrapedLineupPlayer
	AwayLineup []ScrapedLineupPlayer

	HomePlayerStats map[string]ScrapedPlayerMatchStat
	AwayPlayerStats map[string]ScrapedPlayerMatchStat

	Substitutions []ScrapedSubstitution
}

type ScrapedPlayerMatchStat struct {
	Name          string
	Position      string
	Minutes       int
	Goals         *int
	Assists       *int
	Shots         *int
	ShotsOnTarget *int
	Passes        *int
	PassAccuracy  *float64
	Tackles       *int
	Interceptions *int
	Fouls         *int
	Fouled        *int
	Offsides      *int
	Crosses       *int
	YellowCards   *int
	RedCards      *int
	Saves         *int
}

type ScrapedSubstitution struct {
	Minute     int
	PlayerOff  string
	PlayerOn   string
	TeamIsHome bool
}

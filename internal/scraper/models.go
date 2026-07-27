package scraper

import "time"

type ScrapedMatch struct {
	Date         time.Time
	Season       int
	Matchday     int
	HomeTeam     string
	AwayTeam     string
	HomeGoals    int
	AwayGoals    int
	Referee      string
	Stadium      string
	Attendance   int
	Possession   [2]float64
	Shots        [2]int
	ShotsOnTarget [2]int
	Corners      [2]int
	YellowCards  [2]int
	RedCards     [2]int
	Offsides     [2]int
	Fouls        [2]int
	Xg           [2]float64
}

type ScrapedFixture struct {
	Date     time.Time
	Season   int
	Matchday int
	HomeTeam string
	AwayTeam string
}

type ScrapedLineup struct {
	MatchDate    time.Time
	HomeTeam     string
	AwayTeam     string
	PlayerName   string
	TeamName     string
	Position     string
	IsStarter    bool
	MinutesPlayed int
	Goals        int
	Assists      int
	YellowCards  int
	RedCards     int
}

type ScrapedTeam struct {
	Name     string
	ShortName string
	City     string
	Founded  int
	Stadium  string
}

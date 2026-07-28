package scraper

import "time"

type ScrapedMatch struct {
	Season   int
	Round    int
	Date     time.Time
	HomeTeam string
	AwayTeam string

	HomeGoals int
	AwayGoals int

	HomeShots        *int
	AwayShots        *int
	HomeShotsOnTarget *int
	AwayShotsOnTarget *int
	HomeCorners      *int
	AwayCorners      *int
	HomeYellowCards  *int
	AwayYellowCards  *int
	HomeRedCards     *int
	AwayRedCards     *int

	RefereeName string
}

type ScrapedFixture struct {
	Season   int
	Round    int
	Date     time.Time
	HomeTeam string
	AwayTeam string
}

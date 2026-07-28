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

	RefereeName string
}

type ScrapedFixture struct {
	Season   int
	Round    int
	Date     time.Time
	HomeTeam string
	AwayTeam string
}

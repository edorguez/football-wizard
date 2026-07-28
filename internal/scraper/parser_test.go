package scraper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testMatchLogsHTML = `<html><body>
<table id="matchlogs">
<tbody>
<tr data-row="1">
	<td data-stat="date">2025-04-01</td>
	<td data-stat="home_team"><a href="/en/squads/123">Flamengo</a></td>
	<td data-stat="score">2–1</td>
	<td data-stat="away_team"><a href="/en/squads/456">Palmeiras</a></td>
	<td data-stat="referee"><a href="/en/referees/xyz">Anderson Daronco</a></td>
	<td data-stat="home_shots">15</td>
	<td data-stat="away_shots">8</td>
	<td data-stat="home_shots_on_target">5</td>
	<td data-stat="away_shots_on_target">3</td>
	<td data-stat="home_corner_kicks">7</td>
	<td data-stat="away_corner_kicks">2</td>
	<td data-stat="home_yellow_cards">2</td>
	<td data-stat="away_yellow_cards">3</td>
	<td data-stat="home_red_cards">0</td>
	<td data-stat="away_red_cards">1</td>
</tr>
<tr data-row="2">
	<td data-stat="date">2025-04-06</td>
	<td data-stat="home_team"><a href="/en/squads/789">Corinthians</a></td>
	<td data-stat="score">1–1</td>
	<td data-stat="away_team"><a href="/en/squads/012">Santos</a></td>
	<td data-stat="referee"><a href="/en/referees/abc">Wilton Sampaio</a></td>
	<td data-stat="home_shots">10</td>
	<td data-stat="away_shots">12</td>
	<td data-stat="home_shots_on_target">4</td>
	<td data-stat="away_shots_on_target">6</td>
	<td data-stat="home_corner_kicks">5</td>
	<td data-stat="away_corner_kicks">4</td>
	<td data-stat="home_yellow_cards">1</td>
	<td data-stat="away_yellow_cards">2</td>
	<td data-stat="home_red_cards">0</td>
	<td data-stat="away_red_cards">0</td>
</tr>
</tbody>
</table></body></html>`

const testMatchLogsNoStatsHTML = `<html><body>
<table id="matchlogs">
<tbody>
<tr data-row="1">
	<td data-stat="date">2025-05-01</td>
	<td data-stat="home_team"><a href="/en/squads/111">Internacional</a></td>
	<td data-stat="score">3–0</td>
	<td data-stat="away_team"><a href="/en/squads/222">Grêmio</a></td>
	<td data-stat="referee"><a href="/en/referees/r1">Raphael Claus</a></td>
</tr>
</tbody>
</table></body></html>`

const testEmptyHTML = `<html><body><table id="matchlogs"><tbody></tbody></table></body></html>`

func TestParseMatchResults(t *testing.T) {
	t.Parallel()

	matches, err := ParseMatchResults(2025, testMatchLogsHTML)

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	must.Len(matches, 2)

	m := matches[0]

	is.Equal(2025, m.Season)
	is.Equal(1, m.Round)
	is.Equal("Flamengo", m.HomeTeam)
	is.Equal("Palmeiras", m.AwayTeam)
	is.Equal(2, m.HomeGoals)
	is.Equal(1, m.AwayGoals)
	is.Equal("Anderson Daronco", m.RefereeName)

	must.NotNil(m.HomeShots)
	is.Equal(15, *m.HomeShots)

	must.NotNil(m.AwayShots)
	is.Equal(8, *m.AwayShots)

	must.NotNil(m.HomeShotsOnTarget)
	is.Equal(5, *m.HomeShotsOnTarget)

	must.NotNil(m.HomeCorners)
	is.Equal(7, *m.HomeCorners)

	must.NotNil(m.AwayYellowCards)
	is.Equal(3, *m.AwayYellowCards)

	must.NotNil(m.AwayRedCards)
	is.Equal(1, *m.AwayRedCards)
}

func TestParseMatchResults_SecondMatch(t *testing.T) {
	t.Parallel()

	matches, err := ParseMatchResults(2025, testMatchLogsHTML)

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	must.Len(matches, 2)

	m := matches[1]

	is.Equal(2, m.Round)
	is.Equal("Corinthians", m.HomeTeam)
	is.Equal("Santos", m.AwayTeam)
	is.Equal(1, m.HomeGoals)
	is.Equal(1, m.AwayGoals)
	is.Equal("Wilton Sampaio", m.RefereeName)

	must.NotNil(m.HomeShots)
	is.Equal(10, *m.HomeShots)

	must.NotNil(m.AwayShots)
	is.Equal(12, *m.AwayShots)
}

func TestParseMatchResults_NoStats(t *testing.T) {
	t.Parallel()

	matches, err := ParseMatchResults(2025, testMatchLogsNoStatsHTML)

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	must.Len(matches, 1)

	m := matches[0]

	is.Equal("Internacional", m.HomeTeam)
	is.Equal("Grêmio", m.AwayTeam)
	is.Equal(3, m.HomeGoals)
	is.Nil(m.HomeShots)
	is.Nil(m.AwayShots)
	is.Equal("Raphael Claus", m.RefereeName)
}

func TestParseMatchResults_Empty(t *testing.T) {
	t.Parallel()

	_, err := ParseMatchResults(2025, testEmptyHTML)

	assert.Error(t, err)
}

func TestParseFixtures(t *testing.T) {
	t.Parallel()

	const html = `<html><body>
<table id="fixtures">
<tbody>
<tr data-row="1">
	<td data-stat="date">2025-05-10</td>
	<td data-stat="home_team"><a href="/squad/a">Flamengo</a></td>
	<td data-stat="away_team"><a href="/squad/b">Palmeiras</a></td>
</tr>
<tr data-row="2">
	<td data-stat="date">2025-05-11</td>
	<td data-stat="home_team"><a href="/squad/c">Corinthians</a></td>
	<td data-stat="away_team"><a href="/squad/d">Santos</a></td>
</tr>
</tbody>
</table></body></html>`

	fixtures, err := ParseFixtures(2025, html)

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	must.Len(fixtures, 2)

	is.Equal("Flamengo", fixtures[0].HomeTeam)
	is.Equal("Palmeiras", fixtures[0].AwayTeam)
	is.Equal("Corinthians", fixtures[1].HomeTeam)
}

func TestParseMatchResults_IntAttr(t *testing.T) {
	t.Parallel()

	matches, err := ParseMatchResults(2025, testMatchLogsHTML)
	require.NoError(t, err)
	require.Len(t, matches, 2)

	m := matches[0]

	tests := []struct {
		name string
		val  *int
		want int
	}{
		{name: "home_shots", val: m.HomeShots, want: 15},
		{name: "away_shots", val: m.AwayShots, want: 8},
		{name: "home_yellow", val: m.HomeYellowCards, want: 2},
		{name: "away_red", val: m.AwayRedCards, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)

			is.NotNil(tt.val)
			is.Equal(tt.want, *tt.val)
		})
	}
}

func TestParseFixtures_Empty(t *testing.T) {
	t.Parallel()

	const html = `<html><body><table id="fixtures"><tbody></tbody></table></body></html>`

	fixtures, err := ParseFixtures(2025, html)

	is := assert.New(t)

	is.NoError(err)
	is.Empty(fixtures)
}

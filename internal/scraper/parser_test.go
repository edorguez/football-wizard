package scraper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testScheduleHTML = `<html><body>
<table class="stats_table sortable min_width" id="sched_2025_24_1">
<thead>
<tr>
	<th data-stat="gameweek">Wk</th>
	<th data-stat="dayofweek">Day</th>
	<th data-stat="date">Date</th>
	<th data-stat="start_time">Time</th>
	<th data-stat="home_team">Home</th>
	<th data-stat="home_xg">xG</th>
	<th data-stat="score">Score</th>
	<th data-stat="away_xg">xG</th>
	<th data-stat="away_team">Away</th>
	<th data-stat="attendance">Attendance</th>
	<th data-stat="venue">Venue</th>
	<th data-stat="referee">Referee</th>
	<th data-stat="match_report">Match Report</th>
	<th data-stat="notes">Notes</th>
</tr>
</thead>
<tbody>
<tr>
	<th scope="row" class="right" data-stat="gameweek">1</th>
	<td class="left" data-stat="dayofweek" csk="7">Sat</td>
	<td class="left" data-stat="date" csk="20250329"><a href="/en/matches/2025-03-29">2025-03-29</a></td>
	<td class="right" data-stat="start_time" csk="18:30:00">18:30</td>
	<td class="right" data-stat="home_team"><a href="/en/squads/123">Flamengo</a></td>
	<td class="right" data-stat="home_xg">2.4</td>
	<td class="center" data-stat="score"><a href="/en/matches/123">2–1</a></td>
	<td class="right" data-stat="away_xg">0.8</td>
	<td class="left" data-stat="away_team"><a href="/en/squads/456">Palmeiras</a></td>
	<td class="right" data-stat="attendance" csk="37856">37,856</td>
	<td class="left" data-stat="venue">Estádio do Maracanã</td>
	<td class="left" data-stat="referee" csk="Anderson Daronco2025-03-29">Anderson Daronco</td>
	<td class="left" data-stat="match_report"><a href="/en/matches/123">Match Report</a></td>
	<td class="left iz" data-stat="notes"></td>
</tr>
<tr>
	<th scope="row" class="right" data-stat="gameweek">2</th>
	<td class="left" data-stat="dayofweek" csk="7">Sat</td>
	<td class="left" data-stat="date" csk="20250405"><a href="/en/matches/2025-04-05">2025-04-05</a></td>
	<td class="right" data-stat="start_time" csk="18:30:00">18:30</td>
	<td class="right" data-stat="home_team"><a href="/en/squads/789">Corinthians</a></td>
	<td class="right" data-stat="home_xg">1.2</td>
	<td class="center" data-stat="score"><a href="/en/matches/456">1–1</a></td>
	<td class="right" data-stat="away_xg">1.1</td>
	<td class="left" data-stat="away_team"><a href="/en/squads/012">Santos</a></td>
	<td class="right" data-stat="attendance" csk="41200">41,200</td>
	<td class="left" data-stat="venue">Neo Química Arena</td>
	<td class="left" data-stat="referee" csk="Wilton Sampaio2025-04-05">Wilton Sampaio</td>
	<td class="left" data-stat="match_report"><a href="/en/matches/456">Match Report</a></td>
	<td class="left iz" data-stat="notes"></td>
</tr>
<tr>
	<th scope="row" class="right" data-stat="gameweek">3</th>
	<td class="left" data-stat="dayofweek" csk="1">Sun</td>
	<td class="left" data-stat="date" csk="20250510"><a href="/en/matches/2025-05-10">2025-05-10</a></td>
	<td class="right" data-stat="start_time" csk="16:00:00">16:00</td>
	<td class="right" data-stat="home_team"><a href="/en/squads/111">Internacional</a></td>
	<td class="right" data-stat="home_xg"></td>
	<td class="center" data-stat="score">vs</td>
	<td class="right" data-stat="away_xg"></td>
	<td class="left" data-stat="away_team"><a href="/en/squads/222">Grêmio</a></td>
	<td class="right iz" data-stat="attendance"></td>
	<td class="left" data-stat="venue">Estádio Beira-Rio</td>
	<td class="left iz" data-stat="referee"></td>
	<td class="left iz" data-stat="match_report"></td>
	<td class="left iz" data-stat="notes"></td>
</tr>
</tbody>
</table></body></html>`

const testNoMatchesHTML = `<html><body>
<table class="stats_table sortable min_width" id="sched_2025_24_1">
<thead>
<tr>
	<th data-stat="gameweek">Wk</th>
	<th data-stat="date">Date</th>
	<th data-stat="home_team">Home</th>
	<th data-stat="score">Score</th>
	<th data-stat="away_team">Away</th>
	<th data-stat="referee">Referee</th>
</tr>
</thead>
<tbody>
<tr>
	<th scope="row" data-stat="gameweek">1</th>
	<td data-stat="date" csk="20250510">2025-05-10</td>
	<td data-stat="home_team"><a href="/squad/a">Flamengo</a></td>
	<td data-stat="score">vs</td>
	<td data-stat="away_team"><a href="/squad/b">Palmeiras</a></td>
	<td class="left iz" data-stat="referee"></td>
</tr>
</tbody>
</table></body></html>`

const testEmptyTableHTML = `<html><body>
<table class="stats_table sortable min_width" id="sched_2025_24_1">
<thead><tr><th data-stat="gameweek">Wk</th></tr></thead>
<tbody>
</tbody>
</table></body></html>`

func TestParseMatchResults(t *testing.T) {
	t.Parallel()

	matches, err := ParseMatchResults(2025, testScheduleHTML)

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
}

func TestParseMatchResults_SecondMatch(t *testing.T) {
	t.Parallel()

	matches, err := ParseMatchResults(2025, testScheduleHTML)

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
}

func TestParseMatchResults_SkipsFixtures(t *testing.T) {
	t.Parallel()

	matches, err := ParseMatchResults(2025, testScheduleHTML)

	is := assert.New(t)

	is.NoError(err)
	is.Len(matches, 2)

	for _, m := range matches {
		is.NotEqual("Internacional", m.HomeTeam, "fixture row should be skipped")
		is.NotEqual("Grêmio", m.AwayTeam, "fixture row should be skipped")
	}
}

func TestParseMatchResults_Empty(t *testing.T) {
	t.Parallel()

	_, err := ParseMatchResults(2025, testEmptyTableHTML)

	assert.ErrorIs(t, err, errNoMatches)
}

func TestParseFixtures(t *testing.T) {
	t.Parallel()

	fixtures, err := ParseFixtures(2025, testScheduleHTML)

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	must.Len(fixtures, 1)

	is.Equal("Internacional", fixtures[0].HomeTeam)
	is.Equal("Grêmio", fixtures[0].AwayTeam)
	is.Equal(3, fixtures[0].Round)
}

func TestParseFixtures_NoFixtures(t *testing.T) {
	t.Parallel()

	const html = `<html><body>
<table class="stats_table sortable min_width" id="sched_2025_24_1">
<thead><tr><th data-stat="gameweek">Wk</th><th data-stat="score">Score</th></tr></thead>
<tbody>
<tr><th scope="row" data-stat="gameweek">1</th>
	<td data-stat="score"><a href="/m/1">2–1</a></td></tr>
</tbody>
</table></body></html>`

	fixtures, err := ParseFixtures(2025, html)

	is := assert.New(t)

	is.NoError(err)
	is.Empty(fixtures)
}

func TestParseFixtures_EmptyTable(t *testing.T) {
	t.Parallel()

	fixtures, err := ParseFixtures(2025, testEmptyTableHTML)

	is := assert.New(t)

	is.NoError(err)
	is.Empty(fixtures)
}

func TestParseMatchResults_InvalidTableID(t *testing.T) {
	t.Parallel()

	const html = `<html><body>
<table id="wrong_table"><tbody><tr><td data-stat="score">2–1</td></tr></tbody></table>
</body></html>`

	_, err := ParseMatchResults(2025, html)

	assert.Error(t, err)
}

func TestParseMatchResults_NotFound(t *testing.T) {
	t.Parallel()

	_, err := ParseMatchResults(2025, testNoMatchesHTML)

	assert.ErrorIs(t, err, errNoMatches)
}

func TestParseMatchResults_ScoreFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		score      string
		wantHome   int
		wantAway   int
		wantOK     bool
	}{
		{name: "en-dash", score: "2–1", wantHome: 2, wantAway: 1, wantOK: true},
		{name: "hyphen", score: "3-0", wantHome: 3, wantAway: 0, wantOK: true},
		{name: "em-dash", score: "1—2", wantHome: 1, wantAway: 2, wantOK: true},
		{name: "invalid", score: "abc", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			home, away, ok := parseScore(tt.score)

			is := assert.New(t)

			is.Equal(tt.wantOK, ok)
			is.Equal(tt.wantHome, home)
			is.Equal(tt.wantAway, away)
		})
	}
}

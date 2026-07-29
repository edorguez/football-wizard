package scraper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSquadHTML = `<html><body>
<div data-team="abc123">
<h1 itemprop="name">Flamengo</h1>
</div>
<table id="stats_standard_9">
<thead><tr>
	<th data-stat="player">Player</th>
	<th data-stat="nationality">Nation</th>
	<th data-stat="position">Pos</th>
	<th data-stat="age">Age</th>
</tr></thead>
<tbody>
<tr>
	<th data-stat="player"><a href="/en/players/xyz/Gabriel">Gabriel</a></th>
	<td data-stat="nationality"><a href="/en/nationality/bra">BRA</a></td>
	<td data-stat="position">FW</td>
	<td data-stat="age">24</td>
</tr>
<tr>
	<th data-stat="player"><a href="/en/players/abc/Arrascaeta">Arrascaeta</a></th>
	<td data-stat="nationality"><a href="/en/nationality/uru">URU</a></td>
	<td data-stat="position">MF</td>
	<td data-stat="age">28</td>
</tr>
</tbody>
</table>
</body></html>`

func TestParseSquad(t *testing.T) {
	t.Parallel()

	squad, err := ParseSquad(2025, testSquadHTML)

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	is.Equal("abc123", squad.TeamFBRefID)
	is.Len(squad.Players, 2)

	player := squad.Players[0]
	is.Equal("Gabriel", player.Name)
	is.Equal("BRA", player.Nationality)
	is.Equal("FW", player.Position)
	is.Equal(2025, player.Season)
}

func TestParseSquad_Empty(t *testing.T) {
	t.Parallel()

	const html = `<html><body><table id="stats_standard_9"><tbody></tbody></table></body></html>`

	_, err := ParseSquad(2025, html)

	assert.Error(t, err)
}

func TestExtractTeamFBRefID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{name: "valid", html: `<div data-team="abc123">`, expected: "abc123"},
		{name: "not found", html: `<div>no data</div>`, expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)

			is.Equal(tt.expected, extractTeamFBRefID(tt.html))
		})
	}
}

func TestTeamIDFromURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{name: "standard", url: "https://fbref.com/en/squads/abc123/2025/Flamengo-Stats", expected: "abc123"},
		{name: "no match", url: "https://fbref.com/en/teams/123", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)

			is.Equal(tt.expected, teamIDFromURL(tt.url))
		})
	}
}

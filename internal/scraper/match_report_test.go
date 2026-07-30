package scraper

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMatchReport_Basic(t *testing.T) {
	t.Parallel()

	report, err := ParseMatchReport("<html><body><div class=\"scorebox\"></div></body></html>")

	is := assert.New(t)

	is.NoError(err)
	is.Empty(report.HomeTeam)
}

func TestParseMatchReport_WithTeams(t *testing.T) {
	t.Parallel()

	const html = `<html><body>
<div class="scorebox">
<div class="scorebox_team" id="sb_team_0">
<strong><a href="/en/squads/abc">Flamengo</a></strong>
</div>
<div class="scorebox_team" id="sb_team_1">
<strong><a href="/en/squads/xyz">Palmeiras</a></strong>
</div>
</div>
</body></html>`

	report, err := ParseMatchReport(html)

	is := assert.New(t)

	is.NoError(err)
	is.Equal("Flamengo", report.HomeTeam)
	is.Equal("Palmeiras", report.AwayTeam)
}

const testLineupHTML = `<html><body>
<div class="lineup" id="a">
<table>
<tbody>
<tr><th colspan="2">Time A (4-3-3)</th></tr>
<tr><td>1</td><td><a href="/p/1">Player One</a></td></tr>
<tr><td>10</td><td><a href="/p/10">Player Ten</a></td></tr>
<tr><th colspan="2">Bench</th></tr>
<tr><td>12</td><td><a href="/p/12">Sub Twelve</a></td></tr>
<tr><td>15</td><td><a href="/p/15">Sub Fifteen</a></td></tr>
</tbody>
</table>
</div>
<div class="lineup" id="b">
<table>
<tbody>
<tr><th colspan="2">Time B (4-4-2)</th></tr>
<tr><td>2</td><td><a href="/p/2">Player Two</a></td></tr>
<tr><td>7</td><td><a href="/p/7">Player Seven</a></td></tr>
<tr><th colspan="2">Bench</th></tr>
<tr><td>14</td><td><a href="/p/14">Sub Fourteen</a></td></tr>
</tbody>
</table>
</div>
</body></html>`

func TestParseLineups(t *testing.T) {
	t.Parallel()

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(testLineupHTML))
	require.NoError(t, err)

	var report ScrapedMatchReport
	parseLineups(doc, &report)

	is := assert.New(t)

	is.Len(report.HomeLineup, 4)
	is.Len(report.AwayLineup, 3)

	is.Equal("Player One", report.HomeLineup[0].Name)
	is.True(report.HomeLineup[0].IsStarter)
	is.Equal(1, *report.HomeLineup[0].ShirtNum)

	is.Equal("Player Ten", report.HomeLineup[1].Name)
	is.True(report.HomeLineup[1].IsStarter)

	is.Equal("Sub Twelve", report.HomeLineup[2].Name)
	is.False(report.HomeLineup[2].IsStarter)

	is.Equal("Player Two", report.AwayLineup[0].Name)
	is.True(report.AwayLineup[0].IsStarter)

	is.Equal("Sub Fourteen", report.AwayLineup[2].Name)
	is.False(report.AwayLineup[2].IsStarter)
}

func TestParseShotsOnTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		text          string
		wantTotal     *int
		wantOnTarget  *int
	}{
		{name: "normal", text: "5 of 12 — 42%", wantTotal: ip(12), wantOnTarget: ip(5)},
		{name: "reversed order", text: "40% — 7 of 22", wantTotal: ip(22), wantOnTarget: ip(7)},
		{name: "no match", text: "Possession", wantTotal: nil, wantOnTarget: nil},
		{name: "zero", text: "0 of 0 — 0%", wantTotal: ip(0), wantOnTarget: ip(0)},
	}

		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			total, onTarget := parseShotsOnTarget(tt.text)

			is := assert.New(t)

			if tt.wantTotal == nil {
				is.Nil(total)
			} else if total != nil {
				is.Equal(*tt.wantTotal, *total)
			}
			if tt.wantOnTarget == nil {
				is.Nil(onTarget)
			} else if onTarget != nil {
				is.Equal(*tt.wantOnTarget, *onTarget)
			}
		})
	}
}

func TestParseSaves(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want *int
	}{
		{name: "normal", text: "6 of 7 — 85%", want: ip(6)},
		{name: "reversed", text: "75% — 3 of 4", want: ip(3)},
		{name: "no match", text: "Possession", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)

			if tt.want == nil {
				is.Nil(parseSaves(tt.text))
			} else {
				is.Equal(*tt.want, *parseSaves(tt.text))
			}
		})
	}
}

func ip(n int) *int {
	return &n
}

func TestExtractPct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want *int
	}{
		{name: "with percent", s: "32%", want: ip(32)},
		{name: "without percent", s: "68", want: ip(68)},
		{name: "empty", s: "", want: nil},
		{name: "text", s: "abc", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)

			if tt.want == nil {
				is.Nil(extractPct(tt.s))
			} else {
				is.Equal(*tt.want, *extractPct(tt.s))
			}
		})
	}
}



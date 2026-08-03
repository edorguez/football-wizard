package scraper

import (
	"os"
	"path/filepath"
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
		name         string
		text         string
		wantTotal    *int
		wantOnTarget *int
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

func TestParseScoringSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		homeHTML      string
		awayHTML      string
		wantHomeMin   *int
		wantAwayMin   *int
		wantHomeFH    *int
		wantAwayFH    *int
		wantHomeSH    *int
		wantAwaySH    *int
		wantHomeSHMin *int
		wantAwaySHMin *int
	}{
		{
			name:          "goals in both halves",
			homeHTML:      `<div class="event" id="a"><div><a href="/p/1">Hulk</a> · 34’ <div class="event_icon goal"></div></div><div><a href="/p/2">Igor</a> · 75’ <div class="event_icon goal"></div></div></div>`,
			awayHTML:      `<div class="event" id="b"><div><a href="/p/3">Biel</a> · 12’ <div class="event_icon goal"></div></div></div>`,
			wantHomeMin:   ip(34),
			wantAwayMin:   ip(12),
			wantHomeFH:    ip(1),
			wantAwayFH:    ip(1),
			wantHomeSH:    ip(1),
			wantAwaySH:    ip(0),
			wantHomeSHMin: ip(75),
			wantAwaySHMin: nil,
		},
		{
			name:          "stoppage time goal counts as first half",
			homeHTML:      `<div class="event" id="a"><div><a href="/p/1">Hulk</a> · 45+2’ <div class="event_icon goal"></div></div></div>`,
			awayHTML:      ``,
			wantHomeMin:   ip(47),
			wantAwayMin:   nil,
			wantHomeFH:    ip(1),
			wantAwayFH:    nil,
			wantHomeSH:    ip(0),
			wantAwaySH:    nil,
			wantHomeSHMin: nil,
			wantAwaySHMin: nil,
		},
		{
			name:          "second half only",
			homeHTML:      `<div class="event" id="a"><div><a href="/p/1">Hulk</a> · 66’ <div class="event_icon goal"></div></div></div>`,
			awayHTML:      ``,
			wantHomeMin:   ip(66),
			wantAwayMin:   nil,
			wantHomeFH:    ip(0),
			wantAwayFH:    nil,
			wantHomeSH:    ip(1),
			wantAwaySH:    nil,
			wantHomeSHMin: ip(66),
			wantAwaySHMin: nil,
		},
		{
			name:          "no scoring summary",
			homeHTML:      ``,
			awayHTML:      ``,
			wantHomeMin:   nil,
			wantAwayMin:   nil,
			wantHomeFH:    nil,
			wantAwayFH:    nil,
			wantHomeSH:    nil,
			wantAwaySH:    nil,
			wantHomeSHMin: nil,
			wantAwaySHMin: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			html := "<html><body>" + tt.homeHTML + tt.awayHTML + "</body></html>"
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			require.NoError(t, err)

			var report ScrapedMatchReport
			parseScoringSummary(doc, &report)

			assertIntPtr(t, tt.wantHomeMin, report.HomeFirstGoalMinute)
			assertIntPtr(t, tt.wantAwayMin, report.AwayFirstGoalMinute)
			assertIntPtr(t, tt.wantHomeFH, report.HomeGoalsFirstHalf)
			assertIntPtr(t, tt.wantAwayFH, report.AwayGoalsFirstHalf)
			assertIntPtr(t, tt.wantHomeSH, report.HomeGoalsSecondHalf)
			assertIntPtr(t, tt.wantAwaySH, report.AwayGoalsSecondHalf)
			assertIntPtr(t, tt.wantHomeSHMin, report.HomeSecondGoalMinute)
			assertIntPtr(t, tt.wantAwaySHMin, report.AwaySecondGoalMinute)
		})
	}
}

func assertIntPtr(t *testing.T, want, got *int) {
	if want == nil {
		assert.Nil(t, got)
		return
	}
	require.NotNil(t, got)
	assert.Equal(t, *want, *got)
}

func TestParseScoringSummary_RealCachedHTML(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "data", "cache", "2025", "reports", "Atlético Mineiro-vs-Bahia.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("cached HTML not available: %v", err)
	}

	report, err := ParseMatchReport(string(raw))
	require.NoError(t, err)

	is := assert.New(t)
	is.Equal("Atlético Mineiro", report.HomeTeam)
	is.Equal("Bahia", report.AwayTeam)

	require.NotNil(t, report.HomeFirstGoalMinute)
	is.Equal(66, *report.HomeFirstGoalMinute)
	is.Nil(report.AwayFirstGoalMinute)

	require.NotNil(t, report.HomeGoalsFirstHalf)
	is.Equal(0, *report.HomeGoalsFirstHalf)
	require.NotNil(t, report.AwayGoalsFirstHalf)
	is.Equal(0, *report.AwayGoalsFirstHalf)

	// Atlético scored all 3 goals in the second half (66', 75', 77').
	require.NotNil(t, report.HomeGoalsSecondHalf)
	is.Equal(3, *report.HomeGoalsSecondHalf)
	require.NotNil(t, report.AwayGoalsSecondHalf)
	is.Equal(0, *report.AwayGoalsSecondHalf)

	require.NotNil(t, report.HomeSecondGoalMinute)
	is.Equal(66, *report.HomeSecondGoalMinute)
	is.Nil(report.AwaySecondGoalMinute)
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

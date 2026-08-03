package model

import (
	"fmt"
	"strings"
)

// FormatPrediction renders a match forecast as a fixed-width report used by
// both the CLI and the TUI.
func FormatPrediction(p *MatchPrediction) string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s vs %s\n", p.HomeTeam, p.AwayTeam)
	fmt.Fprintf(&b, "expected goals: %.2f - %.2f (total %.2f)\n\n",
		p.ExpectedHomeGoals, p.ExpectedAwayGoals, p.ExpectedTotalGoals)

	fmt.Fprintf(&b, "%-10s %10s  %s\n", "1X2", "PROB", "CONFIDENCE")
	fmt.Fprintf(&b, "%-10s %9.1f%%  %s\n", "Home win", p.HomeWin*100, FromProbability(p.HomeWin))
	fmt.Fprintf(&b, "%-10s %9.1f%%  %s\n", "Draw", p.Draw*100, FromProbability(p.Draw))
	fmt.Fprintf(&b, "%-10s %9.1f%%  %s\n", "Away win", p.AwayWin*100, FromProbability(p.AwayWin))

	b.WriteString("\n")
	fmt.Fprintf(&b, "%-24s %-30s %-6s %9s  %s\n", "MARKET", "OUTCOME", "LINE", "PROB", "CONFIDENCE")
	b.WriteString(strings.Repeat("-", 90))
	b.WriteString("\n")
	for _, m := range p.Markets {
		line := "-"
		if m.Line > 0 {
			line = fmt.Sprintf("%.1f", m.Line)
		}
		fmt.Fprintf(&b, "%-24s %-30s %-6s %8.1f%%  %s\n",
			m.Market, m.Outcome, line, m.Probability*100, m.Confidence)
	}

	return b.String()
}

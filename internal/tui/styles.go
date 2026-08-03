package tui

import "github.com/charmbracelet/lipgloss"

var (
	primary   = lipgloss.Color("#7C3AED")
	secondary = lipgloss.Color("#38BDF8")
	success   = lipgloss.Color("#22C55E")
	warning   = lipgloss.Color("#F59E0B")
	danger    = lipgloss.Color("#EF4444")
	muted     = lipgloss.Color("#6B7280")
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			Padding(0, 1).
			Render

	ListTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primary)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(secondary).
			Padding(0, 1).
			Render

	MenuStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Render

	MenuItemStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Render

	MenuItemSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primary).
				Background(lipgloss.Color("#EDE9FE")).
				Padding(0, 1).
				Render

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondary).
			Padding(1).
			Render

	SuccessStyle = lipgloss.NewStyle().
			Foreground(success).
			Bold(true).
			Render

	WarningStyle = lipgloss.NewStyle().
			Foreground(warning).
			Render

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(warning)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(danger).
			Bold(true).
			Render

	MutedStyle = lipgloss.NewStyle().
			Foreground(muted).
			Render

	HelpStyle = lipgloss.NewStyle().
			Foreground(muted).
			Padding(0, 1).
			Render
)

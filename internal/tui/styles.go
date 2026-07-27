package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle = lipgloss.NewStyle().
		Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFF")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)

	itemStyle = lipgloss.NewStyle().
		PaddingLeft(2)

	dimmedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF4D4D"))

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500"))

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00BFFF"))

	tableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Padding(0, 1)

	tableCellStyle = lipgloss.NewStyle().
		Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Italic(true)

	focusedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)

	blurredStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A49DB5"))

	logPrefixStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4"))

	logTimeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	logTextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E0E0E0"))

	logBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5A5A5A")).
		Padding(0, 1).
		MaxWidth(100)
)

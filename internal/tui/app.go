package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run launches the full-screen TUI and blocks until the user quits.
func Run(deps Deps) error {
	program := tea.NewProgram(NewRoot(deps), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}
	return nil
}

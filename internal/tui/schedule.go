package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type ScheduleModel struct {
	app     *AppContext
	running bool
	message string
}

func NewScheduleModel(app *AppContext) *ScheduleModel {
	running := app.Scheduler.IsRunning()
	return &ScheduleModel{
		app:     app,
		running: running,
	}
}

func (m *ScheduleModel) Init() tea.Cmd {
	return nil
}

func (m *ScheduleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return switchViewMsg(dashboardView) }
		case "enter":
			if m.running {
				go m.app.Scheduler.Stop()
				m.running = false
				m.message = successStyle.Render("scheduler stopped")
			} else {
				err := m.app.Scheduler.Start()
				if err != nil {
					m.message = errorStyle.Render(fmt.Sprintf("error: %v", err))
				} else {
					m.running = true
					m.message = successStyle.Render("scheduler started")
				}
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *ScheduleModel) View() string {
	s := titleStyle.Render("⏰ Scheduler")
	s += "\n\n"

	s += headerStyle.Render("status:")
	s += "\n\n"

	if m.running {
		s += successStyle.Render("  ● Active")
		s += "\n\n"
		s += "  • Daily scrape: 01:00\n"
		s += "  • Weekly retrain: Sunday 03:00\n"
	} else {
		s += dimmedStyle.Render("  ○ Inactive")
	}

	s += "\n\n"

	action := "Start"
	if m.running {
		action = "Stop"
	}
	s += selectedStyle.Render(fmt.Sprintf("  Press Enter to %s scheduler", action))
	s += "\n\n"

	if m.message != "" {
		s += m.message + "\n\n"
	}

	s += helpStyle.Render("Enter to start/stop • Esc to go back")
	return s
}

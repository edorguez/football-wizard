package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type ListModel struct {
	app         *AppContext
	spinner     spinner.Model
	loading     bool
	predictions []string
	message     string
}

func NewListModel(app *AppContext) *ListModel {
	s := spinner.New()
	s.Style = focusedStyle
	s.Spinner = spinner.Dot

	return &ListModel{
		app:     app,
		spinner: s,
	}
}

func (m *ListModel) Init() tea.Cmd {
	return m.loadCmd()
}

func (m *ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return switchViewMsg(dashboardView) }
		}
	case listLoadedMsg:
		m.loading = false
		m.predictions = []string(msg)
		return m, nil
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *ListModel) View() string {
	s := titleStyle.Render("📋 Recent Predictions")
	s += "\n\n"

	if m.loading {
		s += m.spinner.View() + " loading predictions\n"
	}

	if len(m.predictions) == 0 && !m.loading {
		s += dimmedStyle.Render("no predictions yet")
		s += "\n\n" + helpStyle.Render("Esc to go back to menu")
		return s
	}

	for _, p := range m.predictions {
		s += p + "\n"
	}

	s += "\n" + helpStyle.Render("Esc to go back")
	return s
}

func (m *ListModel) loadCmd() tea.Cmd {
	return func() tea.Msg {
		predictions, err := m.app.Predicts.FindLast(10)
		if err != nil {
			return listLoadedMsg([]string{errorStyle.Render("error loading predictions")})
		}

		var lines []string
		for _, p := range predictions {
			line := fmt.Sprintf("%s — Home: %.0f%% Draw: %.0f%% Away: %.0f%% [%s]",
				p.CreatedAt.Format("02/01"),
				p.HomeWinProb*100,
				p.DrawProb*100,
				p.AwayWinProb*100,
				p.ConfidenceLevel,
			)
			lines = append(lines, line)
		}
		return listLoadedMsg(lines)
	}
}

type listLoadedMsg []string

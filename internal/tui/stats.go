package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type StatsModel struct {
	app     *AppContext
	input   textinput.Model
	spinner spinner.Model
	loading bool
	result  string
	message string
}

func NewStatsModel(app *AppContext) *StatsModel {
	ti := textinput.New()
	ti.Placeholder = "Flamengo"
	ti.CharLimit = 50
	ti.Width = 30
	ti.Focus()

	s := spinner.New()
	s.Style = focusedStyle
	s.Spinner = spinner.Dot

	return &StatsModel{
		app:     app,
		input:   ti,
		spinner: s,
	}
}

func (m *StatsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return switchViewMsg(dashboardView) }
		case "enter":
			if !m.loading && m.input.Value() != "" {
				m.loading = true
				m.result = ""
				m.message = ""
				return m, tea.Batch(m.spinner.Tick, m.loadStatsCmd(m.input.Value()))
			}
		}
	case statsLoadedMsg:
		m.loading = false
		m.result = string(msg)
		return m, nil
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *StatsModel) View() string {
	s := titleStyle.Render("📈 Team Stats")
	s += "\n\n"

	if !m.loading && m.result == "" && m.message == "" {
		s += headerStyle.Render("team name:")
		s += "\n\n"
		s += m.input.View()
		s += "\n\n"
		s += helpStyle.Render("Enter to search • Esc to go back")
	}

	if m.loading {
		s += m.spinner.View() + " loading stats\n"
	}

	if m.result != "" {
		s += "\n" + borderStyle.Render(m.result)
		s += "\n\n" + helpStyle.Render("Esc to go back")
	}

	if m.message != "" {
		s += "\n\n" + m.message
		s += "\n\n" + helpStyle.Render("Esc to go back")
	}

	return s
}

func (m *StatsModel) loadStatsCmd(teamName string) tea.Cmd {
	return func() tea.Msg {
		team, err := m.app.TeamRepo.FindByName(teamName)
		if err != nil {
			return statsLoadedMsg(fmt.Sprintf("team %q not found", teamName))
		}

		matches, err := m.app.MatchRepo.FindByTeam(team.ID, 5)
		if err != nil {
			return statsLoadedMsg(fmt.Sprintf("error: %v", err))
		}

		content := fmt.Sprintf("%s\n\n", selectedStyle.Render(team.Name))
		content += fmt.Sprintf("Stadium: %s\n", team.Stadium)
		content += fmt.Sprintf("Founded: %d\n", team.Founded)

		if len(matches) > 0 {
			content += "\nlast 5 matches:\n"
			for _, m := range matches {
				var result string
				if m.HomeTeamID == team.ID {
					result = fmt.Sprintf("  %d - %d vs %s", m.HomeGoals, m.AwayGoals, "(away)")
				} else {
					result = fmt.Sprintf("  %s %d - %d", "(away)", m.AwayGoals, m.HomeGoals)
				}
				content += result + "\n"
			}
		}

		return statsLoadedMsg(content)
	}
}

type statsLoadedMsg string

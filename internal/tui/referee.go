package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type RefereeModel struct {
	app     *AppContext
	input   textinput.Model
	spinner spinner.Model
	loading bool
	result  string
	message string
}

func NewRefereeModel(app *AppContext) *RefereeModel {
	ti := textinput.New()
	ti.Placeholder = "Wilton Sampaio"
	ti.CharLimit = 50
	ti.Width = 30
	ti.Focus()

	s := spinner.New()
	s.Style = focusedStyle
	s.Spinner = spinner.Dot

	return &RefereeModel{
		app:     app,
		input:   ti,
		spinner: s,
	}
}

func (m *RefereeModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *RefereeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return m, tea.Batch(m.spinner.Tick, m.loadRefereeCmd(m.input.Value()))
			}
		}
	case refereeLoadedMsg:
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

func (m *RefereeModel) View() string {
	s := titleStyle.Render("🃏 Referee Profile")
	s += "\n\n"

	if !m.loading && m.result == "" && m.message == "" {
		s += headerStyle.Render("referee name:")
		s += "\n\n"
		s += m.input.View()
		s += "\n\n"
		s += helpStyle.Render("Enter to search • Esc to go back")
	}

	if m.loading {
		s += m.spinner.View() + " searching referee\n"
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

func (m *RefereeModel) loadRefereeCmd(name string) tea.Cmd {
	return func() tea.Msg {
		return refereeLoadedMsg(fmt.Sprintf("referee %q", name))
	}
}

type refereeLoadedMsg string

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/edorguez/football-wizard/internal/database"
)

const historyHeaderHeight = 5

type HistoryModel struct {
	deps     Deps
	viewport viewport.Model
	items    []database.Prediction
	err      error
	ready    bool
}

func NewHistoryModel(deps Deps) tea.Model {
	return &HistoryModel{deps: deps}
}

func (m *HistoryModel) Init() tea.Cmd {
	return m.load()
}

type historyLoadedMsg struct {
	items []database.Prediction
	err   error
}

func (m *HistoryModel) load() tea.Cmd {
	return func() tea.Msg {
		items, err := m.deps.Predictions.ListRecent(20)
		return historyLoadedMsg{items: items, err: err}
	}
}

func (m *HistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-historyHeaderHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - historyHeaderHeight
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, backCmd()
		case "r":
			return m, m.load()
		}

	case historyLoadedMsg:
		m.items = msg.items
		m.err = msg.err
		m.render()
		return m, nil
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *HistoryModel) render() {
	var b strings.Builder
	switch {
	case m.err != nil:
		b.WriteString(ErrorStyle(m.err.Error()))
	case len(m.items) == 0:
		b.WriteString(MutedStyle("no predictions saved yet"))
	default:
		for _, p := range m.items {
			fmt.Fprintf(&b, "%s  %s vs %s\n",
				p.CreatedAt.Format("2006-01-02 15:04"),
				p.HomeTeam.Name,
				p.AwayTeam.Name,
			)
			fmt.Fprintf(&b, "    1X2: %5.1f%% / %5.1f%% / %5.1f%%   xG %.2f - %.2f\n",
				p.HomeWin*100, p.Draw*100, p.AwayWin*100,
				p.ExpectedHomeGoals, p.ExpectedAwayGoals,
			)
			b.WriteString("\n")
		}
	}
	m.viewport.SetContent(b.String())
	m.viewport.GotoTop()
}

func (m *HistoryModel) View() string {
	var b strings.Builder
	b.WriteString(TitleStyle("Recent Predictions"))
	b.WriteString("\n\n")
	if m.ready {
		b.WriteString(m.viewport.View())
	}
	b.WriteString("\n")
	b.WriteString(HelpStyle("r: refresh · pgup/pgdn: scroll · esc: menu"))
	return b.String()
}

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/edorguez/football-wizard/internal/database"
)

const refereeHeaderHeight = 6

type RefereeModel struct {
	deps     Deps
	input    textinput.Model
	results  []database.Referee
	cursor   int
	ref      *database.Referee
	matches  []database.Match
	searched bool
	detail   bool
	viewport viewport.Model
	err      error
	ready    bool
}

func NewRefereeModel(deps Deps) tea.Model {
	input := textinput.New()
	input.Placeholder = "referee name"
	input.Focus()
	return &RefereeModel{deps: deps, input: input}
}

func (m *RefereeModel) Init() tea.Cmd {
	return textinput.Blink
}

type refereeSearchMsg struct {
	results []database.Referee
	err     error
}

func (m *RefereeModel) search() tea.Cmd {
	query := strings.TrimSpace(m.input.Value())
	return func() tea.Msg {
		results, err := m.deps.Refs.Search(query)
		return refereeSearchMsg{results: results, err: err}
	}
}

type refereeMatchesMsg struct {
	matches []database.Match
	err     error
}

func (m *RefereeModel) loadMatches() tea.Cmd {
	return func() tea.Msg {
		matches, err := m.deps.Matches.ListByReferee(m.ref.ID)
		return refereeMatchesMsg{matches: matches, err: err}
	}
}

func (m *RefereeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ready = true
		if m.detail {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - refereeHeaderHeight
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.detail {
				m.detail = false
				m.err = nil
				return m, nil
			}
			return m, backCmd()
		case "enter":
			if m.detail {
				break
			}
			if m.searched {
				if len(m.results) > 0 {
					m.ref = &m.results[m.cursor]
					m.detail = true
					return m, m.loadMatches()
				}
				return m, nil
			}
			return m, m.search()
		case "up":
			if m.searched && !m.detail && len(m.results) > 0 {
				m.cursor = (m.cursor + len(m.results) - 1) % len(m.results)
				return m, nil
			}
		case "down":
			if m.searched && !m.detail && len(m.results) > 0 {
				m.cursor = (m.cursor + 1) % len(m.results)
				return m, nil
			}
		}

	case refereeSearchMsg:
		m.results = msg.results
		m.err = msg.err
		m.searched = true
		m.cursor = 0
		return m, nil

	case refereeMatchesMsg:
		m.matches = msg.matches
		m.err = msg.err
		m.render()
		return m, nil
	}

	if m.detail {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *RefereeModel) render() {
	var b strings.Builder
	if m.err != nil {
		b.WriteString(ErrorStyle(m.err.Error()))
	} else if m.ref != nil {
		fmt.Fprintf(&b, "%s — %d matches\n\n", m.ref.Name, len(m.matches))
		for _, match := range m.matches {
			home, away := match.HomeTeam.Name, match.AwayTeam.Name
			result := "?"
			if match.HomeGoals != nil && match.AwayGoals != nil {
				result = fmt.Sprintf("%d - %d", *match.HomeGoals, *match.AwayGoals)
			}
			fmt.Fprintf(&b, "  %s vs %s   %s   %s\n", home, away, result, match.Date.Format("2006-01-02"))
		}
	}
	m.viewport.SetContent(b.String())
	m.viewport.GotoTop()
}

func (m *RefereeModel) View() string {
	var b strings.Builder

	if m.detail {
		b.WriteString(TitleStyle(fmt.Sprintf("Referee Profile — %s", refereeName(m.ref))))
		b.WriteString("\n\n")
		if m.ready {
			b.WriteString(m.viewport.View())
		}
		b.WriteString("\n")
		b.WriteString(HelpStyle("esc: back to results"))
		return b.String()
	}

	b.WriteString(TitleStyle("Referee Profile"))
	b.WriteString("\n")
	if m.searched {
		if m.err != nil {
			b.WriteString(ErrorStyle(m.err.Error()))
			b.WriteString("\n")
		} else if len(m.results) == 0 {
			b.WriteString(WarningStyle("no referees matched"))
			b.WriteString("\n\n")
		} else {
			b.WriteString("\n")
			for i, ref := range m.results {
				line := fmt.Sprintf("  %s", ref.Name)
				if i == m.cursor {
					line = MenuItemSelectedStyle("→ " + ref.Name)
				}
				b.WriteString(line)
				b.WriteString("\n")
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("Search: ")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")
	b.WriteString(HelpStyle("enter: search/select · ↑/↓: navigate results · esc: menu"))
	return b.String()
}

func refereeName(ref *database.Referee) string {
	if ref == nil {
		return ""
	}
	return ref.Name
}

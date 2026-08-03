package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/edorguez/football-wizard/internal/database"
)

const teamHeaderHeight = 6

type TeamModel struct {
	deps     Deps
	list     list.Model
	team     *database.Team
	matches  []database.Match
	detail   bool
	viewport viewport.Model
	err      error
	ready    bool
}

func NewTeamModel(deps Deps) tea.Model {
	return &TeamModel{deps: deps}
}

func (m *TeamModel) Init() tea.Cmd {
	return m.loadTeams()
}

type teamLoadedMsg struct {
	teams []database.Team
	err   error
}

func (m *TeamModel) loadTeams() tea.Cmd {
	return func() tea.Msg {
		teams, err := m.deps.Teams.ListAll()
		return teamLoadedMsg{teams: teams, err: err}
	}
}

type teamMatchesMsg struct {
	team    *database.Team
	matches []database.Match
	err     error
}

func (m *TeamModel) loadMatches(team *database.Team) tea.Cmd {
	return func() tea.Msg {
		matches, err := m.deps.Matches.ListByTeam(team.ID)
		return teamMatchesMsg{team: team, matches: matches, err: err}
	}
}

func (m *TeamModel) makeList(teams []database.Team) list.Model {
	items := make([]list.Item, 0, len(teams))
	for _, team := range teams {
		items = append(items, teamItem(team))
	}
	w, h := 60, 20
	if m.ready {
		w, h = m.viewport.Width, m.viewport.Height
	}
	l := list.New(items, list.NewDefaultDelegate(), w, h)
	l.Title = "Select a team"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = ListTitleStyle
	return l
}

func (m *TeamModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ready = true
		if m.detail {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - teamHeaderHeight
		} else {
			if w, h := msg.Width-8, msg.Height-8; w > 0 && h > 0 {
				m.list.SetWidth(w)
				m.list.SetHeight(h)
			}
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
			if !m.detail {
				if team := selectedTeamFromList(m.list); team != nil {
					m.team = team
					m.detail = true
					return m, m.loadMatches(team)
				}
			}
		}

	case teamLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.list = m.makeList(msg.teams)
		return m, nil

	case teamMatchesMsg:
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
	if m.list.Items() == nil {
		return m, nil
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *TeamModel) render() {
	var b strings.Builder
	if m.err != nil {
		b.WriteString(ErrorStyle(m.err.Error()))
	} else if m.team != nil {
		fmt.Fprintf(&b, "%s — last %d matches\n\n", m.team.Name, len(m.matches))
		for _, match := range m.matches {
			home, away := match.HomeTeam.Name, match.AwayTeam.Name
			result := "?"
			if match.HomeGoals != nil && match.AwayGoals != nil {
				result = fmt.Sprintf("%d - %d", *match.HomeGoals, *match.AwayGoals)
			}
			fmt.Fprintf(&b, "  %s vs %s   %s   %s\n", home, away, result, match.Date.Format("2006-01-02"))
		}
	} else {
		b.WriteString(MutedStyle("select a team to see its recent form"))
	}
	m.viewport.SetContent(b.String())
	m.viewport.GotoTop()
}

func (m *TeamModel) View() string {
	var b strings.Builder
	if m.detail {
		b.WriteString(TitleStyle(fmt.Sprintf("Team Stats — %s", teamName(m.team))))
		b.WriteString("\n\n")
		if m.ready {
			b.WriteString(m.viewport.View())
		}
		b.WriteString("\n")
		b.WriteString(HelpStyle("esc: back to team list"))
		return b.String()
	}

	b.WriteString(TitleStyle("Team Stats"))
	b.WriteString("\n\n")
	if m.err != nil {
		b.WriteString(ErrorStyle(m.err.Error()))
		b.WriteString("\n")
	}
	if m.list.Items() != nil {
		b.WriteString(m.list.View())
	}
	b.WriteString("\n")
	b.WriteString(HelpStyle("type to filter · enter: select · esc: menu"))
	return b.String()
}

func teamName(team *database.Team) string {
	if team == nil {
		return ""
	}
	return team.Name
}

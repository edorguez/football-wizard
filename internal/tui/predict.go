package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/model"
)

type predictPhase int

const (
	predictSelectHome predictPhase = iota
	predictSelectAway
	predictLoading
	predictDone
)

type teamItem database.Team

func (t teamItem) Title() string       { return t.Name }
func (t teamItem) Description() string { return t.ShortName }
func (t teamItem) FilterValue() string { return t.Name }

type PredictModel struct {
	deps     Deps
	phase    predictPhase
	homeList list.Model
	awayList list.Model
	homeTeam *database.Team
	awayTeam *database.Team
	pred     *model.MatchPrediction
	viewport viewport.Model
	err      error
	ready    bool
}

func NewPredictModel(deps Deps) tea.Model {
	return &PredictModel{deps: deps}
}

func (m *PredictModel) Init() tea.Cmd {
	return m.loadTeams()
}

type teamsLoadedMsg struct {
	teams []database.Team
	err   error
}

func (m *PredictModel) loadTeams() tea.Cmd {
	return func() tea.Msg {
		teams, err := m.deps.Teams.ListAll()
		return teamsLoadedMsg{teams: teams, err: err}
	}
}

func (m *PredictModel) makeList(teams []database.Team) list.Model {
	items := make([]list.Item, 0, len(teams))
	for _, team := range teams {
		items = append(items, teamItem(team))
	}

	width, height := 60, 20
	if m.ready {
		width, height = m.viewport.Width, m.viewport.Height
	}

	l := list.New(items, list.NewDefaultDelegate(), width, height)
	l.Title = "Select a team"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = ListTitleStyle
	return l
}

func selectedTeamFromList(l list.Model) *database.Team {
	item, ok := l.SelectedItem().(teamItem)
	if !ok {
		return nil
	}
	team := database.Team(item)
	return &team
}

func (m *PredictModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ready = true
		if m.phase == predictDone {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 6
		} else {
			w, h := msg.Width-8, msg.Height-8
			if w > 0 {
				m.homeList.SetWidth(w)
				m.awayList.SetWidth(w)
			}
			if h > 0 {
				m.homeList.SetHeight(h)
				m.awayList.SetHeight(h)
			}
		}
		return m, nil

	case teamsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.homeList = m.makeList(msg.teams)
		m.awayList = m.makeList(msg.teams)
		m.phase = predictSelectHome
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, backCmd()
		case "ctrl+c":
			return m, tea.Quit
		}

		if m.phase == predictDone && msg.String() == "n" {
			return m, m.newPrediction()
		}

		if msg.Type == tea.KeyEnter {
			switch m.phase {
			case predictSelectHome:
				if team := selectedTeamFromList(m.homeList); team != nil {
					m.homeTeam = team
					m.phase = predictSelectAway
					m.awayList.Title = "Select away team"
					return m, nil
				}
			case predictSelectAway:
				if team := selectedTeamFromList(m.awayList); team != nil {
					m.awayTeam = team
					m.phase = predictLoading
					m.err = nil
					return m, m.predictCmd()
				}
			}
		}

		switch m.phase {
		case predictSelectHome:
			var cmd tea.Cmd
			m.homeList, cmd = m.homeList.Update(msg)
			return m, cmd
		case predictSelectAway:
			var cmd tea.Cmd
			m.awayList, cmd = m.awayList.Update(msg)
			return m, cmd
		case predictDone:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}

	case predictionReadyMsg:
		m.phase = predictDone
		m.err = msg.err
		m.pred = msg.pred
		if !m.ready {
			m.viewport = viewport.New(80, 30)
			m.ready = true
		}
		m.render()
		return m, nil
	}

	return m, nil
}

type predictionReadyMsg struct {
	pred *model.MatchPrediction
	err  error
}

func (m *PredictModel) predictCmd() tea.Cmd {
	return func() tea.Msg {
		predictor, err := m.deps.Train()
		if err != nil {
			return predictionReadyMsg{err: fmt.Errorf("training models: %w", err)}
		}
		pred := predictor.PredictMatch(m.homeTeam.ID, m.awayTeam.ID, m.homeTeam.Name, m.awayTeam.Name)
		if err := m.savePrediction(pred); err != nil {
			return predictionReadyMsg{err: err}
		}
		return predictionReadyMsg{pred: pred}
	}
}

func (m *PredictModel) savePrediction(pred *model.MatchPrediction) error {
	payload, err := json.Marshal(pred.Markets)
	if err != nil {
		return fmt.Errorf("marshaling prediction markets: %w", err)
	}

	record := &database.Prediction{
		HomeTeamID:        m.homeTeam.ID,
		AwayTeamID:        m.awayTeam.ID,
		HomeWin:           pred.HomeWin,
		Draw:              pred.Draw,
		AwayWin:           pred.AwayWin,
		ExpectedHomeGoals: pred.ExpectedHomeGoals,
		ExpectedAwayGoals: pred.ExpectedAwayGoals,
		Payload:           string(payload),
	}
	if err := m.deps.Predictions.Create(record); err != nil {
		return fmt.Errorf("saving prediction: %w", err)
	}
	return nil
}

func (m *PredictModel) newPrediction() tea.Cmd {
	m.phase = predictSelectHome
	m.homeTeam = nil
	m.awayTeam = nil
	m.pred = nil
	m.err = nil
	return nil
}

func (m *PredictModel) render() {
	if m.err != nil {
		m.viewport.SetContent(ErrorStyle(m.err.Error()))
		return
	}
	m.viewport.SetContent(model.FormatPrediction(m.pred))
	m.viewport.GotoTop()
}

func (m *PredictModel) View() string {
	var b strings.Builder

	switch m.phase {
	case predictSelectHome, predictSelectAway:
		b.WriteString(TitleStyle("Predict a Match"))
		b.WriteString("\n")
		if m.err != nil {
			b.WriteString(ErrorStyle(m.err.Error()))
			b.WriteString("\n")
		}
		if m.homeTeam != nil {
			fmt.Fprintf(&b, "Home: %s\n", m.homeTeam.Name)
		}
		if m.phase == predictSelectAway {
			b.WriteString(SubtitleStyle("Now pick the away team"))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		if m.phase == predictSelectHome {
			b.WriteString(m.homeList.View())
		} else {
			b.WriteString(m.awayList.View())
		}
		b.WriteString("\n")
		b.WriteString(HelpStyle("type to filter · enter: select · esc: menu"))

	case predictLoading:
		b.WriteString(TitleStyle("Predict a Match"))
		b.WriteString("\n\n")
		b.WriteString(WarningStyle(fmt.Sprintf("training models and predicting %s vs %s...",
			m.homeTeam.Name, m.awayTeam.Name)))

	case predictDone:
		b.WriteString(TitleStyle("Prediction"))
		b.WriteString("\n")
		if m.err != nil {
			b.WriteString(ErrorStyle(m.err.Error()))
			b.WriteString("\n")
		} else if m.ready {
			b.WriteString(m.viewport.View())
		}
		b.WriteString("\n")
		b.WriteString(HelpStyle("n: new prediction · pgup/pgdn: scroll · esc: menu"))
	}

	return b.String()
}

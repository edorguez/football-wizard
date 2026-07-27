package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/edorguez/football-wizard/internal/database"
)

type PredictModel struct {
	app        *AppContext
	state      int
	homeInput  textinput.Model
	awayInput  textinput.Model
	spinner    spinner.Model
	predicting bool
	result     string
	message    string
}

func NewPredictModel(app *AppContext) *PredictModel {
	home := textinput.New()
	home.Placeholder = "Flamengo"
	home.CharLimit = 50
	home.Width = 30
	home.Focus()

	away := textinput.New()
	away.Placeholder = "Palmeiras"
	away.CharLimit = 50
	away.Width = 30

	s := spinner.New()
	s.Style = focusedStyle
	s.Spinner = spinner.Dot

	return &PredictModel{
		app:       app,
		state:     0,
		homeInput: home,
		awayInput: away,
		spinner:   s,
	}
}

func (m *PredictModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *PredictModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return switchViewMsg(dashboardView) }
		case "tab":
			if m.state == 0 {
				m.state = 1
				m.homeInput.Blur()
				m.awayInput.Focus()
			} else {
				m.state = 0
				m.awayInput.Blur()
				m.homeInput.Focus()
			}
		case "enter":
			if !m.predicting && m.homeInput.Value() != "" && m.awayInput.Value() != "" {
				m.predicting = true
				m.result = ""
				m.message = ""
				return m, tea.Batch(m.spinner.Tick, m.predictCmd(m.homeInput.Value(), m.awayInput.Value()))
			}
		}
	case predictDoneMsg:
		m.predicting = false
		m.result = string(msg)
		return m, nil
	case predictErrorMsg:
		m.predicting = false
		m.message = errorStyle.Render(fmt.Sprintf("error: %s", string(msg)))
		return m, nil
	case spinner.TickMsg:
		if m.predicting {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	if m.state == 0 {
		m.homeInput, cmd = m.homeInput.Update(msg)
	} else {
		m.awayInput, cmd = m.awayInput.Update(msg)
	}
	return m, cmd
}

func (m *PredictModel) View() string {
	s := titleStyle.Render("🔮 Predict Match")
	s += "\n\n"

	if !m.predicting && m.result == "" && m.message == "" {
		s += headerStyle.Render("teams:")
		s += "\n\n"

		homeLabel := blurredStyle.Render("Home: ")
		if m.state == 0 {
			homeLabel = focusedStyle.Render("Home: ")
		}
		s += homeLabel + m.homeInput.View() + "\n\n"

		awayLabel := blurredStyle.Render("Away: ")
		if m.state == 1 {
			awayLabel = focusedStyle.Render("Away: ")
		}
		s += awayLabel + m.awayInput.View() + "\n\n"

		s += helpStyle.Render("Tab: switch field • Enter: predict • Esc: go back")
	}

	if m.predicting {
		s += m.spinner.View() + " generating prediction\n"
	}

	if m.result != "" {
		s += "\n" + formatPrediction(m.homeInput.Value(), m.awayInput.Value(), m.result)
		s += "\n\n" + helpStyle.Render("Esc to go back to menu")
	}

	if m.message != "" {
		s += "\n\n" + m.message
		s += "\n\n" + helpStyle.Render("Esc to go back")
	}

	return s
}

func (m *PredictModel) predictCmd(home, away string) tea.Cmd {
	return func() tea.Msg {
		homeTeam, err := m.app.TeamRepo.FindByName(home)
		if err != nil {
			return predictErrorMsg(fmt.Sprintf("home team %q not found", home))
		}
		awayTeam, err := m.app.TeamRepo.FindByName(away)
		if err != nil {
			return predictErrorMsg(fmt.Sprintf("away team %q not found", away))
		}

		pred, err := m.app.Predictor.Predict(homeTeam.ID, awayTeam.ID, time.Now())
		if err != nil {
			return predictErrorMsg(err.Error())
		}

		return predictDoneMsg(formatPredictionData(pred))
	}
}

func formatPrediction(home, away, data string) string {
	return borderStyle.Render(data)
}

func formatPredictionData(pred *database.Prediction) string {
	content := fmt.Sprintf("Home: %.1f%% | Draw: %.1f%% | Away: %.1f%%\n\n",
		pred.HomeWinProb*100, pred.DrawProb*100, pred.AwayWinProb*100)
	content += fmt.Sprintf("Over 0.5: %.0f%% | Over 1.5: %.0f%% | Over 2.5: %.0f%% | Over 3.5: %.0f%%\n\n",
		pred.Over05Prob*100, pred.Over15Prob*100, pred.Over25Prob*100, pred.Over35Prob*100)
	content += fmt.Sprintf("BTTS: %.0f%%\n\n", pred.BttsYesProb*100)
	content += fmt.Sprintf("Confidence: %s (%.2f)", pred.ConfidenceLevel, pred.ConfidenceScore)
	return content
}

type predictDoneMsg string
type predictErrorMsg string

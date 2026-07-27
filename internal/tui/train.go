package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type TrainModel struct {
	app      *AppContext
	spinner  spinner.Model
	training bool
	done     bool
	message  string
}

func NewTrainModel(app *AppContext) *TrainModel {
	s := spinner.New()
	s.Style = focusedStyle
	s.Spinner = spinner.Dot

	return &TrainModel{
		app:     app,
		spinner: s,
	}
}

func (m *TrainModel) Init() tea.Cmd {
	return nil
}

func (m *TrainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return switchViewMsg(dashboardView) }
		case "enter":
			if !m.training {
				m.training = true
				m.done = false
				m.message = ""
				return m, tea.Batch(m.spinner.Tick, m.trainCmd())
			}
		}
	case trainDoneMsg:
		m.training = false
		m.done = true
		m.message = successStyle.Render(fmt.Sprintf("model trained — version %s", string(msg)))
		return m, nil
	case trainErrorMsg:
		m.training = false
		m.message = errorStyle.Render(fmt.Sprintf("error: %s", string(msg)))
		return m, nil
	case spinner.TickMsg:
		if m.training {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *TrainModel) View() string {
	s := titleStyle.Render("🧠 Train Model")
	s += "\n\n"

	if !m.training && !m.done {
		s += headerStyle.Render("models to train:")
		s += "\n\n"
		s += "  • Poisson — expected goals (1X2, O/U)\n"
		s += "  • Logistic Regression — BTTS, cards, corners\n"
		s += "  • Decision Tree — confidence level\n"
		s += "\n"
		s += selectedStyle.Render("  Press Enter to start")
		s += "\n\n"
		s += helpStyle.Render("Enter to train • Esc to go back")
	}

	if m.training {
		s += m.spinner.View() + " training model\n"
	}

	if m.message != "" {
		s += "\n\n" + m.message
		s += "\n\n" + helpStyle.Render("Esc to go back to menu")
	}

	return s
}

func (m *TrainModel) trainCmd() tea.Cmd {
	return func() tea.Msg {
		seasons := []int{2024, 2025}
		version, err := m.app.Trainer.Train(seasons)
		if err != nil {
			return trainErrorMsg(err.Error())
		}
		time.Sleep(500 * time.Millisecond)
		return trainDoneMsg(version)
	}
}

type trainDoneMsg string
type trainErrorMsg string

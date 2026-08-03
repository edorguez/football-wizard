package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/edorguez/football-wizard/internal/model"
)

const trainHeaderHeight = 5

type TrainModel struct {
	deps     Deps
	viewport viewport.Model
	running  bool
	done     bool
	report   *model.EvalSummary
	err      error
	ready    bool
}

func NewTrainModel(deps Deps) tea.Model {
	return &TrainModel{deps: deps}
}

func (m *TrainModel) Init() tea.Cmd {
	return nil
}

type trainDoneMsg struct {
	report *model.EvalSummary
	err    error
}

func (m *TrainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-trainHeaderHeight)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - trainHeaderHeight
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, backCmd()
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if !m.running {
				m.running = true
				m.done = false
				m.err = nil
				return m, m.run()
			}
		}

	case trainDoneMsg:
		m.running = false
		m.done = true
		m.report = msg.report
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

func (m *TrainModel) run() tea.Cmd {
	return func() tea.Msg {
		rows, err := m.deps.Matches.ListRows()
		if err != nil {
			return trainDoneMsg{err: fmt.Errorf("loading matches: %w", err)}
		}
		if len(rows) == 0 {
			return trainDoneMsg{err: fmt.Errorf("no completed matches in the database")}
		}
		report, err := model.Evaluate(rows, model.DefaultSplitRound, m.deps.NewTrainer())
		if err != nil {
			return trainDoneMsg{err: err}
		}
		return trainDoneMsg{report: report}
	}
}

func (m *TrainModel) render() {
	var b strings.Builder
	if m.err != nil {
		b.WriteString(ErrorStyle(m.err.Error()))
		b.WriteString("\n")
	} else {
		fmt.Fprintf(&b, "train/test split at round %d (%d train, %d test)\n\n",
			m.report.SplitRound, m.report.TrainCount, m.report.TestCount)
		b.WriteString(m.report.Table())
	}
	m.viewport.SetContent(b.String())
	m.viewport.GotoTop()
}

func (m *TrainModel) View() string {
	var b strings.Builder
	b.WriteString(TitleStyle("Train Models"))
	b.WriteString("\n")

	switch {
	case m.running:
		b.WriteString(WarningStyle("training models..."))
	case m.done && m.err != nil:
		b.WriteString(ErrorStyle("training failed — see details below"))
	case m.done:
		b.WriteString(SuccessStyle("training complete"))
	default:
		b.WriteString(MutedStyle("press enter to train on the current database"))
	}
	b.WriteString("\n\n")

	if m.ready {
		b.WriteString(m.viewport.View())
	}
	b.WriteString("\n")
	b.WriteString(HelpStyle("enter: train · pgup/pgdn: scroll · esc: menu"))
	return b.String()
}

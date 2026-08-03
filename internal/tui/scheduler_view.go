package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type SchedulerModel struct {
	deps     Deps
	running  bool
	nextRuns map[string]time.Time
	err      error
}

func NewSchedulerModel(deps Deps) tea.Model {
	return &SchedulerModel{deps: deps, nextRuns: map[string]time.Time{}}
}

func (m *SchedulerModel) Init() tea.Cmd {
	return m.refresh()
}

type schedulerRefreshMsg struct {
	running bool
	runs    map[string]time.Time
	err     error
}

func (m *SchedulerModel) refresh() tea.Cmd {
	return func() tea.Msg {
		if m.deps.Scheduler == nil {
			return schedulerRefreshMsg{err: fmt.Errorf("scheduler is not available")}
		}
		return schedulerRefreshMsg{
			running: m.deps.Scheduler.Running(),
			runs:    m.deps.Scheduler.NextRuns(),
		}
	}
}

func (m *SchedulerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, backCmd()
		case "s":
			if m.deps.Scheduler != nil {
				if err := m.deps.Scheduler.Start(); err != nil {
					m.err = err
				}
			}
			return m, m.refresh()
		case "x":
			if m.deps.Scheduler != nil {
				m.deps.Scheduler.Stop()
			}
			return m, m.refresh()
		case "r":
			return m, m.refresh()
		}

	case schedulerRefreshMsg:
		m.running = msg.running
		m.nextRuns = msg.runs
		m.err = msg.err
		return m, nil
	}
	return m, nil
}

func (m *SchedulerModel) View() string {
	var b strings.Builder
	b.WriteString(TitleStyle("Scheduler"))
	b.WriteString("\n\n")

	status := MutedStyle("stopped")
	if m.running {
		status = SuccessStyle("running")
	}
	fmt.Fprintf(&b, "Status: %s\n\n", status)

	if len(m.nextRuns) == 0 {
		b.WriteString(MutedStyle("no scheduled jobs"))
	} else {
		b.WriteString("Next runs:\n")
		for label, next := range m.nextRuns {
			if m.running && !next.IsZero() {
				fmt.Fprintf(&b, "  %-12s %s\n", label, next.Local().Format("2006-01-02 15:04"))
			} else {
				fmt.Fprintf(&b, "  %-12s -\n", label)
			}
		}
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(ErrorStyle(m.err.Error()))
	}

	b.WriteString("\n\n")
	b.WriteString(HelpStyle("s: start · x: stop · r: refresh · esc: menu"))
	return b.String()
}

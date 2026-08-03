package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenDashboard screen = iota + 1
	screenScrape
	screenTrain
	screenPredict
	screenHistory
	screenTeam
	screenReferee
	screenScheduler
)

func (s screen) String() string {
	switch s {
	case screenDashboard:
		return "Dashboard"
	case screenScrape:
		return "Scrape"
	case screenTrain:
		return "Train"
	case screenPredict:
		return "Predict"
	case screenHistory:
		return "Recent Predictions"
	case screenTeam:
		return "Team Stats"
	case screenReferee:
		return "Referee Profile"
	case screenScheduler:
		return "Scheduler"
	default:
		return "Unknown"
	}
}

type switchScreenMsg struct{ screen screen }

func switchTo(s screen) tea.Cmd {
	return func() tea.Msg { return switchScreenMsg{screen: s} }
}

func backCmd() tea.Cmd {
	return switchTo(screenDashboard)
}

// Root is the top-level model. It holds the active view and handles navigation
// between screens.
type Root struct {
	deps    Deps
	current tea.Model
	screen  screen
	width   int
	height  int
}

func NewRoot(deps Deps) *Root {
	return &Root{
		deps:    deps,
		screen:  screenDashboard,
		current: NewDashboardModel(deps),
	}
}

func (r *Root) Init() tea.Cmd {
	return r.current.Init()
}

func (r *Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width, r.height = msg.Width, msg.Height
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return r, tea.Quit
		}
	case switchScreenMsg:
		r.screen = msg.screen
		r.current = newScreen(msg.screen, r.deps)

		// New views size their viewports/lists off WindowSizeMsg; apply the last
		// known size so they render immediately instead of hanging on "loading...".
		cmds := []tea.Cmd{r.current.Init()}
		if r.width > 0 {
			var cmd tea.Cmd
			r.current, cmd = r.current.Update(tea.WindowSizeMsg{Width: r.width, Height: r.height})
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return r, tea.Batch(cmds...)
	}

	var cmd tea.Cmd
	r.current, cmd = r.current.Update(msg)
	return r, cmd
}

func (r *Root) View() string {
	return r.current.View()
}

func newScreen(s screen, deps Deps) tea.Model {
	switch s {
	case screenScrape:
		return NewScrapeModel(deps)
	case screenTrain:
		return NewTrainModel(deps)
	case screenPredict:
		return NewPredictModel(deps)
	case screenHistory:
		return NewHistoryModel(deps)
	case screenTeam:
		return NewTeamModel(deps)
	case screenReferee:
		return NewRefereeModel(deps)
	case screenScheduler:
		return NewSchedulerModel(deps)
	default:
		return NewDashboardModel(deps)
	}
}

package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type view int

const (
	dashboardView view = iota
	scrapeView
	trainView
	predictView
	listView
	statsView
	refereeView
	scheduleView
)

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Back  key.Binding
	Quit  key.Binding
	Tab   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Back, k.Enter}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter},
		{k.Tab, k.Back, k.Quit},
	}
}

var keys = keyMap{
	Up:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Back:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "go back")),
	Quit:  key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("ctrl+c/q", "quit")),
	Tab:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
}

type sessionState struct {
	currentView  view
	previousView view
}

type RootModel struct {
	state    sessionState
	help     help.Model
	quitting bool
	app      *AppContext

	dashboard *DashboardModel
	scrape    *ScrapeModel
	train     *TrainModel
	predict   *PredictModel
	list      *ListModel
	stats     *StatsModel
	referee   *RefereeModel
	schedule  *ScheduleModel
}

func NewRootModel(app *AppContext) *RootModel {
	return &RootModel{
		state:    sessionState{currentView: dashboardView},
		help:     help.New(),
		app:      app,
		dashboard: NewDashboardModel(),
		scrape:    NewScrapeModel(app),
		train:     NewTrainModel(app),
		predict:   NewPredictModel(app),
		list:      NewListModel(app),
		stats:     NewStatsModel(app),
		referee:   NewRefereeModel(app),
		schedule:  NewScheduleModel(app),
	}
}

func (m *RootModel) Init() tea.Cmd {
	return nil
}

func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			if m.state.currentView == dashboardView {
				m.quitting = true
				return m, tea.Quit
			}
		case key.Matches(msg, keys.Back):
			return m.switchToView(dashboardView), nil
		}
	case switchViewMsg:
		return m.switchToView(view(msg)), nil
	}

	var cmd tea.Cmd
	switch m.state.currentView {
	case dashboardView:
		updated, c := m.dashboard.Update(msg)
		m.dashboard = updated.(*DashboardModel)
		cmd = c
	case scrapeView:
		updated, c := m.scrape.Update(msg)
		m.scrape = updated.(*ScrapeModel)
		cmd = c
	case trainView:
		updated, c := m.train.Update(msg)
		m.train = updated.(*TrainModel)
		cmd = c
	case predictView:
		updated, c := m.predict.Update(msg)
		m.predict = updated.(*PredictModel)
		cmd = c
	case listView:
		updated, c := m.list.Update(msg)
		m.list = updated.(*ListModel)
		cmd = c
	case statsView:
		updated, c := m.stats.Update(msg)
		m.stats = updated.(*StatsModel)
		cmd = c
	case refereeView:
		updated, c := m.referee.Update(msg)
		m.referee = updated.(*RefereeModel)
		cmd = c
	case scheduleView:
		updated, c := m.schedule.Update(msg)
		m.schedule = updated.(*ScheduleModel)
		cmd = c
	}

	return m, cmd
}

func (m *RootModel) View() string {
	if m.quitting {
		return "\n  Bye!\n\n"
	}

	var content string
	switch m.state.currentView {
	case dashboardView:
		content = m.dashboard.View()
	case scrapeView:
		content = m.scrape.View()
	case trainView:
		content = m.train.View()
	case predictView:
		content = m.predict.View()
	case listView:
		content = m.list.View()
	case statsView:
		content = m.stats.View()
	case refereeView:
		content = m.referee.View()
	case scheduleView:
		content = m.schedule.View()
	}

	helpView := m.help.View(keys)
	return appStyle.Render(content + "\n\n" + helpView)
}

func (m *RootModel) switchToView(v view) tea.Model {
	m.state.previousView = m.state.currentView
	m.state.currentView = v
	return m
}

type switchViewMsg view

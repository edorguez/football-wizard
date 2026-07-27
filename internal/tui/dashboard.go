package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type menuItem struct {
	title       string
	description string
	view        view
}

var menuItems = []menuItem{
	{title: "📊 Scrape Data", description: "Scrape results and stats from FBref", view: scrapeView},
	{title: "🧠 Train Model", description: "Train prediction model with historical data", view: trainView},
	{title: "🔮 Predict Match", description: "Generate prediction for a match", view: predictView},
	{title: "📋 Recent Predictions", description: "View latest predictions", view: listView},
	{title: "📈 Team Stats", description: "View team statistics", view: statsView},
	{title: "🃏 Referee Profile", description: "View referee statistics", view: refereeView},
	{title: "⏰ Scheduler", description: "Control the automatic scheduler", view: scheduleView},
}

type DashboardModel struct {
	selected int
}

func NewDashboardModel() *DashboardModel {
	return &DashboardModel{}
}

func (m *DashboardModel) Init() tea.Cmd {
	return nil
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(menuItems)-1 {
				m.selected++
			}
		case "enter":
			item := menuItems[m.selected]
			return m, func() tea.Msg {
				return switchViewMsg(item.view)
			}
		}
	}
	return m, nil
}

func (m *DashboardModel) View() string {
	s := titleStyle.Render("⚽ Football Wizard — Brasileirão Serie A")
	s += "\n\n"
	s += headerStyle.Render("Main Menu")
	s += "\n\n"

	for i, item := range menuItems {
		cursor := "  "
		if i == m.selected {
			cursor = "▸ "
			s += selectedStyle.Render(fmt.Sprintf("%s%s", cursor, item.title))
		} else {
			s += itemStyle.Render(fmt.Sprintf("%s%s", cursor, item.title))
		}
		s += "\n"
		s += dimmedStyle.Render(fmt.Sprintf("   %s", item.description))
		s += "\n\n"
	}

	return s
}

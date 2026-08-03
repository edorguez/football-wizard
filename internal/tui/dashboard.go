package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type menuItem struct {
	label  string
	desc   string
	key    string
	screen screen
}

var menuItems = []menuItem{
	{label: "Scrape", desc: "Scrape a season with real-time logs", key: "1", screen: screenScrape},
	{label: "Train", desc: "Train models and show held-out accuracy", key: "2", screen: screenTrain},
	{label: "Predict", desc: "Pick two teams and forecast the match", key: "3", screen: screenPredict},
	{label: "Recent Predictions", desc: "Browse predictions saved to the DB", key: "4", screen: screenHistory},
	{label: "Team Stats", desc: "Search a team's recent form", key: "5", screen: screenTeam},
	{label: "Referee Profile", desc: "Search a referee's match history", key: "6", screen: screenReferee},
	{label: "Scheduler", desc: "Start/stop the cron scheduler", key: "7", screen: screenScheduler},
	{label: "Quit", desc: "Exit the application", key: "q"},
}

type DashboardModel struct {
	deps Deps
}

func NewDashboardModel(deps Deps) tea.Model {
	return DashboardModel{deps: deps}
}

func (m DashboardModel) Init() tea.Cmd {
	return nil
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "1", "2", "3", "4", "5", "6", "7":
		for _, item := range menuItems {
			if item.key == keyMsg.String() && item.screen != 0 {
				return m, switchTo(item.screen)
			}
		}
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m DashboardModel) View() string {
	var b strings.Builder
	b.WriteString(TitleStyle("Football Wizard"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle("Predict Brasileirão Série A matches"))
	b.WriteString("\n\n")

	for _, item := range menuItems {
		fmt.Fprintf(&b, "  %s  %-20s %s\n",
			MenuItemStyle(item.key),
			MenuItemStyle(item.label),
			MutedStyle(item.desc),
		)
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle("Select an option (1-7) · q to quit · ctrl+c to force quit"))
	return b.String()
}

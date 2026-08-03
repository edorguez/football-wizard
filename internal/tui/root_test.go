package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootStartsOnDashboard(t *testing.T) {
	t.Parallel()

	root := NewRoot(testDeps(t))
	is := assert.New(t)
	is.Equal(screenDashboard, root.screen)
	is.IsType(DashboardModel{}, root.current)
}

func TestRootSwitchesScreens(t *testing.T) {
	t.Parallel()

	root := NewRoot(testDeps(t))
	updated, _ := root.Update(switchScreenMsg{screen: screenScrape})
	next := updated.(*Root)

	is := assert.New(t)
	is.Equal(screenScrape, next.screen)
	is.IsType(&ScrapeModel{}, next.current)
}

func TestRootForwardsWindowSize(t *testing.T) {
	t.Parallel()

	root := NewRoot(testDeps(t))
	updated, _ := root.Update(switchScreenMsg{screen: screenScrape})
	next := updated.(*Root)

	updated, _ = next.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	after := updated.(*Root)

	scrape, ok := after.current.(*ScrapeModel)
	require.True(t, ok)
	assert.True(t, scrape.ready)
}

// TestRootReSendsWindowSizeOnSwitch guards against the "stuck loading" bug:
// a newly created view must receive the last window size so it stops showing
// its loading placeholder and sizes its viewport/lists.
func TestRootReSendsWindowSizeOnSwitch(t *testing.T) {
	t.Parallel()

	root := NewRoot(testDeps(t))

	// Simulate the app starting: the root sees a window size while on the
	// dashboard.
	updated, _ := root.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	root = updated.(*Root)

	// Navigate to the scrape view: the stored window size must be re-applied.
	updated, cmd := root.Update(switchScreenMsg{screen: screenScrape})
	next := updated.(*Root)
	require.NotNil(t, cmd)

	scrape := next.current.(*ScrapeModel)
	assert.True(t, scrape.ready, "scrape view should be ready after re-dispatch")
	assert.NotContains(t, scrape.View(), "loading...")
}

func TestRootCtrlCQuits(t *testing.T) {
	t.Parallel()

	root := NewRoot(testDeps(t))
	_, cmd := root.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	require.NotNil(t, cmd)
	assert.IsType(t, tea.QuitMsg{}, cmd())
}

func TestDashboardSelectNavigates(t *testing.T) {
	t.Parallel()

	dashboard := NewDashboardModel(testDeps(t))
	updated, cmd := dashboard.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})

	is := assert.New(t)
	is.IsType(DashboardModel{}, updated)
	require.NotNil(t, cmd)

	msg, ok := cmd().(switchScreenMsg)
	require.True(t, ok)
	is.Equal(screenScrape, msg.screen)
}

func TestDashboardQuit(t *testing.T) {
	t.Parallel()

	dashboard := NewDashboardModel(testDeps(t))
	_, cmd := dashboard.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	require.NotNil(t, cmd)
	assert.IsType(t, tea.QuitMsg{}, cmd())
}

func TestDashboardViewRendersMenu(t *testing.T) {
	t.Parallel()

	dashboard := NewDashboardModel(testDeps(t))
	view := dashboard.View()

	is := assert.New(t)
	is.Contains(view, "Football Wizard")
	is.Contains(view, "Scrape")
	is.Contains(view, "Scheduler")
}

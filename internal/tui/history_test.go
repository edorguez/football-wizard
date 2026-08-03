package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/edorguez/football-wizard/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryEmpty(t *testing.T) {
	t.Parallel()

	model := NewHistoryModel(testDeps(t)).(*HistoryModel)
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model.Update(cmdMsg(model.Init()))

	assert.Contains(t, model.View(), "no predictions saved yet")
}

func TestHistoryRendersSavedPredictions(t *testing.T) {
	t.Parallel()

	deps := testDeps(t)
	home := insertTeam(t, deps, "Flamengo")
	away := insertTeam(t, deps, "Palmeiras")

	model := NewHistoryModel(deps).(*HistoryModel)
	require.NoError(t, deps.Predictions.Create(&database.Prediction{
		HomeTeamID: home.ID,
		AwayTeamID: away.ID,
		HomeWin:    0.6,
		Draw:       0.25,
		AwayWin:    0.15,
	}))

	model.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	model.Update(cmdMsg(model.Init()))

	is := assert.New(t)
	is.Contains(model.View(), "Flamengo vs Palmeiras")
	is.Contains(model.View(), "60.0%")
}

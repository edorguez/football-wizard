package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/edorguez/football-wizard/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPredictPhaseFlow(t *testing.T) {
	t.Parallel()

	deps := testDeps(t)
	insertTeam(t, deps, "Flamengo")
	insertTeam(t, deps, "Palmeiras")

	model := NewPredictModel(deps).(*PredictModel)

	// Load teams.
	updated, cmd := model.Update(cmdMsg(model.Init()))
	m := updated.(*PredictModel)
	require.Equal(t, predictSelectHome, m.phase)
	require.NotNil(t, m.homeList.Items())
	_ = cmd

	// Select home team (cursor starts on first item).
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*PredictModel)
	require.Equal(t, predictSelectAway, m.phase)
	require.NotNil(t, m.homeTeam)

	// Select away team -> prediction starts.
	updated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*PredictModel)
	require.Equal(t, predictLoading, m.phase)
	require.NotNil(t, cmd)

	// No matches in the DB, so the prediction reports an error.
	ready, ok := cmd().(predictionReadyMsg)
	require.True(t, ok)
	require.Error(t, ready.err)
}

func TestPredictNoTeams(t *testing.T) {
	t.Parallel()

	model := NewPredictModel(testDeps(t)).(*PredictModel)
	updated, _ := model.Update(cmdMsg(model.Init()))
	m := updated.(*PredictModel)

	assert.Equal(t, predictSelectHome, m.phase)
	assert.Empty(t, m.homeList.Items())
}

func TestPredictNewPredictionResets(t *testing.T) {
	t.Parallel()

	deps := testDeps(t)
	insertTeam(t, deps, "Flamengo")
	insertTeam(t, deps, "Palmeiras")

	model := NewPredictModel(deps).(*PredictModel)
	model.Update(cmdMsg(model.Init()))
	model.phase = predictDone
	model.homeTeam = &database.Team{Name: "Flamengo"}
	model.awayTeam = &database.Team{Name: "Palmeiras"}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m := updated.(*PredictModel)

	assert.Equal(t, predictSelectHome, m.phase)
	assert.Nil(t, m.homeTeam)
	assert.Nil(t, m.awayTeam)
}

func cmdMsg(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

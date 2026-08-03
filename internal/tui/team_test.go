package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamLoadsAndSelects(t *testing.T) {
	t.Parallel()

	deps := testDeps(t)
	insertTeam(t, deps, "Flamengo")
	insertTeam(t, deps, "Palmeiras")

	model := NewTeamModel(deps).(*TeamModel)
	model.Update(cmdMsg(model.Init()))
	require.NotNil(t, model.list.Items())

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(*TeamModel)
	require.True(t, m.detail)
	require.NotNil(t, m.team)
	require.NotNil(t, cmd)

	// The team has no matches yet.
	msg := cmd()
	loaded, ok := msg.(teamMatchesMsg)
	require.True(t, ok)
	require.NoError(t, loaded.err)
	assert.Empty(t, loaded.matches)

	// Esc returns to the team list.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(*TeamModel)
	assert.False(t, m.detail)
}

func TestTeamViewEmptyState(t *testing.T) {
	t.Parallel()

	model := NewTeamModel(testDeps(t)).(*TeamModel)
	model.Update(cmdMsg(model.Init()))
	model.ready = true

	assert.Contains(t, model.View(), "Team Stats")
}

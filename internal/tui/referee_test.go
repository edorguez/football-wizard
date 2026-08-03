package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/edorguez/football-wizard/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefereeSearchFlow(t *testing.T) {
	t.Parallel()

	deps := testDeps(t)
	insertReferee(t, deps, "Raphael Claus")
	insertReferee(t, deps, "Rafael Traci")

	model := NewRefereeModel(deps).(*RefereeModel)
	model.input.SetValue("Rapha")

	// Enter triggers a search.
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(*RefereeModel)
	require.NotNil(t, cmd)

	msg := cmd()
	search, ok := msg.(refereeSearchMsg)
	require.True(t, ok)
	require.NoError(t, search.err)

	updated, _ = m.Update(search)
	m = updated.(*RefereeModel)
	assert.True(t, m.searched)
	require.Len(t, m.results, 1)
	assert.Equal(t, "Raphael Claus", m.results[0].Name)
}

func TestRefereeSelectLoadsMatches(t *testing.T) {
	t.Parallel()

	deps := testDeps(t)
	insertReferee(t, deps, "Raphael Claus")

	model := NewRefereeModel(deps).(*RefereeModel)
	model.results = []database.Referee{{Name: "Raphael Claus"}}
	model.searched = true

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(*RefereeModel)
	require.True(t, m.detail)
	require.NotNil(t, m.ref)
	require.NotNil(t, cmd)

	msg := cmd()
	loaded, ok := msg.(refereeMatchesMsg)
	require.True(t, ok)
	require.NoError(t, loaded.err)
}

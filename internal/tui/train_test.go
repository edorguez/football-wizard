package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrainStartsOnEnter(t *testing.T) {
	t.Parallel()

	model := NewTrainModel(testDeps(t)).(*TrainModel)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(*TrainModel)

	require.NotNil(t, cmd)
	assert.True(t, m.running)

	// The database is empty, so training reports a "no matches" error.
	msg := cmd()
	done, ok := msg.(trainDoneMsg)
	require.True(t, ok)
	require.Error(t, done.err)

	updated, _ = m.Update(done)
	m = updated.(*TrainModel)
	assert.False(t, m.running)
	assert.True(t, m.done)
}

func TestTrainRendersReport(t *testing.T) {
	t.Parallel()

	model := NewTrainModel(testDeps(t)).(*TrainModel)
	model.ready = true
	model.done = true

	model.Update(trainDoneMsg{err: assert.AnError})
	assert.Contains(t, model.View(), "training failed")

	model = NewTrainModel(testDeps(t)).(*TrainModel)
	model.ready = true
	model.done = true
	assert.Contains(t, model.View(), "training complete")
}

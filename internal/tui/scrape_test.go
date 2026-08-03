package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScrapeInvalidSeason(t *testing.T) {
	t.Parallel()

	model := NewScrapeModel(testDeps(t)).(*ScrapeModel)
	model.input.SetValue("abc")

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(*ScrapeModel)

	require.Nil(t, cmd)
	require.Error(t, m.err)
}

func TestScrapeStartCmd(t *testing.T) {
	t.Parallel()

	model := NewScrapeModel(testDeps(t)).(*ScrapeModel)
	model.input.SetValue("2024")

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := updated.(*ScrapeModel)

	require.NotNil(t, cmd)
	msg, ok := cmd().(scrapeStartMsg)
	require.True(t, ok)
	assert.Equal(t, 2024, msg.season)
	assert.False(t, msg.full)
	assert.False(t, m.running)
}

func TestScrapeFullToggle(t *testing.T) {
	t.Parallel()

	model := NewScrapeModel(testDeps(t)).(*ScrapeModel)
	model.input.SetValue("2024")

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	m := updated.(*ScrapeModel)
	assert.True(t, m.full)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*ScrapeModel)
	require.NotNil(t, cmd)
	msg, ok := cmd().(scrapeStartMsg)
	require.True(t, ok)
	assert.True(t, msg.full)
}

func TestScrapeRunningBlocksInput(t *testing.T) {
	t.Parallel()

	model := NewScrapeModel(testDeps(t)).(*ScrapeModel)
	model.running = true

	// Esc is ignored while running.
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m := updated.(*ScrapeModel)
	assert.Nil(t, cmd)
	assert.True(t, m.running)
}

func TestScrapeDoneClearsRunning(t *testing.T) {
	t.Parallel()

	model := NewScrapeModel(testDeps(t)).(*ScrapeModel)
	model.running = true
	model.ready = true

	updated, _ := model.Update(scrapeDoneMsg{})
	m := updated.(*ScrapeModel)

	assert.False(t, m.running)
	assert.True(t, m.done)
	assert.NoError(t, m.err)
}

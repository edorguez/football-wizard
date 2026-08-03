package tui

import (
	"io"
	"log/slog"
	"testing"

	"github.com/edorguez/football-wizard/internal/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchedulerViewUnavailable(t *testing.T) {
	t.Parallel()

	model := NewSchedulerModel(testDeps(t)).(*SchedulerModel)
	model.Update(cmdMsg(model.Init()))

	assert.Contains(t, model.View(), "scheduler is not available")
}

func TestSchedulerViewStartStop(t *testing.T) {
	t.Parallel()

	deps := testDeps(t)
	s, err := scheduler.New([]scheduler.Job{
		{Label: "scrape", Spec: "0 1 * * *", Run: func() error { return nil }},
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	require.NoError(t, err)
	deps.Scheduler = s

	model := NewSchedulerModel(deps).(*SchedulerModel)

	// Start.
	updated, _ := model.Update(cmdMsg(model.Init()))
	m := updated.(*SchedulerModel)
	require.False(t, m.running)

	updated, cmd := m.Update(teaKey("s"))
	m = updated.(*SchedulerModel)
	require.True(t, deps.Scheduler.Running())

	updated, _ = m.Update(cmdMsg(cmd))
	m = updated.(*SchedulerModel)
	assert.True(t, m.running)

	// Stop.
	updated, cmd = m.Update(teaKey("x"))
	m = updated.(*SchedulerModel)
	require.False(t, deps.Scheduler.Running())

	updated, _ = m.Update(cmdMsg(cmd))
	m = updated.(*SchedulerModel)
	assert.False(t, m.running)
}

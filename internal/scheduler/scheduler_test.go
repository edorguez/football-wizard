package scheduler

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDailySpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		hhmm    string
		want    string
		wantErr bool
	}{
		{name: "valid", hhmm: "01:00", want: "0 1 * * *"},
		{name: "zero padded", hhmm: "03:30", want: "30 3 * * *"},
		{name: "missing minutes", hhmm: "01", wantErr: true},
		{name: "bad hour", hhmm: "25:00", wantErr: true},
		{name: "bad minute", hhmm: "01:99", wantErr: true},
		{name: "text", hhmm: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DailySpec(tt.hhmm)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWeeklySpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		day     string
		hhmm    string
		want    string
		wantErr bool
	}{
		{name: "sunday", day: "Sunday", hhmm: "03:00", want: "0 3 * * 0"},
		{name: "monday", day: "monday", hhmm: "09:15", want: "15 9 * * 1"},
		{name: "saturday", day: "Saturday", hhmm: "12:00", want: "0 12 * * 6"},
		{name: "unknown day", day: "Funday", hhmm: "03:00", wantErr: true},
		{name: "bad time", day: "Sunday", hhmm: "3pm", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WeeklySpec(tt.day, tt.hhmm)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewRejectsInvalidJob(t *testing.T) {
	t.Parallel()

	_, err := New([]Job{{Label: "broken", Spec: "not-a-spec", Run: func() error { return nil }}}, slog.New(slog.NewTextHandler(nil, nil)))
	assert.Error(t, err)
}

func TestNewRejectsEmptyLabel(t *testing.T) {
	t.Parallel()

	_, err := New([]Job{{Spec: "0 1 * * *", Run: func() error { return nil }}}, slog.New(slog.NewTextHandler(nil, nil)))
	assert.Error(t, err)
}

func TestSchedulerLifecycle(t *testing.T) {
	t.Parallel()

	s, err := New([]Job{
		{Label: "scrape", Spec: "0 1 * * *", Run: func() error { return nil }},
		{Label: "retrain", Spec: "0 3 * * 0", Run: func() error { return nil }},
	}, slog.New(slog.NewTextHandler(nil, nil)))
	require.NoError(t, err)

	is := assert.New(t)
	is.False(s.Running())

	require.NoError(t, s.Start())
	is.True(s.Running())

	runs := s.NextRuns()
	is.Len(runs, 2)
	for _, label := range []string{"scrape", "retrain"} {
		_, ok := runs[label]
		is.True(ok, "expected next run for %s", label)
	}

	s.Stop()
	is.False(s.Running())

	// Stopping again is a no-op.
	s.Stop()
	is.False(s.Running())
}

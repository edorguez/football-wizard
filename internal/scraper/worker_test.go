package scraper

import (
	"log/slog"
	"testing"
	"time"

	"github.com/edorguez/football-wizard/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorkerPool_ClampsMinWorkers(t *testing.T) {
	t.Parallel()

	pool := NewWorkerPool(nil, nil, nil, 0, time.Second, 0.5)

	is := assert.New(t)

	is.Equal(1, pool.workers)
}

func TestNewWorkerPool_DefaultDelay(t *testing.T) {
	t.Parallel()

	pool := NewWorkerPool(nil, nil, nil, 5, 2*time.Second, 0.5)

	is := assert.New(t)

	is.Equal(5, pool.workers)
	is.Equal(2*time.Second, pool.rateLimitDelay)
	is.Equal(0.5, pool.jitter)
}

func TestWorkerProgressLogging(t *testing.T) {
	t.Parallel()

	buf := logger.NewRingBuffer(100)
	pool := NewWorkerPool(nil, nil, slog.New(slog.NewTextHandler(buf, nil)), 1, 0, 0)

	// Before the first step boundary no progress is logged.
	pool.logProgress("match reports", 5, 380)
	assert.Empty(t, buf.Lines())

	// Every 10 completed items produces one progress line.
	pool.logProgress("match reports", 10, 380)
	lines := buf.Lines()
	require.Len(t, lines, 1)
	assert.Contains(t, lines[0], "match reports")
	assert.Contains(t, lines[0], "completed=10")
	assert.Contains(t, lines[0], "total=380")

	// Zero total never logs.
	pool.logProgress("squads", 10, 0)
	require.Len(t, buf.Lines(), 1)
}

func TestThrottle_NoDelay(t *testing.T) {
	t.Parallel()

	pool := NewWorkerPool(nil, nil, nil, 1, 0, 0)

	start := time.Now()
	pool.throttle()
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 100*time.Millisecond)
}

func TestThrottle_WithDelay(t *testing.T) {
	t.Parallel()

	pool := NewWorkerPool(nil, nil, nil, 1, 50*time.Millisecond, 0)

	start := time.Now()
	pool.throttle()
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 45*time.Millisecond)
}

func TestThrottle_WithJitter(t *testing.T) {
	t.Parallel()

	pool := NewWorkerPool(nil, nil, nil, 1, 100*time.Millisecond, 1.0)

	var times []time.Duration
	for range 5 {
		start := time.Now()
		pool.throttle()
		times = append(times, time.Since(start))
	}

	var min, max time.Duration
	for i, d := range times {
		if i == 0 || d < min {
			min = d
		}
		if i == 0 || d > max {
			max = d
		}
	}

	assert.Greater(t, max-min, time.Duration(0), "jitter should produce varying delays")
}

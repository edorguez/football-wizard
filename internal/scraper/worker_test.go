package scraper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

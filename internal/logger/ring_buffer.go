package logger

import (
	"strings"
	"sync"
)

// RingBuffer is a thread-safe, fixed-capacity line buffer. It implements
// io.Writer so it can be used as the output of a slog handler, which lets the
// TUI stream log output into a viewport.
type RingBuffer struct {
	mu    sync.Mutex
	lines []string
	max   int
}

// NewRingBuffer returns a ring buffer holding at most max lines.
func NewRingBuffer(max int) *RingBuffer {
	if max <= 0 {
		max = 1000
	}
	return &RingBuffer{lines: []string{}, max: max}
}

// Write appends the incoming bytes, splitting on newlines.
func (b *RingBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, part := range strings.Split(string(p), "\n") {
		if part == "" {
			continue
		}
		b.lines = append(b.lines, part)
	}
	b.trim()

	return len(p), nil
}

// Lines returns a copy of all buffered lines.
func (b *RingBuffer) Lines() []string {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := make([]string, len(b.lines))
	copy(out, b.lines)
	return out
}

// Clear empties the buffer.
func (b *RingBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lines = []string{}
}

func (b *RingBuffer) trim() {
	if len(b.lines) > b.max {
		b.lines = b.lines[len(b.lines)-b.max:]
	}
}

package logger

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRingBufferWriteAndRead(t *testing.T) {
	t.Parallel()

	buf := NewRingBuffer(10)
	is := assert.New(t)

	_, err := buf.Write([]byte("first line\nsecond line\n"))
	require.NoError(t, err)

	lines := buf.Lines()
	is.Equal([]string{"first line", "second line"}, lines)
}

func TestRingBufferIgnoresEmptyLines(t *testing.T) {
	t.Parallel()

	buf := NewRingBuffer(10)
	_, err := buf.Write([]byte("\n\nhello\n\n"))
	require.NoError(t, err)

	is := assert.New(t)
	is.Equal([]string{"hello"}, buf.Lines())
}

func TestRingBufferTrimsToMax(t *testing.T) {
	t.Parallel()

	buf := NewRingBuffer(3)
	for i := 0; i < 10; i++ {
		_, err := buf.Write([]byte(fmt.Sprintf("line-%d\n", i)))
		require.NoError(t, err)
	}

	is := assert.New(t)
	is.Equal([]string{"line-7", "line-8", "line-9"}, buf.Lines())
}

func TestRingBufferClear(t *testing.T) {
	t.Parallel()

	buf := NewRingBuffer(5)
	_, err := buf.Write([]byte("hello\n"))
	require.NoError(t, err)

	buf.Clear()

	is := assert.New(t)
	is.Empty(buf.Lines())
}

func TestRingBufferConcurrentWrites(t *testing.T) {
	t.Parallel()

	buf := NewRingBuffer(100)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_, _ = buf.Write([]byte(fmt.Sprintf("w%d-%d\n", n, j)))
			}
		}(i)
	}
	wg.Wait()

	is := assert.New(t)
	is.LessOrEqual(len(buf.Lines()), 100)
}

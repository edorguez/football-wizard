package logger

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_DefaultFormat(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(Config{
		Level:  "info",
		Format: "",
		Output: &buf,
	})

	log.Info("hello")

	output := buf.String()
	assert.Contains(t, output, "hello")
	assert.Contains(t, output, "[INF]")
}

func TestNew_JSONFormat(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(Config{
		Level:  "info",
		Format: "json",
		Output: &buf,
	})

	log.Info("hello world")

	output := buf.String()
	assert.Contains(t, output, "hello world")
	assert.Contains(t, output, `"msg"`)
}

func TestNew_TextFormat(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(Config{
		Level:  "debug",
		Format: "text",
		Output: &buf,
	})

	log.Debug("debug msg")
	log.Info("info msg")

	output := buf.String()
	assert.Contains(t, output, "debug msg")
	assert.Contains(t, output, "info msg")
}

func TestNew_LevelFiltering(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	log := New(Config{
		Level:  "error",
		Format: "text",
		Output: &buf,
	})

	log.Info("should not appear")
	log.Error("should appear")

	assert.NotContains(t, buf.String(), "should not appear")
	assert.Contains(t, buf.String(), "should appear")
}

func TestNew_LevelParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		level    string
		expected slog.Level
	}{
		{name: "debug", level: "debug", expected: slog.LevelDebug},
		{name: "info", level: "info", expected: slog.LevelInfo},
		{name: "warn", level: "warn", expected: slog.LevelWarn},
		{name: "error", level: "error", expected: slog.LevelError},
		{name: "unknown defaults to info", level: "unknown", expected: slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			log := New(Config{Level: tt.level, Output: &buf})

			is := assert.New(t)

			is.NotNil(log)
			is.True(log.Handler().Enabled(nil, tt.expected))
		})
	}
}

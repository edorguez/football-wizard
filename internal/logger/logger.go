package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

const (
	LevelSuccess = slog.Level(2)
)

type Config struct {
	Level  string
	Format string
	Output io.Writer
}

func New(cfg Config) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{Level: level})
	case "text":
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{Level: level})
	default:
		handler = newColoredHandler(output, level)
	}

	return slog.New(handler)
}

func Success(log *slog.Logger, msg string, args ...any) {
	log.Log(context.Background(), LevelSuccess, msg, args...)
}

type coloredHandler struct {
	level  slog.Level
	output io.Writer
}

func newColoredHandler(output io.Writer, level slog.Level) *coloredHandler {
	return &coloredHandler{output: output, level: level}
}

func (h *coloredHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *coloredHandler) Handle(_ context.Context, r slog.Record) error {
	timeStr := r.Time.Format("15:04:05")

	levelPrefix, color := levelFormat(r.Level)
	reset := "\033[0m"

	msg := r.Message

	var attrs string
	r.Attrs(func(a slog.Attr) bool {
		if a.Value.Kind() == slog.KindString {
			attrs += fmt.Sprintf(" %s=%s", a.Key, a.Value.String())
		} else {
			attrs += fmt.Sprintf(" %s=%v", a.Key, a.Value.Any())
		}
		return true
	})

	_, err := fmt.Fprintf(h.output, "%s[%s]%s %s %s%s\n", color, levelPrefix, reset, timeStr, msg, attrs)
	return err
}

func (h *coloredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *coloredHandler) WithGroup(name string) slog.Handler {
	return h
}

func levelFormat(level slog.Level) (string, string) {
	switch {
	case level < slog.LevelInfo:
		return "DEB", "\033[90m"
	case level == LevelSuccess:
		return "SUC", "\033[32m"
	case level < slog.LevelWarn:
		return "INF", "\033[34m"
	case level < slog.LevelError:
		return "WAR", "\033[33m"
	default:
		return "ERR", "\033[31m"
	}
}

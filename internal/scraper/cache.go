package scraper

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var cacheDir = "data/cache"

type Cache struct {
	logger *slog.Logger
}

func NewCache(logger *slog.Logger) *Cache {
	return &Cache{logger: logger}
}

func cachePath(season int, parts ...string) string {
	elements := append([]string{cacheDir, fmt.Sprintf("%d", season)}, parts...)
	return filepath.Join(elements...)
}

func (c *Cache) Read(season int, parts ...string) (string, bool) {
	path := cachePath(season, parts...)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	c.logger.Debug("cache hit", "path", path, "size", len(data))
	return string(data), true
}

func (c *Cache) Write(season int, data string, parts ...string) {
	path := cachePath(season, parts...)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.logger.Error("creating cache directory", "path", dir, "error", err)
		return
	}
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		c.logger.Error("writing cache", "path", path, "error", err)
		return
	}
	c.logger.Debug("cache written", "path", path, "size", len(data))
}

func (c *Cache) SchedulePath(season int) string {
	return cachePath(season, "schedule.html")
}

func (c *Cache) SquadPath(season int, teamID string) string {
	name := fmt.Sprintf("squad_%s.html", sanitizeFilename(teamID))
	return cachePath(season, "squads", name)
}

func (c *Cache) ReportPath(season int, matchID string) string {
	name := fmt.Sprintf("report_%s.html", sanitizeFilename(matchID))
	return cachePath(season, "reports", name)
}

func sanitizeFilename(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	return s
}

package scraper

import (
	"os"
	"path/filepath"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
)

func TestCacheReadWrite(t *testing.T) {
	tmpDir := t.TempDir()
	origCacheDir := cacheDir

	t.Cleanup(func() {
		cacheDir = origCacheDir
	})

	cacheDir = tmpDir

	cache := NewCache(slog.Default())

	cache.Write(2025, "test data", "testfile.html")

	got, ok := cache.Read(2025, "testfile.html")

	is := assert.New(t)

	is.True(ok)
	is.Equal("test data", got)
}

func TestCacheRead_Miss(t *testing.T) {
	tmpDir := t.TempDir()
	origCacheDir := cacheDir

	t.Cleanup(func() {
		cacheDir = origCacheDir
	})

	cacheDir = tmpDir

	cache := NewCache(slog.Default())

	_, ok := cache.Read(2025, "nonexistent.html")

	assert.False(t, ok)
}

func TestCacheDirCreated(t *testing.T) {
	tmpDir := t.TempDir()
	origCacheDir := cacheDir

	t.Cleanup(func() {
		cacheDir = origCacheDir
	})

	cacheDir = tmpDir

	cache := NewCache(slog.Default())

	cache.Write(2025, "data", "subdir/test.html")

	fullPath := filepath.Join(tmpDir, "2025", "subdir", "test.html")
	_, err := os.Stat(fullPath)

	assert.NoError(t, err)
}

func TestCachePaths(t *testing.T) {
	tmpDir := t.TempDir()
	origCacheDir := cacheDir

	t.Cleanup(func() {
		cacheDir = origCacheDir
	})

	cacheDir = tmpDir

	cache := NewCache(slog.Default())

	schedPath := cache.SchedulePath(2025)
	is := assert.New(t)

	is.Contains(schedPath, "2025")
	is.Contains(schedPath, "schedule.html")

	squadPath := cache.SquadPath(2025, "abc123")
	is.Contains(squadPath, "abc123")
	is.Contains(squadPath, "squads")

	reportPath := cache.ReportPath(2025, "match1")
	is.Contains(reportPath, "match1")
	is.Contains(reportPath, "reports")
}

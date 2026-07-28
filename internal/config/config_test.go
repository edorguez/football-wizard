package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ValidConfig(t *testing.T) {
	t.Parallel()

	content := []byte(`
database:
  path: /tmp/test.db
log:
  level: debug
  format: text
`)
	path := writeTempConfig(t, content)
	defer os.Remove(path)

	cfg, err := Load(path)

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	must.NotNil(cfg)

	is.Equal("/tmp/test.db", cfg.Database.Path)
	is.Equal("debug", cfg.Log.Level)
	is.Equal("text", cfg.Log.Format)
}

func TestLoad_Defaults(t *testing.T) {
	t.Parallel()

	content := []byte(``)
	path := writeTempConfig(t, content)
	defer os.Remove(path)

	cfg, err := Load(path)

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	must.NotNil(cfg)

	is.Equal("data/football-wizard.db", cfg.Database.Path)
	is.Equal("01:00", cfg.Scheduler.ScrapeTime)
	is.Equal("Sunday", cfg.Scheduler.TrainDay)
	is.Equal("03:00", cfg.Scheduler.TrainTime)
	is.Equal("info", cfg.Log.Level)
	is.Equal("json", cfg.Log.Format)
}

func TestLoad_InvalidYAML(t *testing.T) {
	t.Parallel()

	content := []byte(`invalid: [yaml: broken`)
	path := writeTempConfig(t, content)
	defer os.Remove(path)

	_, err := Load(path)

	assert.Error(t, err)
}

func writeTempConfig(t *testing.T, content []byte) string {
	t.Helper()

	f, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)

	_, err = f.Write(content)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	return f.Name()
}

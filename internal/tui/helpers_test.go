package tui

import (
	"io"
	"log/slog"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/model"
	"github.com/edorguez/football-wizard/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func teaKey(s string) tea.Msg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func testDeps(t *testing.T) Deps {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&database.Team{},
		&database.Referee{},
		&database.Match{},
		&database.MatchStat{},
		&database.Fixture{},
		&database.Prediction{},
	))

	quiet := slog.New(slog.NewTextHandler(io.Discard, nil))

	return Deps{
		Log:         quiet,
		Teams:       repository.NewTeamRepository(db),
		Matches:     repository.NewMatchRepository(db),
		Refs:        repository.NewRefereeRepository(db),
		Predictions: repository.NewPredictionRepository(db),
		Train: func() (*model.Predictor, error) {
			return nil, assert.AnError
		},
	}
}

func insertTeam(t *testing.T, deps Deps, name string) database.Team {
	t.Helper()
	team := &database.Team{Name: name, Country: "Brazil"}
	require.NoError(t, deps.Teams.Upsert(team))
	return *team
}

func insertReferee(t *testing.T, deps Deps, name string) database.Referee {
	t.Helper()
	ref := &database.Referee{Name: name}
	require.NoError(t, deps.Refs.Upsert(ref))
	return *ref
}

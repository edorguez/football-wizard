package repository

import (
	"testing"
	"time"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixtureRepository_BulkCreate(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	fixtureRepo := NewFixtureRepository(db)
	teamRepo := NewTeamRepository(db)

	home := &database.Team{Name: "Flamengo", Country: "Brazil"}
	away := &database.Team{Name: "Vasco", Country: "Brazil"}
	require.NoError(t, teamRepo.Upsert(home))
	require.NoError(t, teamRepo.Upsert(away))

	fixtures := []database.Fixture{
		{Season: 2025, Round: 1, HomeTeamID: home.ID, AwayTeamID: away.ID, Date: time.Now()},
		{Season: 2025, Round: 2, HomeTeamID: away.ID, AwayTeamID: home.ID, Date: time.Now().Add(7 * 24 * time.Hour)},
	}

	err := fixtureRepo.BulkCreate(fixtures)

	assert.NoError(t, err)
}

func TestFixtureRepository_ListUpcoming(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	fixtureRepo := NewFixtureRepository(db)
	teamRepo := NewTeamRepository(db)

	home := &database.Team{Name: "Corinthians", Country: "Brazil"}
	away := &database.Team{Name: "Santos", Country: "Brazil"}
	require.NoError(t, teamRepo.Upsert(home))
	require.NoError(t, teamRepo.Upsert(away))

	fixtures := []database.Fixture{
		{Season: 2025, Round: 1, HomeTeamID: home.ID, AwayTeamID: away.ID, Date: time.Now().Add(24 * time.Hour), Status: "scheduled"},
		{Season: 2025, Round: 2, HomeTeamID: away.ID, AwayTeamID: home.ID, Date: time.Now().Add(8 * 24 * time.Hour), Status: "scheduled"},
	}
	require.NoError(t, fixtureRepo.BulkCreate(fixtures))

	upcoming, err := fixtureRepo.ListUpcoming()

	is := assert.New(t)

	is.NoError(err)
	is.Len(upcoming, 2)
}

func TestFixtureRepository_ListBySeason(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	fixtureRepo := NewFixtureRepository(db)
	teamRepo := NewTeamRepository(db)

	home := &database.Team{Name: "Botafogo", Country: "Brazil"}
	away := &database.Team{Name: "Fluminense", Country: "Brazil"}
	require.NoError(t, teamRepo.Upsert(home))
	require.NoError(t, teamRepo.Upsert(away))

	fixtures := []database.Fixture{
		{Season: 2025, Round: 1, HomeTeamID: home.ID, AwayTeamID: away.ID, Date: time.Now()},
	}
	require.NoError(t, fixtureRepo.BulkCreate(fixtures))

	list, err := fixtureRepo.ListBySeason(2025)

	is := assert.New(t)

	is.NoError(err)
	is.Len(list, 1)
}

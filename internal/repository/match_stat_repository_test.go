package repository

import (
	"testing"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchStatRepository_Create(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	teamRepo := NewTeamRepository(db)
	matchRepo := NewMatchRepository(db)
	statRepo := NewMatchStatRepository(db)

	home := &database.Team{Name: "Flamengo", Country: "Brazil"}
	away := &database.Team{Name: "Palmeiras", Country: "Brazil"}
	require.NoError(t, teamRepo.Upsert(home))
	require.NoError(t, teamRepo.Upsert(away))

	match := &database.Match{
		Season:     2025,
		HomeTeamID: home.ID,
		AwayTeamID: away.ID,
		Status:     "completed",
	}
	require.NoError(t, matchRepo.Create(match))

	shots := 10
	stat := &database.MatchStat{
		MatchID:   match.ID,
		HomeShots: &shots,
	}

	err := statRepo.Create(stat)

	is := assert.New(t)

	is.NoError(err)
	is.NotZero(stat.ID)
}

func TestMatchStatRepository_FindByMatchID(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	teamRepo := NewTeamRepository(db)
	matchRepo := NewMatchRepository(db)
	statRepo := NewMatchStatRepository(db)

	home := &database.Team{Name: "Santos", Country: "Brazil"}
	away := &database.Team{Name: "São Paulo", Country: "Brazil"}
	require.NoError(t, teamRepo.Upsert(home))
	require.NoError(t, teamRepo.Upsert(away))

	match := &database.Match{
		Season:     2025,
		HomeTeamID: home.ID,
		AwayTeamID: away.ID,
		Status:     "completed",
	}
	require.NoError(t, matchRepo.Create(match))

	s := 5
	stat := &database.MatchStat{MatchID: match.ID, HomeShots: &s}
	require.NoError(t, statRepo.Create(stat))

	found, err := statRepo.FindByMatchID(match.ID)

	is := assert.New(t)

	is.NoError(err)
	is.NotNil(found)
	is.Equal(match.ID, found.MatchID)
}

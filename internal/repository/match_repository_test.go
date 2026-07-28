package repository

import (
	"testing"
	"time"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchRepository_Create(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	matchRepo := NewMatchRepository(db)
	teamRepo := NewTeamRepository(db)

	home := &database.Team{Name: "Flamengo", Country: "Brazil"}
	away := &database.Team{Name: "Palmeiras", Country: "Brazil"}
	require.NoError(t, teamRepo.Upsert(home))
	require.NoError(t, teamRepo.Upsert(away))

	homeGoals := 2
	awayGoals := 1

	match := &database.Match{
		Season:     2025,
		Round:      1,
		Date:       time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
		HomeTeamID: home.ID,
		AwayTeamID: away.ID,
		HomeGoals:  &homeGoals,
		AwayGoals:  &awayGoals,
		Status:     "completed",
	}

	err := matchRepo.Create(match)

	is := assert.New(t)

	is.NoError(err)
	is.NotZero(match.ID)
}

func TestMatchRepository_ListBySeason(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	matchRepo := NewMatchRepository(db)
	teamRepo := NewTeamRepository(db)

	home := &database.Team{Name: "Santos", Country: "Brazil"}
	away := &database.Team{Name: "São Paulo", Country: "Brazil"}
	require.NoError(t, teamRepo.Upsert(home))
	require.NoError(t, teamRepo.Upsert(away))

	for i := 1; i <= 3; i++ {
		match := &database.Match{
			Season:     2025,
			Round:      i,
			HomeTeamID: home.ID,
			AwayTeamID: away.ID,
			Status:     "completed",
		}
		require.NoError(t, matchRepo.Create(match))
	}

	matches, err := matchRepo.ListBySeason(2025)

	is := assert.New(t)

	is.NoError(err)
	is.Len(matches, 3)
}

func TestMatchRepository_ListByTeam(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	matchRepo := NewMatchRepository(db)
	teamRepo := NewTeamRepository(db)

	team := &database.Team{Name: "Grêmio", Country: "Brazil"}
	opponent := &database.Team{Name: "Internacional", Country: "Brazil"}
	require.NoError(t, teamRepo.Upsert(team))
	require.NoError(t, teamRepo.Upsert(opponent))

	for i := 1; i <= 3; i++ {
		match := &database.Match{
			Season:     2025,
			HomeTeamID: team.ID,
			AwayTeamID: opponent.ID,
			Status:     "completed",
			Date:       time.Date(2025, time.Month(i), 1, 0, 0, 0, 0, time.UTC),
		}
		require.NoError(t, matchRepo.Create(match))
	}

	matches, err := matchRepo.ListByTeam(team.ID)

	is := assert.New(t)

	is.NoError(err)
	is.NotEmpty(matches)
}

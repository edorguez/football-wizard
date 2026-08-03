package repository

import (
	"testing"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchRepository_ListByReferee(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	teamRepo := NewTeamRepository(db)
	refRepo := NewRefereeRepository(db)
	matchRepo := NewMatchRepository(db)

	home := &database.Team{Name: "Santos", Country: "Brazil"}
	away := &database.Team{Name: "São Paulo", Country: "Brazil"}
	ref := &database.Referee{Name: "Raphael Claus"}
	require.NoError(t, teamRepo.Upsert(home))
	require.NoError(t, teamRepo.Upsert(away))
	require.NoError(t, refRepo.Upsert(ref))

	for i := 1; i <= 3; i++ {
		match := &database.Match{
			Season:     2025,
			Round:      i,
			HomeTeamID: home.ID,
			AwayTeamID: away.ID,
			RefereeID:  &ref.ID,
			Status:     "completed",
		}
		require.NoError(t, matchRepo.Upsert(match))
	}

	matches, err := matchRepo.ListByReferee(ref.ID)
	require.NoError(t, err)

	is := assert.New(t)
	is.Len(matches, 3)
	is.Equal("Santos", matches[0].HomeTeam.Name)
}

func TestMatchRepository_ListRows(t *testing.T) {
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
		Round:      1,
		HomeTeamID: home.ID,
		AwayTeamID: away.ID,
		Status:     "completed",
	}
	require.NoError(t, matchRepo.Upsert(match))

	shots, sot, saves := 14, 6, 2
	firstHalf, secondHalf := 1, 1
	firstMinute, secondMinute := 20, 66
	stat := &database.MatchStat{
		MatchID:              match.ID,
		HomeShots:            &shots,
		HomeShotsOnTarget:    &sot,
		HomeSaves:            &saves,
		HomeGoalsFirstHalf:   &firstHalf,
		HomeGoalsSecondHalf:  &secondHalf,
		HomeFirstGoalMinute:  &firstMinute,
		HomeSecondGoalMinute: &secondMinute,
	}
	require.NoError(t, statRepo.Upsert(stat))

	rows, err := matchRepo.ListRows()
	require.NoError(t, err)
	require.Len(t, rows, 1)

	row := rows[0]
	is := assert.New(t)
	is.Equal(match.ID, row.ID)
	is.Equal(uint(home.ID), row.HomeTeamID)
	is.Equal(uint(away.ID), row.AwayTeamID)
	require.NotNil(t, row.HomeShots)
	is.Equal(14, *row.HomeShots)
	require.NotNil(t, row.HomeShotsOnTarget)
	is.Equal(6, *row.HomeShotsOnTarget)
	require.NotNil(t, row.HomeSaves)
	is.Equal(2, *row.HomeSaves)
	require.NotNil(t, row.HomeGoalsFirstHalf)
	is.Equal(1, *row.HomeGoalsFirstHalf)
	require.NotNil(t, row.HomeGoalsSecondHalf)
	is.Equal(1, *row.HomeGoalsSecondHalf)
	require.NotNil(t, row.HomeFirstGoalMinute)
	is.Equal(20, *row.HomeFirstGoalMinute)
	require.NotNil(t, row.HomeSecondGoalMinute)
	is.Equal(66, *row.HomeSecondGoalMinute)
}

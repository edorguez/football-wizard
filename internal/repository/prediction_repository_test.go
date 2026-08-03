package repository

import (
	"testing"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPredictionRepository_CreateAndListRecent(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	teamRepo := NewTeamRepository(db)
	repo := NewPredictionRepository(db)

	home := &database.Team{Name: "Flamengo", Country: "Brazil"}
	away := &database.Team{Name: "Palmeiras", Country: "Brazil"}
	require.NoError(t, teamRepo.Upsert(home))
	require.NoError(t, teamRepo.Upsert(away))

	for i := 0; i < 5; i++ {
		prediction := &database.Prediction{
			HomeTeamID: home.ID,
			AwayTeamID: away.ID,
			HomeWin:    0.5,
			Draw:       0.3,
			AwayWin:    0.2,
			Payload:    `[{"market":"btts","outcome":"Yes","probability":0.4}]`,
		}
		require.NoError(t, repo.Create(prediction))
	}

	recent, err := repo.ListRecent(3)
	require.NoError(t, err)

	is := assert.New(t)
	is.Len(recent, 3)
	is.Equal("Flamengo", recent[0].HomeTeam.Name)
	is.Equal("Palmeiras", recent[0].AwayTeam.Name)
}

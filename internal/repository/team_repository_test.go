package repository

import (
	"testing"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupRepoDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(&database.Team{}, &database.Referee{}, &database.Match{}, &database.MatchStat{}, &database.Fixture{}))

	return db
}

func TestTeamRepository_Upsert(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	repo := NewTeamRepository(db)

	team := &database.Team{
		Name:      "Flamengo",
		ShortName: "FLA",
		Country:   "Brazil",
	}

	err := repo.Upsert(team)

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	is.NotZero(team.ID)
}

func TestTeamRepository_Upsert_Duplicate(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	repo := NewTeamRepository(db)

	team := &database.Team{Name: "Flamengo", Country: "Brazil"}
	require.NoError(t, repo.Upsert(team))

	team.ID = 0
	err := repo.Upsert(team)

	assert.NoError(t, err)
}

func TestTeamRepository_FindByName(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	repo := NewTeamRepository(db)

	require.NoError(t, repo.Upsert(&database.Team{Name: "Corinthians", Country: "Brazil"}))

	found, err := repo.FindByName("Corinthians")

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	must.NotNil(found)
	is.Equal("Corinthians", found.Name)
}

func TestTeamRepository_FindByName_NotFound(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	repo := NewTeamRepository(db)

	_, err := repo.FindByName("NonExistent")

	assert.Error(t, err)
}

func TestTeamRepository_ListAll(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	repo := NewTeamRepository(db)

	teams := []string{"Flamengo", "Palmeiras", "Corinthians"}
	for _, name := range teams {
		require.NoError(t, repo.Upsert(&database.Team{Name: name, Country: "Brazil"}))
	}

	all, err := repo.ListAll()

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	is.Len(all, 3)
}

func TestTeamRepository_FindByID(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	repo := NewTeamRepository(db)

	team := &database.Team{Name: "São Paulo", Country: "Brazil"}
	require.NoError(t, repo.Upsert(team))

	found, err := repo.FindByID(team.ID)

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	must.NotNil(found)
	is.Equal(team.Name, found.Name)
}

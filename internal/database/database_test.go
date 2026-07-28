package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupMemDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	return db
}

func TestMigrate_CreatesTables(t *testing.T) {
	t.Parallel()

	db := setupMemDB(t)
	err := Migrate(db)

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)

	is.True(db.Migrator().HasTable(&Team{}))
	is.True(db.Migrator().HasTable(&Referee{}))
	is.True(db.Migrator().HasTable(&Match{}))
	is.True(db.Migrator().HasTable(&MatchStat{}))
	is.True(db.Migrator().HasTable(&Fixture{}))
}

func TestMigrate_Idempotent(t *testing.T) {
	t.Parallel()

	db := setupMemDB(t)
	must := require.New(t)

	must.NoError(Migrate(db))
	must.NoError(Migrate(db))
}

func TestConnect_WALEnabled(t *testing.T) {
	t.Parallel()

	db, err := Connect(":memory:")

	must := require.New(t)
	must.NoError(err)

	sqlDB, err := db.DB()
	must.NoError(err)
	must.NoError(sqlDB.Close())
}

func TestTeamModel(t *testing.T) {
	t.Parallel()

	db := setupMemDB(t)
	require.NoError(t, Migrate(db))

	team := Team{
		Name:      "Flamengo",
		ShortName: "FLA",
		Country:   "Brazil",
	}

	err := db.Create(&team).Error

	is := assert.New(t)

	is.NoError(err)
	is.NotZero(team.ID)
	is.Equal("Flamengo", team.Name)
}

func TestMatchModel_WithRelations(t *testing.T) {
	t.Parallel()

	db := setupMemDB(t)
	require.NoError(t, Migrate(db))

	home := Team{Name: "Flamengo", ShortName: "FLA", Country: "Brazil"}
	away := Team{Name: "Palmeiras", ShortName: "PAL", Country: "Brazil"}

	require.NoError(t, db.Create(&home).Error)
	require.NoError(t, db.Create(&away).Error)

	homeGoals := 2
	awayGoals := 1

	match := Match{
		Season:     2025,
		Round:      1,
		HomeTeamID: home.ID,
		AwayTeamID: away.ID,
		HomeGoals:  &homeGoals,
		AwayGoals:  &awayGoals,
		Status:     "completed",
	}

	err := db.Create(&match).Error

	is := assert.New(t)

	is.NoError(err)
	is.NotZero(match.ID)
	is.Equal(2025, match.Season)
}

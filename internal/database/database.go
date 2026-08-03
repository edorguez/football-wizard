package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var dbLogger = logger.New(
	log.New(os.Stderr, "\r", 0),
	logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Error,
		IgnoreRecordNotFoundError: true,
	},
)

func Connect(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(1)

	if _, err := sqlDB.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("enabling WAL: %w", err)
	}
	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}
	if _, err := sqlDB.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		return nil, fmt.Errorf("setting busy timeout: %w", err)
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Team{},
		&Referee{},
		&Match{},
		&MatchStat{},
		&Player{},
		&TeamSquadMember{},
		&MatchLineup{},
		&MatchPlayerStat{},
		&MatchSubstitution{},
		&Fixture{},
		&Prediction{},
	)
}

package database

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB, log *zap.Logger) error {
	models := []interface{}{
		&Team{},
		&Player{},
		&Match{},
		&MatchStat{},
		&Lineup{},
		&Referee{},
		&RefereeStat{},
		&Fixture{},
		&Prediction{},
		&ModelFeature{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("migrating %T: %w", m, err)
		}
	}

	log.Info("database migrations completed")
	return nil
}

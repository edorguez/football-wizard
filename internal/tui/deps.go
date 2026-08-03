package tui

import (
	"log/slog"

	"github.com/edorguez/football-wizard/internal/config"
	"github.com/edorguez/football-wizard/internal/model"
	"github.com/edorguez/football-wizard/internal/repository"
	"github.com/edorguez/football-wizard/internal/scheduler"
	"github.com/edorguez/football-wizard/internal/scraper"
)

// Deps carries every dependency the TUI views need. It is wired once in
// main.go and injected into each view constructor.
type Deps struct {
	Cfg         *config.Config
	Log         *slog.Logger
	Teams       *repository.TeamRepository
	Matches     *repository.MatchRepository
	Refs        *repository.RefereeRepository
	Predictions *repository.PredictionRepository

	// NewScraper builds a scraper that writes logs to the given logger. Views
	// use a per-run ring-buffer logger so scrape output streams into a view.
	NewScraper func(log *slog.Logger) *scraper.Scraper

	// NewTrainer returns a trainer configured from config.yaml.
	NewTrainer func() *model.Trainer

	// Train builds a predictor by training on the current database contents.
	Train func() (*model.Predictor, error)

	// Scheduler is the shared cron scheduler controlled from the TUI.
	Scheduler *scheduler.Scheduler
}

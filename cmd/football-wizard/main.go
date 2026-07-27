package main

import (
	"fmt"
	"os"

	"github.com/edorguez/football-wizard/internal/config"
	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/logger"
	"github.com/edorguez/football-wizard/internal/model"
	"github.com/edorguez/football-wizard/internal/predictor"
	"github.com/edorguez/football-wizard/internal/repository"
	"github.com/edorguez/football-wizard/internal/scheduler"
	"github.com/edorguez/football-wizard/internal/scraper"
	"github.com/edorguez/football-wizard/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	log, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}
	defer log.Sync()

	db, err := database.Connect(cfg.Database.Path, log)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	err = database.AutoMigrate(db, log)
	if err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	teamRepo := repository.NewTeamRepo(db)
	matchRepo := repository.NewMatchRepo(db)
	fixtureRepo := repository.NewFixtureRepo(db)
	predictionRepo := repository.NewPredictionRepo(db)
	refereeRepo := repository.NewRefereeRepo(db)

	headlessxURL := cfg.Scraper.HeadlessXURL
	if envURL := os.Getenv("HEADLESSX_URL"); envURL != "" {
		headlessxURL = envURL
	}
	headlessxKey := cfg.Scraper.APIKey
	if envKey := os.Getenv("HEADLESSX_API_KEY"); envKey != "" {
		headlessxKey = envKey
	}
	scraperClient := scraper.NewClient(headlessxURL, headlessxKey)
	scraperParser := scraper.NewParser(scraperClient, log)
	scraperSaver := scraper.NewSaver(teamRepo, matchRepo, refereeRepo)

	featureEngine := model.NewFeatureEngine(db, matchRepo, teamRepo)

	trainer := model.NewTrainer(featureEngine, matchRepo, teamRepo, log)

	predictorEngine := predictor.NewEngine(trainer, featureEngine, fixtureRepo, predictionRepo, teamRepo, log)

	sched := scheduler.NewScheduler(
		scraperClient,
		teamRepo,
		matchRepo,
		fixtureRepo,
		predictorEngine,
		log,
		cfg.Scheduler.ScrapeTime,
		cfg.Scheduler.TrainDay,
		cfg.Scheduler.TrainTime,
	)

	if len(os.Args) > 1 && os.Args[1] == "daemon" {
		return tui.RunDaemon(sched, log)
	}

	appCtx := &tui.AppContext{
		DB:          db,
		Config:      cfg,
		Log:         log,
		TeamRepo:    teamRepo,
		MatchRepo:   matchRepo,
		Fixtures:    fixtureRepo,
		Predicts:    predictionRepo,
		Scraper:     scraperClient,
		Parser:      scraperParser,
		ScrapeSaver: scraperSaver,
		Features:    featureEngine,
		Trainer:     trainer,
		Predictor:   predictorEngine,
		Scheduler:   sched,
	}

	program := tea.NewProgram(tui.NewRootModel(appCtx), tea.WithAltScreen())
	_, err = program.Run()
	if err != nil {
		return fmt.Errorf("running tui: %w", err)
	}

	return nil
}

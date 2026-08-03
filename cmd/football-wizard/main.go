package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/edorguez/football-wizard/internal/config"
	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/logger"
	"github.com/edorguez/football-wizard/internal/model"
	"github.com/edorguez/football-wizard/internal/repository"
	"github.com/edorguez/football-wizard/internal/scheduler"
	"github.com/edorguez/football-wizard/internal/scraper"
	"github.com/edorguez/football-wizard/internal/tui"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	season := flag.Int("season", 2025, "season year to scrape")
	full := flag.Bool("full", false, "scrape squads and match reports too")
	workers := flag.Int("workers", 2, "concurrent scrape workers (max 7)")
	delay := flag.Int("delay", 2, "rate limit delay in seconds between requests per worker")
	train := flag.Bool("train", false, "train models and report accuracy on held-out matches")
	predictHome := flag.String("predict-home", "", "predict a fixture: home team name")
	predictAway := flag.String("predict-away", "", "predict a fixture: away team name")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
	})

	log.Info("starting football-wizard", "season", *season, "full", *full)

	db, err := database.Connect(cfg.Database.Path)
	if err != nil {
		log.Error("connecting to database", "error", err)
		os.Exit(1)
	}

	if err := database.Migrate(db); err != nil {
		log.Error("running migrations", "error", err)
		os.Exit(1)
	}

	teamsRepo := repository.NewTeamRepository(db)
	refsRepo := repository.NewRefereeRepository(db)
	matchesRepo := repository.NewMatchRepository(db)
	playersRepo := repository.NewPlayerRepository(db)
	matchStatRepo := repository.NewMatchStatRepository(db)
	lineupRepo := repository.NewLineupRepository(db)
	fixturesRepo := repository.NewFixtureRepository(db)
	predictionRepo := repository.NewPredictionRepository(db)

	newScraper := func(log *slog.Logger) *scraper.Scraper {
		client := scraper.NewClient(cfg.HeadlessX.APIURL, cfg.HeadlessX.APIKey, log)
		cache := scraper.NewCache(log)
		pool := scraper.NewWorkerPool(client, cache, log, *workers, time.Duration(*delay)*time.Second, 0.5)
		saver := scraper.NewSaver(teamsRepo, refsRepo, matchesRepo, playersRepo, matchStatRepo, lineupRepo, fixturesRepo, log)
		return scraper.NewScraper(client, cache, saver, pool, matchesRepo, log)
	}

	newTrainer := func() *model.Trainer {
		return buildTrainer(cfg)
	}

	trainAll := func() (*model.Predictor, error) {
		rows, err := matchesRepo.ListRows()
		if err != nil {
			return nil, fmt.Errorf("loading matches: %w", err)
		}
		if len(rows) == 0 {
			return nil, fmt.Errorf("no completed matches in the database")
		}
		return newTrainer().Train(rows)
	}

	sched, err := buildScheduler(cfg, log, *season, newScraper, trainAll)
	if err != nil {
		log.Error("building scheduler", "error", err)
		os.Exit(1)
	}

	args := flag.Args()
	action := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { action[f.Name] = true })

	hasAction := action["season"] || action["full"] || action["workers"] ||
		action["delay"] || action["train"] || action["predict-home"] || action["predict-away"]

	switch {
	case len(args) > 0 && args[0] == "daemon":
		if err := runDaemon(log, sched); err != nil {
			log.Error("daemon failed", "error", err)
			os.Exit(1)
		}

	case len(args) > 0 && args[0] == "ui":
		if err := runTUI(cfg, log, teamsRepo, matchesRepo, refsRepo, predictionRepo, newScraper, newTrainer, trainAll, sched); err != nil {
			log.Error("TUI failed", "error", err)
			os.Exit(1)
		}

	case len(args) == 0 && !hasAction:
		if err := runTUI(cfg, log, teamsRepo, matchesRepo, refsRepo, predictionRepo, newScraper, newTrainer, trainAll, sched); err != nil {
			log.Error("TUI failed", "error", err)
			os.Exit(1)
		}

	case *train:
		if err := runTrain(matchesRepo, cfg); err != nil {
			log.Error("training models", "error", err)
			os.Exit(1)
		}
		logger.Success(log, "training done")

	case *predictHome != "" || *predictAway != "":
		if *predictHome == "" || *predictAway == "" {
			log.Error("both -predict-home and -predict-away are required together")
			os.Exit(1)
		}
		if err := runPredict(matchesRepo, teamsRepo, cfg, *predictHome, *predictAway); err != nil {
			log.Error("predicting fixture", "error", err)
			os.Exit(1)
		}
		logger.Success(log, "prediction done")

	default:
		if cfg.HeadlessX.APIKey == "" {
			log.Warn("headlessx.api_key is empty; scraping will fail")
		}
		sc := newScraper(log)
		if *full {
			if err := sc.ScrapeSeasonFull(*season); err != nil {
				log.Error("full scrape failed", "season", *season, "error", err)
				os.Exit(1)
			}
		} else if err := sc.ScrapeSeason(*season); err != nil {
			log.Error("scraping season", "season", *season, "error", err)
			os.Exit(1)
		}
		logger.Success(log, "done")
	}
}

func runTUI(
	cfg *config.Config,
	log *slog.Logger,
	teamsRepo *repository.TeamRepository,
	matchesRepo *repository.MatchRepository,
	refsRepo *repository.RefereeRepository,
	predictionRepo *repository.PredictionRepository,
	newScraper func(log *slog.Logger) *scraper.Scraper,
	newTrainer func() *model.Trainer,
	train func() (*model.Predictor, error),
	sched *scheduler.Scheduler,
) error {
	deps := tui.Deps{
		Cfg:         cfg,
		Log:         log,
		Teams:       teamsRepo,
		Matches:     matchesRepo,
		Refs:        refsRepo,
		Predictions: predictionRepo,
		NewScraper:  newScraper,
		NewTrainer:  newTrainer,
		Train:       train,
		Scheduler:   sched,
	}
	return tui.Run(deps)
}

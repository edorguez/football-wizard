package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/edorguez/football-wizard/internal/config"
	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/logger"
	"github.com/edorguez/football-wizard/internal/repository"
	"github.com/edorguez/football-wizard/internal/scraper"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	season := flag.Int("season", 2025, "season year to scrape")
	full := flag.Bool("full", false, "scrape squads and match reports too")
	workers := flag.Int("workers", 2, "concurrent scrape workers (max 7)")
	delay := flag.Int("delay", 2, "rate limit delay in seconds between requests per worker")
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

	log.Info("database migrated", "path", cfg.Database.Path)

	teamsRepo := repository.NewTeamRepository(db)
	refsRepo := repository.NewRefereeRepository(db)
	matchesRepo := repository.NewMatchRepository(db)
	playersRepo := repository.NewPlayerRepository(db)
	matchStatRepo := repository.NewMatchStatRepository(db)
	lineupRepo := repository.NewLineupRepository(db)
	fixturesRepo := repository.NewFixtureRepository(db)

	if cfg.HeadlessX.APIKey == "" {
		log.Error("headlessx.api_key is required")
		os.Exit(1)
	}

	client := scraper.NewClient(cfg.HeadlessX.APIURL, cfg.HeadlessX.APIKey, log)
	defer client.Close()

	if *workers > 7 {
		log.Warn("worker count capped to 7", "requested", *workers)
		*workers = 7
	}

	cache := scraper.NewCache(log)

	pool := scraper.NewWorkerPool(client, cache, log, *workers, time.Duration(*delay)*time.Second, 0.5)

	saver := scraper.NewSaver(teamsRepo, refsRepo, matchesRepo, playersRepo, matchStatRepo, lineupRepo, fixturesRepo, log)

	sc := scraper.NewScraper(client, cache, saver, pool, matchesRepo, log)

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

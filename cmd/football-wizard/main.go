package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/edorguez/football-wizard/internal/config"
	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/logger"
	"github.com/edorguez/football-wizard/internal/repository"
	"github.com/edorguez/football-wizard/internal/scraper"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	season := flag.Int("season", 2025, "season year to scrape")
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

	log.Info("starting football-wizard", "season", *season)

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
	matchStatsRepo := repository.NewMatchStatRepository(db)
	fixturesRepo := repository.NewFixtureRepository(db)

	client, err := scraper.NewClient(log)
	if err != nil {
		log.Error("creating browser client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Error("closing browser client", "error", err)
		}
	}()

	saver := scraper.NewSaver(teamsRepo, refsRepo, matchesRepo, matchStatsRepo, fixturesRepo, log)

	sc := scraper.NewScraper(client, saver, log)

	if err := sc.ScrapeSeason(*season); err != nil {
		log.Error("scraping season", "season", *season, "error", err)
		os.Exit(1)
	}

	log.Info("done")
}

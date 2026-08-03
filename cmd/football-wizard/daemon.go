package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/edorguez/football-wizard/internal/config"
	"github.com/edorguez/football-wizard/internal/model"
	"github.com/edorguez/football-wizard/internal/scheduler"
	"github.com/edorguez/football-wizard/internal/scraper"
)

// buildScheduler wires the daily scrape and weekly retrain jobs from config.
func buildScheduler(
	cfg *config.Config,
	log *slog.Logger,
	season int,
	newScraper func(log *slog.Logger) *scraper.Scraper,
	train func() (*model.Predictor, error),
) (*scheduler.Scheduler, error) {
	scrapeSpec, err := scheduler.DailySpec(cfg.Scheduler.ScrapeTime)
	if err != nil {
		return nil, fmt.Errorf("building scrape schedule: %w", err)
	}
	retrainSpec, err := scheduler.WeeklySpec(cfg.Scheduler.TrainDay, cfg.Scheduler.TrainTime)
	if err != nil {
		return nil, fmt.Errorf("building retrain schedule: %w", err)
	}

	jobs := []scheduler.Job{
		{
			Label: "scrape",
			Spec:  scrapeSpec,
			Run: func() error {
				return newScraper(log).ScrapeSeason(season)
			},
		},
		{
			Label: "retrain",
			Spec:  retrainSpec,
			Run: func() error {
				_, err := train()
				return err
			},
		},
	}

	return scheduler.New(jobs, log)
}

// runDaemon runs the scheduler headless until interrupted.
func runDaemon(log *slog.Logger, sched *scheduler.Scheduler) error {
	log.Info("starting daemon")

	if err := sched.Start(); err != nil {
		return fmt.Errorf("starting scheduler: %w", err)
	}
	defer sched.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	sig := <-sigCh
	log.Info("shutting down", "signal", sig.String())
	return nil
}

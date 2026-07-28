package scraper

import (
	"fmt"
	"log/slog"
)

type Scraper struct {
	client *Client
	saver  *Saver
	logger *slog.Logger
}

func NewScraper(client *Client, saver *Saver, logger *slog.Logger) *Scraper {
	return &Scraper{
		client: client,
		saver:  saver,
		logger: logger,
	}
}

func (s *Scraper) ScrapeSeason(season int) error {
	s.logger.Info("starting scrape", "season", season)

	url := fmt.Sprintf("https://fbref.com/en/comps/24/%d/schedule/%d-Serie-A-schedule", season, season)

	html, err := s.client.FetchHTML(url)
	if err != nil {
		return fmt.Errorf("fetching season %d: %w", season, err)
	}

	matches, err := ParseMatchResults(season, html)
	if err != nil {
		return fmt.Errorf("parsing season %d: %w", season, err)
	}

	s.logger.Info("parsed matches", "season", season, "count", len(matches))

	if err := s.saver.SaveMatches(matches); err != nil {
		return fmt.Errorf("saving season %d: %w", season, err)
	}

	s.logger.Info("season scraped successfully", "season", season, "matches", len(matches))

	return nil
}

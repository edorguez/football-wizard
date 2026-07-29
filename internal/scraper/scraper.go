package scraper

import (
	"fmt"
	"log/slog"

	"github.com/edorguez/football-wizard/internal/logger"
)

type Scraper struct {
	client   *Client
	cache    *Cache
	saver    *Saver
	pool     *WorkerPool
	logger   *slog.Logger
}

func NewScraper(client *Client, cache *Cache, saver *Saver, pool *WorkerPool, logger *slog.Logger) *Scraper {
	return &Scraper{
		client: client,
		cache:  cache,
		saver:  saver,
		pool:   pool,
		logger: logger,
	}
}

func (s *Scraper) ScrapeSeason(season int) error {
	s.logger.Info("starting scrape", "season", season)

	html, err := s.fetchSchedule(season)
	if err != nil {
		return fmt.Errorf("fetching schedule: %w", err)
	}

	matches, err := ParseMatchResults(season, html)
	if err != nil {
		return fmt.Errorf("parsing schedule: %w", err)
	}

	s.logger.Info("parsed matches", "season", season, "count", len(matches))

	if err := s.saver.SaveMatches(matches); err != nil {
		return fmt.Errorf("saving matches: %w", err)
	}

	logger.Success(s.logger, "season scraped successfully", "season", season, "matches", len(matches))

	return nil
}

func (s *Scraper) ScrapeSeasonFull(season int) error {
	if err := s.ScrapeSeason(season); err != nil {
		return err
	}

	if err := s.ScrapeSquads(season); err != nil {
		s.logger.Error("squad scraping had errors", "season", season, "error", err)
	}

	if err := s.ScrapeMatchReports(season); err != nil {
		s.logger.Error("match report scraping had errors", "season", season, "error", err)
	}

	return nil
}

func (s *Scraper) fetchSchedule(season int) (string, error) {
	if cached, ok := s.cache.Read(season, "schedule.html"); ok {
		s.logger.Info("using cached schedule", "season", season)
		return cached, nil
	}

	url := fmt.Sprintf(
		"https://fbref.com/en/comps/24/%d/schedule/%d-Serie-A-Scores-and-Fixtures",
		season, season,
	)

	html, err := s.client.FetchHTMLWithJS(url)
	if err != nil {
		return "", fmt.Errorf("fetching season %d: %w", season, err)
	}

	s.cache.Write(season, html, "schedule.html")

	return html, nil
}

func (s *Scraper) ScrapeSquads(season int) error {
	html, err := s.fetchSchedule(season)
	if err != nil {
		return fmt.Errorf("fetching schedule for squads: %w", err)
	}

	teamURLs, err := ParseSquadURLs(season, html)
	if err != nil {
		return fmt.Errorf("parsing team URLs: %w", err)
	}

	s.logger.Info("found teams for squad scraping", "count", len(teamURLs))

	var jobs []FetchJob
	for teamName, url := range teamURLs {
		name := teamName
		jobs = append(jobs, FetchJob{
			Label:      fmt.Sprintf("squad-%s", name),
			URL:        url,
			Season:     season,
			RequiresJS: false,
			CacheParts: []string{"squads", fmt.Sprintf("%s.html", name)},
			ParseFn: func(_ int, html string) error {
				return s.parseSquadAndSave(name, html)
			},
		})
	}

	results := s.pool.Run(jobs)

	var errCount int
	for _, r := range results {
		if r.Err != nil {
			s.logger.Error("squad scrape failed", "label", r.Label, "error", r.Err)
			errCount++
		}
	}

	logger.Success(s.logger, "squad scraping complete", "total", len(jobs), "errors", errCount)

	if errCount > 0 {
		return fmt.Errorf("%d squad(s) failed to scrape", errCount)
	}

	return nil
}

func (s *Scraper) parseSquadAndSave(teamName string, html string) error {
	squad := ScrapedSquad{TeamName: teamName}

	players, err := ParseSquadPlayers(0, html)
	if err != nil {
		return err
	}
	squad.Players = players

	return s.saver.SaveSquad(squad)
}

func (s *Scraper) ScrapeMatchReports(season int) error {
	html, err := s.fetchSchedule(season)
	if err != nil {
		return fmt.Errorf("fetching schedule for match reports: %w", err)
	}

	matches, err := ParseMatchResults(season, html)
	if err != nil {
		return fmt.Errorf("parsing schedule: %w", err)
	}

	var jobs []FetchJob
	for _, m := range matches {
		if m.MatchReportURL == "" {
			continue
		}
		jobs = append(jobs, FetchJob{
			Label:      fmt.Sprintf("%s-vs-%s", m.HomeTeam, m.AwayTeam),
			URL:        m.MatchReportURL,
			Season:     season,
			RequiresJS: false,
			CacheParts: []string{"reports", fmt.Sprintf("%s-vs-%s.html", m.HomeTeam, m.AwayTeam)},
			ParseFn:    s.parseAndSaveReport,
		})
	}

	s.logger.Info("match reports to scrape", "count", len(jobs))

	results := s.pool.Run(jobs)

	var errCount int
	for _, r := range results {
		if r.Err != nil {
			s.logger.Error("match report scrape failed", "label", r.Label, "error", r.Err)
			errCount++
		}
	}

	logger.Success(s.logger, "match report scraping complete", "total", len(jobs), "errors", errCount)

	if errCount > 0 {
		return fmt.Errorf("%d match report(s) failed to scrape", errCount)
	}

	return nil
}

func (s *Scraper) parseAndSaveReport(_ int, html string) error {
	_, err := ParseMatchReport(html)
	if err != nil {
		return fmt.Errorf("parsing match report: %w", err)
	}

	return nil
}

package scraper

import (
	"fmt"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/repository"
)

type Saver struct {
	teamRepo    *repository.TeamRepo
	matchRepo   *repository.MatchRepo
	refereeRepo *repository.RefereeRepo
}

func NewSaver(teamRepo *repository.TeamRepo, matchRepo *repository.MatchRepo, refereeRepo *repository.RefereeRepo) *Saver {
	return &Saver{
		teamRepo:    teamRepo,
		matchRepo:   matchRepo,
		refereeRepo: refereeRepo,
	}
}

func (s *Saver) Save(scraped []ScrapedMatch) (int, error) {
	if len(scraped) == 0 {
		return 0, nil
	}

	if err := s.upsertTeams(scraped); err != nil {
		return 0, fmt.Errorf("saving teams: %w", err)
	}

	if err := s.upsertReferees(scraped); err != nil {
		return 0, fmt.Errorf("saving referees: %w", err)
	}

	matches, err := s.buildMatches(scraped)
	if err != nil {
		return 0, fmt.Errorf("building matches: %w", err)
	}

	if err := s.matchRepo.BulkCreate(matches); err != nil {
		return 0, fmt.Errorf("saving matches: %w", err)
	}

	return len(matches), nil
}

func (s *Saver) upsertTeams(scraped []ScrapedMatch) error {
	seen := map[string]bool{}
	var teams []database.Team
	for _, m := range scraped {
		if !seen[m.HomeTeam] {
			seen[m.HomeTeam] = true
			teams = append(teams, database.Team{Name: m.HomeTeam})
		}
		if !seen[m.AwayTeam] {
			seen[m.AwayTeam] = true
			teams = append(teams, database.Team{Name: m.AwayTeam})
		}
	}
	return s.teamRepo.BulkUpsert(teams)
}

func (s *Saver) upsertReferees(scraped []ScrapedMatch) error {
	seen := map[string]bool{}
	var refs []database.Referee
	for _, m := range scraped {
		if m.Referee == "" || seen[m.Referee] {
			continue
		}
		seen[m.Referee] = true
		refs = append(refs, database.Referee{Name: m.Referee})
	}
	if len(refs) == 0 {
		return nil
	}
	return s.refereeRepo.BulkUpsert(refs)
}

func (s *Saver) buildMatches(scraped []ScrapedMatch) ([]database.Match, error) {
	var matches []database.Match
	for _, sm := range scraped {
		home, err := s.teamRepo.FindByName(sm.HomeTeam)
		if err != nil {
			return nil, fmt.Errorf("looking up home team %q: %w", sm.HomeTeam, err)
		}
		away, err := s.teamRepo.FindByName(sm.AwayTeam)
		if err != nil {
			return nil, fmt.Errorf("looking up away team %q: %w", sm.AwayTeam, err)
		}
		var refereeID int64
		if sm.Referee != "" {
			ref, err := s.refereeRepo.FindByName(sm.Referee)
			if err != nil {
				return nil, fmt.Errorf("looking up referee %q: %w", sm.Referee, err)
			}
			refereeID = ref.ID
		}

		matches = append(matches, database.Match{
			Date:       sm.Date,
			Season:     sm.Season,
			Matchday:   sm.Matchday,
			HomeTeamID: home.ID,
			AwayTeamID: away.ID,
			HomeGoals:  sm.HomeGoals,
			AwayGoals:  sm.AwayGoals,
			RefereeID:  refereeID,
			Stadium:    sm.Stadium,
			Attendance: sm.Attendance,
		})
	}
	return matches, nil
}

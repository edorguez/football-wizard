package scraper

import (
	"fmt"
	"log/slog"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/repository"
)

type Saver struct {
	teamsRepo   *repository.TeamRepository
	refsRepo    *repository.RefereeRepository
	matchesRepo *repository.MatchRepository
	matchStats  *repository.MatchStatRepository
	fixtures    *repository.FixtureRepository
	logger      *slog.Logger
}

func NewSaver(
	teamsRepo *repository.TeamRepository,
	refsRepo *repository.RefereeRepository,
	matchesRepo *repository.MatchRepository,
	matchStats *repository.MatchStatRepository,
	fixtures *repository.FixtureRepository,
	logger *slog.Logger,
) *Saver {
	return &Saver{
		teamsRepo:   teamsRepo,
		refsRepo:    refsRepo,
		matchesRepo: matchesRepo,
		matchStats:  matchStats,
		fixtures:    fixtures,
		logger:      logger,
	}
}

func (s *Saver) SaveMatches(matches []ScrapedMatch) error {
	s.logger.Info("saving matches", "count", len(matches))

	teamCache := map[string]uint{}
	refereeCache := map[string]*uint{}

	for _, sm := range matches {
		teamID, err := s.resolveTeam(sm.HomeTeam, teamCache)
		if err != nil {
			return fmt.Errorf("resolving team %q: %w", sm.HomeTeam, err)
		}
		teamCache[sm.HomeTeam] = teamID

		teamID, err = s.resolveTeam(sm.AwayTeam, teamCache)
		if err != nil {
			return fmt.Errorf("resolving team %q: %w", sm.AwayTeam, err)
		}
		teamCache[sm.AwayTeam] = teamID

		if sm.RefereeName != "" {
			refID, err := s.resolveReferee(sm.RefereeName, refereeCache)
			if err != nil {
				return fmt.Errorf("resolving referee %q: %w", sm.RefereeName, err)
			}
			refereeCache[sm.RefereeName] = refID
		}
	}

	s.logger.Info("teams and referees resolved", "teams", len(teamCache))

	for _, sm := range matches {
		homeGoals := sm.HomeGoals
		awayGoals := sm.AwayGoals

		match := database.Match{
			Season:     sm.Season,
			Round:      sm.Round,
			Date:       sm.Date,
			HomeTeamID: teamCache[sm.HomeTeam],
			AwayTeamID: teamCache[sm.AwayTeam],
			HomeGoals:  &homeGoals,
			AwayGoals:  &awayGoals,
			RefereeID:  refereeCache[sm.RefereeName],
			Status:     "completed",
		}

		if err := s.matchesRepo.Create(&match); err != nil {
			return fmt.Errorf("creating match %s vs %s: %w", sm.HomeTeam, sm.AwayTeam, err)
		}

		if sm.HomeShots != nil || sm.AwayShots != nil {
			stat := database.MatchStat{
				MatchID:           match.ID,
				HomeShots:         sm.HomeShots,
				AwayShots:         sm.AwayShots,
				HomeShotsOnTarget: sm.HomeShotsOnTarget,
				AwayShotsOnTarget: sm.AwayShotsOnTarget,
				HomeCorners:       sm.HomeCorners,
				AwayCorners:       sm.AwayCorners,
				HomeYellowCards:   sm.HomeYellowCards,
				AwayYellowCards:   sm.AwayYellowCards,
				HomeRedCards:      sm.HomeRedCards,
				AwayRedCards:      sm.AwayRedCards,
			}
			if err := s.matchStats.Create(&stat); err != nil {
				return fmt.Errorf("creating match stats for match %d: %w", match.ID, err)
			}
		}
	}

	s.logger.Info("matches saved successfully", "count", len(matches))

	return nil
}

func (s *Saver) SaveFixtures(fixtures []ScrapedFixture) error {
	s.logger.Info("saving fixtures", "count", len(fixtures))

	teamCache := map[string]uint{}

	dbFixtures := make([]database.Fixture, 0, len(fixtures))

	for _, sf := range fixtures {
		teamID, err := s.resolveTeam(sf.HomeTeam, teamCache)
		if err != nil {
			return fmt.Errorf("resolving team %q: %w", sf.HomeTeam, err)
		}
		teamCache[sf.HomeTeam] = teamID

		teamID, err = s.resolveTeam(sf.AwayTeam, teamCache)
		if err != nil {
			return fmt.Errorf("resolving team %q: %w", sf.AwayTeam, err)
		}
		teamCache[sf.AwayTeam] = teamID

		dbFixtures = append(dbFixtures, database.Fixture{
			Season:     sf.Season,
			Round:      sf.Round,
			Date:       sf.Date,
			HomeTeamID: teamCache[sf.HomeTeam],
			AwayTeamID: teamCache[sf.AwayTeam],
			Status:     "scheduled",
		})
	}

	return s.fixtures.BulkCreate(dbFixtures)
}

func (s *Saver) resolveTeam(name string, cache map[string]uint) (uint, error) {
	if id, ok := cache[name]; ok {
		return id, nil
	}

	team := &database.Team{
		Name:      name,
		ShortName: name,
		Country:   "Brazil",
	}

	if err := s.teamsRepo.Upsert(team); err != nil {
		return 0, fmt.Errorf("upserting team %q: %w", name, err)
	}

	s.logger.Debug("team upserted", "name", name, "id", team.ID)

	return team.ID, nil
}

func (s *Saver) resolveReferee(name string, cache map[string]*uint) (*uint, error) {
	if id, ok := cache[name]; ok {
		return id, nil
	}

	referee := &database.Referee{Name: name}

	if err := s.refsRepo.Upsert(referee); err != nil {
		return nil, fmt.Errorf("upserting referee %q: %w", name, err)
	}

	s.logger.Debug("referee upserted", "name", name, "id", referee.ID)

	return &referee.ID, nil
}

package scraper

import (
	"fmt"
	"log/slog"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/logger"
	"github.com/edorguez/football-wizard/internal/repository"
)

type Saver struct {
	teamsRepo   *repository.TeamRepository
	refsRepo    *repository.RefereeRepository
	matchesRepo *repository.MatchRepository
	playersRepo *repository.PlayerRepository
	lineupRepo  *repository.LineupRepository
	fixtures    *repository.FixtureRepository
	logger      *slog.Logger
}

func NewSaver(
	teamsRepo *repository.TeamRepository,
	refsRepo *repository.RefereeRepository,
	matchesRepo *repository.MatchRepository,
	playersRepo *repository.PlayerRepository,
	lineupRepo *repository.LineupRepository,
	fixtures *repository.FixtureRepository,
	logger *slog.Logger,
) *Saver {
	return &Saver{
		teamsRepo:   teamsRepo,
		refsRepo:    refsRepo,
		matchesRepo: matchesRepo,
		playersRepo: playersRepo,
		lineupRepo:  lineupRepo,
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
			Season:        sm.Season,
			Round:         sm.Round,
			Date:          sm.Date,
			HomeTeamID:    teamCache[sm.HomeTeam],
			AwayTeamID:    teamCache[sm.AwayTeam],
			HomeGoals:     &homeGoals,
			AwayGoals:     &awayGoals,
			HomeXG:        sm.HomeXG,
			AwayXG:        sm.AwayXG,
			Venue:         sm.Venue,
			Attendance:    sm.Attendance,
			RefereeID:     refereeCache[sm.RefereeName],
			MatchReportURL: sm.MatchReportURL,
			Status:        "completed",
		}

		if err := s.matchesRepo.Upsert(&match); err != nil {
			return fmt.Errorf("creating match %s vs %s: %w", sm.HomeTeam, sm.AwayTeam, err)
		}
	}

	logger.Success(s.logger, "matches saved successfully", "count", len(matches))

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

func (s *Saver) SaveSquad(squad ScrapedSquad) error {
	s.logger.Info("saving squad", "team", squad.TeamName, "players", len(squad.Players))

	team, err := s.teamsRepo.FindByName(squad.TeamName)
	if err != nil {
		return fmt.Errorf("finding team %q: %w", squad.TeamName, err)
	}

	for _, sp := range squad.Players {
		player := &database.Player{
			Name:        sp.Name,
			Nationality: sp.Nationality,
			Position:    sp.Position,
		}

		if err := s.playersRepo.Upsert(player); err != nil {
			return fmt.Errorf("upserting player %q: %w", sp.Name, err)
		}

		member := &database.TeamSquadMember{
			TeamID:   team.ID,
			PlayerID: player.ID,
			Season:   sp.Season,
			ShirtNum: sp.ShirtNum,
		}

		if err := s.playersRepo.UpsertSquadMember(member); err != nil {
			return fmt.Errorf("upserting squad member %q for %q: %w", sp.Name, squad.TeamName, err)
		}
	}

	s.logger.Info("squad saved", "team", squad.TeamName, "players", len(squad.Players))

	return nil
}

func (s *Saver) SaveMatchReport(report ScrapedMatchReport, matchID uint) error {
	s.logger.Info("saving match report", "match_id", matchID)

	s.logger.Debug("report data not yet implemented", "match_id", matchID, "home", report.HomeTeam, "away", report.AwayTeam)

	return nil
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

func (s *Saver) resolvePlayer(name string) (*database.Player, error) {
	player := &database.Player{Name: name}
	if err := s.playersRepo.Upsert(player); err != nil {
		return nil, fmt.Errorf("upserting player %q: %w", name, err)
	}
	return player, nil
}

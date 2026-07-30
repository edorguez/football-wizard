package scraper

import (
	"fmt"
	"log/slog"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/edorguez/football-wizard/internal/logger"
	"github.com/edorguez/football-wizard/internal/repository"
)

type Saver struct {
	teamsRepo    *repository.TeamRepository
	refsRepo     *repository.RefereeRepository
	matchesRepo  *repository.MatchRepository
	playersRepo  *repository.PlayerRepository
	matchStatRepo *repository.MatchStatRepository
	lineupRepo   *repository.LineupRepository
	fixtures     *repository.FixtureRepository
	logger       *slog.Logger
}

func NewSaver(
	teamsRepo *repository.TeamRepository,
	refsRepo *repository.RefereeRepository,
	matchesRepo *repository.MatchRepository,
	playersRepo *repository.PlayerRepository,
	matchStatRepo *repository.MatchStatRepository,
	lineupRepo *repository.LineupRepository,
	fixtures *repository.FixtureRepository,
	logger *slog.Logger,
) *Saver {
	return &Saver{
		teamsRepo:    teamsRepo,
		refsRepo:     refsRepo,
		matchesRepo:  matchesRepo,
		playersRepo:  playersRepo,
		matchStatRepo: matchStatRepo,
		lineupRepo:   lineupRepo,
		fixtures:     fixtures,
		logger:       logger,
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
			Season:         sm.Season,
			Round:          sm.Round,
			Date:           sm.Date,
			HomeTeamID:     teamCache[sm.HomeTeam],
			AwayTeamID:     teamCache[sm.AwayTeam],
			HomeGoals:      &homeGoals,
			AwayGoals:      &awayGoals,
			HomeXG:         sm.HomeXG,
			AwayXG:         sm.AwayXG,
			Venue:          sm.Venue,
			Attendance:     sm.Attendance,
			RefereeID:      refereeCache[sm.RefereeName],
			MatchReportURL: sm.MatchReportURL,
			Status:         "completed",
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

	match, err := s.matchesRepo.FindByID(matchID)
	if err != nil {
		return fmt.Errorf("finding match %d: %w", matchID, err)
	}

	if err := s.saveMatchStats(report, match.ID); err != nil {
		return fmt.Errorf("saving match stats: %w", err)
	}

	if err := s.saveLineupsAndStats(report, match); err != nil {
		return fmt.Errorf("saving lineups and player stats: %w", err)
	}

	if err := s.aggregateDerivedStats(match.ID); err != nil {
		s.logger.Warn("aggregating derived stats", "match_id", matchID, "error", err)
	}

	logger.Success(s.logger, "match report saved", "match_id", matchID)

	return nil
}

func (s *Saver) saveMatchStats(report ScrapedMatchReport, matchID uint) error {
	stat := &database.MatchStat{
		MatchID:            matchID,
		HomePossession:     report.HomePossession,
		AwayPossession:     report.AwayPossession,
		HomeShots:          report.HomeShots,
		AwayShots:          report.AwayShots,
		HomeShotsOnTarget:  report.HomeShotsOnTarget,
		AwayShotsOnTarget:  report.AwayShotsOnTarget,
		HomeShotsOffTarget: report.HomeShotsOffTarget,
		AwayShotsOffTarget: report.AwayShotsOffTarget,
		HomeSaves:          report.HomeSaves,
		AwaySaves:          report.AwaySaves,
		HomeFouls:          report.HomeFouls,
		AwayFouls:          report.AwayFouls,
		HomeCorners:        report.HomeCorners,
		AwayCorners:        report.AwayCorners,
		HomeCrosses:        report.HomeCrosses,
		AwayCrosses:        report.AwayCrosses,
		HomeOffsides:       report.HomeOffsides,
		AwayOffsides:       report.AwayOffsides,
	}

	return s.matchStatRepo.Upsert(stat)
}

func (s *Saver) saveLineupsAndStats(report ScrapedMatchReport, match *database.Match) error {
	var allLineups []database.MatchLineup
	var allStats []database.MatchPlayerStat
	var allSubs []database.MatchSubstitution

	s.collectTeamData(match.ID, match.HomeTeamID, report.HomeLineup, report.HomePlayerStats, &allLineups, &allStats, true)
	s.collectTeamData(match.ID, match.AwayTeamID, report.AwayLineup, report.AwayPlayerStats, &allLineups, &allStats, false)

	allSubs = s.deriveSubstitutions(match.ID, match.HomeTeamID, match.AwayTeamID, allLineups, allStats)

	if len(allLineups) > 0 {
		if err := s.lineupRepo.UpsertLineups(allLineups); err != nil {
			return fmt.Errorf("saving lineups: %w", err)
		}
	}
	if len(allStats) > 0 {
		if err := s.lineupRepo.UpsertPlayerStats(allStats); err != nil {
			return fmt.Errorf("saving player stats: %w", err)
		}
	}
	if len(allSubs) > 0 {
		if err := s.lineupRepo.UpsertSubstitutions(allSubs); err != nil {
			return fmt.Errorf("saving substitutions: %w", err)
		}
	}

	return nil
}

func (s *Saver) deriveSubstitutions(matchID, homeTeamID, awayTeamID uint, allLineups []database.MatchLineup, allStats []database.MatchPlayerStat) []database.MatchSubstitution {
	statsByPlayer := map[uint]database.MatchPlayerStat{}
	for _, st := range allStats {
		statsByPlayer[st.PlayerID] = st
	}

	lineupsByTeam := map[uint][]database.MatchLineup{}
	for _, lu := range allLineups {
		lineupsByTeam[lu.TeamID] = append(lineupsByTeam[lu.TeamID], lu)
	}

	var subs []database.MatchSubstitution

	for _, teamID := range []uint{homeTeamID, awayTeamID} {
		teamLineups := lineupsByTeam[teamID]
		var subbedOn []uint
		var subbedOff []uint

		for _, lu := range teamLineups {
			if !lu.IsStarter && lu.PlayerID != 0 {
				subbedOn = append(subbedOn, lu.PlayerID)
				continue
			}
			if lu.IsStarter {
				if st, ok := statsByPlayer[lu.PlayerID]; ok && st.MinutesPlayed != nil && *st.MinutesPlayed < 90 {
					subbedOff = append(subbedOff, lu.PlayerID)
				}
			}
		}

		count := len(subbedOn)
		if len(subbedOff) < count {
			count = len(subbedOff)
		}
		for i := range count {
			subs = append(subs, database.MatchSubstitution{
				MatchID:    matchID,
				TeamID:     uint(teamID),
				PlayerOffID: subbedOff[i],
				PlayerOnID:  subbedOn[i],
				Minute:     0,
			})
		}
	}

	return subs
}

func (s *Saver) collectTeamData(matchID, teamID uint, lineups []ScrapedLineupPlayer, stats map[string]ScrapedPlayerMatchStat, allLineups *[]database.MatchLineup, allStats *[]database.MatchPlayerStat, isHome bool) {
	seenPlayers := map[string]bool{}

	for _, lp := range lineups {
		if lp.Name == "" || seenPlayers[lp.Name] {
			continue
		}
		seenPlayers[lp.Name] = true

		if !lp.IsStarter && !lp.HasSubIcon {
			continue
		}

		player, err := s.resolvePlayer(lp.Name)
		if err != nil {
			s.logger.Error("resolving player for lineup", "name", lp.Name, "error", err)
			continue
		}

		*allLineups = append(*allLineups, database.MatchLineup{
			MatchID:   matchID,
			TeamID:    teamID,
			PlayerID:  player.ID,
			IsStarter: lp.IsStarter,
			Position:  lp.Position,
			ShirtNum:  lp.ShirtNum,
		})

		if ps, ok := stats[lp.Name]; ok {
			*allStats = append(*allStats, statToDB(matchID, teamID, player.ID, ps))
		}
	}

	for name, ps := range stats {
		if seenPlayers[name] {
			continue
		}

		player, err := s.resolvePlayer(name)
		if err != nil {
			s.logger.Error("resolving player for stats", "name", name, "error", err)
			continue
		}

		*allStats = append(*allStats, statToDB(matchID, teamID, player.ID, ps))
	}
}

func statToDB(matchID, teamID, playerID uint, ps ScrapedPlayerMatchStat) database.MatchPlayerStat {
	minutes := ps.Minutes
	return database.MatchPlayerStat{
		MatchID:       matchID,
		TeamID:        teamID,
		PlayerID:      playerID,
		MinutesPlayed: &minutes,
		Goals:         ps.Goals,
		Assists:       ps.Assists,
		Shots:         ps.Shots,
		ShotsOnTarget: ps.ShotsOnTarget,
		Passes:        ps.Passes,
		Tackles:       ps.Tackles,
		Interceptions: ps.Interceptions,
		Fouls:         ps.Fouls,
		Fouled:        ps.Fouled,
		Offsides:      ps.Offsides,
		Crosses:       ps.Crosses,
		YellowCards:   ps.YellowCards,
		RedCards:      ps.RedCards,
		Saves:         ps.Saves,
	}
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

func (s *Saver) aggregateDerivedStats(matchID uint) error {
	s.logger.Debug("aggregating derived stats", "match_id", matchID)

	err := s.matchesRepo.DB().Exec(`
		UPDATE match_stats
		SET home_tackles = (SELECT COALESCE(SUM(tackles), 0) FROM match_player_stats WHERE match_id = ? AND team_id = (SELECT home_team_id FROM matches WHERE id = ?)),
		    away_tackles = (SELECT COALESCE(SUM(tackles), 0) FROM match_player_stats WHERE match_id = ? AND team_id = (SELECT away_team_id FROM matches WHERE id = ?)),
		    home_interceptions = (SELECT COALESCE(SUM(interceptions), 0) FROM match_player_stats WHERE match_id = ? AND team_id = (SELECT home_team_id FROM matches WHERE id = ?)),
		    away_interceptions = (SELECT COALESCE(SUM(interceptions), 0) FROM match_player_stats WHERE match_id = ? AND team_id = (SELECT away_team_id FROM matches WHERE id = ?)),
		    home_yellow_cards = (SELECT COALESCE(SUM(yellow_cards), 0) FROM match_player_stats WHERE match_id = ? AND team_id = (SELECT home_team_id FROM matches WHERE id = ?)),
		    away_yellow_cards = (SELECT COALESCE(SUM(yellow_cards), 0) FROM match_player_stats WHERE match_id = ? AND team_id = (SELECT away_team_id FROM matches WHERE id = ?)),
		    home_red_cards = (SELECT COALESCE(SUM(red_cards), 0) FROM match_player_stats WHERE match_id = ? AND team_id = (SELECT home_team_id FROM matches WHERE id = ?)),
		    away_red_cards = (SELECT COALESCE(SUM(red_cards), 0) FROM match_player_stats WHERE match_id = ? AND team_id = (SELECT away_team_id FROM matches WHERE id = ?))
		WHERE match_id = ?
	`, matchID, matchID, matchID, matchID, matchID, matchID, matchID, matchID, matchID, matchID, matchID, matchID, matchID, matchID, matchID, matchID, matchID).Error

	return err
}

func (s *Saver) resolvePlayer(name string) (*database.Player, error) {
	player := &database.Player{Name: name}
	if err := s.playersRepo.Upsert(player); err != nil {
		return nil, fmt.Errorf("upserting player %q: %w", name, err)
	}
	return player, nil
}

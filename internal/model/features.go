package model

import (
	"math"
	"time"
)

// MatchRow is a database-agnostic view of a completed match plus its
// aggregated team stats, used both for training and for prediction.
type MatchRow struct {
	ID         uint
	Season     int
	Round      int
	Date       time.Time
	HomeTeamID uint
	AwayTeamID uint
	HomeGoals  *int
	AwayGoals  *int
	HomeXG     *float64
	AwayXG     *float64

	HomeCorners     *int
	AwayCorners     *int
	HomeOffsides    *int
	AwayOffsides    *int
	HomeYellowCards *int
	AwayYellowCards *int
	HomeRedCards    *int
	AwayRedCards    *int

	HomeShots         *int
	AwayShots         *int
	HomeShotsOnTarget *int
	AwayShotsOnTarget *int
	HomeSaves         *int
	AwaySaves         *int

	HomeGoalsFirstHalf   *int
	AwayGoalsFirstHalf   *int
	HomeGoalsSecondHalf  *int
	AwayGoalsSecondHalf  *int
	HomeFirstGoalMinute  *int
	AwayFirstGoalMinute  *int
	HomeSecondGoalMinute *int
	AwaySecondGoalMinute *int
}

// Features is the engineered vector for a single fixture. Every value is
// computed strictly from matches played before this fixture (causal), so
// the same builder is safe for both training and live prediction.
type Features struct {
	HomeGoalsPerGame    float64
	AwayGoalsPerGame    float64
	HomeConcededPerGame float64
	AwayConcededPerGame float64

	HomeFormPoints5 float64
	AwayFormPoints5 float64
	HomeGoals5      float64
	AwayGoals5      float64

	HomeXGPerGame       float64
	AwayXGPerGame       float64
	HomeCornersPerGame  float64
	AwayCornersPerGame  float64
	HomeCardsPerGame    float64
	AwayCardsPerGame    float64
	HomeOffsidesPerGame float64
	AwayOffsidesPerGame float64

	HomeShotsPerGame         float64
	AwayShotsPerGame         float64
	HomeShotsOnTargetPerGame float64
	AwayShotsOnTargetPerGame float64
	HomeSavesPerGame         float64
	AwaySavesPerGame         float64

	EloDiff float64
}

// FeatureNames mirrors the order of Features for reporting/debugging.
var FeatureNames = []string{
	"home_goals_per_game",
	"away_goals_per_game",
	"home_conceded_per_game",
	"away_conceded_per_game",
	"home_form_points_5",
	"away_form_points_5",
	"home_goals_5",
	"away_goals_5",
	"home_xg_per_game",
	"away_xg_per_game",
	"home_corners_per_game",
	"away_corners_per_game",
	"home_cards_per_game",
	"away_cards_per_game",
	"home_offsides_per_game",
	"away_offsides_per_game",
	"home_shots_per_game",
	"away_shots_per_game",
	"home_shots_on_target_per_game",
	"away_shots_on_target_per_game",
	"home_saves_per_game",
	"away_saves_per_game",
	"elo_diff",
}

func (f *Features) Vector() []float64 {
	return []float64{
		f.HomeGoalsPerGame,
		f.AwayGoalsPerGame,
		f.HomeConcededPerGame,
		f.AwayConcededPerGame,
		f.HomeFormPoints5,
		f.AwayFormPoints5,
		f.HomeGoals5,
		f.AwayGoals5,
		f.HomeXGPerGame,
		f.AwayXGPerGame,
		f.HomeCornersPerGame,
		f.AwayCornersPerGame,
		f.HomeCardsPerGame,
		f.AwayCardsPerGame,
		f.HomeOffsidesPerGame,
		f.AwayOffsidesPerGame,
		f.HomeShotsPerGame,
		f.AwayShotsPerGame,
		f.HomeShotsOnTargetPerGame,
		f.AwayShotsOnTargetPerGame,
		f.HomeSavesPerGame,
		f.AwaySavesPerGame,
		f.EloDiff,
	}
}

const (
	eloInit    = 1500.0
	eloHomeAdv = 50.0
	eloK       = 25.0
)

type recentGame struct {
	goalsFor     int
	goalsAgainst int
	points       int
}

type teamAgg struct {
	games         int
	goalsFor      int
	goalsAgainst  int
	xgFor         float64
	corners       int
	cards         int
	offsides      int
	shots         int
	shotsOnTarget int
	saves         int
	points        int
	elo           float64
	recent        []recentGame
}

func (t *teamAgg) rate(value float64) float64 {
	if t.games == 0 {
		return 0
	}
	return value / float64(t.games)
}

func (t *teamAgg) recentAvg(sel func(recentGame) float64) float64 {
	if len(t.recent) == 0 {
		return 0
	}
	var sum float64
	for _, g := range t.recent {
		sum += sel(g)
	}
	return sum / float64(len(t.recent))
}

func (t *teamAgg) form() (points, goalsFor float64) {
	return t.recentAvg(func(g recentGame) float64 { return float64(g.points) }),
		t.recentAvg(func(g recentGame) float64 { return float64(g.goalsFor) })
}

func winPoints(goalsFor, goalsAgainst int) int {
	switch {
	case goalsFor > goalsAgainst:
		return 3
	case goalsFor == goalsAgainst:
		return 1
	default:
		return 0
	}
}

// Engine maintains causal per-team aggregates while walking matches in
// chronological order.
type Engine struct {
	teams map[uint]*teamAgg
}

func NewEngine() *Engine {
	return &Engine{teams: map[uint]*teamAgg{}}
}

func (e *Engine) team(id uint) *teamAgg {
	if t, ok := e.teams[id]; ok {
		return t
	}
	t := &teamAgg{elo: eloInit}
	e.teams[id] = t
	return t
}

// Features returns the engineered vector for a hypothetical fixture between
// home and away based on the current aggregate state (matches seen so far).
func (e *Engine) Features(homeID, awayID uint) *Features {
	home := e.team(homeID)
	away := e.team(awayID)

	homePoints, homeGoals5 := home.form()
	awayPoints, awayGoals5 := away.form()

	return &Features{
		HomeGoalsPerGame:         home.rate(float64(home.goalsFor)),
		AwayGoalsPerGame:         away.rate(float64(away.goalsFor)),
		HomeConcededPerGame:      home.rate(float64(home.goalsAgainst)),
		AwayConcededPerGame:      away.rate(float64(away.goalsAgainst)),
		HomeFormPoints5:          homePoints,
		AwayFormPoints5:          awayPoints,
		HomeGoals5:               homeGoals5,
		AwayGoals5:               awayGoals5,
		HomeXGPerGame:            home.rate(home.xgFor),
		AwayXGPerGame:            away.rate(away.xgFor),
		HomeCornersPerGame:       home.rate(float64(home.corners)),
		AwayCornersPerGame:       away.rate(float64(away.corners)),
		HomeCardsPerGame:         home.rate(float64(home.cards)),
		AwayCardsPerGame:         away.rate(float64(away.cards)),
		HomeOffsidesPerGame:      home.rate(float64(home.offsides)),
		AwayOffsidesPerGame:      away.rate(float64(away.offsides)),
		HomeShotsPerGame:         home.rate(float64(home.shots)),
		AwayShotsPerGame:         away.rate(float64(away.shots)),
		HomeShotsOnTargetPerGame: home.rate(float64(home.shotsOnTarget)),
		AwayShotsOnTargetPerGame: away.rate(float64(away.shotsOnTarget)),
		HomeSavesPerGame:         home.rate(float64(home.saves)),
		AwaySavesPerGame:         away.rate(float64(away.saves)),
		EloDiff:                  home.elo - away.elo,
	}
}

// Apply folds a completed match into both teams' aggregate state.
func (e *Engine) Apply(m MatchRow) {
	home := e.team(m.HomeTeamID)
	away := e.team(m.AwayTeamID)

	hg, ag := 0, 0
	if m.HomeGoals != nil {
		hg = *m.HomeGoals
	}
	if m.AwayGoals != nil {
		ag = *m.AwayGoals
	}

	home.games++
	away.games++
	home.goalsFor += hg
	home.goalsAgainst += ag
	away.goalsFor += ag
	away.goalsAgainst += hg

	if m.HomeXG != nil {
		home.xgFor += *m.HomeXG
	}
	if m.AwayXG != nil {
		away.xgFor += *m.AwayXG
	}
	if m.HomeCorners != nil {
		home.corners += *m.HomeCorners
	}
	if m.AwayCorners != nil {
		away.corners += *m.AwayCorners
	}
	home.cards += cardCount(m.HomeYellowCards, m.HomeRedCards)
	away.cards += cardCount(m.AwayYellowCards, m.AwayRedCards)
	if m.HomeOffsides != nil {
		home.offsides += *m.HomeOffsides
	}
	if m.AwayOffsides != nil {
		away.offsides += *m.AwayOffsides
	}
	if m.HomeShots != nil {
		home.shots += *m.HomeShots
	}
	if m.AwayShots != nil {
		away.shots += *m.AwayShots
	}
	if m.HomeShotsOnTarget != nil {
		home.shotsOnTarget += *m.HomeShotsOnTarget
	}
	if m.AwayShotsOnTarget != nil {
		away.shotsOnTarget += *m.AwayShotsOnTarget
	}
	if m.HomeSaves != nil {
		home.saves += *m.HomeSaves
	}
	if m.AwaySaves != nil {
		away.saves += *m.AwaySaves
	}

	home.points += winPoints(hg, ag)
	away.points += winPoints(ag, hg)

	home.recent = appendRecent(home.recent, hg, ag, winPoints(hg, ag))
	away.recent = appendRecent(away.recent, ag, hg, winPoints(ag, hg))

	applyElo(home, away, hg, ag)
}

func cardCount(yellow, red *int) int {
	n := 0
	if yellow != nil {
		n += *yellow
	}
	if red != nil {
		n += *red
	}
	return n
}

const recentWindow = 5

func appendRecent(history []recentGame, gf, ga, pts int) []recentGame {
	history = append(history, recentGame{goalsFor: gf, goalsAgainst: ga, points: pts})
	if len(history) > recentWindow {
		history = history[1:]
	}
	return history
}

func applyElo(home, away *teamAgg, hg, ag int) {
	homeBonus := home.elo + eloHomeAdv
	expHome := 1 / (1 + math.Pow(10, (away.elo-homeBonus)/400))
	expAway := 1 - expHome

	homeScore, awayScore := 0.5, 0.5
	if hg > ag {
		homeScore, awayScore = 1, 0
	} else if ag > hg {
		homeScore, awayScore = 0, 1
	}

	home.elo += eloK * (homeScore - expHome)
	away.elo += eloK * (awayScore - expAway)
}

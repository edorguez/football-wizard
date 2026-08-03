package model

// Market identifies a prediction market. Binary markets are each served by a
// single logistic classifier; FirstScorer is served by two (home first, away
// first) and TotalGoals by the Poisson model.
type Market string

const (
	MarketTotalGoals      Market = "total_goals"
	MarketBTTS            Market = "btts"
	MarketCardsOU         Market = "cards_over_under"
	MarketCornersOU       Market = "corners_over_under"
	MarketOffsidesOU      Market = "offsides_over_under"
	MarketShotsOU         Market = "shots_over_under"
	MarketShotsOnTargetOU Market = "shots_on_target_over_under"
	MarketSavesOU         Market = "saves_over_under"
	MarketHomeFirstHalf   Market = "home_first_half_goals"
	MarketAwayFirstHalf   Market = "away_first_half_goals"
	MarketHomeSecondHalf  Market = "home_second_half_goals"
	MarketAwaySecondHalf  Market = "away_second_half_goals"
	MarketFirstScorer     Market = "first_scorer"
)

// DefaultThresholds is the fallback O/U line per market. Config overrides it.
var DefaultThresholds = map[Market]float64{
	MarketCardsOU:         3.5,
	MarketCornersOU:       9.5,
	MarketOffsidesOU:      3.5,
	MarketShotsOU:         24.5,
	MarketShotsOnTargetOU: 7.5,
	MarketSavesOU:         5.5,
	MarketTotalGoals:      2.5,
}

// DefaultGoalLines are the supported total-goals lines offered by the Poisson
// model (0.5 means "at least one goal").
var DefaultGoalLines = []float64{0.5, 1.5, 2.5, 3.5, 4.5}

// MarketSpec describes how to derive a market's label from a completed match.
type MarketSpec struct {
	Market       Market
	HasThreshold bool
	// Label returns 1 or 0 for the market outcome, plus false if the match
	// lacks the required data.
	Label func(m MatchRow, threshold float64) (float64, bool)
}

func (m MatchRow) completed() bool {
	return m.HomeGoals != nil && m.AwayGoals != nil
}

func (m MatchRow) btts() (float64, bool) {
	if !m.completed() {
		return 0, false
	}
	if *m.HomeGoals > 0 && *m.AwayGoals > 0 {
		return 1, true
	}
	return 0, true
}

func (m MatchRow) totalCards() (float64, bool) {
	if m.HomeYellowCards == nil || m.AwayYellowCards == nil {
		return 0, false
	}
	sum := float64(*m.HomeYellowCards + *m.AwayYellowCards)
	if m.HomeRedCards != nil {
		sum += float64(*m.HomeRedCards)
	}
	if m.AwayRedCards != nil {
		sum += float64(*m.AwayRedCards)
	}
	return sum, true
}

func (m MatchRow) corners() (float64, bool) {
	if m.HomeCorners == nil || m.AwayCorners == nil {
		return 0, false
	}
	return float64(*m.HomeCorners + *m.AwayCorners), true
}

func (m MatchRow) offsides() (float64, bool) {
	if m.HomeOffsides == nil || m.AwayOffsides == nil {
		return 0, false
	}
	return float64(*m.HomeOffsides + *m.AwayOffsides), true
}

func (m MatchRow) shots() (float64, bool) {
	if m.HomeShots == nil || m.AwayShots == nil {
		return 0, false
	}
	return float64(*m.HomeShots + *m.AwayShots), true
}

func (m MatchRow) shotsOnTarget() (float64, bool) {
	if m.HomeShotsOnTarget == nil || m.AwayShotsOnTarget == nil {
		return 0, false
	}
	return float64(*m.HomeShotsOnTarget + *m.AwayShotsOnTarget), true
}

func (m MatchRow) saves() (float64, bool) {
	if m.HomeSaves == nil || m.AwaySaves == nil {
		return 0, false
	}
	return float64(*m.HomeSaves + *m.AwaySaves), true
}

func (m MatchRow) firstHalfGoals(sideIsHome bool) (float64, bool) {
	if sideIsHome {
		return halfGoals(m.HomeGoalsFirstHalf)
	}
	return halfGoals(m.AwayGoalsFirstHalf)
}

func (m MatchRow) secondHalfGoals(sideIsHome bool) (float64, bool) {
	if sideIsHome {
		return halfGoals(m.HomeGoalsSecondHalf)
	}
	return halfGoals(m.AwayGoalsSecondHalf)
}

func halfGoals(value *int) (float64, bool) {
	if value == nil {
		return 0, false
	}
	if *value > 0 {
		return 1, true
	}
	return 0, true
}

// FirstScorer returns whether the home team scored first (1), the away team
// scored first (0), or neither did (ok=false).
func (m MatchRow) homeScoredFirst() (float64, bool) {
	if !m.completed() || m.HomeFirstGoalMinute == nil {
		return 0, false
	}
	if m.AwayFirstGoalMinute == nil || *m.HomeFirstGoalMinute < *m.AwayFirstGoalMinute {
		return 1, true
	}
	return 0, true
}

// awayScoredFirst returns whether the away team scored first.
func (m MatchRow) awayScoredFirst() (float64, bool) {
	if !m.completed() || m.AwayFirstGoalMinute == nil {
		return 0, false
	}
	if m.HomeFirstGoalMinute == nil || *m.AwayFirstGoalMinute < *m.HomeFirstGoalMinute {
		return 1, true
	}
	return 0, true
}

// Specs returns the binary market definitions. FirstScorer and TotalGoals are
// handled separately (see Trainer).
var binaryMarketSpecs = []MarketSpec{
	{
		Market: MarketBTTS, HasThreshold: false, Label: func(m MatchRow, _ float64) (float64, bool) {
			return m.btts()
		},
	},
	{
		Market: MarketCardsOU, HasThreshold: true, Label: func(m MatchRow, t float64) (float64, bool) {
			v, ok := m.totalCards()
			if !ok {
				return 0, false
			}
			if v > t {
				return 1, true
			}
			return 0, true
		},
	},
	{
		Market: MarketCornersOU, HasThreshold: true, Label: func(m MatchRow, t float64) (float64, bool) {
			v, ok := m.corners()
			if !ok {
				return 0, false
			}
			if v > t {
				return 1, true
			}
			return 0, true
		},
	},
	{
		Market: MarketOffsidesOU, HasThreshold: true, Label: func(m MatchRow, t float64) (float64, bool) {
			v, ok := m.offsides()
			if !ok {
				return 0, false
			}
			if v > t {
				return 1, true
			}
			return 0, true
		},
	},
	{
		Market: MarketShotsOU, HasThreshold: true, Label: func(m MatchRow, t float64) (float64, bool) {
			v, ok := m.shots()
			if !ok {
				return 0, false
			}
			if v > t {
				return 1, true
			}
			return 0, true
		},
	},
	{
		Market: MarketShotsOnTargetOU, HasThreshold: true, Label: func(m MatchRow, t float64) (float64, bool) {
			v, ok := m.shotsOnTarget()
			if !ok {
				return 0, false
			}
			if v > t {
				return 1, true
			}
			return 0, true
		},
	},
	{
		Market: MarketSavesOU, HasThreshold: true, Label: func(m MatchRow, t float64) (float64, bool) {
			v, ok := m.saves()
			if !ok {
				return 0, false
			}
			if v > t {
				return 1, true
			}
			return 0, true
		},
	},
	{
		Market: MarketHomeFirstHalf, HasThreshold: false, Label: func(m MatchRow, _ float64) (float64, bool) {
			return m.firstHalfGoals(true)
		},
	},
	{
		Market: MarketAwayFirstHalf, HasThreshold: false, Label: func(m MatchRow, _ float64) (float64, bool) {
			return m.firstHalfGoals(false)
		},
	},
	{
		Market: MarketHomeSecondHalf, HasThreshold: false, Label: func(m MatchRow, _ float64) (float64, bool) {
			return m.secondHalfGoals(true)
		},
	},
	{
		Market: MarketAwaySecondHalf, HasThreshold: false, Label: func(m MatchRow, _ float64) (float64, bool) {
			return m.secondHalfGoals(false)
		},
	},
}

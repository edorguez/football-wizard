package model

import (
	"math"

	"github.com/edorguez/football-wizard/pkg/utils"
	"gonum.org/v1/gonum/stat/distuv"
)

type PoissonModel struct {
	HomeAttack  float64
	HomeDefense float64
	AwayAttack  float64
	AwayDefense float64
	LeagueAvg   float64
}

func NewPoissonModel(homeGoalsFor, homeGoalsAgainst, awayGoalsFor, awayGoalsAgainst, totalGoals, matches float64) *PoissonModel {
	if matches == 0 {
		matches = 1
	}
	if totalGoals == 0 {
		totalGoals = 1
	}

	leagueAvg := totalGoals / matches

	return &PoissonModel{
		HomeAttack:  homeGoalsFor / matches / leagueAvg,
		HomeDefense: homeGoalsAgainst / matches / leagueAvg,
		AwayAttack:  awayGoalsFor / matches / leagueAvg,
		AwayDefense: awayGoalsAgainst / matches / leagueAvg,
		LeagueAvg:   leagueAvg,
	}
}

func (m *PoissonModel) ExpectedGoals() (homeExpected, awayExpected float64) {
	homeExpected = m.HomeAttack * m.AwayDefense * m.LeagueAvg
	awayExpected = m.AwayAttack * m.HomeDefense * m.LeagueAvg
	return utils.Clamp(homeExpected, 0.1, 5.0), utils.Clamp(awayExpected, 0.1, 5.0)
}

func (m *PoissonModel) ScoreProbability(homeGoals, awayGoals int) float64 {
	homeExp, awayExp := m.ExpectedGoals()

	homeDist := distuv.Poisson{Lambda: homeExp}
	awayDist := distuv.Poisson{Lambda: awayExp}

	return homeDist.Prob(float64(homeGoals)) * awayDist.Prob(float64(awayGoals))
}

func (m *PoissonModel) HomeWinProb() float64 {
	homeExp, awayExp := m.ExpectedGoals()
	homeDist := distuv.Poisson{Lambda: homeExp}
	awayDist := distuv.Poisson{Lambda: awayExp}

	prob := 0.0
	for h := 0; h <= 10; h++ {
		for a := 0; a <= 10; a++ {
			if h > a {
				prob += homeDist.Prob(float64(h)) * awayDist.Prob(float64(a))
			}
		}
	}
	return prob
}

func (m *PoissonModel) DrawProb() float64 {
	homeExp, awayExp := m.ExpectedGoals()
	homeDist := distuv.Poisson{Lambda: homeExp}
	awayDist := distuv.Poisson{Lambda: awayExp}

	prob := 0.0
	for i := 0; i <= 10; i++ {
		prob += homeDist.Prob(float64(i)) * awayDist.Prob(float64(i))
	}
	return prob
}

func (m *PoissonModel) AwayWinProb() float64 {
	homeExp, awayExp := m.ExpectedGoals()
	homeDist := distuv.Poisson{Lambda: homeExp}
	awayDist := distuv.Poisson{Lambda: awayExp}

	prob := 0.0
	for h := 0; h <= 10; h++ {
		for a := 0; a <= 10; a++ {
			if h < a {
				prob += homeDist.Prob(float64(h)) * awayDist.Prob(float64(a))
			}
		}
	}
	return prob
}

func (m *PoissonModel) OverProb(threshold float64) float64 {
	homeExp, awayExp := m.ExpectedGoals()
	homeDist := distuv.Poisson{Lambda: homeExp}
	awayDist := distuv.Poisson{Lambda: awayExp}

	prob := 0.0
	for h := 0; h <= 10; h++ {
		for a := 0; a <= 10; a++ {
			if float64(h+a) > threshold {
				prob += homeDist.Prob(float64(h)) * awayDist.Prob(float64(a))
			}
		}
	}
	return prob
}

func (m *PoissonModel) Over05Prob() float64  { return m.OverProb(0.5) }
func (m *PoissonModel) Over15Prob() float64  { return m.OverProb(1.5) }
func (m *PoissonModel) Over25Prob() float64  { return m.OverProb(2.5) }
func (m *PoissonModel) Over35Prob() float64  { return m.OverProb(3.5) }

func (m *PoissonModel) BttsProb() float64 {
	homeExp, awayExp := m.ExpectedGoals()
	homeDist := distuv.Poisson{Lambda: homeExp}
	awayDist := distuv.Poisson{Lambda: awayExp}

	homeScore := 1.0 - homeDist.Prob(0)
	awayScore := 1.0 - awayDist.Prob(0)
	return homeScore * awayScore
}

func poissonProb(lambda float64, k int) float64 {
	return math.Exp(-lambda) * math.Pow(lambda, float64(k)) / float64(factorial(k))
}

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}

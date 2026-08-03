package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarketLabels(t *testing.T) {
	t.Parallel()

	m := makeMatch(2025, 1, 1, 2, 2, 1)
	m.HomeCorners = intPtr(5)
	m.AwayCorners = intPtr(4)
	m.HomeYellowCards = intPtr(2)
	m.AwayYellowCards = intPtr(1)
	m.HomeGoalsFirstHalf = intPtr(1)
	m.AwayGoalsFirstHalf = intPtr(0)
	m.HomeFirstGoalMinute = intPtr(10)
	m.AwayFirstGoalMinute = intPtr(30)
	m.HomeShots = intPtr(14)
	m.AwayShots = intPtr(8)
	m.HomeShotsOnTarget = intPtr(6)
	m.AwayShotsOnTarget = intPtr(2)
	m.HomeSaves = intPtr(1)
	m.AwaySaves = intPtr(4)

	is := assert.New(t)

	v, ok := m.btts()
	is.True(ok)
	is.Equal(1.0, v)

	v, ok = m.totalCards()
	is.True(ok)
	is.Equal(3.0, v)

	v, ok = m.corners()
	is.True(ok)
	is.Equal(9.0, v)

	v, ok = m.firstHalfGoals(true)
	is.True(ok)
	is.Equal(1.0, v)

	v, ok = m.firstHalfGoals(false)
	is.True(ok)
	is.Equal(0.0, v)

	v, ok = m.homeScoredFirst()
	is.True(ok)
	is.Equal(1.0, v)

	v, ok = m.awayScoredFirst()
	is.True(ok)
	is.Equal(0.0, v)

	v, ok = m.shots()
	is.True(ok)
	is.Equal(22.0, v)

	v, ok = m.shotsOnTarget()
	is.True(ok)
	is.Equal(8.0, v)

	v, ok = m.saves()
	is.True(ok)
	is.Equal(5.0, v)
}

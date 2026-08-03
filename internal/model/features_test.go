package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEngineFeaturesCausal(t *testing.T) {
	t.Parallel()

	engine := NewEngine()

	// Team 1 plays its first game: no prior form, so rates are zero.
	f := engine.Features(1, 2)
	is := assert.New(t)
	is.Equal(0.0, f.HomeGoalsPerGame)
	is.Equal(0.0, f.HomeFormPoints5)
	is.Zero(f.EloDiff)

	// Team 1 wins 2-1.
	engine.Apply(makeMatch(2025, 1, 1, 2, 2, 1))

	f = engine.Features(1, 2)
	is.Equal(2.0, f.HomeGoalsPerGame)
	is.Equal(1.0, f.HomeConcededPerGame)
	is.Equal(3.0, f.HomeFormPoints5)
	is.Equal(2.0, f.HomeGoals5)
}

func TestEngineRecentWindow(t *testing.T) {
	t.Parallel()

	engine := NewEngine()
	for i := 0; i < 8; i++ {
		engine.Apply(makeMatch(2025, i, 1, 2, 1, 0))
	}

	// Average points over the last 5 wins: 3.0 per game (all wins).
	is := assert.New(t)
	is.Equal(3.0, engine.Features(1, 2).HomeFormPoints5)
	is.Equal(1.0, engine.Features(1, 2).HomeGoals5)
}

func TestEngineShotsSavesFeatures(t *testing.T) {
	t.Parallel()

	engine := NewEngine()

	m := makeMatch(2025, 1, 1, 2, 1, 0)
	m.HomeShots = intPtr(12)
	m.AwayShots = intPtr(5)
	m.HomeShotsOnTarget = intPtr(4)
	m.AwayShotsOnTarget = intPtr(1)
	m.HomeSaves = intPtr(2)
	m.AwaySaves = intPtr(3)
	engine.Apply(m)

	f := engine.Features(1, 2)
	is := assert.New(t)
	is.Equal(12.0, f.HomeShotsPerGame)
	is.Equal(5.0, f.AwayShotsPerGame)
	is.Equal(4.0, f.HomeShotsOnTargetPerGame)
	is.Equal(1.0, f.AwayShotsOnTargetPerGame)
	is.Equal(2.0, f.HomeSavesPerGame)
	is.Equal(3.0, f.AwaySavesPerGame)
}

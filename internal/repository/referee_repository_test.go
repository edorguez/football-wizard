package repository

import (
	"testing"

	"github.com/edorguez/football-wizard/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefereeRepository_Upsert(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	repo := NewRefereeRepository(db)

	ref := &database.Referee{Name: "Anderson Daronco"}
	err := repo.Upsert(ref)

	is := assert.New(t)

	is.NoError(err)
	is.NotZero(ref.ID)
}

func TestRefereeRepository_Upsert_Duplicate(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	repo := NewRefereeRepository(db)

	ref := &database.Referee{Name: "Wilton Sampaio"}
	require.NoError(t, repo.Upsert(ref))

	ref.ID = 0
	err := repo.Upsert(ref)

	assert.NoError(t, err)
}

func TestRefereeRepository_FindByName(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	repo := NewRefereeRepository(db)

	require.NoError(t, repo.Upsert(&database.Referee{Name: "Raphael Claus"}))

	found, err := repo.FindByName("Raphael Claus")

	is := assert.New(t)
	must := require.New(t)

	must.NoError(err)
	must.NotNil(found)
	is.Equal("Raphael Claus", found.Name)
}

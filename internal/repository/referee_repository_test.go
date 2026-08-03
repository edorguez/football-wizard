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

func TestRefereeRepository_Search(t *testing.T) {
	t.Parallel()

	db := setupRepoDB(t)
	repo := NewRefereeRepository(db)

	for _, name := range []string{"Raphael Claus", "Rafael Traci", "Anderson Daronco"} {
		require.NoError(t, repo.Upsert(&database.Referee{Name: name}))
	}

	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "partial", query: "Raph", want: 1},
		{name: "lowercase", query: "rafael", want: 1},
		{name: "no match", query: "zzz", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := repo.Search(tt.query)
			require.NoError(t, err)
			assert.Len(t, results, tt.want)
		})
	}
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

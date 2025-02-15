package migrations

import (
	"testing"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/migrations"
	"github.com/stretchr/testify/require"
)

func TestBadgerStorage_SupportedStatusesAreSavedAndReturned(t *testing.T) {
	testCases := []struct {
		Name   string
		Status migrations.Status
	}{
		{
			Name:   "finished",
			Status: migrations.StatusFinished,
		},
		{
			Name:   "failed",
			Status: migrations.StatusFailed,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db := fixtures.Badger(t)
			storage := NewBadgerStorage(db)

			someName1 := fixtures.SomeString()
			someName2 := fixtures.SomeString()

			_, err := storage.LoadStatus(someName1)
			require.ErrorIs(t, err, migrations.ErrStatusNotFound)

			_, err = storage.LoadStatus(someName2)
			require.ErrorIs(t, err, migrations.ErrStatusNotFound)

			err = storage.SaveStatus(someName1, testCase.Status)
			require.NoError(t, err)

			status, err := storage.LoadStatus(someName1)
			require.NoError(t, err)
			require.Equal(t, status, testCase.Status)

			_, err = storage.LoadStatus(someName2)
			require.ErrorIs(t, err, migrations.ErrStatusNotFound)
		})
	}
}

func TestBadgerStorage_StateIsSavedAndReturned(t *testing.T) {
	db := fixtures.Badger(t)
	storage := NewBadgerStorage(db)

	someName1 := fixtures.SomeString()
	someName2 := fixtures.SomeString()

	_, err := storage.LoadState(someName1)
	require.ErrorIs(t, err, migrations.ErrStateNotFound)

	_, err = storage.LoadState(someName2)
	require.ErrorIs(t, err, migrations.ErrStateNotFound)

	state := migrations.State{
		fixtures.SomeString(): fixtures.SomeString(),
	}

	err = storage.SaveState(someName1, state)
	require.NoError(t, err)

	retrievedState, err := storage.LoadState(someName1)
	require.NoError(t, err)
	require.Equal(t, state, retrievedState)

	_, err = storage.LoadState(someName2)
	require.ErrorIs(t, err, migrations.ErrStateNotFound)
}

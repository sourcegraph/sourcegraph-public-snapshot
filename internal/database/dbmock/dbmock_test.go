package dbmock_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock/internal"
	"github.com/stretchr/testify/require"
)

func TestWithDB(t *testing.T) {
	mockDB := database.NewMockDB()
	mockStore := internal.NewMockStore()
	db := dbmock.New(mockDB, mockStore)

	err := internal.NewStore().WithDB(db).Create(context.Background())
	require.NoError(t, err)
}

// TestMockedStoreIsReturnedFromWithDB asserts that, when a mocked store is
// embedded into a database.DB, it is returned when its corresponding store
// is initialized with .WithDB(db).
func TestMockedStoreIsReturnedFromWithDB(t *testing.T) {
	const mockedErr = "mocked return error"
	mockStore := internal.NewMockStore()
	mockStore.CreateFunc.SetDefaultReturn(errors.New(mockedErr))

	db := dbmock.New(database.NewMockDB(), mockStore)

	err := internal.NewStore().WithDB(db).Create(context.Background())
	require.ErrorContains(t, err, mockedErr)
}

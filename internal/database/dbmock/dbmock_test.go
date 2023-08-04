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

// TestUninitializedStorePanics demonstrates that a MockableStore
// that has not been accessed with `.WithDB(db)` will panic.
// This prevents accidentally using uninitialized stores.
func TestUninitializedStorePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf(`Expected store on which .WithDB(db) has not been called to panic`)
		}
	}()

	internal.NewStore().Create(context.Background())
}

// TestInitializedStoreDoesNotPanic demonstrates how to embed a MockableStore
// in a database.DB.
func TestInitializedStoreDoesNotPanic(t *testing.T) {
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

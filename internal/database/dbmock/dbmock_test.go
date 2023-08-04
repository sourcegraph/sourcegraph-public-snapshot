package dbmock_test

import (
	"context"
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
			t.Errorf("Expected store on which .WithDB(db) has not been called to panic")
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

	require.NoError(t, internal.NewStore().WithDB(db).Create(context.Background()))
}

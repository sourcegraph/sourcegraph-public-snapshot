package dbmock_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock/internal"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock/internal/internalmock"
	"github.com/stretchr/testify/require"
)

// TestMockedStoreIsReturnedFromWithDB asserts that, when a mocked store is
// embedded into a database.DB, it is returned when its corresponding store
// is initialized with .WithDB(db).
func TestMockedStoreIsReturnedFromWithDB(t *testing.T) {
	const mockedErr = "mocked return error"
	mockStore := internalmock.NewMockDBStore()
	mockStore.CreateFunc.SetDefaultReturn(errors.New(mockedErr))

	db := dbmock.New(database.NewMockDB(), mockStore)

	err := internal.Store.WithDB(db).Create(context.Background())
	require.ErrorContains(t, err, mockedErr)
}

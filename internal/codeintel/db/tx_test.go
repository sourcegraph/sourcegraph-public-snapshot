package db

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestSavepointNotInTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	if err := db.Savepoint(context.Background(), "sp_test"); err != ErrNoTransaction {
		t.Errorf("unexpected error. want=%q have=%q", ErrNoTransaction, err)
	}
}

func TestRollbackToSavepointNotInTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	if err := db.RollbackToSavepoint(context.Background(), "sp_test"); err != ErrNoTransaction {
		t.Errorf("unexpected error. want=%q have=%q", ErrNoTransaction, err)
	}
}

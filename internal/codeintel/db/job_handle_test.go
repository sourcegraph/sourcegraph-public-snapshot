package db

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestJobHandleUnmarkedClose(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	ctx := context.Background()
	tx, _, err := db.transact(ctx)
	if err != nil {
		t.Fatalf("unexpected error creating transaction: %s", err)
	}

	jobHandle := &jobHandleImpl{db: tx, id: 1}
	if err := jobHandle.Done(nil); err == nil {
		t.Fatalf("unexpected nil error")
	} else if !strings.Contains(err.Error(), ErrJobNotFinalized.Error()) {
		t.Errorf("unexpected error. want=%q have=%q", ErrJobNotFinalized, err)
	}
}

func TestJobHandleRollbackNoSavepoint(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	ctx := context.Background()
	tx, _, err := db.transact(ctx)
	if err != nil {
		t.Fatalf("unexpected error creating transaction: %s", err)
	}

	jobHandle := &jobHandleImpl{db: tx, id: 1}
	if err := jobHandle.RollbackToLastSavepoint(ctx); err == nil {
		t.Fatalf("unexpected nil error")
	} else if !strings.Contains(err.Error(), ErrNoSavepoint.Error()) {
		t.Errorf("unexpected error. want=%q have=%q", ErrNoSavepoint, err)
	}
}

func TestJobHandleSavepointRollback(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	ctx := context.Background()
	tx, _, err := db.transact(ctx)
	if err != nil {
		t.Fatalf("unexpected error creating transaction: %s", err)
	}

	jobHandle := &jobHandleImpl{db: tx, id: 1}
	if err := jobHandle.Savepoint(ctx); err != nil {
		t.Fatalf("unexpected error creating savepoint: %s", err)
	}
	if err := jobHandle.MarkComplete(ctx); err != nil {
		t.Fatalf("unexpected error marking upload complete: %s", err)
	}
	if err := jobHandle.RollbackToLastSavepoint(ctx); err != nil {
		t.Fatalf("unexpected error rolling back to savepoint: %s", err)
	}

	if err := jobHandle.Done(nil); err == nil {
		t.Fatalf("unexpected nil error")
	} else if !strings.Contains(err.Error(), ErrJobNotFinalized.Error()) {
		t.Errorf("unexpected error. want=%q have=%q", ErrJobNotFinalized, err)
	}
}

func TestJobHandlePartialSavepointRollback(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	ctx := context.Background()
	tx, _, err := db.transact(ctx)
	if err != nil {
		t.Fatalf("unexpected error creating transaction: %s", err)
	}

	jobHandle := &jobHandleImpl{db: tx, id: 1}
	if err := jobHandle.Savepoint(ctx); err != nil {
		t.Fatalf("unexpected error creating savepoint: %s", err)
	}
	if err := jobHandle.MarkComplete(ctx); err != nil {
		t.Fatalf("unexpected error marking upload complete: %s", err)
	}
	if err := jobHandle.Savepoint(ctx); err != nil {
		t.Fatalf("unexpected error creating savepoint: %s", err)
	}
	if err := jobHandle.RollbackToLastSavepoint(ctx); err != nil {
		t.Fatalf("unexpected error rolling back to savepoint: %s", err)
	}

	if err := jobHandle.Done(nil); err != nil {
		t.Fatalf("unexpected error closing transaction: %s", err)
	}
}

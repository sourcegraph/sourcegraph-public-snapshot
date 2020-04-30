package db

import (
	"context"
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
	tw, err := db.beginTx(ctx)
	if err != nil {
		t.Fatalf("unexpected error creating transaction: %s", err)
	}

	jobHandle := &jobHandleImpl{
		ctx:      ctx,
		id:       1,
		tw:       tw,
		txCloser: &txCloser{tw.tx},
	}

	if err := jobHandle.CloseTx(nil); err == nil {
		t.Fatalf("unexpected nil error")
	} else if err != ErrJobNotFinalized {
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
	tw, err := db.beginTx(ctx)
	if err != nil {
		t.Fatalf("unexpected error creating transaction: %s", err)
	}

	jobHandle := &jobHandleImpl{
		ctx:      ctx,
		id:       1,
		tw:       tw,
		txCloser: &txCloser{tw.tx},
	}

	if err := jobHandle.RollbackToLastSavepoint(); err == nil {
		t.Fatalf("unexpected nil error")
	} else if err != ErrNoSavepoint {
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
	tw, err := db.beginTx(ctx)
	if err != nil {
		t.Fatalf("unexpected error creating transaction: %s", err)
	}

	jobHandle := &jobHandleImpl{
		ctx:      ctx,
		id:       1,
		tw:       tw,
		txCloser: &txCloser{tw.tx},
	}

	if err := jobHandle.Savepoint(); err != nil {
		t.Fatalf("unexpected error creating savepoint: %s", err)
	}
	if err := jobHandle.MarkComplete(); err != nil {
		t.Fatalf("unexpected error marking upload complete: %s", err)
	}
	if err := jobHandle.RollbackToLastSavepoint(); err != nil {
		t.Fatalf("unexpected error rolling back to savepoint: %s", err)
	}

	if err := jobHandle.CloseTx(nil); err == nil {
		t.Fatalf("unexpected nil error")
	} else if err != ErrJobNotFinalized {
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
	tw, err := db.beginTx(ctx)
	if err != nil {
		t.Fatalf("unexpected error creating transaction: %s", err)
	}

	jobHandle := &jobHandleImpl{
		ctx:      ctx,
		id:       1,
		tw:       tw,
		txCloser: &txCloser{tw.tx},
	}

	if err := jobHandle.Savepoint(); err != nil {
		t.Fatalf("unexpected error creating savepoint: %s", err)
	}
	if err := jobHandle.MarkComplete(); err != nil {
		t.Fatalf("unexpected error marking upload complete: %s", err)
	}
	if err := jobHandle.Savepoint(); err != nil {
		t.Fatalf("unexpected error creating savepoint: %s", err)
	}
	if err := jobHandle.RollbackToLastSavepoint(); err != nil {
		t.Fatalf("unexpected error rolling back to savepoint: %s", err)
	}

	if err := jobHandle.CloseTx(nil); err != nil {
		t.Fatalf("unexpected error closing transaction: %s", err)
	}
}

package db

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestDequeueRecordConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := rawTestDB()

	// Add dequeueable upload
	insertUploads(t, dbconn.Global, Upload{ID: 1, State: "queued"})

	_, tx1, ok1, err1 := db.dequeueByID(context.Background(), "lsif_uploads", uploadColumnsWithNullRank, scanFirstUploadInterface, 1)
	if ok1 {
		defer func() { _ = tx1.Done(nil) }()
	}

	_, tx2, ok2, err2 := db.dequeueByID(context.Background(), "lsif_uploads", uploadColumnsWithNullRank, scanFirstUploadInterface, 1)
	if ok2 {
		defer func() { _ = tx2.Done(nil) }()
	}

	if err1 != ErrDequeueRace && err2 != ErrDequeueRace {
		t.Errorf("expected error. want=%q have=%q and %q", ErrDequeueRace, err1, err2)
	}
}

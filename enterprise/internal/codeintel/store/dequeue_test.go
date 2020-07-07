package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestDequeueMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := rawTestStore()

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(-time.Minute)
	t3 := t2.Add(-time.Minute)
	insertUploads(
		t,
		dbconn.Global,
		Upload{ID: 1, State: "queued", UploadedAt: t1},
		Upload{ID: 2, State: "queued", UploadedAt: t2},
		Upload{ID: 3, State: "queued", UploadedAt: t3},
	)

	var ids []int
	for i := 0; i < 3; i++ {
		record, tx, ok, err := store.dequeueRecord(
			context.Background(),
			"lsif_uploads_with_repository_name",
			"lsif_uploads",
			uploadColumnsWithNullRank,
			sqlf.Sprintf("uploaded_at desc"),
			nil,
			scanFirstUploadInterface,
		)

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !ok {
			t.Fatalf("expected record")
		}
		defer func() { _ = tx.Done(nil) }()

		ids = append(ids, record.(Upload).ID)
	}

	if diff := cmp.Diff([]int{1, 2, 3}, ids); diff != "" {
		t.Errorf("unexpected dump (-want +got):\n%s", diff)
	}
}

func TestDequeueByIDRace(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := rawTestStore()

	insertUploads(t, dbconn.Global, Upload{ID: 1, State: "queued"})

	_, tx1, ok1, err1 := store.dequeueByID(
		context.Background(),
		"lsif_uploads_with_repository_name",
		"lsif_uploads",
		uploadColumnsWithNullRank,
		scanFirstUploadInterface,
		1,
	)
	if ok1 {
		defer func() { _ = tx1.Done(nil) }()
	}

	_, tx2, ok2, err2 := store.dequeueByID(
		context.Background(),
		"lsif_uploads_with_repository_name",
		"lsif_uploads",
		uploadColumnsWithNullRank,
		scanFirstUploadInterface,
		1,
	)
	if ok2 {
		defer func() { _ = tx2.Done(nil) }()
	}

	if err1 != ErrDequeueRace && err2 != ErrDequeueRace {
		t.Errorf("expected error. want=%q have=%q and %q", ErrDequeueRace, err1, err2)
	}
}

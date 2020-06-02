package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestGetIndexByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	// Index does not exist initially
	if _, exists, err := db.GetIndexByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting index: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}

	queuedAt := time.Unix(1587396557, 0).UTC()
	startedAt := queuedAt.Add(time.Minute)
	expected := Index{
		ID:                1,
		Commit:            makeCommit(1),
		QueuedAt:          queuedAt,
		State:             "processing",
		FailureSummary:    nil,
		FailureStacktrace: nil,
		StartedAt:         &startedAt,
		FinishedAt:        nil,
		RepositoryID:      123,
		Rank:              nil,
	}

	insertIndexes(t, dbconn.Global, expected)

	if index, exists, err := db.GetIndexByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting index: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if diff := cmp.Diff(expected, index); diff != "" {
		t.Errorf("unexpected index (-want +got):\n%s", diff)
	}
}

func TestGetQueuedIndexRank(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(+time.Minute * 5)
	t3 := t1.Add(+time.Minute * 3)
	t4 := t1.Add(+time.Minute * 1)
	t5 := t1.Add(+time.Minute * 4)
	t6 := t1.Add(+time.Minute * 2)

	insertIndexes(t, dbconn.Global,
		Index{ID: 1, QueuedAt: t1, State: "queued"},
		Index{ID: 2, QueuedAt: t2, State: "queued"},
		Index{ID: 3, QueuedAt: t3, State: "queued"},
		Index{ID: 4, QueuedAt: t4, State: "queued"},
		Index{ID: 5, QueuedAt: t5, State: "queued"},
		Index{ID: 6, QueuedAt: t6, State: "processing"},
	)

	if index, _, _ := db.GetIndexByID(context.Background(), 1); index.Rank == nil || *index.Rank != 1 {
		t.Errorf("unexpected rank. want=%d have=%s", 1, printableRank{index.Rank})
	}
	if index, _, _ := db.GetIndexByID(context.Background(), 2); index.Rank == nil || *index.Rank != 5 {
		t.Errorf("unexpected rank. want=%d have=%s", 5, printableRank{index.Rank})
	}
	if index, _, _ := db.GetIndexByID(context.Background(), 3); index.Rank == nil || *index.Rank != 3 {
		t.Errorf("unexpected rank. want=%d have=%s", 3, printableRank{index.Rank})
	}
	if index, _, _ := db.GetIndexByID(context.Background(), 4); index.Rank == nil || *index.Rank != 2 {
		t.Errorf("unexpected rank. want=%d have=%s", 2, printableRank{index.Rank})
	}
	if index, _, _ := db.GetIndexByID(context.Background(), 5); index.Rank == nil || *index.Rank != 4 {
		t.Errorf("unexpected rank. want=%d have=%s", 4, printableRank{index.Rank})
	}

	// Only considers queued indexes to determine rank
	if index, _, _ := db.GetIndexByID(context.Background(), 6); index.Rank != nil {
		t.Errorf("unexpected rank. want=%s have=%s", "nil", printableRank{index.Rank})
	}
}

func TestIndexQueueSize(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	insertIndexes(t, dbconn.Global,
		Index{ID: 1, State: "queued"},
		Index{ID: 2, State: "errored"},
		Index{ID: 3, State: "processing"},
		Index{ID: 4, State: "completed"},
		Index{ID: 5, State: "completed"},
		Index{ID: 6, State: "queued"},
		Index{ID: 7, State: "processing"},
		Index{ID: 8, State: "completed"},
		Index{ID: 9, State: "processing"},
		Index{ID: 10, State: "queued"},
	)

	count, err := db.IndexQueueSize(context.Background())
	if err != nil {
		t.Fatalf("unexpected error getting index queue size: %s", err)
	}
	if count != 3 {
		t.Errorf("unexpected count. want=%d have=%d", 3, count)
	}
}

func TestIsQueued(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	insertIndexes(t, dbconn.Global, Index{ID: 1, RepositoryID: 1, Commit: makeCommit(1)})
	insertUploads(t, dbconn.Global, Upload{ID: 2, RepositoryID: 2, Commit: makeCommit(2)})

	testCases := []struct {
		repositoryID int
		commit       string
		expected     bool
	}{
		{1, makeCommit(1), true},
		{1, makeCommit(2), false},
		{2, makeCommit(1), false},
		{2, makeCommit(2), true},
		{3, makeCommit(1), false},
		{3, makeCommit(2), false},
	}

	for _, testCase := range testCases {
		name := fmt.Sprintf("repositoryId=%d commit=%s", testCase.repositoryID, testCase.commit)

		t.Run(name, func(t *testing.T) {
			queued, err := db.IsQueued(context.Background(), testCase.repositoryID, testCase.commit)
			if err != nil {
				t.Fatalf("unexpected error checking if commit is queued: %s", err)
			}
			if queued != testCase.expected {
				t.Errorf("unexpected state. want=%v have=%v", testCase.expected, queued)
			}
		})
	}
}

func TestInsertIndex(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	id, err := db.InsertIndex(context.Background(), Index{
		Commit:       makeCommit(1),
		State:        "queued",
		RepositoryID: 50,
	})
	if err != nil {
		t.Fatalf("unexpected error enqueueing index: %s", err)
	}

	rank := 1
	expected := Index{
		ID:                id,
		Commit:            makeCommit(1),
		QueuedAt:          time.Time{},
		State:             "queued",
		FailureSummary:    nil,
		FailureStacktrace: nil,
		StartedAt:         nil,
		FinishedAt:        nil,
		RepositoryID:      50,
		Rank:              &rank,
	}

	if index, exists, err := db.GetIndexByID(context.Background(), id); err != nil {
		t.Fatalf("unexpected error getting index: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else {
		// Update auto-generated timestamp
		expected.QueuedAt = index.QueuedAt

		if diff := cmp.Diff(expected, index); diff != "" {
			t.Errorf("unexpected index (-want +got):\n%s", diff)
		}
	}
}

func TestMarkIndexComplete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	insertIndexes(t, dbconn.Global, Index{ID: 1, State: "queued"})

	if err := db.MarkIndexComplete(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error marking index as complete: %s", err)
	}

	if index, exists, err := db.GetIndexByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting index: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if index.State != "completed" {
		t.Errorf("unexpected state. want=%q have=%q", "completed", index.State)
	}
}

func TestMarkIndexErrored(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	insertIndexes(t, dbconn.Global, Index{ID: 1, State: "queued"})

	if err := db.MarkIndexErrored(context.Background(), 1, "oops", ""); err != nil {
		t.Fatalf("unexpected error marking index as complete: %s", err)
	}

	if index, exists, err := db.GetIndexByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting index: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if index.State != "errored" {
		t.Errorf("unexpected state. want=%q have=%q", "errored", index.State)
	}
}

func TestDequeueIndexProcessSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	// Add dequeueable index
	insertIndexes(t, dbconn.Global, Index{ID: 1, State: "queued"})

	index, tx, ok, err := db.DequeueIndex(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing index: %s", err)
	}
	if !ok {
		t.Fatalf("expected something to be dequeueable")
	}

	if index.ID != 1 {
		t.Errorf("unexpected index id. want=%d have=%d", 1, index.ID)
	}
	if index.State != "processing" {
		t.Errorf("unexpected state. want=%s have=%s", "processing", index.State)
	}

	if state, _, err := scanFirstString(dbconn.Global.Query("SELECT state FROM lsif_indexes WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting state: %s", err)
	} else if state != "processing" {
		t.Errorf("unexpected state outside of txn. want=%s have=%s", "processing", state)
	}

	if err := tx.MarkIndexComplete(context.Background(), index.ID); err != nil {
		t.Fatalf("unexpected error marking index complete: %s", err)
	}
	_ = tx.Done(nil)

	if state, _, err := scanFirstString(dbconn.Global.Query("SELECT state FROM lsif_indexes WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting state: %s", err)
	} else if state != "completed" {
		t.Errorf("unexpected state outside of txn. want=%s have=%s", "completed", state)
	}
}

func TestDequeueIndexProcessError(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	// Add dequeueable index
	insertIndexes(t, dbconn.Global, Index{ID: 1, State: "queued"})

	index, tx, ok, err := db.DequeueIndex(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing index: %s", err)
	}
	if !ok {
		t.Fatalf("expected something to be dequeueable")
	}

	if index.ID != 1 {
		t.Errorf("unexpected index id. want=%d have=%d", 1, index.ID)
	}
	if index.State != "processing" {
		t.Errorf("unexpected state. want=%s have=%s", "processing", index.State)
	}

	if state, _, err := scanFirstString(dbconn.Global.Query("SELECT state FROM lsif_indexes WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting state: %s", err)
	} else if state != "processing" {
		t.Errorf("unexpected state outside of txn. want=%s have=%s", "processing", state)
	}

	if err := tx.MarkIndexErrored(context.Background(), index.ID, "test summary", "test stacktrace"); err != nil {
		t.Fatalf("unexpected error marking index complete: %s", err)
	}
	_ = tx.Done(nil)

	if state, _, err := scanFirstString(dbconn.Global.Query("SELECT state FROM lsif_indexes WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting state: %s", err)
	} else if state != "errored" {
		t.Errorf("unexpected state outside of txn. want=%s have=%s", "errored", state)
	}

	if summary, _, err := scanFirstString(dbconn.Global.Query("SELECT failure_summary FROM lsif_indexes WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting failure_summary: %s", err)
	} else if summary != "test summary" {
		t.Errorf("unexpected failure summary outside of txn. want=%s have=%s", "test summary", summary)
	}

	if stacktrace, _, err := scanFirstString(dbconn.Global.Query("SELECT failure_stacktrace FROM lsif_indexes WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting failure_stacktrace: %s", err)
	} else if stacktrace != "test stacktrace" {
		t.Errorf("unexpected failure stacktrace outside of txn. want=%s have=%s", "test stacktrace", stacktrace)
	}
}

func TestDequeueIndexSkipsLocked(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(time.Minute)
	t3 := t2.Add(time.Minute)
	insertIndexes(
		t,
		dbconn.Global,
		Index{ID: 1, State: "queued", QueuedAt: t1},
		Index{ID: 2, State: "processing", QueuedAt: t2},
		Index{ID: 3, State: "queued", QueuedAt: t3},
	)

	tx1, err := dbconn.Global.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx1.Rollback() }()

	// Row lock index 1 in a transaction which should be skipped by ResetStalled
	if _, err := tx1.Query(`SELECT * FROM lsif_indexes WHERE id = 1 FOR UPDATE`); err != nil {
		t.Fatal(err)
	}

	index, tx2, ok, err := db.DequeueIndex(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing index: %s", err)
	}
	if !ok {
		t.Fatalf("expected something to be dequeueable")
	}
	defer func() { _ = tx2.Done(nil) }()

	if index.ID != 3 {
		t.Errorf("unexpected index id. want=%d have=%d", 3, index.ID)
	}
	if index.State != "processing" {
		t.Errorf("unexpected state. want=%s have=%s", "processing", index.State)
	}
}

func TestDequeueIndexEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	_, tx, ok, err := db.DequeueIndex(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing index: %s", err)
	}
	if ok {
		_ = tx.Done(nil)
		t.Fatalf("unexpected dequeue")
	}
}

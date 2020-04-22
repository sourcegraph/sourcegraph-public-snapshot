package db

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

type printableRank struct{ value *int }

func (r printableRank) String() string {
	if r.value == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", r.value)
}

func TestGetUploadByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	// Upload does not exist initially
	if _, exists, err := db.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}

	uploadedAt := time.Unix(1587396557, 0).UTC()
	startedAt := uploadedAt.Add(time.Minute)
	expected := Upload{
		ID:                1,
		Commit:            makeCommit(1),
		Root:              "sub/",
		VisibleAtTip:      true,
		UploadedAt:        uploadedAt,
		State:             "processing",
		FailureSummary:    nil,
		FailureStacktrace: nil,
		StartedAt:         &startedAt,
		FinishedAt:        nil,
		TracingContext:    `{"id": 42}`,
		RepositoryID:      123,
		Indexer:           "lsif-go",
		Rank:              nil,
	}

	insertUploads(t, db.db, expected)

	if upload, exists, err := db.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if diff := cmp.Diff(expected, upload); diff != "" {
		t.Errorf("unexpected upload (-want +got):\n%s", diff)
	}
}

func TestGetQueuedUploadRank(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(+time.Minute * 5)
	t3 := t1.Add(+time.Minute * 3)
	t4 := t1.Add(+time.Minute * 1)
	t5 := t1.Add(+time.Minute * 4)
	t6 := t1.Add(+time.Minute * 2)

	insertUploads(t, db.db,
		Upload{ID: 1, UploadedAt: t1, State: "queued"},
		Upload{ID: 2, UploadedAt: t2, State: "queued"},
		Upload{ID: 3, UploadedAt: t3, State: "queued"},
		Upload{ID: 4, UploadedAt: t4, State: "queued"},
		Upload{ID: 5, UploadedAt: t5, State: "queued"},
		Upload{ID: 6, UploadedAt: t6, State: "processing"},
	)

	if upload, _, _ := db.GetUploadByID(context.Background(), 1); upload.Rank == nil || *upload.Rank != 1 {
		t.Errorf("unexpected rank. want=%d have=%s", 1, printableRank{upload.Rank})
	}
	if upload, _, _ := db.GetUploadByID(context.Background(), 2); upload.Rank == nil || *upload.Rank != 5 {
		t.Errorf("unexpected rank. want=%d have=%s", 5, printableRank{upload.Rank})
	}
	if upload, _, _ := db.GetUploadByID(context.Background(), 3); upload.Rank == nil || *upload.Rank != 3 {
		t.Errorf("unexpected rank. want=%d have=%s", 3, printableRank{upload.Rank})
	}
	if upload, _, _ := db.GetUploadByID(context.Background(), 4); upload.Rank == nil || *upload.Rank != 2 {
		t.Errorf("unexpected rank. want=%d have=%s", 2, printableRank{upload.Rank})
	}
	if upload, _, _ := db.GetUploadByID(context.Background(), 5); upload.Rank == nil || *upload.Rank != 4 {
		t.Errorf("unexpected rank. want=%d have=%s", 4, printableRank{upload.Rank})
	}

	// Only considers queued uploads to determine rank
	if upload, _, _ := db.GetUploadByID(context.Background(), 6); upload.Rank != nil {
		t.Errorf("unexpected rank. want=%s have=%s", "nil", printableRank{upload.Rank})
	}
}

func TestGetUploadsByRepo(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(-time.Minute * 1)
	t3 := t1.Add(-time.Minute * 2)
	t4 := t1.Add(-time.Minute * 3)
	t5 := t1.Add(-time.Minute * 4)
	t6 := t1.Add(-time.Minute * 5)
	t7 := t1.Add(-time.Minute * 6)
	t8 := t1.Add(-time.Minute * 7)
	t9 := t1.Add(-time.Minute * 8)
	t10 := t1.Add(-time.Minute * 9)
	failureSummary := "unlucky 333"

	insertUploads(t, db.db,
		Upload{ID: 1, Commit: makeCommit(3331), UploadedAt: t1, Root: "sub1/", State: "queued"},
		Upload{ID: 2, UploadedAt: t2, VisibleAtTip: true, State: "errored", FailureSummary: &failureSummary, Indexer: "lsif-tsc"},
		Upload{ID: 3, Commit: makeCommit(3333), UploadedAt: t3, Root: "sub2/", State: "queued"},
		Upload{ID: 4, UploadedAt: t4, State: "queued", RepositoryID: 51},
		Upload{ID: 5, Commit: makeCommit(3333), UploadedAt: t5, Root: "sub1/", VisibleAtTip: true, State: "processing", Indexer: "lsif-tsc"},
		Upload{ID: 6, UploadedAt: t6, Root: "sub2/", State: "processing"},
		Upload{ID: 7, UploadedAt: t7, Root: "sub1/", VisibleAtTip: true, Indexer: "lsif-tsc"},
		Upload{ID: 8, UploadedAt: t8, VisibleAtTip: true, Indexer: "lsif-tsc"},
		Upload{ID: 9, UploadedAt: t9, State: "queued"},
		Upload{ID: 10, UploadedAt: t10, Root: "sub1/", Indexer: "lsif-tsc"},
	)

	testCases := []struct {
		state        string
		term         string
		visibleAtTip bool
		expectedIDs  []int
	}{
		{expectedIDs: []int{1, 2, 3, 5, 6, 7, 8, 9, 10}},
		{state: "completed", expectedIDs: []int{7, 8, 10}},
		{term: "sub", expectedIDs: []int{1, 3, 5, 6, 7, 10}}, // searches root
		{term: "003", expectedIDs: []int{1, 3, 5}},           // searches commits
		{term: "333", expectedIDs: []int{1, 2, 3, 5}},        // searches commits and failure summary
		{term: "tsc", expectedIDs: []int{2, 5, 7, 8, 10}},    // searches indexer
		{visibleAtTip: true, expectedIDs: []int{2, 5, 7, 8}},
	}

	for _, testCase := range testCases {
		name := fmt.Sprintf("state=%s term=%s visibleAtTip=%v", testCase.state, testCase.term, testCase.visibleAtTip)

		t.Run(name, func(t *testing.T) {
			for lo := 0; lo < len(testCase.expectedIDs); lo++ {
				hi := lo + 3
				if hi > len(testCase.expectedIDs) {
					hi = len(testCase.expectedIDs)
				}

				uploads, totalCount, err := db.GetUploadsByRepo(context.Background(), 50, testCase.state, testCase.term, testCase.visibleAtTip, 3, lo)
				if err != nil {
					t.Fatalf("unexpected error getting uploads for repo: %s", err)
				}
				if totalCount != len(testCase.expectedIDs) {
					t.Errorf("unexpected total count. want=%d have=%d", len(testCase.expectedIDs), totalCount)
				}

				var ids []int
				for _, upload := range uploads {
					ids = append(ids, upload.ID)
				}

				if diff := cmp.Diff(testCase.expectedIDs[lo:hi], ids); diff != "" {
					t.Errorf("unexpected upload ids at offset %d (-want +got):\n%s", lo, diff)
				}
			}
		})
	}
}

func TestEnqueue(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	id, closer, err := db.Enqueue(context.Background(), makeCommit(1), "sub/", `{"id": 42}`, 50, "lsif-go")
	if err != nil {
		t.Fatalf("unexpected error enqueueing upload: %s", err)
	}

	// Upload does not exist before transaction commit
	if _, exists, err := db.GetUploadByID(context.Background(), id); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}

	// Commit transaction
	_ = closer.CloseTx(nil)

	rank := 1
	expected := Upload{
		ID:                id,
		Commit:            makeCommit(1),
		Root:              "sub/",
		VisibleAtTip:      false,
		UploadedAt:        time.Time{},
		State:             "queued",
		FailureSummary:    nil,
		FailureStacktrace: nil,
		StartedAt:         nil,
		FinishedAt:        nil,
		TracingContext:    `{"id": 42}`,
		RepositoryID:      50,
		Indexer:           "lsif-go",
		Rank:              &rank,
	}

	if upload, exists, err := db.GetUploadByID(context.Background(), id); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else {
		// Update auto-generated timestamp
		expected.UploadedAt = upload.UploadedAt

		if diff := cmp.Diff(expected, upload); diff != "" {
			t.Errorf("unexpected upload (-want +got):\n%s", diff)
		}
	}
}

func TestEnqueueRollback(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	id, closer, err := db.Enqueue(context.Background(), makeCommit(1), "sub/", `{"id": 42}`, 50, "lsif-go")
	if err != nil {
		t.Fatalf("unexpected error enqueueing upload: %s", err)
	}
	_ = closer.CloseTx(errors.New(""))

	// Upload does not exist after rollback
	if _, exists, err := db.GetUploadByID(context.Background(), id); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}
}

func TestGetStates(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	insertUploads(t, db.db,
		Upload{ID: 1, State: "queued"},
		Upload{ID: 2},
		Upload{ID: 3, State: "processing"},
		Upload{ID: 4, State: "errored"},
	)

	expected := map[int]string{
		1: "queued",
		2: "completed",
		4: "errored",
	}

	if states, err := db.GetStates(context.Background(), []int{1, 2, 4, 6}); err != nil {
		t.Fatalf("unexpected error getting states: %s", err)
	} else if diff := cmp.Diff(expected, states); diff != "" {
		t.Errorf("unexpected upload states (-want +got):\n%s", diff)
	}
}

func TestDeleteUploadByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	insertUploads(t, db.db,
		Upload{ID: 1},
	)

	var called bool
	getTipCommit := func(repositoryID int) (string, error) {
		called = true
		return "", nil
	}

	if found, err := db.DeleteUploadByID(context.Background(), 1, getTipCommit); err != nil {
		t.Fatalf("unexpected error deleting upload: %s", err)
	} else if !found {
		t.Fatalf("expected record to exist")
	} else if called {
		t.Fatalf("unexpected call to getTipCommit")
	}

	// Upload no longer exists
	if _, exists, err := db.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if exists {
		t.Fatal("unexpected record")
	}
}

func TestDeleteUploadByIDMissingRow(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	getTipCommit := func(repositoryID int) (string, error) {
		return "", nil
	}

	if found, err := db.DeleteUploadByID(context.Background(), 1, getTipCommit); err != nil {
		t.Fatalf("unexpected error deleting upload: %s", err)
	} else if found {
		t.Fatalf("unexpected record")
	}
}

func TestDeleteUploadByIDUpdatesVisibility(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	insertUploads(t, db.db,
		Upload{ID: 1, Commit: makeCommit(4), Root: "sub1/", VisibleAtTip: true},
		Upload{ID: 2, Commit: makeCommit(3), Root: "sub2/", VisibleAtTip: true},
		Upload{ID: 3, Commit: makeCommit(2), Root: "sub1/", VisibleAtTip: false},
		Upload{ID: 4, Commit: makeCommit(1), Root: "sub2/", VisibleAtTip: false},
	)

	insertCommits(t, db.db, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(2)},
		makeCommit(4): {makeCommit(3)},
	})

	var called bool
	getTipCommit := func(repositoryID int) (string, error) {
		called = true
		return makeCommit(4), nil
	}

	if found, err := db.DeleteUploadByID(context.Background(), 1, getTipCommit); err != nil {
		t.Fatalf("unexpected error deleting upload: %s", err)
	} else if !found {
		t.Fatalf("expected record to exist")
	} else if !called {
		t.Fatalf("expected call to getTipCommit")
	}

	expected := map[int]bool{2: true, 3: true, 4: false}
	visibilities := getDumpVisibilities(t, db.db)
	if diff := cmp.Diff(expected, visibilities); diff != "" {
		t.Errorf("unexpected visibility (-want +got):\n%s", diff)
	}
}

func TestResetStalled(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	now := time.Unix(1587396557, 0).UTC()
	t1 := now.Add(-time.Second * 6) // old
	t2 := now.Add(-time.Second * 2) // new enough
	t3 := now.Add(-time.Second * 3) // new enough
	t4 := now.Add(-time.Second * 8) // old
	t5 := now.Add(-time.Second * 8) // old

	insertUploads(t, db.db,
		Upload{ID: 1, State: "processing", StartedAt: &t1},
		Upload{ID: 2, State: "processing", StartedAt: &t2},
		Upload{ID: 3, State: "processing", StartedAt: &t3},
		Upload{ID: 4, State: "processing", StartedAt: &t4},
		Upload{ID: 5, State: "processing", StartedAt: &t5},
	)

	tx, err := db.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	// Row lock upload 5 in a transaction which should be skipped by ResetStalled
	if _, err := tx.Query(`SELECT * FROM lsif_uploads WHERE id = 5 FOR UPDATE`); err != nil {
		t.Fatal(err)
	}

	expected := []int{1, 4}

	if ids, err := db.ResetStalled(context.Background(), now); err != nil {
		t.Fatalf("unexpected error resetting stalled uploads: %s", err)
	} else if diff := cmp.Diff(expected, ids); diff != "" {
		t.Errorf("unexpected ids (-want +got):\n%s", diff)
	}
}

package db

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestGetUploadByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

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
		RepositoryID:      123,
		Indexer:           "lsif-go",
		NumParts:          1,
		UploadedParts:     []int{},
		Rank:              nil,
	}

	insertUploads(t, dbconn.Global, expected)

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
	db := testDB()

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(+time.Minute * 5)
	t3 := t1.Add(+time.Minute * 3)
	t4 := t1.Add(+time.Minute * 1)
	t5 := t1.Add(+time.Minute * 4)
	t6 := t1.Add(+time.Minute * 2)

	insertUploads(t, dbconn.Global,
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
	db := testDB()

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

	insertUploads(t, dbconn.Global,
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

func TestQueueSize(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, State: "queued"},
		Upload{ID: 2, State: "errored"},
		Upload{ID: 3, State: "processing"},
		Upload{ID: 4, State: "completed"},
		Upload{ID: 5, State: "completed"},
		Upload{ID: 6, State: "queued"},
		Upload{ID: 7, State: "processing"},
		Upload{ID: 8, State: "completed"},
		Upload{ID: 9, State: "processing"},
		Upload{ID: 10, State: "queued"},
	)

	count, err := db.QueueSize(context.Background())
	if err != nil {
		t.Fatalf("unexpected error getting queue size: %s", err)
	}
	if count != 3 {
		t.Errorf("unexpected count. want=%d have=%d", 3, count)
	}
}

func TestInsertUploadUploading(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	id, err := db.InsertUpload(context.Background(), Upload{
		Commit:       makeCommit(1),
		Root:         "sub/",
		State:        "uploading",
		RepositoryID: 50,
		Indexer:      "lsif-go",
		NumParts:     3,
	})
	if err != nil {
		t.Fatalf("unexpected error enqueueing upload: %s", err)
	}

	expected := Upload{
		ID:                id,
		Commit:            makeCommit(1),
		Root:              "sub/",
		VisibleAtTip:      false,
		UploadedAt:        time.Time{},
		State:             "uploading",
		FailureSummary:    nil,
		FailureStacktrace: nil,
		StartedAt:         nil,
		FinishedAt:        nil,
		RepositoryID:      50,
		Indexer:           "lsif-go",
		NumParts:          3,
		UploadedParts:     []int{},
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

func TestInsertUploadQueued(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	id, err := db.InsertUpload(context.Background(), Upload{
		Commit:        makeCommit(1),
		Root:          "sub/",
		State:         "queued",
		RepositoryID:  50,
		Indexer:       "lsif-go",
		NumParts:      1,
		UploadedParts: []int{0},
	})
	if err != nil {
		t.Fatalf("unexpected error enqueueing upload: %s", err)
	}

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
		RepositoryID:      50,
		Indexer:           "lsif-go",
		NumParts:          1,
		UploadedParts:     []int{0},
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

func TestMarkQueued(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	insertUploads(t, dbconn.Global, Upload{ID: 1, State: "uploading"})

	if err := db.MarkQueued(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error marking upload as queued: %s", err)
	}

	if upload, exists, err := db.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if upload.State != "queued" {
		t.Errorf("unexpected state. want=%q have=%q", "queued", upload.State)
	}
}

func TestAddUploadPart(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	insertUploads(t, dbconn.Global, Upload{ID: 1, State: "uploading"})

	for _, part := range []int{1, 5, 2, 3, 2, 2, 1, 6} {
		if err := db.AddUploadPart(context.Background(), 1, part); err != nil {
			t.Fatalf("unexpected error adding upload part: %s", err)
		}
	}
	if upload, exists, err := db.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else {
		sort.Ints(upload.UploadedParts)
		if diff := cmp.Diff([]int{1, 2, 3, 5, 6}, upload.UploadedParts); diff != "" {
			t.Errorf("unexpected upload parts (-want +got):\n%s", diff)
		}
	}
}

func TestMarkComplete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	insertUploads(t, dbconn.Global, Upload{ID: 1, State: "queued"})

	if err := db.MarkComplete(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error marking upload as completed: %s", err)
	}

	if upload, exists, err := db.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if upload.State != "completed" {
		t.Errorf("unexpected state. want=%q have=%q", "completed", upload.State)
	}
}

func TestMarkErrored(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := &dbImpl{db: dbconn.Global}

	insertUploads(t, dbconn.Global, Upload{ID: 1, State: "queued"})

	if err := db.MarkErrored(context.Background(), 1, "oops", ""); err != nil {
		t.Fatalf("unexpected error marking upload as errored: %s", err)
	}

	if upload, exists, err := db.GetUploadByID(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error getting upload: %s", err)
	} else if !exists {
		t.Fatal("expected record to exist")
	} else if upload.State != "errored" {
		t.Errorf("unexpected state. want=%q have=%q", "errored", upload.State)
	}
}

func TestDequeueConversionSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	// Add dequeueable upload
	insertUploads(t, dbconn.Global, Upload{ID: 1, State: "queued"})

	upload, tx, ok, err := db.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing upload: %s", err)
	}
	if !ok {
		t.Fatalf("expected something to be dequeueable")
	}

	if upload.ID != 1 {
		t.Errorf("unexpected upload id. want=%d have=%d", 1, upload.ID)
	}
	if upload.State != "processing" {
		t.Errorf("unexpected state. want=%s have=%s", "processing", upload.State)
	}

	if state, _, err := scanFirstString(dbconn.Global.Query("SELECT state FROM lsif_uploads WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting state: %s", err)
	} else if state != "processing" {
		t.Errorf("unexpected state outside of txn. want=%s have=%s", "processing", state)
	}

	if err := tx.MarkComplete(context.Background(), upload.ID); err != nil {
		t.Fatalf("unexpected error marking upload complete: %s", err)
	}
	_ = tx.Done(nil)

	if state, _, err := scanFirstString(dbconn.Global.Query("SELECT state FROM lsif_uploads WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting state: %s", err)
	} else if state != "completed" {
		t.Errorf("unexpected state outside of txn. want=%s have=%s", "completed", state)
	}
}

func TestDequeueConversionError(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	// Add dequeueable upload
	insertUploads(t, dbconn.Global, Upload{ID: 1, State: "queued"})

	upload, tx, ok, err := db.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing upload: %s", err)
	}
	if !ok {
		t.Fatalf("expected something to be dequeueable")
	}

	if upload.ID != 1 {
		t.Errorf("unexpected upload id. want=%d have=%d", 1, upload.ID)
	}
	if upload.State != "processing" {
		t.Errorf("unexpected state. want=%s have=%s", "processing", upload.State)
	}

	if state, _, err := scanFirstString(dbconn.Global.Query("SELECT state FROM lsif_uploads WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting state: %s", err)
	} else if state != "processing" {
		t.Errorf("unexpected state outside of txn. want=%s have=%s", "processing", state)
	}

	if err := tx.MarkErrored(context.Background(), upload.ID, "test summary", "test stacktrace"); err != nil {
		t.Fatalf("unexpected error marking upload complete: %s", err)
	}
	_ = tx.Done(nil)

	if state, _, err := scanFirstString(dbconn.Global.Query("SELECT state FROM lsif_uploads WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting state: %s", err)
	} else if state != "errored" {
		t.Errorf("unexpected state outside of txn. want=%s have=%s", "errored", state)
	}

	if summary, _, err := scanFirstString(dbconn.Global.Query("SELECT failure_summary FROM lsif_uploads WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting failure_summary: %s", err)
	} else if summary != "test summary" {
		t.Errorf("unexpected failure summary outside of txn. want=%s have=%s", "test summary", summary)
	}

	if stacktrace, _, err := scanFirstString(dbconn.Global.Query("SELECT failure_stacktrace FROM lsif_uploads WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting failure_stacktrace: %s", err)
	} else if stacktrace != "test stacktrace" {
		t.Errorf("unexpected failure stacktrace outside of txn. want=%s have=%s", "test stacktrace", stacktrace)
	}
}

func TestDequeueWithSavepointRollback(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	// Add dequeueable upload
	insertUploads(t, dbconn.Global, Upload{ID: 1, State: "queued", Indexer: "lsif-go"})

	ctx := context.Background()
	upload, tx, ok, err := db.Dequeue(ctx)
	if err != nil {
		t.Fatalf("unexpected error dequeueing upload: %s", err)
	}
	if !ok {
		t.Fatalf("expected something to be dequeueable")
	}

	savepointID, err := tx.Savepoint(ctx)
	if err != nil {
		t.Fatalf("unexpected error creating savepoint: %s", err)
	}

	// alter record in the underlying transacted db
	if err := unwrapDB(tx).exec(ctx, sqlf.Sprintf(`UPDATE lsif_uploads SET indexer = 'lsif-tsc' WHERE id = 1`)); err != nil {
		t.Fatalf("unexpected error altering record: %s", err)
	}

	// undo alteration
	if err := tx.RollbackToSavepoint(ctx, savepointID); err != nil {
		t.Fatalf("unexpected error rolling back to savepoint: %s", err)
	}

	if err := tx.MarkComplete(ctx, upload.ID); err != nil {
		t.Fatalf("unexpected error marking upload complete: %s", err)
	}
	if err := tx.Done(nil); err != nil {
		t.Fatalf("unexpected error closing transaction: %s", err)
	}

	if indexerName, _, err := scanFirstString(dbconn.Global.Query("SELECT indexer FROM lsif_uploads WHERE id = 1")); err != nil {
		t.Errorf("unexpected error getting indexer: %s", err)
	} else if indexerName != "lsif-go" {
		t.Errorf("unexpected failure summary outside of txn. want=%s have=%s", "lsif-go", indexerName)
	}
}

func TestDequeueSkipsLocked(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	t1 := time.Unix(1587396557, 0).UTC()
	t2 := t1.Add(time.Minute)
	t3 := t2.Add(time.Minute)
	insertUploads(
		t,
		dbconn.Global,
		Upload{ID: 1, State: "queued", UploadedAt: t1},
		Upload{ID: 2, State: "processing", UploadedAt: t2},
		Upload{ID: 3, State: "queued", UploadedAt: t3},
	)

	tx1, err := dbconn.Global.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx1.Rollback() }()

	// Row lock upload 1 in a transaction which should be skipped by ResetStalled
	if _, err := tx1.Query(`SELECT * FROM lsif_uploads WHERE id = 1 FOR UPDATE`); err != nil {
		t.Fatal(err)
	}

	upload, tx2, ok, err := db.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing upload: %s", err)
	}
	if !ok {
		t.Fatalf("expected something to be dequeueable")
	}
	defer func() { _ = tx2.Done(nil) }()

	if upload.ID != 3 {
		t.Errorf("unexpected upload id. want=%d have=%d", 3, upload.ID)
	}
	if upload.State != "processing" {
		t.Errorf("unexpected state. want=%s have=%s", "processing", upload.State)
	}
}

func TestDequeueEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	_, tx, ok, err := db.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("unexpected error dequeueing upload: %s", err)
	}
	if ok {
		_ = tx.Done(nil)
		t.Fatalf("unexpected dequeue")
	}
}

func TestGetStates(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	insertUploads(t, dbconn.Global,
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
	db := testDB()

	insertUploads(t, dbconn.Global,
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
	db := testDB()

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
	db := testDB()

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, Commit: makeCommit(4), Root: "sub1/", VisibleAtTip: true},
		Upload{ID: 2, Commit: makeCommit(3), Root: "sub2/", VisibleAtTip: true},
		Upload{ID: 3, Commit: makeCommit(2), Root: "sub1/", VisibleAtTip: false},
		Upload{ID: 4, Commit: makeCommit(1), Root: "sub2/", VisibleAtTip: false},
	)

	if err := db.UpdateCommits(context.Background(), 50, map[string][]string{
		makeCommit(1): {},
		makeCommit(2): {makeCommit(1)},
		makeCommit(3): {makeCommit(2)},
		makeCommit(4): {makeCommit(3)},
	}); err != nil {
		t.Fatalf("unexpected error updating commits: %s", err)
	}

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
	visibilities := getDumpVisibilities(t, dbconn.Global)
	if diff := cmp.Diff(expected, visibilities); diff != "" {
		t.Errorf("unexpected visibility (-want +got):\n%s", diff)
	}
}

func TestResetStalled(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := testDB()

	now := time.Unix(1587396557, 0).UTC()
	t1 := now.Add(-time.Second * 6) // old
	t2 := now.Add(-time.Second * 2) // new enough
	t3 := now.Add(-time.Second * 3) // new enough
	t4 := now.Add(-time.Second * 8) // old
	t5 := now.Add(-time.Second * 8) // old

	insertUploads(t, dbconn.Global,
		Upload{ID: 1, State: "processing", StartedAt: &t1},
		Upload{ID: 2, State: "processing", StartedAt: &t2},
		Upload{ID: 3, State: "processing", StartedAt: &t3},
		Upload{ID: 4, State: "processing", StartedAt: &t4},
		Upload{ID: 5, State: "processing", StartedAt: &t5},
	)

	tx, err := dbconn.Global.BeginTx(context.Background(), nil)
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

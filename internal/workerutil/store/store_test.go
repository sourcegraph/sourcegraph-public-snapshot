package store

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestStoreDequeueState(t *testing.T) {
	setupStoreTest(t)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state, uploaded_at)
		VALUES
			(1, 'queued', NOW() - '1 minute'::interval),
			(2, 'queued', NOW() - '2 minute'::interval),
			(3, 'state2', NOW() - '3 minute'::interval),
			(4, 'queued', NOW() - '4 minute'::interval),
			(5, 'state2', NOW() - '5 minute'::interval)
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	record, tx, ok, err := testStore(defaultTestStoreOptions).Dequeue(context.Background(), nil)
	assertDequeueRecordResult(t, 4, record, tx, ok, err)
}

func TestStoreDequeueOrder(t *testing.T) {
	setupStoreTest(t)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state, uploaded_at)
		VALUES
			(1, 'queued', NOW() - '2 minute'::interval),
			(2, 'queued', NOW() - '5 minute'::interval),
			(3, 'queued', NOW() - '3 minute'::interval),
			(4, 'queued', NOW() - '1 minute'::interval),
			(5, 'queued', NOW() - '4 minute'::interval)
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	record, tx, ok, err := testStore(defaultTestStoreOptions).Dequeue(context.Background(), nil)
	assertDequeueRecordResult(t, 2, record, tx, ok, err)
}

func TestStoreDequeueConditions(t *testing.T) {
	setupStoreTest(t)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state, uploaded_at)
		VALUES
			(1, 'queued', NOW() - '1 minute'::interval),
			(2, 'queued', NOW() - '2 minute'::interval),
			(3, 'queued', NOW() - '3 minute'::interval),
			(4, 'queued', NOW() - '4 minute'::interval),
			(5, 'queued', NOW() - '5 minute'::interval)
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	conditions := []*sqlf.Query{sqlf.Sprintf("w.id < 4")}
	record, tx, ok, err := testStore(defaultTestStoreOptions).Dequeue(context.Background(), conditions)
	assertDequeueRecordResult(t, 3, record, tx, ok, err)
}

func TestStoreDequeueDelay(t *testing.T) {
	setupStoreTest(t)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state, uploaded_at, process_after)
		VALUES
			(1, 'queued', NOW() - '1 minute'::interval, NULL),
			(2, 'queued', NOW() - '2 minute'::interval, NULL),
			(3, 'queued', NOW() - '3 minute'::interval, NOW() + '2 minute'::interval),
			(4, 'queued', NOW() - '4 minute'::interval, NULL),
			(5, 'queued', NOW() - '5 minute'::interval, NOW() + '1 minute'::interval)
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	record, tx, ok, err := testStore(defaultTestStoreOptions).Dequeue(context.Background(), nil)
	assertDequeueRecordResult(t, 4, record, tx, ok, err)
}

func TestStoreDequeueView(t *testing.T) {
	setupStoreTest(t)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state, uploaded_at)
		VALUES
			(1, 'queued', NOW() - '1 minute'::interval),
			(2, 'queued', NOW() - '2 minute'::interval),
			(3, 'queued', NOW() - '3 minute'::interval),
			(4, 'queued', NOW() - '4 minute'::interval),
			(5, 'queued', NOW() - '5 minute'::interval)
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	options := StoreOptions{
		TableName:         "workerutil_test w",
		ViewName:          "workerutil_test_view v",
		Scan:              testScanFirstRecordView,
		OrderByExpression: sqlf.Sprintf("v.uploaded_at"),
		ColumnExpressions: []*sqlf.Query{
			sqlf.Sprintf("v.id"),
			sqlf.Sprintf("v.state"),
			sqlf.Sprintf("v.new_field"),
		},
		StalledMaxAge: time.Second * 5,
		MaxNumResets:  5,
	}

	conditions := []*sqlf.Query{sqlf.Sprintf("v.new_field < 15")}
	record, tx, ok, err := testStore(options).Dequeue(context.Background(), conditions)
	assertDequeueRecordViewResult(t, 2, 14, record, tx, ok, err)
}

func TestStoreDequeueConcurrent(t *testing.T) {
	setupStoreTest(t)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state, uploaded_at)
		VALUES
			(1, 'queued', NOW() - '2 minute'::interval),
			(2, 'queued', NOW() - '1 minute'::interval)
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	store := testStore(defaultTestStoreOptions)

	// Worker A
	record1, tx1, ok, err := store.Dequeue(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !ok {
		t.Fatalf("expected a dequeueable record")
	}
	defer func() { _ = tx1.Done(nil) }()

	// Worker B
	record2, tx2, ok, err := store.Dequeue(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !ok {
		t.Fatalf("expected a second dequeueable record")
	}
	defer func() { _ = tx2.Done(nil) }()

	if val := record1.(TestRecord).ID; val != 1 {
		t.Errorf("unexpected id. want=%d have=%d", 1, val)
	}
	if val := record2.(TestRecord).ID; val != 2 {
		t.Errorf("unexpected id. want=%d have=%d", 2, val)
	}

	// Worker C
	_, _, ok, err = store.Dequeue(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if ok {
		t.Fatalf("did not expect a third dequeueable record")
	}
}

func TestStoreRequeue(t *testing.T) {
	setupStoreTest(t)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state)
		VALUES
			(1, 'processing')
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	after := testNow().Add(time.Hour)

	if err := testStore(defaultTestStoreOptions).Requeue(context.Background(), 1, after); err != nil {
		t.Fatalf("unexpected error requeueing index: %s", err)
	}

	rows, err := dbconn.Global.Query(`SELECT state, process_after FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fatalf("unexpected error querying record: %s", err)
	}
	defer func() { _ = basestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fatal("expected record to exist")
	}

	var state string
	var processAfter *time.Time

	if err := rows.Scan(&state, &processAfter); err != nil {
		t.Fatalf("unexpected error scanning record: %s", err)
	}
	if state != "queued" {
		t.Errorf("unexpected state. want=%q have=%q", "queued", state)
	}
	if processAfter == nil || *processAfter != after {
		t.Errorf("unexpected process after. want=%s have=%s", after, processAfter)
	}
}

func TestStoreMarkComplete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore(defaultTestStoreOptions)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state)
		VALUES
			(1, 'processing')
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	marked, err := store.MarkComplete(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error marking upload as completed: %s", err)
	}
	if !marked {
		t.Fatalf("expected record to be marked")
	}

	rows, err := dbconn.Global.Query(`SELECT state, failure_message FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fatalf("unexpected error querying record: %s", err)
	}
	defer func() { _ = basestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fatal("expected record to exist")
	}

	var state string
	var failureMessage *string
	if err := rows.Scan(&state, &failureMessage); err != nil {
		t.Fatalf("unexpected error scanning record: %s", err)
	}
	if state != "completed" {
		t.Errorf("unexpected state. want=%q have=%q", "completed", state)
	}
	if failureMessage != nil {
		t.Errorf("unexpected failure message. want=%v have=%v", nil, failureMessage)
	}
}

func TestStoreMarkCompleteNotProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore(defaultTestStoreOptions)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state, failure_message)
		VALUES
			(1, 'errored', 'old message')
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	marked, err := store.MarkComplete(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error marking upload as completed: %s", err)
	}
	if marked {
		t.Fatalf("expected record not to be marked")
	}

	rows, err := dbconn.Global.Query(`SELECT state, failure_message FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fatalf("unexpected error querying record: %s", err)
	}
	defer func() { _ = basestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fatal("expected record to exist")
	}

	var state string
	var failureMessage *string
	if err := rows.Scan(&state, &failureMessage); err != nil {
		t.Fatalf("unexpected error scanning record: %s", err)
	}
	if state != "errored" {
		t.Errorf("unexpected state. want=%q have=%q", "errored", state)
	}
	if failureMessage == nil || *failureMessage != "old message" {
		t.Errorf("unexpected failure message. want=%v have=%v", "old message", failureMessage)
	}
}

func TestStoreMarkErrored(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore(defaultTestStoreOptions)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state)
		VALUES
			(1, 'processing')
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	marked, err := store.MarkErrored(context.Background(), 1, "new message")
	if err != nil {
		t.Fatalf("unexpected error marking upload as completed: %s", err)
	}
	if !marked {
		t.Fatalf("expected record to be marked")
	}

	rows, err := dbconn.Global.Query(`SELECT state, failure_message FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fatalf("unexpected error querying record: %s", err)
	}
	defer func() { _ = basestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fatal("expected record to exist")
	}

	var state string
	var failureMessage *string
	if err := rows.Scan(&state, &failureMessage); err != nil {
		t.Fatalf("unexpected error scanning record: %s", err)
	}
	if state != "errored" {
		t.Errorf("unexpected state. want=%q have=%q", "errored", state)
	}
	if failureMessage == nil || *failureMessage != "new message" {
		t.Errorf("unexpected failure message. want=%v have=%v", "new message", failureMessage)
	}
}

func TestStoreMarkErroredAlreadyCompleted(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore(defaultTestStoreOptions)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state)
		VALUES
			(1, 'completed')
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	marked, err := store.MarkErrored(context.Background(), 1, "new message")
	if err != nil {
		t.Fatalf("unexpected error marking upload as completed: %s", err)
	}
	if !marked {
		t.Fatalf("expected record to be marked")
	}

	rows, err := dbconn.Global.Query(`SELECT state, failure_message FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fatalf("unexpected error querying record: %s", err)
	}
	defer func() { _ = basestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fatal("expected record to exist")
	}

	var state string
	var failureMessage *string
	if err := rows.Scan(&state, &failureMessage); err != nil {
		t.Fatalf("unexpected error scanning record: %s", err)
	}
	if state != "errored" {
		t.Errorf("unexpected state. want=%q have=%q", "errored", state)
	}
	if failureMessage == nil || *failureMessage != "new message" {
		t.Errorf("unexpected failure message. want=%v have=%v", "new message", failureMessage)
	}
}

func TestStoreMarkErroredAlreadyErrored(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore(defaultTestStoreOptions)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state, failure_message)
		VALUES
			(1, 'errored', 'old message')
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	marked, err := store.MarkErrored(context.Background(), 1, "new message")
	if err != nil {
		t.Fatalf("unexpected error marking upload as completed: %s", err)
	}
	if marked {
		t.Fatalf("expected record not to be marked")
	}

	rows, err := dbconn.Global.Query(`SELECT state, failure_message FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fatalf("unexpected error querying record: %s", err)
	}
	defer func() { _ = basestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fatal("expected record to exist")
	}

	var state string
	var failureMessage *string
	if err := rows.Scan(&state, &failureMessage); err != nil {
		t.Fatalf("unexpected error scanning record: %s", err)
	}
	if state != "errored" {
		t.Errorf("unexpected state. want=%q have=%q", "errored", state)
	}
	if failureMessage == nil || *failureMessage != "old message" {
		t.Errorf("unexpected failure message. want=%v have=%v", "old message", failureMessage)
	}
}

func TestStoreResetStalled(t *testing.T) {
	setupStoreTest(t)

	if _, err := dbconn.Global.Exec(`
		INSERT INTO workerutil_test (id, state, started_at, num_resets)
		VALUES
			(1, 'processing', NOW() - '6 second'::interval, 1),
			(2, 'processing', NOW() - '2 second'::interval, 0),
			(3, 'processing', NOW() - '3 second'::interval, 0),
			(4, 'processing', NOW() - '8 second'::interval, 0),
			(5, 'processing', NOW() - '8 second'::interval, 0),
			(6, 'processing', NOW() - '6 second'::interval, 5),
			(7, 'processing', NOW() - '8 second'::interval, 5)
	`); err != nil {
		t.Fatalf("unexpected error inserting records: %s", err)
	}

	tx, err := dbconn.Global.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	// Row lock record 5 in a transaction which should be skipped by ResetStalled
	if _, err := tx.Exec(`SELECT * FROM workerutil_test WHERE id = 5 FOR UPDATE`); err != nil {
		t.Fatal(err)
	}

	resetIDs, erroredIDs, err := testStore(defaultTestStoreOptions).ResetStalled(context.Background())
	if err != nil {
		t.Fatalf("unexpected error resetting stalled records: %s", err)
	}
	sort.Ints(resetIDs)
	sort.Ints(erroredIDs)

	if diff := cmp.Diff([]int{1, 4}, resetIDs); diff != "" {
		t.Errorf("unexpected reset ids (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff([]int{6, 7}, erroredIDs); diff != "" {
		t.Errorf("unexpected errored ids (-want +got):\n%s", diff)
	}

	rows, err := dbconn.Global.Query(`SELECT state, num_resets FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fatalf("unexpected error querying record: %s", err)
	}
	defer func() { _ = basestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fatal("expected record to exist")
	}

	var state string
	var numResets int
	if err := rows.Scan(&state, &numResets); err != nil {
		t.Fatalf("unexpected error scanning record: %s", err)
	}
	if state != "queued" {
		t.Errorf("unexpected state. want=%q have=%q", "queued", state)
	}
	if numResets != 2 {
		t.Errorf("unexpected num resets. want=%d have=%d", 2, numResets)
	}

	rows, err = dbconn.Global.Query(`SELECT state FROM workerutil_test WHERE id = 6`)
	if err != nil {
		t.Fatalf("unexpected error querying record: %s", err)
	}
	defer func() { _ = basestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fatal("expected record to exist")
	}

	if err := rows.Scan(&state); err != nil {
		t.Fatalf("unexpected error scanning record: %s", err)
	}
	if state != "errored" {
		t.Errorf("unexpected state. want=%q have=%q", "errored", state)
	}
}

package store

import (
	"database/sql"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func testStore(db *sql.DB, options Options) *store {
	return newStore(basestore.NewHandleWithDB(db, sql.TxOptions{}), options, &observation.TestContext)
}

type TestRecord struct {
	ID            int
	State         string
	ExecutionLogs []ExecutionLogEntry
}

func (v TestRecord) RecordID() int {
	return v.ID
}

func testScanFirstRecord(rows *sql.Rows, queryErr error) (v workerutil.Record, _ bool, err error) {
	if queryErr != nil {
		return nil, false, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		var record TestRecord
		if err := rows.Scan(&record.ID, &record.State, pq.Array(&record.ExecutionLogs)); err != nil {
			return nil, false, err
		}

		return record, true, nil
	}

	return nil, false, nil
}

type TestRecordView struct {
	ID       int
	State    string
	NewField int
}

func (v TestRecordView) RecordID() int {
	return v.ID
}

func testScanFirstRecordView(rows *sql.Rows, queryErr error) (v workerutil.Record, exists bool, err error) {
	if queryErr != nil {
		return nil, false, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		var record TestRecordView
		if err := rows.Scan(&record.ID, &record.State, &record.NewField); err != nil {
			return nil, false, err
		}

		return record, true, nil
	}

	return nil, false, nil
}

type TestRecordRetry struct {
	ID        int
	State     string
	NumResets int
}

func (v TestRecordRetry) RecordID() int {
	return v.ID
}

func testScanFirstRecordRetry(rows *sql.Rows, queryErr error) (v workerutil.Record, exists bool, err error) {
	if queryErr != nil {
		return nil, false, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if rows.Next() {
		var record TestRecordRetry
		if err := rows.Scan(&record.ID, &record.State, &record.NumResets); err != nil {
			return nil, false, err
		}

		return record, true, nil
	}

	return nil, false, nil
}

func setupStoreTest(t *testing.T) *sql.DB {
	db := dbtest.NewDB(t)

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS workerutil_test (
			id                integer NOT NULL,
			state             text NOT NULL,
			failure_message   text,
			started_at        timestamp with time zone,
			last_heartbeat_at timestamp with time zone,
			finished_at       timestamp with time zone,
			process_after     timestamp with time zone,
			num_resets        integer NOT NULL default 0,
			num_failures      integer NOT NULL default 0,
			created_at        timestamp with time zone NOT NULL default NOW(),
			execution_logs    json[],
			worker_hostname   text NOT NULL default ''
		)
	`); err != nil {
		t.Fatalf("unexpected error creating test table: %s", err)
	}

	if _, err := db.Exec(`
		CREATE OR REPLACE VIEW workerutil_test_view AS (
			SELECT w.*, (w.id * 7) as new_field FROM workerutil_test w
		)
	`); err != nil {
		t.Fatalf("unexpected error creating test table: %s", err)
	}
	return db
}

func defaultTestStoreOptions(clock glock.Clock) Options {
	return Options{
		Name:              "test",
		TableName:         "workerutil_test w",
		Scan:              testScanFirstRecord,
		OrderByExpression: sqlf.Sprintf("w.created_at"),
		ColumnExpressions: []*sqlf.Query{
			sqlf.Sprintf("w.id"),
			sqlf.Sprintf("w.state"),
			sqlf.Sprintf("w.execution_logs"),
		},
		AlternateColumnNames: map[string]string{
			"queued_at": "created_at",
		},
		StalledMaxAge: time.Second * 5,
		MaxNumResets:  5,
		MaxNumRetries: 3,
		clock:         clock,
	}
}

func assertDequeueRecordResult(t *testing.T, expectedID int, record any, ok bool, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !ok {
		t.Fatalf("expected a dequeueable record")
	}

	if val := record.(TestRecord).ID; val != expectedID {
		t.Errorf("unexpected id. want=%d have=%d", expectedID, val)
	}
	if val := record.(TestRecord).State; val != "processing" {
		t.Errorf("unexpected state. want=%s have=%s", "processing", val)
	}
}

func assertDequeueRecordResultLogCount(t *testing.T, expectedLogCount int, record any) {
	if val := len(record.(TestRecord).ExecutionLogs); val != expectedLogCount {
		t.Errorf("unexpected count of logs. want=%d have=%d", expectedLogCount, val)
	}
}

func assertDequeueRecordViewResult(t *testing.T, expectedID, expectedNewField int, record any, ok bool, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !ok {
		t.Fatalf("expected a dequeueable record")
	}

	if val := record.(TestRecordView).ID; val != expectedID {
		t.Errorf("unexpected id. want=%d have=%d", expectedID, val)
	}
	if val := record.(TestRecordView).State; val != "processing" {
		t.Errorf("unexpected state. want=%s have=%s", "processing", val)
	}
	if val := record.(TestRecordView).NewField; val != expectedNewField {
		t.Errorf("unexpected new field. want=%d have=%d", expectedNewField, val)
	}
}

func assertDequeueRecordRetryResult(t *testing.T, expectedID, record any, ok bool, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !ok {
		t.Fatalf("expected a dequeueable record")
	}

	if val := record.(TestRecordRetry).ID; val != expectedID {
		t.Errorf("unexpected id. want=%d have=%d", expectedID, val)
	}
	if val := record.(TestRecordRetry).State; val != "processing" {
		t.Errorf("unexpected state. want=%s have=%s", "processing", val)
	}
}

func testNow() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

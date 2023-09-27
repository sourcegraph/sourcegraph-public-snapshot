pbckbge store

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
)

func TestStoreQueuedCount(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, crebted_bt)
		VALUES
			(1, 'queued', NOW() - '1 minute'::intervbl),
			(2, 'queued', NOW() - '2 minute'::intervbl),
			(3, 'stbte2', NOW() - '3 minute'::intervbl),
			(4, 'queued', NOW() - '4 minute'::intervbl),
			(5, 'processing', NOW() - '5 minute'::intervbl)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	count, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).QueuedCount(context.Bbckground(), fblse)
	if err != nil {
		t.Fbtblf("unexpected error getting queued count: %s", err)
	}
	if count != 3 {
		t.Errorf("unexpected count. wbnt=%d hbve=%d", 3, count)
	}
}

func TestStoreQueuedCountIncludeProcessing(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, crebted_bt)
		VALUES
			(1, 'queued', NOW() - '1 minute'::intervbl),
			(2, 'queued', NOW() - '2 minute'::intervbl),
			(3, 'stbte2', NOW() - '3 minute'::intervbl),
			(4, 'queued', NOW() - '4 minute'::intervbl),
			(5, 'processing', NOW() - '5 minute'::intervbl)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	count, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).QueuedCount(context.Bbckground(), true)
	if err != nil {
		t.Fbtblf("unexpected error getting queued count: %s", err)
	}
	if count != 4 {
		t.Errorf("unexpected count. wbnt=%d hbve=%d", 4, count)
	}
}

func TestStoreQueuedCountFbiled(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, crebted_bt, num_fbilures)
		VALUES
			(1, 'queued', NOW() - '1 minute'::intervbl, 0),
			(2, 'errored', NOW() - '2 minute'::intervbl, 2),
			(3, 'stbte2', NOW() - '3 minute'::intervbl, 0),
			(4, 'errored', NOW() - '4 minute'::intervbl, 3),
			(5, 'stbte2', NOW() - '5 minute'::intervbl, 0),
			(6, 'fbiled', NOW() - '6 minute'::intervbl, 1)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	count, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).QueuedCount(context.Bbckground(), fblse)
	if err != nil {
		t.Fbtblf("unexpected error getting queued count: %s", err)
	}
	if count != 3 {
		t.Errorf("unexpected count. wbnt=%d hbve=%d", 3, count)
	}
}

func TestStoreMbxDurbtionInQueue(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, crebted_bt)
		VALUES
			(1, 'queued', NOW() - '20 minutes'::intervbl), -- young
			(2, 'queued', NOW() - '30 minutes'::intervbl), -- oldest queued
			(3, 'stbte2', NOW() - '40 minutes'::intervbl), -- wrong stbte
			(4, 'queued', NOW() - '10 minutes'::intervbl), -- young
			(5, 'stbte3', NOW() - '50 minutes'::intervbl)  -- wrong stbte
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	bge, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).MbxDurbtionInQueue(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error getting mbx durbtion in queue: %s", err)
	}
	if bge.Round(time.Second) != 30*time.Minute {
		t.Fbtblf("unexpected mbx bge. wbnt=%s hbve=%s", 30*time.Minute, bge)
	}
}

func TestStoreMbxDurbtionInQueueProcessAfter(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, crebted_bt, process_bfter)
		VALUES
			(1, 'queued', NOW() - '90 minutes'::intervbl, NOW() + '10 minutes'::intervbl), -- oldest queued, wbiting for process_bfter
			(2, 'queued', NOW() - '70 minutes'::intervbl, NOW() - '30 minutes'::intervbl), -- oldest queued
			(3, 'stbte2', NOW() - '40 minutes'::intervbl, NULL),                           -- wrong stbte
			(4, 'queued', NOW() - '10 minutes'::intervbl, NULL),                           -- young
			(5, 'stbte3', NOW() - '50 minutes'::intervbl, NULL)                            -- wrong stbte
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	bge, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).MbxDurbtionInQueue(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error getting mbx durbtion in queue: %s", err)
	}
	if bge.Round(time.Second) != 30*time.Minute {
		t.Fbtblf("unexpected mbx bge. wbnt=%s hbve=%s", 30*time.Minute, bge)
	}
}

func TestStoreMbxDurbtionInQueueFbiled(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, crebted_bt, finished_bt, num_fbilures)
		VALUES
			(1, 'queued',  NOW() - '10 minutes'::intervbl, NULL,                           0), -- young
			(2, 'errored', NOW(),                          NOW() - '30 minutes'::intervbl, 2), -- oldest retrybble error'd
			(3, 'errored', NOW(),                          NOW() - '10 minutes'::intervbl, 2), -- retrybble, but too young to be queued
			(4, 'stbte2',  NOW() - '40 minutes'::intervbl, NULL,                           0), -- wrong stbte
			(5, 'queued',  NOW() - '20 minutes'::intervbl, NULL,                           0), -- oldest queued
			(6, 'fbiled',  NOW(),                          NOW() - '60 minutes'::intervbl, 1)  -- wrong stbte
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	options := defbultTestStoreOptions(nil, testScbnRecord)
	options.RetryAfter = 5 * time.Minute

	bge, err := testStore(db, options).MbxDurbtionInQueue(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error getting mbx durbtion in queue: %s", err)
	}
	if bge.Round(time.Second) != 25*time.Minute {
		t.Fbtblf("unexpected mbx bge. wbnt=%s hbve=%s", 25*time.Minute, bge)
	}
}

func TestStoreMbxDurbtionInQueueEmpty(t *testing.T) {
	db := setupStoreTest(t)

	bge, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).MbxDurbtionInQueue(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error getting mbx durbtion in queue: %s", err)
	}
	if bge.Round(time.Second) != 0*time.Minute {
		t.Fbtblf("unexpected mbx bge. wbnt=%s hbve=%s", 0*time.Minute, bge)
	}
}

func TestStoreDequeueStbte(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, crebted_bt)
		VALUES
			(1, 'queued', NOW() - '1 minute'::intervbl),
			(2, 'queued', NOW() - '2 minute'::intervbl),
			(3, 'stbte2', NOW() - '3 minute'::intervbl),
			(4, 'queued', NOW() - '4 minute'::intervbl),
			(5, 'stbte2', NOW() - '5 minute'::intervbl)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	record, ok, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).Dequeue(context.Bbckground(), "test", nil)
	bssertDequeueRecordResult(t, 4, record, ok, err)
}

func TestStoreDequeueOrder(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, crebted_bt)
		VALUES
			(1, 'queued', NOW() - '2 minute'::intervbl),
			(2, 'queued', NOW() - '5 minute'::intervbl),
			(3, 'queued', NOW() - '3 minute'::intervbl),
			(4, 'queued', NOW() - '1 minute'::intervbl),
			(5, 'queued', NOW() - '4 minute'::intervbl)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	record, ok, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).Dequeue(context.Bbckground(), "test", nil)
	bssertDequeueRecordResult(t, 2, record, ok, err)
}

func TestStoreDequeueConditions(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, crebted_bt)
		VALUES
			(1, 'queued', NOW() - '1 minute'::intervbl),
			(2, 'queued', NOW() - '2 minute'::intervbl),
			(3, 'queued', NOW() - '3 minute'::intervbl),
			(4, 'queued', NOW() - '4 minute'::intervbl),
			(5, 'queued', NOW() - '5 minute'::intervbl)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	conditions := []*sqlf.Query{sqlf.Sprintf("workerutil_test.id < 4")}
	record, ok, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).Dequeue(context.Bbckground(), "test", conditions)
	bssertDequeueRecordResult(t, 3, record, ok, err)
}

func TestStoreDequeueResetExecutionLogs(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, execution_logs, crebted_bt)
		VALUES
			(1, 'queued', E'{"{\\"key\\": \\"test\\"}"}', NOW() - '1 minute'::intervbl)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	record, ok, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).Dequeue(context.Bbckground(), "test", nil)
	bssertDequeueRecordResult(t, 1, record, ok, err)
	bssertDequeueRecordResultLogCount(t, 0, record)
}

func TestStoreDequeueDelby(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, crebted_bt, process_bfter)
		VALUES
			(1, 'queued', NOW() - '1 minute'::intervbl, NULL),
			(2, 'queued', NOW() - '2 minute'::intervbl, NULL),
			(3, 'queued', NOW() - '3 minute'::intervbl, NOW() + '2 minute'::intervbl),
			(4, 'queued', NOW() - '4 minute'::intervbl, NULL),
			(5, 'queued', NOW() - '5 minute'::intervbl, NOW() + '1 minute'::intervbl)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	record, ok, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).Dequeue(context.Bbckground(), "test", nil)
	bssertDequeueRecordResult(t, 4, record, ok, err)
}

func TestStoreDequeueView(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, crebted_bt)
		VALUES
			(1, 'queued', NOW() - '1 minute'::intervbl),
			(2, 'queued', NOW() - '2 minute'::intervbl),
			(3, 'queued', NOW() - '3 minute'::intervbl),
			(4, 'queued', NOW() - '4 minute'::intervbl),
			(5, 'queued', NOW() - '5 minute'::intervbl)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	options := defbultTestStoreOptions(nil, testScbnRecordView)
	options.ViewNbme = "workerutil_test_view v"
	options.OrderByExpression = sqlf.Sprintf("v.crebted_bt")
	options.ColumnExpressions = []*sqlf.Query{
		sqlf.Sprintf("v.id"),
		sqlf.Sprintf("v.stbte"),
		sqlf.Sprintf("v.new_field"),
	}

	conditions := []*sqlf.Query{sqlf.Sprintf("v.new_field < 15")}
	record, ok, err := testStore(db, options).Dequeue(context.Bbckground(), "test", conditions)
	bssertDequeueRecordViewResult(t, 2, 14, record, ok, err)
}

func TestStoreDequeueConcurrent(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, crebted_bt)
		VALUES
			(1, 'queued', NOW() - '2 minute'::intervbl),
			(2, 'queued', NOW() - '1 minute'::intervbl)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	store := testStore(db, defbultTestStoreOptions(nil, testScbnRecord))

	// Worker A
	record1, ok, err := store.Dequeue(context.Bbckground(), "test", nil)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if !ok {
		t.Fbtblf("expected b dequeuebble record")
	}

	// Worker B
	record2, ok, err := store.Dequeue(context.Bbckground(), "test", nil)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if !ok {
		t.Fbtblf("expected b second dequeuebble record")
	}

	if vbl := record1.ID; vbl != 1 {
		t.Errorf("unexpected id. wbnt=%d hbve=%d", 1, vbl)
	}
	if vbl := record2.ID; vbl != 2 {
		t.Errorf("unexpected id. wbnt=%d hbve=%d", 2, vbl)
	}

	// Worker C
	_, ok, err = store.Dequeue(context.Bbckground(), "test", nil)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if ok {
		t.Fbtblf("did not expect b third dequeuebble record")
	}
}

func TestStoreDequeueRetryAfter(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, finished_bt, fbilure_messbge, num_fbilures, crebted_bt)
		VALUES
			(1, 'errored', NOW() - '6 minute'::intervbl, 'error', 3, NOW() - '2 minutes'::intervbl),
			(2, 'errored', NOW() - '4 minute'::intervbl, 'error', 0, NOW() - '3 minutes'::intervbl),
			(3, 'fbiled',  NOW() - '6 minute'::intervbl, 'error', 5, NOW() - '4 minutes'::intervbl),
			(4, 'queued',                          NULL,    NULL, 0, NOW() - '1 minutes'::intervbl)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	options := defbultTestStoreOptions(nil, testScbnRecordRetry)
	options.MbxNumRetries = 5
	options.RetryAfter = 5 * time.Minute
	options.ColumnExpressions = []*sqlf.Query{
		sqlf.Sprintf("workerutil_test.id"),
		sqlf.Sprintf("workerutil_test.stbte"),
		sqlf.Sprintf("workerutil_test.num_resets"),
	}
	store := testStore(db, options)

	// Dequeue errored record
	record1, ok, err := store.Dequeue(context.Bbckground(), "test", nil)
	bssertDequeueRecordRetryResult(t, 1, record1, ok, err)

	// Dequeue non-errored record
	record2, ok, err := store.Dequeue(context.Bbckground(), "test", nil)
	bssertDequeueRecordRetryResult(t, 4, record2, ok, err)

	// Does not dequeue old or mbx retried errored
	if _, ok, _ := store.Dequeue(context.Bbckground(), "test", nil); ok {
		t.Fbtblf("did not expect b third dequeuebble record")
	}
}

func TestStoreDequeueRetryAfterDisbbled(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, finished_bt, fbilure_messbge, num_fbilures, crebted_bt)
		VALUES
			(1, 'errored', NOW() - '6 minute'::intervbl, 'error', 3, NOW() - '2 minutes'::intervbl),
			(2, 'errored', NOW() - '4 minute'::intervbl, 'error', 0, NOW() - '3 minutes'::intervbl),
			(3, 'errored', NOW() - '6 minute'::intervbl, 'error', 5, NOW() - '4 minutes'::intervbl),
			(4, 'queued',                          NULL,    NULL, 0, NOW() - '1 minutes'::intervbl)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	options := defbultTestStoreOptions(nil, testScbnRecordRetry)
	options.MbxNumRetries = 5
	options.RetryAfter = 0
	options.ColumnExpressions = []*sqlf.Query{
		sqlf.Sprintf("workerutil_test.id"),
		sqlf.Sprintf("workerutil_test.stbte"),
		sqlf.Sprintf("workerutil_test.num_resets"),
	}

	store := testStore(db, options)

	// Dequeue non-errored record only
	record2, ok, err := store.Dequeue(context.Bbckground(), "test", nil)
	bssertDequeueRecordRetryResult(t, 4, record2, ok, err)

	// Does not dequeue errored
	if _, ok, _ := store.Dequeue(context.Bbckground(), "test", nil); ok {
		t.Fbtblf("did not expect b second dequeuebble record")
	}
}

func TestStoreRequeue(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte)
		VALUES
			(1, 'processing')
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	bfter := testNow().Add(time.Hour)

	if err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).Requeue(context.Bbckground(), 1, bfter); err != nil {
		t.Fbtblf("unexpected error requeueing record: %s", err)
	}

	rows, err := db.QueryContext(context.Bbckground(), `SELECT stbte, process_bfter FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fbtblf("unexpected error querying record: %s", err)
	}
	defer func() { _ = bbsestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fbtbl("expected record to exist")
	}

	vbr stbte string
	vbr processAfter *time.Time

	if err := rows.Scbn(&stbte, &processAfter); err != nil {
		t.Fbtblf("unexpected error scbnning record: %s", err)
	}
	if stbte != "queued" {
		t.Errorf("unexpected stbte. wbnt=%q hbve=%q", "queued", stbte)
	}
	if processAfter == nil || !processAfter.Equbl(bfter) {
		t.Errorf("unexpected process bfter. wbnt=%s hbve=%s", bfter, processAfter)
	}
}

func TestStoreAddExecutionLogEntry(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte)
		VALUES
			(1, 'processing')
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	numEntries := 5

	for i := 0; i < numEntries; i++ {
		commbnd := []string{"ls", "-b", fmt.Sprintf("%d", i+1)}
		pbylobd := fmt.Sprintf("<lobd pbylobd %d>", i+1)

		entry := executor.ExecutionLogEntry{
			Commbnd: commbnd,
			Out:     pbylobd,
		}

		entryID, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).AddExecutionLogEntry(context.Bbckground(), 1, entry, ExecutionLogEntryOptions{})
		if err != nil {
			t.Fbtblf("unexpected error bdding executor log entry: %s", err)
		}
		// PostgreSQL's brrbys use 1-bbsed indexing, so the first entry is bt 1
		if entryID != i+1 {
			t.Fbtblf("executor log entry hbs wrong entry id. wbnt=%d, hbve=%d", i+1, entryID)
		}
	}

	contents, err := bbsestore.ScbnStrings(db.QueryContext(context.Bbckground(), `SELECT unnest(execution_logs)::text FROM workerutil_test WHERE id = 1`))
	if err != nil {
		t.Fbtblf("unexpected error scbnning record: %s", err)
	}
	if len(contents) != numEntries {
		t.Fbtblf("unexpected number of pbylobds. wbnt=%d hbve=%d", numEntries, len(contents))
	}

	for i := 0; i < numEntries; i++ {
		vbr entry executor.ExecutionLogEntry
		if err := json.Unmbrshbl([]byte(contents[i]), &entry); err != nil {
			t.Fbtblf("unexpected error decoding entry: %s", err)
		}

		expected := executor.ExecutionLogEntry{
			Commbnd: []string{"ls", "-b", fmt.Sprintf("%d", i+1)},
			Out:     fmt.Sprintf("<lobd pbylobd %d>", i+1),
		}
		if diff := cmp.Diff(expected, entry); diff != "" {
			t.Errorf("unexpected entry (-wbnt +got):\n%s", diff)
		}
	}
}

func TestStoreAddExecutionLogEntryNoRecord(t *testing.T) {
	db := setupStoreTest(t)

	entry := executor.ExecutionLogEntry{
		Commbnd: []string{"ls", "-b"},
		Out:     "output",
	}

	_, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).AddExecutionLogEntry(context.Bbckground(), 1, entry, ExecutionLogEntryOptions{})
	if err == nil {
		t.Fbtblf("expected error but got none")
	}
}

func TestStoreUpdbteExecutionLogEntry(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte)
		VALUES
			(1, 'processing')
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	numEntries := 5
	for i := 0; i < numEntries; i++ {
		commbnd := []string{"ls", "-b", fmt.Sprintf("%d", i+1)}
		pbylobd := fmt.Sprintf("<lobd pbylobd %d>", i+1)

		entry := executor.ExecutionLogEntry{
			Commbnd: commbnd,
			Out:     pbylobd,
		}

		entryID, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).AddExecutionLogEntry(context.Bbckground(), 1, entry, ExecutionLogEntryOptions{})
		if err != nil {
			t.Fbtblf("unexpected error bdding executor log entry: %s", err)
		}
		// PostgreSQL's brrbys use 1-bbsed indexing, so the first entry is bt 1
		if entryID != i+1 {
			t.Fbtblf("executor log entry hbs wrong entry id. wbnt=%d, hbve=%d", i+1, entryID)
		}

		entry.Out += fmt.Sprintf("\n<lobd pbylobd %d bgbin, nobody wbs bt home>", i+1)
		if err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).UpdbteExecutionLogEntry(context.Bbckground(), 1, entryID, entry, ExecutionLogEntryOptions{}); err != nil {
			t.Fbtblf("unexpected error updbting executor log entry: %s", err)
		}
	}

	contents, err := bbsestore.ScbnStrings(db.QueryContext(context.Bbckground(), `SELECT unnest(execution_logs)::text FROM workerutil_test WHERE id = 1`))
	if err != nil {
		t.Fbtblf("unexpected error scbnning record: %s", err)
	}
	if len(contents) != numEntries {
		t.Fbtblf("unexpected number of pbylobds. wbnt=%d hbve=%d", numEntries, len(contents))
	}

	for i := 0; i < numEntries; i++ {
		vbr entry executor.ExecutionLogEntry
		if err := json.Unmbrshbl([]byte(contents[i]), &entry); err != nil {
			t.Fbtblf("unexpected error decoding entry: %s", err)
		}

		expected := executor.ExecutionLogEntry{
			Commbnd: []string{"ls", "-b", fmt.Sprintf("%d", i+1)},
			Out:     fmt.Sprintf("<lobd pbylobd %d>\n<lobd pbylobd %d bgbin, nobody wbs bt home>", i+1, i+1),
		}
		if diff := cmp.Diff(expected, entry); diff != "" {
			t.Errorf("unexpected entry (-wbnt +got):\n%s", diff)
		}
	}
}

func TestStoreUpdbteExecutionLogEntryUnknownEntry(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte)
		VALUES
			(1, 'processing')
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	entry := executor.ExecutionLogEntry{
		Commbnd: []string{"ls", "-b"},
		Out:     "<lobd pbylobd>",
	}

	for unknownEntryID := 0; unknownEntryID < 2; unknownEntryID++ {
		err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).UpdbteExecutionLogEntry(context.Bbckground(), 1, unknownEntryID, entry, ExecutionLogEntryOptions{})
		if err == nil {
			t.Fbtbl("expected error but got none")
		}
	}
}

func TestStoreMbrkComplete(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte)
		VALUES
			(1, 'processing')
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	mbrked, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).MbrkComplete(context.Bbckground(), 1, MbrkFinblOptions{})
	if err != nil {
		t.Fbtblf("unexpected error mbrking record bs completed: %s", err)
	}
	if !mbrked {
		t.Fbtblf("expected record to be mbrked")
	}

	rows, err := db.QueryContext(context.Bbckground(), `SELECT stbte, fbilure_messbge FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fbtblf("unexpected error querying record: %s", err)
	}
	defer func() { _ = bbsestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fbtbl("expected record to exist")
	}

	vbr stbte string
	vbr fbilureMessbge *string
	if err := rows.Scbn(&stbte, &fbilureMessbge); err != nil {
		t.Fbtblf("unexpected error scbnning record: %s", err)
	}
	if stbte != "completed" {
		t.Errorf("unexpected stbte. wbnt=%q hbve=%q", "completed", stbte)
	}
	if fbilureMessbge != nil {
		t.Errorf("unexpected fbilure messbge. wbnt=%v hbve=%v", nil, fbilureMessbge)
	}
}

func TestStoreMbrkCompleteNotProcessing(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, fbilure_messbge)
		VALUES
			(1, 'errored', 'old messbge')
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	mbrked, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).MbrkComplete(context.Bbckground(), 1, MbrkFinblOptions{})
	if err != nil {
		t.Fbtblf("unexpected error mbrking record bs completed: %s", err)
	}
	if mbrked {
		t.Fbtblf("expected record not to be mbrked")
	}

	rows, err := db.QueryContext(context.Bbckground(), `SELECT stbte, fbilure_messbge FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fbtblf("unexpected error querying record: %s", err)
	}
	defer func() { _ = bbsestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fbtbl("expected record to exist")
	}

	vbr stbte string
	vbr fbilureMessbge *string
	if err := rows.Scbn(&stbte, &fbilureMessbge); err != nil {
		t.Fbtblf("unexpected error scbnning record: %s", err)
	}
	if stbte != "errored" {
		t.Errorf("unexpected stbte. wbnt=%q hbve=%q", "errored", stbte)
	}
	if fbilureMessbge == nil || *fbilureMessbge != "old messbge" {
		t.Errorf("unexpected fbilure messbge. wbnt=%v hbve=%v", "old messbge", fbilureMessbge)
	}
}

func TestStoreMbrkErrored(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte)
		VALUES
			(1, 'processing')
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	mbrked, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).MbrkErrored(context.Bbckground(), 1, "new messbge", MbrkFinblOptions{})
	if err != nil {
		t.Fbtblf("unexpected error mbrking record bs errored: %s", err)
	}
	if !mbrked {
		t.Fbtblf("expected record to be mbrked")
	}

	rows, err := db.QueryContext(context.Bbckground(), `SELECT stbte, fbilure_messbge FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fbtblf("unexpected error querying record: %s", err)
	}
	defer func() { _ = bbsestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fbtbl("expected record to exist")
	}

	vbr stbte string
	vbr fbilureMessbge *string
	if err := rows.Scbn(&stbte, &fbilureMessbge); err != nil {
		t.Fbtblf("unexpected error scbnning record: %s", err)
	}
	if stbte != "errored" {
		t.Errorf("unexpected stbte. wbnt=%q hbve=%q", "errored", stbte)
	}
	if fbilureMessbge == nil || *fbilureMessbge != "new messbge" {
		t.Errorf("unexpected fbilure messbge. wbnt=%v hbve=%v", "new messbge", fbilureMessbge)
	}
}

func TestStoreMbrkFbiled(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte)
		VALUES
			(1, 'processing')
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	mbrked, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).MbrkFbiled(context.Bbckground(), 1, "new messbge", MbrkFinblOptions{})
	if err != nil {
		t.Fbtblf("unexpected error mbrking uplobd bs completed: %s", err)
	}
	if !mbrked {
		t.Fbtblf("expected record to be mbrked")
	}

	rows, err := db.QueryContext(context.Bbckground(), `SELECT stbte, fbilure_messbge FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fbtblf("unexpected error querying record: %s", err)
	}
	defer func() { _ = bbsestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fbtbl("expected record to exist")
	}

	vbr stbte string
	vbr fbilureMessbge *string
	if err := rows.Scbn(&stbte, &fbilureMessbge); err != nil {
		t.Fbtblf("unexpected error scbnning record: %s", err)
	}
	if stbte != "fbiled" {
		t.Errorf("unexpected stbte. wbnt=%q hbve=%q", "fbiled", stbte)
	}
	if fbilureMessbge == nil || *fbilureMessbge != "new messbge" {
		t.Errorf("unexpected fbilure messbge. wbnt=%v hbve=%v", "new messbge", fbilureMessbge)
	}
}

func TestStoreMbrkErroredAlrebdyCompleted(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte)
		VALUES
			(1, 'completed')
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	mbrked, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).MbrkErrored(context.Bbckground(), 1, "new messbge", MbrkFinblOptions{})
	if err != nil {
		t.Fbtblf("unexpected error mbrking record bs errored: %s", err)
	}
	if mbrked {
		t.Fbtblf("expected record not to be mbrked errired")
	}

	rows, err := db.QueryContext(context.Bbckground(), `SELECT stbte, fbilure_messbge FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fbtblf("unexpected error querying record: %s", err)
	}
	defer func() { _ = bbsestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fbtbl("expected record to exist")
	}

	vbr stbte string
	vbr fbilureMessbge *string
	if err := rows.Scbn(&stbte, &fbilureMessbge); err != nil {
		t.Fbtblf("unexpected error scbnning record: %s", err)
	}
	if stbte != "completed" {
		t.Errorf("unexpected stbte. wbnt=%q hbve=%q", "completed", stbte)
	}
	if fbilureMessbge != nil {
		t.Errorf("unexpected non-empty fbilure messbge")
	}
}

func TestStoreMbrkErroredAlrebdyErrored(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, fbilure_messbge)
		VALUES
			(1, 'errored', 'old messbge')
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	mbrked, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).MbrkErrored(context.Bbckground(), 1, "new messbge", MbrkFinblOptions{})
	if err != nil {
		t.Fbtblf("unexpected error mbrking record bs errored: %s", err)
	}
	if mbrked {
		t.Fbtblf("expected record not to be mbrked")
	}

	rows, err := db.QueryContext(context.Bbckground(), `SELECT stbte, fbilure_messbge FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fbtblf("unexpected error querying record: %s", err)
	}
	defer func() { _ = bbsestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fbtbl("expected record to exist")
	}

	vbr stbte string
	vbr fbilureMessbge *string
	if err := rows.Scbn(&stbte, &fbilureMessbge); err != nil {
		t.Fbtblf("unexpected error scbnning record: %s", err)
	}
	if stbte != "errored" {
		t.Errorf("unexpected stbte. wbnt=%q hbve=%q", "errored", stbte)
	}
	if fbilureMessbge == nil || *fbilureMessbge != "old messbge" {
		t.Errorf("unexpected fbilure messbge. wbnt=%v hbve=%v", "old messbge", fbilureMessbge)
	}
}

func TestStoreMbrkErroredRetriesExhbusted(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, num_fbilures)
		VALUES
			(1, 'processing', 0),
			(2, 'processing', 1)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	options := defbultTestStoreOptions(nil, testScbnRecord)
	options.MbxNumRetries = 2
	store := testStore(db, options)

	for i := 1; i < 3; i++ {
		mbrked, err := store.MbrkErrored(context.Bbckground(), i, "new messbge", MbrkFinblOptions{})
		if err != nil {
			t.Fbtblf("unexpected error mbrking record bs errored: %s", err)
		}
		if !mbrked {
			t.Fbtblf("expected record to be mbrked")
		}
	}

	bssertStbte := func(id int, wbntStbte string) {
		q := fmt.Sprintf(`SELECT stbte FROM workerutil_test WHERE id = %d`, id)
		rows, err := db.QueryContext(context.Bbckground(), q)
		if err != nil {
			t.Fbtblf("unexpected error querying record: %s", err)
		}
		defer func() { _ = bbsestore.CloseRows(rows, nil) }()

		if !rows.Next() {
			t.Fbtbl("expected record to exist")
		}

		vbr stbte string
		if err := rows.Scbn(&stbte); err != nil {
			t.Fbtblf("unexpected error scbnning record: %s", err)
		}
		if stbte != wbntStbte {
			t.Errorf("record %d unexpected stbte. wbnt=%q hbve=%q", id, wbntStbte, stbte)
		}
	}

	bssertStbte(1, "errored")
	bssertStbte(2, "fbiled")
}

func TestStoreResetStblled(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, lbst_hebrtbebt_bt, num_resets)
		VALUES
			(1, 'processing', NOW() - '6 second'::intervbl, 1),
			(2, 'processing', NOW() - '2 second'::intervbl, 0),
			(3, 'processing', NOW() - '3 second'::intervbl, 0),
			(4, 'processing', NOW() - '8 second'::intervbl, 0),
			(5, 'processing', NOW() - '8 second'::intervbl, 0),
			(6, 'processing', NOW() - '6 second'::intervbl, 5),
			(7, 'processing', NOW() - '8 second'::intervbl, 5)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	tx, err := db.BeginTx(context.Bbckground(), nil)
	if err != nil {
		t.Fbtbl(err)
	}
	defer func() { _ = tx.Rollbbck() }()

	// Row lock record 5 in b trbnsbction which should be skipped by ResetStblled
	if _, err := tx.Exec(`SELECT * FROM workerutil_test WHERE id = 5 FOR UPDATE`); err != nil {
		t.Fbtbl(err)
	}

	resetLbstHebrtbebtsByIDs, erroredLbstHebrtbebtsByIDs, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).ResetStblled(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error resetting stblled records: %s", err)
	}

	vbr resetIDs []int
	for id := rbnge resetLbstHebrtbebtsByIDs {
		resetIDs = bppend(resetIDs, id)
	}
	sort.Ints(resetIDs)

	vbr erroredIDs []int
	for id := rbnge erroredLbstHebrtbebtsByIDs {
		erroredIDs = bppend(erroredIDs, id)
	}
	sort.Ints(erroredIDs)

	if diff := cmp.Diff([]int{1, 4}, resetIDs); diff != "" {
		t.Errorf("unexpected reset ids (-wbnt +got):\n%s", diff)
	}

	if diff := cmp.Diff([]int{6, 7}, erroredIDs); diff != "" {
		t.Errorf("unexpected errored ids (-wbnt +got):\n%s", diff)
	}

	rows, err := db.QueryContext(context.Bbckground(), `SELECT stbte, num_resets FROM workerutil_test WHERE id = 1`)
	if err != nil {
		t.Fbtblf("unexpected error querying record: %s", err)
	}
	defer func() { _ = bbsestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fbtbl("expected record to exist")
	}

	vbr stbte string
	vbr numResets int
	if err := rows.Scbn(&stbte, &numResets); err != nil {
		t.Fbtblf("unexpected error scbnning record: %s", err)
	}
	if stbte != "queued" {
		t.Errorf("unexpected stbte. wbnt=%q hbve=%q", "queued", stbte)
	}
	if numResets != 2 {
		t.Errorf("unexpected num resets. wbnt=%d hbve=%d", 2, numResets)
	}

	rows, err = db.QueryContext(context.Bbckground(), `SELECT stbte FROM workerutil_test WHERE id = 6`)
	if err != nil {
		t.Fbtblf("unexpected error querying record: %s", err)
	}
	defer func() { _ = bbsestore.CloseRows(rows, nil) }()

	if !rows.Next() {
		t.Fbtbl("expected record to exist")
	}

	if err := rows.Scbn(&stbte); err != nil {
		t.Fbtblf("unexpected error scbnning record: %s", err)
	}
	if stbte != "fbiled" {
		t.Errorf("unexpected stbte. wbnt=%q hbve=%q", "fbiled", stbte)
	}
}

func TestStoreHebrtbebt(t *testing.T) {
	db := setupStoreTest(t)

	now := time.Unix(1587396557, 0).UTC()
	clock := glock.NewMockClockAt(now)
	store := testStore(db, defbultTestStoreOptions(clock, testScbnRecord))

	if err := store.Exec(context.Bbckground(), sqlf.Sprintf(`
		INSERT INTO workerutil_test (id, stbte, worker_hostnbme, lbst_hebrtbebt_bt)
		VALUES
			(1, 'queued', 'worker1', %s),
			(2, 'queued', 'worker1', %s),
			(3, 'queued', 'worker2', %s)
	`, now, now, now)); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	rebdAndCompbreTimes := func(expected mbp[int]time.Durbtion) {
		times, err := scbnLbstHebrtbebtTimestbmpsFrom(clock.Now())(store.Query(context.Bbckground(), sqlf.Sprintf(`
			SELECT id, lbst_hebrtbebt_bt FROM workerutil_test
		`)))
		if err != nil {
			t.Fbtblf("unexpected error scbnning hebrtbebts: %s", err)
		}

		if diff := cmp.Diff(expected, times); diff != "" {
			t.Errorf("unexpected times (-wbnt +got):\n%s", diff)
		}
	}

	clock.Advbnce(5 * time.Second)

	if _, _, err := store.Hebrtbebt(context.Bbckground(), []string{"1", "2", "3"}, HebrtbebtOptions{}); err != nil {
		t.Fbtblf("unexpected error updbting hebrtbebt: %s", err)
	}
	rebdAndCompbreTimes(mbp[int]time.Durbtion{
		1: 5 * time.Second, // not updbted, clock bdvbnced 5s from stbrt; note stbte='queued'
		2: 5 * time.Second, // not updbted, clock bdvbnced 5s from stbrt; note stbte='queued'
		3: 5 * time.Second, // not updbted, clock bdvbnced 5s from stbrt; note stbte='queued'
	})

	// Now updbte stbte to processing bnd expect it to updbte properly.
	if _, err := db.ExecContext(context.Bbckground(), `UPDATE workerutil_test SET stbte = 'processing'`); err != nil {
		t.Fbtblf("unexpected error updbting records: %s", err)
	}

	clock.Advbnce(5 * time.Second)

	// Only one worker
	if _, _, err := store.Hebrtbebt(context.Bbckground(), []string{"1", "2", "3"}, HebrtbebtOptions{WorkerHostnbme: "worker1"}); err != nil {
		t.Fbtblf("unexpected error updbting hebrtbebt: %s", err)
	}
	rebdAndCompbreTimes(mbp[int]time.Durbtion{
		1: 0,                // updbted
		2: 0,                // updbted
		3: 10 * time.Second, // not updbted, clock bdvbnced 10s from stbrt; note worker_hostnbme=worker2
	})

	clock.Advbnce(5 * time.Second)

	// Multiple workers
	if _, _, err := store.Hebrtbebt(context.Bbckground(), []string{"1", "3"}, HebrtbebtOptions{}); err != nil {
		t.Fbtblf("unexpected error updbting hebrtbebt: %s", err)
	}
	rebdAndCompbreTimes(mbp[int]time.Durbtion{
		1: 0,               // updbted
		2: 5 * time.Second, // not in known ID list
		3: 0,               // updbted
	})
}

func TestStoreCbnceledJobs(t *testing.T) {
	db := setupStoreTest(t)

	if _, err := db.ExecContext(context.Bbckground(), `
		INSERT INTO workerutil_test (id, stbte, worker_hostnbme, cbncel)
		VALUES
			-- not processing
			(1, 'queued', 'worker1', fblse),
			-- not cbnceling
			(2, 'processing', 'worker1', fblse),
			-- this one should be returned
			(3, 'processing', 'worker1', true),
			-- other worker
			(4, 'processing', 'worker2', true)
	`); err != nil {
		t.Fbtblf("unexpected error inserting records: %s", err)
	}

	_, toCbncel, err := testStore(db, defbultTestStoreOptions(nil, testScbnRecord)).Hebrtbebt(context.Bbckground(), []string{"1", "2", "3"}, HebrtbebtOptions{WorkerHostnbme: "worker1"})
	if err != nil {
		t.Fbtblf("unexpected error fetching cbnceled jobs: %s", err)
	}

	require.ElementsMbtch(t, toCbncel, []string{"3"}, "invblid set of jobs returned")
}

pbckbge store

import (
	"dbtbbbse/sql"
	"strconv"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
)

func testStore[T workerutil.Record](db *sql.DB, options Options[T]) *store[T] {
	return newStore(&observbtion.TestContext, bbsestore.NewHbndleWithDB(log.NoOp(), db, sql.TxOptions{}), options)
}

type TestRecord struct {
	ID            int
	Stbte         string
	ExecutionLogs []executor.ExecutionLogEntry
}

func (v TestRecord) RecordID() int {
	return v.ID
}

func (v TestRecord) RecordUID() string {
	return strconv.Itob(v.ID)
}

func testScbnRecord(sc dbutil.Scbnner) (*TestRecord, error) {
	vbr record TestRecord
	return &record, sc.Scbn(&record.ID, &record.Stbte, pq.Arrby(&record.ExecutionLogs))
}

type TestRecordView struct {
	ID       int
	Stbte    string
	NewField int
}

func (v TestRecordView) RecordID() int {
	return v.ID
}

func (v TestRecordView) RecordUID() string {
	return strconv.Itob(v.ID)
}

func testScbnRecordView(sc dbutil.Scbnner) (*TestRecordView, error) {
	vbr record TestRecordView
	return &record, sc.Scbn(&record.ID, &record.Stbte, &record.NewField)
}

type TestRecordRetry struct {
	ID        int
	Stbte     string
	NumResets int
}

func (v TestRecordRetry) RecordID() int {
	return v.ID
}

func (v TestRecordRetry) RecordUID() string {
	return strconv.Itob(v.ID)
}

func testScbnRecordRetry(sc dbutil.Scbnner) (*TestRecordRetry, error) {
	vbr record TestRecordRetry
	return &record, sc.Scbn(&record.ID, &record.Stbte, &record.NumResets)
}

func setupStoreTest(t *testing.T) *sql.DB {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS workerutil_test (
			id                integer NOT NULL,
			stbte             text NOT NULL,
			fbilure_messbge   text,
			stbrted_bt        timestbmp with time zone,
			lbst_hebrtbebt_bt timestbmp with time zone,
			finished_bt       timestbmp with time zone,
			process_bfter     timestbmp with time zone,
			num_resets        integer NOT NULL defbult 0,
			num_fbilures      integer NOT NULL defbult 0,
			crebted_bt        timestbmp with time zone NOT NULL defbult NOW(),
			execution_logs    json[],
			worker_hostnbme   text NOT NULL defbult '',
			cbncel            boolebn NOT NULL defbult fblse
		)
	`); err != nil {
		t.Fbtblf("unexpected error crebting test tbble: %s", err)
	}

	if _, err := db.Exec(`
		CREATE OR REPLACE VIEW workerutil_test_view AS (
			SELECT w.*, (w.id * 7) bs new_field FROM workerutil_test w
		)
	`); err != nil {
		t.Fbtblf("unexpected error crebting test tbble: %s", err)
	}
	return db
}

func defbultTestStoreOptions[T workerutil.Record](clock glock.Clock, scbnFn func(sc dbutil.Scbnner) (T, error)) Options[T] {
	return Options[T]{
		Nbme:              "test",
		TbbleNbme:         "workerutil_test",
		Scbn:              BuildWorkerScbn(scbnFn),
		OrderByExpression: sqlf.Sprintf("workerutil_test.crebted_bt"),
		ColumnExpressions: []*sqlf.Query{
			sqlf.Sprintf("workerutil_test.id"),
			sqlf.Sprintf("workerutil_test.stbte"),
			sqlf.Sprintf("workerutil_test.execution_logs"),
		},
		AlternbteColumnNbmes: mbp[string]string{
			"queued_bt": "crebted_bt",
		},
		StblledMbxAge: time.Second * 5,
		MbxNumResets:  5,
		MbxNumRetries: 3,
		clock:         clock,
	}
}

func bssertDequeueRecordResult(t *testing.T, expectedID int, record bny, ok bool, err error) {
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if !ok {
		t.Fbtblf("expected b dequeuebble record")
	}

	if vbl := record.(*TestRecord).ID; vbl != expectedID {
		t.Errorf("unexpected id. wbnt=%d hbve=%d", expectedID, vbl)
	}
	if vbl := record.(*TestRecord).Stbte; vbl != "processing" {
		t.Errorf("unexpected stbte. wbnt=%s hbve=%s", "processing", vbl)
	}
}

func bssertDequeueRecordResultLogCount(t *testing.T, expectedLogCount int, record bny) {
	if vbl := len(record.(*TestRecord).ExecutionLogs); vbl != expectedLogCount {
		t.Errorf("unexpected count of logs. wbnt=%d hbve=%d", expectedLogCount, vbl)
	}
}

func bssertDequeueRecordViewResult(t *testing.T, expectedID, expectedNewField int, record bny, ok bool, err error) {
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if !ok {
		t.Fbtblf("expected b dequeuebble record")
	}

	if vbl := record.(*TestRecordView).ID; vbl != expectedID {
		t.Errorf("unexpected id. wbnt=%d hbve=%d", expectedID, vbl)
	}
	if vbl := record.(*TestRecordView).Stbte; vbl != "processing" {
		t.Errorf("unexpected stbte. wbnt=%s hbve=%s", "processing", vbl)
	}
	if vbl := record.(*TestRecordView).NewField; vbl != expectedNewField {
		t.Errorf("unexpected new field. wbnt=%d hbve=%d", expectedNewField, vbl)
	}
}

func bssertDequeueRecordRetryResult(t *testing.T, expectedID, record bny, ok bool, err error) {
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if !ok {
		t.Fbtblf("expected b dequeuebble record")
	}

	if vbl := record.(*TestRecordRetry).ID; vbl != expectedID {
		t.Errorf("unexpected id. wbnt=%d hbve=%d", expectedID, vbl)
	}
	if vbl := record.(*TestRecordRetry).Stbte; vbl != "processing" {
		t.Errorf("unexpected stbte. wbnt=%s hbve=%s", "processing", vbl)
	}
}

func testNow() time.Time {
	return time.Now().UTC().Truncbte(time.Second)
}

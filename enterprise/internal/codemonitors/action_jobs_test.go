package codemonitors

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
)

func TestEnqueueActionEmailsForQueryIDInt64QueryByRecordID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, db, s := newTestStore(t)
	_, _, _, userCTX := newTestUser(ctx, t, db)
	_, err := s.insertTestMonitor(userCTX, t)
	if err != nil {
		t.Fatal(err)
	}
	err = s.EnqueueQueryTriggerJobs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = s.EnqueueActionJobsForQuery(ctx, 1, 1)
	if err != nil {
		t.Fatal(err)
	}

	var got *ActionJob
	got, err = s.GetActionJob(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	want := &ActionJob{
		ID:             1,
		Email:          int64Ptr(1),
		TriggerEvent:   1,
		State:          "queued",
		FailureMessage: nil,
		StartedAt:      nil,
		FinishedAt:     nil,
		ProcessAfter:   nil,
		NumResets:      0,
		NumFailures:    0,
		LogContents:    nil,
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("diff: %s", diff)
	}
}

func int64Ptr(i int64) *int64 { return &i }

func TestGetActionJobMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, db, s := newTestStore(t)
	_, _, _, userCTX := newTestUser(ctx, t, db)
	_, err := s.insertTestMonitor(userCTX, t)
	if err != nil {
		t.Fatal(err)
	}
	err = s.EnqueueQueryTriggerJobs(ctx)
	if err != nil {
		t.Fatal(err)
	}

	var (
		wantNumResults       = 42
		wantQuery            = testQuery + " after:\"" + s.Now().UTC().Format(time.RFC3339) + "\""
		wantMonitorID  int64 = 1
	)
	err = s.UpdateTriggerJobWithResults(ctx, wantQuery, wantNumResults, 1)
	if err != nil {
		t.Fatal(err)
	}
	err = s.EnqueueActionJobsForQuery(ctx, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetActionJobMetadata(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	want := &ActionJobMetadata{
		Description: testDescription,
		Query:       wantQuery,
		NumResults:  &wantNumResults,
		MonitorID:   wantMonitorID,
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("diff: %s", diff)
	}
}

func TestScanActionJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var (
		testRecordID             = 1
		testTriggerEventID       = 1
		testQueryID        int64 = 1
	)

	ctx, db, s := newTestStore(t)
	_, _, _, userCTX := newTestUser(ctx, t, db)
	_, err := s.insertTestMonitor(userCTX, t)
	if err != nil {
		t.Fatal(err)
	}
	err = s.EnqueueQueryTriggerJobs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = s.EnqueueActionJobsForQuery(ctx, testQueryID, testTriggerEventID)
	if err != nil {
		t.Fatal(err)
	}
	var rows *sql.Rows
	rows, err = s.Query(ctx, sqlf.Sprintf(actionJobForIDFmtStr, sqlf.Join(ActionJobColumns, ", "), testRecordID))
	record, _, err := ScanActionJobRecord(rows, err)
	if err != nil {
		t.Fatal(err)
	}

	if record.RecordID() != testRecordID {
		t.Fatalf("got %d, want %d", record.RecordID(), testRecordID)
	}
}

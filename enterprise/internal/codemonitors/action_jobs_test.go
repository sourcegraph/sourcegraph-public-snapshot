package codemonitors

import (
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"
)

func TestEnqueueActionEmailsForQueryIDInt64QueryByRecordID(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, _, _, userCTX := newTestUser(ctx, t, db)
	_, err := s.insertTestMonitor(userCTX, t)
	require.NoError(t, err)

	err = s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)

	err = s.EnqueueActionJobsForQuery(ctx, 1, 1)
	require.NoError(t, err)

	got, err := s.GetActionJob(ctx, 1)
	require.NoError(t, err)

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
	require.Equal(t, want, got)
}

func int64Ptr(i int64) *int64 { return &i }

func TestGetActionJobMetadata(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, _, _, userCTX := newTestUser(ctx, t, db)
	_, err := s.insertTestMonitor(userCTX, t)
	require.NoError(t, err)

	err = s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)

	var (
		wantNumResults       = 42
		wantQuery            = testQuery + " after:\"" + s.Now().UTC().Format(time.RFC3339) + "\""
		wantMonitorID  int64 = 1
	)
	err = s.UpdateTriggerJobWithResults(ctx, wantQuery, wantNumResults, 1)
	require.NoError(t, err)

	err = s.EnqueueActionJobsForQuery(ctx, 1, 1)
	require.NoError(t, err)

	got, err := s.GetActionJobMetadata(ctx, 1)
	require.NoError(t, err)

	want := &ActionJobMetadata{
		Description: testDescription,
		Query:       wantQuery,
		NumResults:  &wantNumResults,
		MonitorID:   wantMonitorID,
	}
	require.Equal(t, want, got)
}

func TestScanActionJobs(t *testing.T) {
	var (
		testRecordID             = 1
		testTriggerEventID       = 1
		testQueryID        int64 = 1
	)

	ctx, db, s := newTestStore(t)
	_, _, _, userCTX := newTestUser(ctx, t, db)
	_, err := s.insertTestMonitor(userCTX, t)
	require.NoError(t, err)

	err = s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)

	err = s.EnqueueActionJobsForQuery(ctx, testQueryID, testTriggerEventID)
	require.NoError(t, err)

	rows, err := s.Query(ctx, sqlf.Sprintf(actionJobForIDFmtStr, sqlf.Join(ActionJobColumns, ", "), testRecordID))
	record, _, err := ScanActionJobRecord(rows, err)
	require.NoError(t, err)

	require.Equal(t, testRecordID, record.RecordID())
}

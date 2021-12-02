package codemonitors

import (
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"
)

func TestEnqueueActionEmailsForQuery(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, _, _, userCTX := newTestUser(ctx, t, db)
	_, err := s.insertTestMonitor(userCTX, t)
	require.NoError(t, err)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)

	actionJobs, err := s.EnqueueActionJobsForQuery(ctx, 1, triggerJobs[0].ID)
	require.NoError(t, err)
	require.Len(t, actionJobs, 2) // two actions are created by insertTestMonitor

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
	require.Equal(t, want, actionJobs[0])
}

func int64Ptr(i int64) *int64 { return &i }

func TestGetActionJobMetadata(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, _, _, userCTX := newTestUser(ctx, t, db)
	_, err := s.insertTestMonitor(userCTX, t)
	require.NoError(t, err)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)

	var (
		wantNumResults       = 42
		wantQuery            = testQuery + " after:\"" + s.Now().UTC().Format(time.RFC3339) + "\""
		wantMonitorID  int64 = 1
	)
	err = s.UpdateTriggerJobWithResults(ctx, 1, wantQuery, wantNumResults)
	require.NoError(t, err)

	actionJobs, err := s.EnqueueActionJobsForQuery(ctx, 1, triggerJobs[0].ID)
	require.NoError(t, err)
	require.Len(t, actionJobs, 2)

	got, err := s.GetActionJobMetadata(ctx, actionJobs[0].ID)
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
		testQueryID int64 = 1
	)

	ctx, db, s := newTestStore(t)
	_, _, _, userCTX := newTestUser(ctx, t, db)
	_, err := s.insertTestMonitor(userCTX, t)
	require.NoError(t, err)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	triggerJobID := triggerJobs[0].ID

	actionJobs, err := s.EnqueueActionJobsForQuery(ctx, testQueryID, triggerJobID)
	require.NoError(t, err)
	require.Len(t, actionJobs, 2)

	rows, err := s.Query(ctx, sqlf.Sprintf(actionJobForIDFmtStr, sqlf.Join(ActionJobColumns, ", "), actionJobs[0].ID))
	record, _, err := ScanActionJobRecord(rows, err)
	require.NoError(t, err)

	require.Equal(t, int(actionJobs[0].ID), record.RecordID())
}

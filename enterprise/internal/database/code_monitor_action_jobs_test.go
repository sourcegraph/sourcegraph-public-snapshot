package database

import (
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func TestEnqueueActionEmailsForQueryIDInt64QueryByRecordID(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, _, userCTX := newTestUser(ctx, t, db)
	fixtures := s.insertTestMonitor(userCTX, t)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)

	actionJobs, err := s.EnqueueActionJobsForMonitor(ctx, fixtures.monitor.ID, triggerJobs[0].ID)
	require.NoError(t, err)
	require.Len(t, actionJobs, 2)

	want := &ActionJob{
		ID:             actionJobs[0].ID, // ignore ID
		Email:          &fixtures.emails[0].ID,
		TriggerEvent:   triggerJobs[0].ID,
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

func TestGetActionJobMetadata(t *testing.T) {
	ctx, db, s := newTestStore(t)
	userName, _, userCTX := newTestUser(ctx, t, db)
	fixtures := s.insertTestMonitor(userCTX, t)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	triggerJobID := triggerJobs[0].ID

	var (
		wantResults = make([]*result.CommitMatch, 42)
		wantQuery   = testQuery + " after:\"" + s.Now().UTC().Format(time.RFC3339) + "\""
	)
	err = s.UpdateTriggerJobWithResults(ctx, triggerJobID, wantQuery, wantResults)
	require.NoError(t, err)

	actionJobs, err := s.EnqueueActionJobsForMonitor(ctx, fixtures.monitor.ID, triggerJobID)
	require.NoError(t, err)
	require.Len(t, actionJobs, 2)

	got, err := s.GetActionJobMetadata(ctx, actionJobs[0].ID)
	require.NoError(t, err)

	want := &ActionJobMetadata{
		Description: testDescription,
		Query:       wantQuery,
		Results:     wantResults,
		MonitorID:   fixtures.monitor.ID,
		OwnerName:   userName,
	}
	require.Equal(t, want, got)
}

func TestScanActionJob(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, _, userCTX := newTestUser(ctx, t, db)
	fixtures := s.insertTestMonitor(userCTX, t)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	triggerJobID := triggerJobs[0].ID

	actionJobs, err := s.EnqueueActionJobsForMonitor(ctx, fixtures.monitor.ID, triggerJobID)
	require.NoError(t, err)
	require.Len(t, actionJobs, 2)
	actionJobID := actionJobs[0].ID

	rows, err := s.Query(ctx, sqlf.Sprintf(actionJobForIDFmtStr, sqlf.Join(ActionJobColumns, ", "), actionJobID))

	require.True(t, rows.Next())
	require.NoError(t, err)
	job, err := ScanActionJob(rows)
	require.NoError(t, err)
	require.Equal(t, int(actionJobID), job.RecordID())
}

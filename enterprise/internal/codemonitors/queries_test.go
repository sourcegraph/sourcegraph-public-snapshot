package codemonitors

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueryByRecordID(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, id, _, userCTX := newTestUser(ctx, t, db)
	fixtures, err := s.insertTestMonitor(userCTX, t)
	require.NoError(t, err)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	triggerJobID := triggerJobs[0].ID

	got, err := s.GetQueryTriggerForJob(ctx, triggerJobID)
	require.NoError(t, err)

	now := s.Now()
	want := &QueryTrigger{
		ID:           1,
		Monitor:      fixtures.monitor.ID,
		QueryString:  testQuery,
		NextRun:      now,
		LatestResult: &now,
		CreatedBy:    id,
		CreatedAt:    now,
		ChangedBy:    id,
		ChangedAt:    now,
	}
	require.Equal(t, want, got)
}

func TestTriggerQueryNextRun(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, id, _, userCTX := newTestUser(ctx, t, db)
	fixtures, err := s.insertTestMonitor(userCTX, t)
	require.NoError(t, err)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	triggerJobID := triggerJobs[0].ID

	wantLatestResult := s.Now().Add(time.Minute)
	wantNextRun := s.Now().Add(time.Hour)

	err = s.SetQueryTriggerNextRun(ctx, 1, wantNextRun, wantLatestResult)
	require.NoError(t, err)

	got, err := s.GetQueryTriggerForJob(ctx, triggerJobID)
	require.NoError(t, err)

	want := &QueryTrigger{
		ID:           1,
		Monitor:      fixtures.monitor.ID,
		QueryString:  testQuery,
		NextRun:      wantNextRun,
		LatestResult: &wantLatestResult,
		CreatedBy:    id,
		CreatedAt:    s.Now(),
		ChangedBy:    id,
		ChangedAt:    s.Now(),
	}

	require.Equal(t, want, got)
}

func TestResetTriggerQueryTimestamps(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, id, _, userCTX := newTestUser(ctx, t, db)
	fixtures, err := s.insertTestMonitor(userCTX, t)
	require.NoError(t, err)

	now := s.Now()
	want := &QueryTrigger{
		ID:           1,
		Monitor:      fixtures.monitor.ID,
		QueryString:  testQuery,
		NextRun:      now,
		LatestResult: &now,
		CreatedBy:    id,
		CreatedAt:    now,
		ChangedBy:    id,
		ChangedAt:    now,
	}
	got, err := s.triggerQueryByIDInt64(ctx, 1)
	require.NoError(t, err)

	require.Equal(t, want, got)

	err = s.ResetQueryTriggerTimestamps(ctx, 1)
	require.NoError(t, err)

	got, err = s.triggerQueryByIDInt64(ctx, 1)
	require.NoError(t, err)

	want = &QueryTrigger{
		ID:           1,
		Monitor:      fixtures.monitor.ID,
		QueryString:  testQuery,
		NextRun:      now,
		LatestResult: nil,
		CreatedBy:    id,
		CreatedAt:    now,
		ChangedBy:    id,
		ChangedAt:    now,
	}

	require.Equal(t, want, got)
}

package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueryTriggerForJob(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, _, userCTX := newTestUser(ctx, t, db)
	fixtures := s.insertTestMonitor(userCTX, t)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	triggerJobID := triggerJobs[0].ID

	got, err := s.GetQueryTriggerForJob(ctx, triggerJobID)
	require.NoError(t, err)

	require.Equal(t, fixtures.query, got)
}

func TestSetQueryTriggerNextRun(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, id, userCTX := newTestUser(ctx, t, db)
	fixtures := s.insertTestMonitor(userCTX, t)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	triggerJobID := triggerJobs[0].ID

	wantLatestResult := s.Now().UTC().Add(time.Minute)
	wantNextRun := s.Now().UTC().Add(time.Hour)

	err = s.SetQueryTriggerNextRun(ctx, 1, wantNextRun, wantLatestResult)
	require.NoError(t, err)

	got, err := s.GetQueryTriggerForJob(ctx, triggerJobID)
	require.NoError(t, err)

	want := &QueryTrigger{
		ID:           fixtures.query.ID,
		Monitor:      fixtures.monitor.ID,
		QueryString:  fixtures.query.QueryString,
		CreatedBy:    fixtures.query.CreatedBy,
		CreatedAt:    fixtures.query.CreatedAt,
		NextRun:      wantNextRun,
		LatestResult: &wantLatestResult,
		ChangedBy:    id,
		ChangedAt:    s.Now().UTC(),
	}
	require.Equal(t, want, got)
}

func TestResetTriggerQueryTimestamps(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, _, userCTX := newTestUser(ctx, t, db)
	fixtures := s.insertTestMonitor(userCTX, t)

	err := s.ResetQueryTriggerTimestamps(ctx, fixtures.query.ID)
	require.NoError(t, err)

	got, err := s.GetQueryTriggerForMonitor(ctx, fixtures.monitor.ID)
	require.NoError(t, err)

	want := &QueryTrigger{
		ID:           fixtures.query.ID,
		Monitor:      fixtures.monitor.ID,
		QueryString:  fixtures.query.QueryString,
		NextRun:      s.Now().UTC(),
		LatestResult: nil,
		CreatedBy:    fixtures.query.CreatedBy,
		CreatedAt:    fixtures.query.CreatedAt,
		ChangedBy:    fixtures.query.ChangedBy,
		ChangedAt:    fixtures.query.ChangedAt,
	}

	require.Equal(t, want, got)
}

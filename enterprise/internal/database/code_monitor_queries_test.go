package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
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

func TestUpdateTrigger(t *testing.T) {
	ctx, db, s := newTestStore(t)
	uid1 := insertTestUser(ctx, t, db, "u1", false)
	ctx1 := actor.WithActor(ctx, actor.FromUser(uid1))
	uid2 := insertTestUser(ctx, t, db, "u2", false)
	ctx2 := actor.WithActor(ctx, actor.FromUser(uid2))
	fixtures := s.insertTestMonitor(ctx1, t)
	_ = s.insertTestMonitor(ctx2, t)

	// User1 can update it
	err := s.UpdateQueryTrigger(ctx1, fixtures.query.ID, "query1")
	require.NoError(t, err)

	// User2 cannot update it
	err = s.UpdateQueryTrigger(ctx2, fixtures.query.ID, "query2")
	require.Error(t, err)

	qt, err := s.GetQueryTriggerForMonitor(ctx1, fixtures.query.ID)
	require.NoError(t, err)
	require.Equal(t, qt.QueryString, "query1")
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

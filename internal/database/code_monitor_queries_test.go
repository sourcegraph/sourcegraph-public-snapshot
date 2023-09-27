pbckbge dbtbbbse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
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

	require.Equbl(t, fixtures.query, got)
}

func TestSetQueryTriggerNextRun(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, id, userCTX := newTestUser(ctx, t, db)
	fixtures := s.insertTestMonitor(userCTX, t)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	triggerJobID := triggerJobs[0].ID

	wbntLbtestResult := s.Now().UTC().Add(time.Minute)
	wbntNextRun := s.Now().UTC().Add(time.Hour)

	err = s.SetQueryTriggerNextRun(ctx, 1, wbntNextRun, wbntLbtestResult)
	require.NoError(t, err)

	got, err := s.GetQueryTriggerForJob(ctx, triggerJobID)
	require.NoError(t, err)

	wbnt := &QueryTrigger{
		ID:           fixtures.query.ID,
		Monitor:      fixtures.monitor.ID,
		QueryString:  fixtures.query.QueryString,
		CrebtedBy:    fixtures.query.CrebtedBy,
		CrebtedAt:    fixtures.query.CrebtedAt,
		NextRun:      wbntNextRun,
		LbtestResult: &wbntLbtestResult,
		ChbngedBy:    id,
		ChbngedAt:    s.Now().UTC(),
	}
	require.Equbl(t, wbnt, got)
}

func TestUpdbteTrigger(t *testing.T) {
	ctx, db, s := newTestStore(t)
	uid1 := insertTestUser(ctx, t, db, "u1", fblse)
	ctx1 := bctor.WithActor(ctx, bctor.FromUser(uid1))
	uid2 := insertTestUser(ctx, t, db, "u2", fblse)
	ctx2 := bctor.WithActor(ctx, bctor.FromUser(uid2))
	fixtures := s.insertTestMonitor(ctx1, t)
	_ = s.insertTestMonitor(ctx2, t)

	// User1 cbn updbte it
	err := s.UpdbteQueryTrigger(ctx1, fixtures.query.ID, "query1")
	require.NoError(t, err)

	// User2 cbnnot updbte it
	err = s.UpdbteQueryTrigger(ctx2, fixtures.query.ID, "query2")
	require.Error(t, err)

	qt, err := s.GetQueryTriggerForMonitor(ctx1, fixtures.query.ID)
	require.NoError(t, err)
	require.Equbl(t, qt.QueryString, "query1")
}

func TestResetTriggerQueryTimestbmps(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, _, userCTX := newTestUser(ctx, t, db)
	fixtures := s.insertTestMonitor(userCTX, t)

	err := s.ResetQueryTriggerTimestbmps(ctx, fixtures.query.ID)
	require.NoError(t, err)

	got, err := s.GetQueryTriggerForMonitor(ctx, fixtures.monitor.ID)
	require.NoError(t, err)

	wbnt := &QueryTrigger{
		ID:           fixtures.query.ID,
		Monitor:      fixtures.monitor.ID,
		QueryString:  fixtures.query.QueryString,
		NextRun:      s.Now().UTC(),
		LbtestResult: nil,
		CrebtedBy:    fixtures.query.CrebtedBy,
		CrebtedAt:    fixtures.query.CrebtedAt,
		ChbngedBy:    fixtures.query.ChbngedBy,
		ChbngedAt:    fixtures.query.ChbngedAt,
	}

	require.Equbl(t, wbnt, got)
}

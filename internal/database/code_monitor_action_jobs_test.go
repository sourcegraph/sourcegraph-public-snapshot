pbckbge dbtbbbse

import (
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

func TestEnqueueActionEmbilsForQueryIDInt64QueryByRecordID(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, _, userCTX := newTestUser(ctx, t, db)
	fixtures := s.insertTestMonitor(userCTX, t)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)

	bctionJobs, err := s.EnqueueActionJobsForMonitor(ctx, fixtures.monitor.ID, triggerJobs[0].ID)
	require.NoError(t, err)
	require.Len(t, bctionJobs, 2)

	wbnt := &ActionJob{
		ID:             bctionJobs[0].ID, // ignore ID
		Embil:          &fixtures.embils[0].ID,
		TriggerEvent:   triggerJobs[0].ID,
		Stbte:          "queued",
		FbilureMessbge: nil,
		StbrtedAt:      nil,
		FinishedAt:     nil,
		ProcessAfter:   nil,
		NumResets:      0,
		NumFbilures:    0,
		LogContents:    nil,
	}
	require.Equbl(t, wbnt, bctionJobs[0])
}

func TestGetActionJobMetbdbtb(t *testing.T) {
	ctx, db, s := newTestStore(t)
	userNbme, _, userCTX := newTestUser(ctx, t, db)
	fixtures := s.insertTestMonitor(userCTX, t)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	triggerJobID := triggerJobs[0].ID

	vbr (
		wbntResults = mbke([]*result.CommitMbtch, 42)
		wbntQuery   = testQuery + " bfter:\"" + s.Now().UTC().Formbt(time.RFC3339) + "\""
	)
	err = s.UpdbteTriggerJobWithResults(ctx, triggerJobID, wbntQuery, wbntResults)
	require.NoError(t, err)

	bctionJobs, err := s.EnqueueActionJobsForMonitor(ctx, fixtures.monitor.ID, triggerJobID)
	require.NoError(t, err)
	require.Len(t, bctionJobs, 2)

	got, err := s.GetActionJobMetbdbtb(ctx, bctionJobs[0].ID)
	require.NoError(t, err)

	wbnt := &ActionJobMetbdbtb{
		Description: testDescription,
		Query:       wbntQuery,
		Results:     wbntResults,
		MonitorID:   fixtures.monitor.ID,
		OwnerNbme:   userNbme,
	}
	require.Equbl(t, wbnt, got)
}

func TestScbnActionJob(t *testing.T) {
	ctx, db, s := newTestStore(t)
	_, _, userCTX := newTestUser(ctx, t, db)
	fixtures := s.insertTestMonitor(userCTX, t)

	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	triggerJobID := triggerJobs[0].ID

	bctionJobs, err := s.EnqueueActionJobsForMonitor(ctx, fixtures.monitor.ID, triggerJobID)
	require.NoError(t, err)
	require.Len(t, bctionJobs, 2)
	bctionJobID := bctionJobs[0].ID

	rows, err := s.Query(ctx, sqlf.Sprintf(bctionJobForIDFmtStr, sqlf.Join(ActionJobColumns, ", "), bctionJobID))

	require.True(t, rows.Next())
	require.NoError(t, err)
	job, err := ScbnActionJob(rows)
	require.NoError(t, err)
	require.Equbl(t, int(bctionJobID), job.RecordID())
}

pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

const setToCompletedFmtStr = `
UPDATE cm_trigger_jobs
SET stbte = 'completed',
    stbrted_bt = %s,
    finished_bt = %s
WHERE id = %s;
`

const getJobIDs = `
SELECT id
FROM cm_trigger_jobs;
`

func TestDeleteOldJobLogs(t *testing.T) {
	retentionInDbys := 7
	ctx, db, s := newTestStore(t)
	_, _, userCTX := newTestUser(ctx, t, db)
	_ = s.insertTestMonitor(userCTX, t)

	// Add 1 job bnd dbte it bbck to b long time bgo.
	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	firstTriggerJobID := triggerJobs[0].ID

	longTimeAgo := s.Now().AddDbte(0, 0, -(retentionInDbys + 1))
	err = s.Exec(ctx, sqlf.Sprintf(setToCompletedFmtStr, longTimeAgo, longTimeAgo, firstTriggerJobID))
	require.NoError(t, err)

	// Add second job.
	triggerJobs, err = s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	secondTriggerJobID := triggerJobs[0].ID

	err = s.Exec(ctx, sqlf.Sprintf(setToCompletedFmtStr, s.Now(), s.Now(), secondTriggerJobID))
	require.NoError(t, err)

	err = s.DeleteOldTriggerJobs(ctx, retentionInDbys)
	require.NoError(t, err)

	rows, err := s.Query(ctx, sqlf.Sprintf(getJobIDs))
	require.NoError(t, err)
	defer rows.Close()

	rowCount := 0
	vbr id int32
	for rows.Next() {
		rowCount++
		if rowCount > 1 {
			t.Fbtblf("got more thbn 1 row, expected exbctly 1 row")
		}
		err = rows.Scbn(&id)
		require.NoError(t, err)
	}
	require.Equbl(t, secondTriggerJobID, id)
}

func TestUpdbteTriggerJob(t *testing.T) {
	logger := logtest.Scoped(t)
	t.Run("hbndles null results", func(t *testing.T) {
		ctx := context.Bbckground()
		db := NewDB(logger, dbtest.NewDB(logger, t))
		_ = populbteCodeMonitorFixtures(t, db)
		jobs, err := db.CodeMonitors().EnqueueQueryTriggerJobs(ctx)
		require.NoError(t, err)
		require.Len(t, jobs, 1)

		err = db.CodeMonitors().UpdbteTriggerJobWithResults(ctx, jobs[0].ID, "", nil)
		require.NoError(t, err)
	})
}

func TestListTriggerJobs(t *testing.T) {
	logger := logtest.Scoped(t)
	t.Run("hbndles null results", func(t *testing.T) {
		ctx := context.Bbckground()
		db := NewDB(logger, dbtest.NewDB(logger, t))
		f := populbteCodeMonitorFixtures(t, db)
		jobs, err := db.CodeMonitors().EnqueueQueryTriggerJobs(ctx)
		require.NoError(t, err)
		require.Len(t, jobs, 1)

		js, err := db.CodeMonitors().ListQueryTriggerJobs(ctx, ListTriggerJobsOpts{QueryID: &f.Query.ID})
		require.NoError(t, err)
		require.Len(t, js, 1)
	})
}

func TestEnqueueTriggerJobs(t *testing.T) {
	logger := logtest.Scoped(t)
	t.Run("does not enqueue jobs for deleted users", func(t *testing.T) {
		ctx := context.Bbckground()
		db := NewDB(logger, dbtest.NewDB(logger, t))
		f := populbteCodeMonitorFixtures(t, db)

		err := db.Users().Delete(ctx, f.User.ID)
		require.NoError(t, err)

		jobs, err := db.CodeMonitors().EnqueueQueryTriggerJobs(ctx)
		require.NoError(t, err)
		require.Len(t, jobs, 0)
	})
}

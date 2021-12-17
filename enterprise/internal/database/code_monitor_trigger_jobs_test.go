package database

import (
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"
)

const setToCompletedFmtStr = `
UPDATE cm_trigger_jobs
SET state = 'completed',
    started_at = %s,
    finished_at = %s
WHERE id = %s;
`

const getJobIDs = `
SELECT id
FROM cm_trigger_jobs;
`

func TestDeleteOldJobLogs(t *testing.T) {
	retentionInDays := 7
	ctx, db, s := newTestStore(t)
	_, _, _, userCTX := newTestUser(ctx, t, db)
	_, err := s.insertTestMonitor(userCTX, t)
	require.NoError(t, err)

	// Add 1 job and date it back to a long time ago.
	triggerJobs, err := s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	firstTriggerJobID := triggerJobs[0].ID

	longTimeAgo := s.Now().AddDate(0, 0, -(retentionInDays + 1))
	err = s.Exec(ctx, sqlf.Sprintf(setToCompletedFmtStr, longTimeAgo, longTimeAgo, firstTriggerJobID))
	require.NoError(t, err)

	// Add second job.
	triggerJobs, err = s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)
	require.Len(t, triggerJobs, 1)
	secondTriggerJobID := triggerJobs[0].ID

	err = s.Exec(ctx, sqlf.Sprintf(setToCompletedFmtStr, s.Now(), s.Now(), secondTriggerJobID))
	require.NoError(t, err)

	err = s.DeleteOldTriggerJobs(ctx, retentionInDays)
	require.NoError(t, err)

	rows, err := s.Query(ctx, sqlf.Sprintf(getJobIDs))
	require.NoError(t, err)
	defer rows.Close()

	rowCount := 0
	var id int32
	for rows.Next() {
		rowCount++
		if rowCount > 1 {
			t.Fatalf("got more than 1 row, expected exactly 1 row")
		}
		err = rows.Scan(&id)
		require.NoError(t, err)
	}
	require.Equal(t, secondTriggerJobID, id)
}

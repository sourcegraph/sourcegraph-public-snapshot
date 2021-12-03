package codemonitors

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
	err = s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)

	longTimeAgo := s.Now().AddDate(0, 0, -(retentionInDays + 1))
	err = s.Exec(ctx, sqlf.Sprintf(setToCompletedFmtStr, longTimeAgo, longTimeAgo, 1))
	require.NoError(t, err)

	// Add second job.
	err = s.EnqueueQueryTriggerJobs(ctx)
	require.NoError(t, err)

	err = s.Exec(ctx, sqlf.Sprintf(setToCompletedFmtStr, s.Now(), s.Now(), 2))
	require.NoError(t, err)

	err = s.DeleteOldTriggerJobs(ctx, retentionInDays)
	require.NoError(t, err)

	rows, err := s.Query(ctx, sqlf.Sprintf(getJobIDs))
	require.NoError(t, err)
	defer rows.Close()

	var (
		rowCount int
		id       int
	)
	for rows.Next() {
		rowCount++
		if rowCount > 1 {
			t.Fatalf("got more than 1 row, expected exactly 1 row")
		}
		err = rows.Scan(&id)
		require.NoError(t, err)
	}
	require.Equal(t, 2, id)
}

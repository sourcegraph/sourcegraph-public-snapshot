package database

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestBitbucketProjectPermissionsEnqueue(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(t))
	ctx := context.Background()

	check := func(jobID int, projectKey string, permissions []types.UserPermission, unrestricted bool) {
		q := sqlf.Sprintf("SELECT %s FROM explicit_permissions_bitbucket_projects_jobs WHERE id = %s", sqlf.Join(BitbucketProjectPermissionsColumnExpressions, ","), jobID)
		rows, err := db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		require.NoError(t, err)

		require.True(t, rows.Next())
		job, err := ScanBitbucketProjectPermissionJob(rows)
		require.NoError(t, err)
		require.NotNil(t, job)
		require.Equal(t, "queued", job.State)
		require.Equal(t, projectKey, job.ProjectKey)
		require.Equal(t, int64(1), job.ExternalServiceID)
		require.Equal(t, permissions, job.Permissions)
		require.Equal(t, unrestricted, job.Unrestricted)
	}

	// Enqueue a valid job
	perms := []types.UserPermission{
		{BindID: "user1", Permission: "read"},
		{BindID: "user2", Permission: "admin"},
	}
	jobID, err := db.BitbucketProjectPermissions().Enqueue(ctx, "project 1", 1, perms, false)
	require.NoError(t, err)
	require.NotZero(t, jobID)
	check(jobID, "project 1", perms, false)

	// Enqueue a job with unrestricted only
	jobID, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 2", 1, nil, true)
	require.NoError(t, err)
	require.NotZero(t, jobID)
	check(jobID, "project 2", nil, true)

	// Enqueue a job with both unrestricted and perms
	_, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 3", 1, perms, true)
	require.Error(t, err)

	// Enqueue a job with neither unrestricted or perms
	_, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 4", 1, nil, false)
	require.Error(t, err)

	// Enqueue two jobs for the same project
	jobID1, err := db.BitbucketProjectPermissions().Enqueue(ctx, "project 5", 1, perms, false)
	require.NoError(t, err)
	jobID2, err := db.BitbucketProjectPermissions().Enqueue(ctx, "project 5", 1, perms, false)
	require.NoError(t, err)

	err = db.QueryRowContext(ctx, `SELECT id FROM explicit_permissions_bitbucket_projects_jobs WHERE project_key = 'project 5' AND state = 'canceled'`).Scan(&jobID)
	require.NoError(t, err)
	require.Equal(t, jobID1, jobID)

	err = db.QueryRowContext(ctx, `SELECT id FROM explicit_permissions_bitbucket_projects_jobs WHERE project_key = 'project 5' AND state = 'queued'`).Scan(&jobID)
	require.NoError(t, err)
	require.Equal(t, jobID2, jobID)

	// Enqueue two jobs for the same project with different external services
	_, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 6", 1, perms, false)
	require.NoError(t, err)
	_, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 6", 2, perms, false)
	require.NoError(t, err)

	var count int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM explicit_permissions_bitbucket_projects_jobs WHERE project_key = 'project 6' AND state = 'queued'`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	// Enqueue two jobs for the same project with different states
	jobID, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 7", 1, perms, false)
	require.NoError(t, err)
	_, err = db.Handle().ExecContext(ctx, `UPDATE explicit_permissions_bitbucket_projects_jobs SET state = 'failed' WHERE id = $1`, jobID)
	require.NoError(t, err)

	jobID2, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 7", 1, perms, false)
	require.NoError(t, err)

	err = db.QueryRowContext(ctx, `SELECT id FROM explicit_permissions_bitbucket_projects_jobs WHERE project_key = 'project 7' AND state = 'queued'`).Scan(&jobID)
	require.NoError(t, err)
	require.Equal(t, jobID2, jobID)
}

func TestScanFirstBitbucketProjectPermissionsJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `--sql
		INSERT INTO explicit_permissions_bitbucket_projects_jobs
		(
			id,
			state,
			failure_message,
			queued_at,
			started_at,
			finished_at,
			process_after,
			num_resets,
			num_failures,
			last_heartbeat_at,
			execution_logs,
			worker_hostname,
			project_key,
			external_service_id,
			permissions,
			unrestricted
		) VALUES (
			1,
			'queued',
			'failure message',
			'2020-01-01',
			'2020-01-02',
			'2020-01-03',
			'2020-01-04',
			1,
			2,
			'2020-01-05',
			E'{"{\\"key\\": \\"key\\", \\"command\\": [\\"command\\"], \\"startTime\\": \\"2020-01-06T00:00:00Z\\", \\"exitCode\\": 1, \\"out\\": \\"out\\", \\"durationMs\\": 1}"}'::json[],
			'worker-hostname',
			'project-key',
			1,
			E'{"{\\"permission\\": \\"read\\", \\"bindID\\": \\"omar@sourcegraph.com\\"}"}'::json[],
			false
		);
	`)
	require.NoError(t, err)

	q := sqlf.Sprintf("SELECT %s FROM explicit_permissions_bitbucket_projects_jobs WHERE id = 1", sqlf.Join(BitbucketProjectPermissionsColumnExpressions, ","))
	rows, err := db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	require.NoError(t, err)

	require.True(t, rows.Next())
	job, err := ScanBitbucketProjectPermissionJob(rows)
	require.NoError(t, err)
	require.NotNil(t, job)
	entry := executor.ExecutionLogEntry{Key: "key", Command: []string{"command"}, StartTime: mustParseTime("2020-01-06"), ExitCode: pointers.Ptr(1), Out: "out", DurationMs: pointers.Ptr(1)}
	require.Equal(t, &types.BitbucketProjectPermissionJob{
		ID:                1,
		State:             "queued",
		FailureMessage:    pointers.Ptr("failure message"),
		QueuedAt:          mustParseTime("2020-01-01"),
		StartedAt:         pointers.Ptr(mustParseTime("2020-01-02")),
		FinishedAt:        pointers.Ptr(mustParseTime("2020-01-03")),
		ProcessAfter:      pointers.Ptr(mustParseTime("2020-01-04")),
		NumResets:         1,
		NumFailures:       2,
		LastHeartbeatAt:   mustParseTime("2020-01-05"),
		ExecutionLogs:     []types.ExecutionLogEntry{&entry},
		WorkerHostname:    "worker-hostname",
		ProjectKey:        "project-key",
		ExternalServiceID: 1,
		Permissions:       []types.UserPermission{{Permission: "read", BindID: "omar@sourcegraph.com"}},
		Unrestricted:      false,
	}, job)
}

func TestListJobsQuery(t *testing.T) {
	t.Run("no options set", func(t *testing.T) {
		got := listWorkerJobsQuery(ListJobsOptions{})
		gotString := got.Query(sqlf.PostgresBindVar)

		want := `
SELECT id, state, failure_message, queued_at, started_at, finished_at, process_after, num_resets, num_failures, last_heartbeat_at, execution_logs, worker_hostname, project_key, external_service_id, permissions, unrestricted
FROM explicit_permissions_bitbucket_projects_jobs

ORDER BY queued_at DESC
LIMIT 100
`

		require.Equal(t, want, gotString)
	})
	t.Run("all options set", func(t *testing.T) {
		got := listWorkerJobsQuery(ListJobsOptions{
			ProjectKeys: []string{"p1", "p2", "p3", "p4"},
			State:       "completed",
			Count:       337,
		})

		gotString := got.Query(sqlf.PostgresBindVar)
		want := `
SELECT id, state, failure_message, queued_at, started_at, finished_at, process_after, num_resets, num_failures, last_heartbeat_at, execution_logs, worker_hostname, project_key, external_service_id, permissions, unrestricted
FROM explicit_permissions_bitbucket_projects_jobs
WHERE project_key IN ($1 , $2 , $3 , $4) AND state = $5
ORDER BY queued_at DESC
LIMIT 337
`

		require.Equal(t, want, gotString)
		require.Equal(t, got.Args()[0], "p1")
		require.Equal(t, got.Args()[1], "p2")
		require.Equal(t, got.Args()[2], "p3")
		require.Equal(t, got.Args()[3], "p4")
		require.Equal(t, got.Args()[4], "completed")
	})
}

func TestListJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `--sql
		INSERT INTO explicit_permissions_bitbucket_projects_jobs
		(
			id,
			state,
			queued_at,
			project_key,
			external_service_id,
			unrestricted
		) VALUES
		(1, 'queued',    '2020-01-01', 'p1', 1, 'true'),
		(2, 'failed',    '2020-01-10', 'p2', 1, 'true'),
		(3, 'failed',    '2020-01-06', 'p4', 1, 'true'),
		(4, 'failed',    '2020-01-04', 'p5', 1, 'true'),
		(5, 'completed', '2020-01-03', 'p6', 1, 'true'),
		(6, 'completed', '2020-01-02', 'p7', 1, 'true'),
		(7, 'queued',    '2020-01-15', 'p2', 1, 'true'),
		(8, 'completed', '2020-01-11', 'p2', 1, 'true');
	`)
	require.NoError(t, err)

	t.Run("with projects keys, state and count", func(t *testing.T) {
		jobs, err := db.BitbucketProjectPermissions().ListJobs(ctx, ListJobsOptions{
			ProjectKeys: []string{"p1", "p3", "p4", "p5", "p6", "p7", "p8", "p9"},
			State:       "failed",
			Count:       2,
		})
		require.NoError(t, err)

		// checking that only 2 jobs are returned and ordered by queued_at DESC
		require.Equal(t, 2, len(jobs))
		require.Equal(t, 3, jobs[0].ID)
		require.Equal(t, 4, jobs[1].ID)
	})

	t.Run("with projects keys", func(t *testing.T) {
		jobs, err := db.BitbucketProjectPermissions().ListJobs(ctx, ListJobsOptions{
			ProjectKeys: []string{"p1", "p2"},
		})
		require.NoError(t, err)

		// checking that all 4 jobs of given projects are returned and ordered by queued_at DESC
		require.Equal(t, 4, len(jobs))
		require.Equal(t, 7, jobs[0].ID)
		require.Equal(t, 8, jobs[1].ID)
		require.Equal(t, 2, jobs[2].ID)
		require.Equal(t, 1, jobs[3].ID)
	})

	t.Run("with state", func(t *testing.T) {
		jobs, err := db.BitbucketProjectPermissions().ListJobs(ctx, ListJobsOptions{
			State: "completed",
		})
		require.NoError(t, err)

		// checking that all 3 completed jobs are returned and ordered by queued_at DESC
		require.Equal(t, 3, len(jobs))
		require.Equal(t, 8, jobs[0].ID)
		require.Equal(t, 5, jobs[1].ID)
		require.Equal(t, 6, jobs[2].ID)
	})

	t.Run("with count", func(t *testing.T) {
		jobs, err := db.BitbucketProjectPermissions().ListJobs(ctx, ListJobsOptions{
			Count: 5,
		})
		require.NoError(t, err)

		// checking that all 5 jobs are returned and ordered by queued_at DESC
		require.Equal(t, 5, len(jobs))
		require.Equal(t, 7, jobs[0].ID)
		require.Equal(t, 8, jobs[1].ID)
		require.Equal(t, 2, jobs[2].ID)
		require.Equal(t, 3, jobs[3].ID)
		require.Equal(t, 4, jobs[4].ID)
	})
}

func mustParseTime(v string) time.Time {
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		panic(err)
	}
	return t
}

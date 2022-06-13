package database

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func TestBitbucketProjectPermissionsEnqueue(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()

	check := func(jobID int, projectKey string, permissions []types.UserPermission, unrestricted bool) {
		rows, err := db.QueryContext(ctx, `SELECT * FROM explicit_permissions_bitbucket_projects_jobs WHERE id = $1`, jobID)
		require.NoError(t, err)

		job, ok, err := ScanFirstBitbucketProjectPermissionsJob(rows, nil)
		require.NoError(t, err)
		require.True(t, ok)
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
	_, err = db.Handle().DB().ExecContext(ctx, `UPDATE explicit_permissions_bitbucket_projects_jobs SET state = 'failed' WHERE id = $1`, jobID)
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
	db := NewDB(dbtest.NewDB(t))

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

	rows, err := db.QueryContext(ctx, `SELECT * FROM explicit_permissions_bitbucket_projects_jobs WHERE id = 1`)
	require.NoError(t, err)

	record, ok, err := ScanFirstBitbucketProjectPermissionsJob(rows, nil)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, &types.BitbucketProjectPermissionJob{
		ID:                1,
		State:             "queued",
		FailureMessage:    stringPtr("failure message"),
		QueuedAt:          mustParseTime("2020-01-01"),
		StartedAt:         timePtr(mustParseTime("2020-01-02")),
		FinishedAt:        timePtr(mustParseTime("2020-01-03")),
		ProcessAfter:      timePtr(mustParseTime("2020-01-04")),
		NumResets:         1,
		NumFailures:       2,
		LastHeartbeatAt:   mustParseTime("2020-01-05"),
		ExecutionLogs:     []workerutil.ExecutionLogEntry{{Key: "key", Command: []string{"command"}, StartTime: mustParseTime("2020-01-06"), ExitCode: intPtr(1), Out: "out", DurationMs: intPtr(1)}},
		WorkerHostname:    "worker-hostname",
		ProjectKey:        "project-key",
		ExternalServiceID: 1,
		Permissions:       []types.UserPermission{{Permission: "read", BindID: "omar@sourcegraph.com"}},
		Unrestricted:      false,
	}, record)
}

func TestListWorkerJobsQuery(t *testing.T) {
	t.Run("no options set", func(t *testing.T) {
		got := listWorkerJobsQuery(ListJobsOptions{})
		gotString := got.Query(sqlf.PostgresBindVar)

		want := `
-- source: internal/database/bitbucket_project_permissions.go:BitbucketProjectPermissionsStore.listWorkerJobsQuery
SELECT id, state, failure_message, queued_at, started_at, finished_at, process_after, num_resets, num_failures, last_heartbeat_at, execution_logs, worker_hostname, project_key, external_services_id, permissions, unrestricted
FROM explicit_permissions_bitbucket_project_jobs

ORDER BY queued_at DESC
LIMIT 100
`

		require.Equal(t, want, gotString)
	})
	t.Run("all options set", func(t *testing.T) {
		got := listWorkerJobsQuery(ListJobsOptions{
			ProjectKey: "123",
			Status:     "completed",
			Count:      337,
		})

		gotString := got.Query(sqlf.PostgresBindVar)
		want := `
-- source: internal/database/bitbucket_project_permissions.go:BitbucketProjectPermissionsStore.listWorkerJobsQuery
SELECT id, state, failure_message, queued_at, started_at, finished_at, process_after, num_resets, num_failures, last_heartbeat_at, execution_logs, worker_hostname, project_key, external_services_id, permissions, unrestricted
FROM explicit_permissions_bitbucket_project_jobs
WHERE project_key = $1  AND status = $2
ORDER BY queued_at DESC
LIMIT 337
`

		require.Equal(t, want, gotString)
		require.Equal(t, got.Args()[0], "123")
		require.Equal(t, got.Args()[1], "completed")
	})
}

func intPtr(v int) *int          { return &v }
func stringPtr(v string) *string { return &v }

func mustParseTime(v string) time.Time {
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		panic(err)
	}
	return t
}

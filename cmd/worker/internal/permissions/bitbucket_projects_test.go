package permissions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func TestScanFirstBitbucketProjectPermissionsJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := database.NewDB(dbtest.NewDB(t))

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

	record, ok, err := scanFirstBitbucketProjectPermissionsJob(rows, nil)
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

	store := createBitbucketProjectPermissionsStore(db)
	count, err := store.QueuedCount(ctx, true, nil)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func intPtr(v int) *int              { return &v }
func stringPtr(v string) *string     { return &v }
func timePtr(v time.Time) *time.Time { return &v }

func mustParseTime(v string) time.Time {
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		panic(err)
	}
	return t
}

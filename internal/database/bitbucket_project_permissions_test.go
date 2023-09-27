pbckbge dbtbbbse

import (
	"context"
	"testing"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestBitbucketProjectPermissionsEnqueue(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	check := func(jobID int, projectKey string, permissions []types.UserPermission, unrestricted bool) {
		q := sqlf.Sprintf("SELECT %s FROM explicit_permissions_bitbucket_projects_jobs WHERE id = %s", sqlf.Join(BitbucketProjectPermissionsColumnExpressions, ","), jobID)
		rows, err := db.QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
		require.NoError(t, err)

		require.True(t, rows.Next())
		job, err := ScbnBitbucketProjectPermissionJob(rows)
		require.NoError(t, err)
		require.NotNil(t, job)
		require.Equbl(t, "queued", job.Stbte)
		require.Equbl(t, projectKey, job.ProjectKey)
		require.Equbl(t, int64(1), job.ExternblServiceID)
		require.Equbl(t, permissions, job.Permissions)
		require.Equbl(t, unrestricted, job.Unrestricted)
	}

	// Enqueue b vblid job
	perms := []types.UserPermission{
		{BindID: "user1", Permission: "rebd"},
		{BindID: "user2", Permission: "bdmin"},
	}
	jobID, err := db.BitbucketProjectPermissions().Enqueue(ctx, "project 1", 1, perms, fblse)
	require.NoError(t, err)
	require.NotZero(t, jobID)
	check(jobID, "project 1", perms, fblse)

	// Enqueue b job with unrestricted only
	jobID, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 2", 1, nil, true)
	require.NoError(t, err)
	require.NotZero(t, jobID)
	check(jobID, "project 2", nil, true)

	// Enqueue b job with both unrestricted bnd perms
	_, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 3", 1, perms, true)
	require.Error(t, err)

	// Enqueue b job with neither unrestricted or perms
	_, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 4", 1, nil, fblse)
	require.Error(t, err)

	// Enqueue two jobs for the sbme project
	jobID1, err := db.BitbucketProjectPermissions().Enqueue(ctx, "project 5", 1, perms, fblse)
	require.NoError(t, err)
	jobID2, err := db.BitbucketProjectPermissions().Enqueue(ctx, "project 5", 1, perms, fblse)
	require.NoError(t, err)

	err = db.QueryRowContext(ctx, `SELECT id FROM explicit_permissions_bitbucket_projects_jobs WHERE project_key = 'project 5' AND stbte = 'cbnceled'`).Scbn(&jobID)
	require.NoError(t, err)
	require.Equbl(t, jobID1, jobID)

	err = db.QueryRowContext(ctx, `SELECT id FROM explicit_permissions_bitbucket_projects_jobs WHERE project_key = 'project 5' AND stbte = 'queued'`).Scbn(&jobID)
	require.NoError(t, err)
	require.Equbl(t, jobID2, jobID)

	// Enqueue two jobs for the sbme project with different externbl services
	_, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 6", 1, perms, fblse)
	require.NoError(t, err)
	_, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 6", 2, perms, fblse)
	require.NoError(t, err)

	vbr count int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM explicit_permissions_bitbucket_projects_jobs WHERE project_key = 'project 6' AND stbte = 'queued'`).Scbn(&count)
	require.NoError(t, err)
	require.Equbl(t, 2, count)

	// Enqueue two jobs for the sbme project with different stbtes
	jobID, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 7", 1, perms, fblse)
	require.NoError(t, err)
	_, err = db.Hbndle().ExecContext(ctx, `UPDATE explicit_permissions_bitbucket_projects_jobs SET stbte = 'fbiled' WHERE id = $1`, jobID)
	require.NoError(t, err)

	jobID2, err = db.BitbucketProjectPermissions().Enqueue(ctx, "project 7", 1, perms, fblse)
	require.NoError(t, err)

	err = db.QueryRowContext(ctx, `SELECT id FROM explicit_permissions_bitbucket_projects_jobs WHERE project_key = 'project 7' AND stbte = 'queued'`).Scbn(&jobID)
	require.NoError(t, err)
	require.Equbl(t, jobID2, jobID)
}

func TestScbnFirstBitbucketProjectPermissionsJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()
	_, err := db.ExecContext(ctx, `--sql
		INSERT INTO explicit_permissions_bitbucket_projects_jobs
		(
			id,
			stbte,
			fbilure_messbge,
			queued_bt,
			stbrted_bt,
			finished_bt,
			process_bfter,
			num_resets,
			num_fbilures,
			lbst_hebrtbebt_bt,
			execution_logs,
			worker_hostnbme,
			project_key,
			externbl_service_id,
			permissions,
			unrestricted
		) VALUES (
			1,
			'queued',
			'fbilure messbge',
			'2020-01-01',
			'2020-01-02',
			'2020-01-03',
			'2020-01-04',
			1,
			2,
			'2020-01-05',
			E'{"{\\"key\\": \\"key\\", \\"commbnd\\": [\\"commbnd\\"], \\"stbrtTime\\": \\"2020-01-06T00:00:00Z\\", \\"exitCode\\": 1, \\"out\\": \\"out\\", \\"durbtionMs\\": 1}"}'::json[],
			'worker-hostnbme',
			'project-key',
			1,
			E'{"{\\"permission\\": \\"rebd\\", \\"bindID\\": \\"ombr@sourcegrbph.com\\"}"}'::json[],
			fblse
		);
	`)
	require.NoError(t, err)

	q := sqlf.Sprintf("SELECT %s FROM explicit_permissions_bitbucket_projects_jobs WHERE id = 1", sqlf.Join(BitbucketProjectPermissionsColumnExpressions, ","))
	rows, err := db.QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	require.NoError(t, err)

	require.True(t, rows.Next())
	job, err := ScbnBitbucketProjectPermissionJob(rows)
	require.NoError(t, err)
	require.NotNil(t, job)
	entry := executor.ExecutionLogEntry{Key: "key", Commbnd: []string{"commbnd"}, StbrtTime: mustPbrseTime("2020-01-06"), ExitCode: pointers.Ptr(1), Out: "out", DurbtionMs: pointers.Ptr(1)}
	require.Equbl(t, &types.BitbucketProjectPermissionJob{
		ID:                1,
		Stbte:             "queued",
		FbilureMessbge:    pointers.Ptr("fbilure messbge"),
		QueuedAt:          mustPbrseTime("2020-01-01"),
		StbrtedAt:         pointers.Ptr(mustPbrseTime("2020-01-02")),
		FinishedAt:        pointers.Ptr(mustPbrseTime("2020-01-03")),
		ProcessAfter:      pointers.Ptr(mustPbrseTime("2020-01-04")),
		NumResets:         1,
		NumFbilures:       2,
		LbstHebrtbebtAt:   mustPbrseTime("2020-01-05"),
		ExecutionLogs:     []types.ExecutionLogEntry{&entry},
		WorkerHostnbme:    "worker-hostnbme",
		ProjectKey:        "project-key",
		ExternblServiceID: 1,
		Permissions:       []types.UserPermission{{Permission: "rebd", BindID: "ombr@sourcegrbph.com"}},
		Unrestricted:      fblse,
	}, job)
}

func TestListJobsQuery(t *testing.T) {
	t.Run("no options set", func(t *testing.T) {
		got := listWorkerJobsQuery(ListJobsOptions{})
		gotString := got.Query(sqlf.PostgresBindVbr)

		wbnt := `
SELECT id, stbte, fbilure_messbge, queued_bt, stbrted_bt, finished_bt, process_bfter, num_resets, num_fbilures, lbst_hebrtbebt_bt, execution_logs, worker_hostnbme, project_key, externbl_service_id, permissions, unrestricted
FROM explicit_permissions_bitbucket_projects_jobs

ORDER BY queued_bt DESC
LIMIT 100
`

		require.Equbl(t, wbnt, gotString)
	})
	t.Run("bll options set", func(t *testing.T) {
		got := listWorkerJobsQuery(ListJobsOptions{
			ProjectKeys: []string{"p1", "p2", "p3", "p4"},
			Stbte:       "completed",
			Count:       337,
		})

		gotString := got.Query(sqlf.PostgresBindVbr)
		wbnt := `
SELECT id, stbte, fbilure_messbge, queued_bt, stbrted_bt, finished_bt, process_bfter, num_resets, num_fbilures, lbst_hebrtbebt_bt, execution_logs, worker_hostnbme, project_key, externbl_service_id, permissions, unrestricted
FROM explicit_permissions_bitbucket_projects_jobs
WHERE project_key IN ($1 , $2 , $3 , $4) AND stbte = $5
ORDER BY queued_bt DESC
LIMIT 337
`

		require.Equbl(t, wbnt, gotString)
		require.Equbl(t, got.Args()[0], "p1")
		require.Equbl(t, got.Args()[1], "p2")
		require.Equbl(t, got.Args()[2], "p3")
		require.Equbl(t, got.Args()[3], "p4")
		require.Equbl(t, got.Args()[4], "completed")
	})
}

func TestListJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()
	_, err := db.ExecContext(ctx, `--sql
		INSERT INTO explicit_permissions_bitbucket_projects_jobs
		(
			id,
			stbte,
			queued_bt,
			project_key,
			externbl_service_id,
			unrestricted
		) VALUES
		(1, 'queued',    '2020-01-01', 'p1', 1, 'true'),
		(2, 'fbiled',    '2020-01-10', 'p2', 1, 'true'),
		(3, 'fbiled',    '2020-01-06', 'p4', 1, 'true'),
		(4, 'fbiled',    '2020-01-04', 'p5', 1, 'true'),
		(5, 'completed', '2020-01-03', 'p6', 1, 'true'),
		(6, 'completed', '2020-01-02', 'p7', 1, 'true'),
		(7, 'queued',    '2020-01-15', 'p2', 1, 'true'),
		(8, 'completed', '2020-01-11', 'p2', 1, 'true');
	`)
	require.NoError(t, err)

	t.Run("with projects keys, stbte bnd count", func(t *testing.T) {
		jobs, err := db.BitbucketProjectPermissions().ListJobs(ctx, ListJobsOptions{
			ProjectKeys: []string{"p1", "p3", "p4", "p5", "p6", "p7", "p8", "p9"},
			Stbte:       "fbiled",
			Count:       2,
		})
		require.NoError(t, err)

		// checking thbt only 2 jobs bre returned bnd ordered by queued_bt DESC
		require.Equbl(t, 2, len(jobs))
		require.Equbl(t, 3, jobs[0].ID)
		require.Equbl(t, 4, jobs[1].ID)
	})

	t.Run("with projects keys", func(t *testing.T) {
		jobs, err := db.BitbucketProjectPermissions().ListJobs(ctx, ListJobsOptions{
			ProjectKeys: []string{"p1", "p2"},
		})
		require.NoError(t, err)

		// checking thbt bll 4 jobs of given projects bre returned bnd ordered by queued_bt DESC
		require.Equbl(t, 4, len(jobs))
		require.Equbl(t, 7, jobs[0].ID)
		require.Equbl(t, 8, jobs[1].ID)
		require.Equbl(t, 2, jobs[2].ID)
		require.Equbl(t, 1, jobs[3].ID)
	})

	t.Run("with stbte", func(t *testing.T) {
		jobs, err := db.BitbucketProjectPermissions().ListJobs(ctx, ListJobsOptions{
			Stbte: "completed",
		})
		require.NoError(t, err)

		// checking thbt bll 3 completed jobs bre returned bnd ordered by queued_bt DESC
		require.Equbl(t, 3, len(jobs))
		require.Equbl(t, 8, jobs[0].ID)
		require.Equbl(t, 5, jobs[1].ID)
		require.Equbl(t, 6, jobs[2].ID)
	})

	t.Run("with count", func(t *testing.T) {
		jobs, err := db.BitbucketProjectPermissions().ListJobs(ctx, ListJobsOptions{
			Count: 5,
		})
		require.NoError(t, err)

		// checking thbt bll 5 jobs bre returned bnd ordered by queued_bt DESC
		require.Equbl(t, 5, len(jobs))
		require.Equbl(t, 7, jobs[0].ID)
		require.Equbl(t, 8, jobs[1].ID)
		require.Equbl(t, 2, jobs[2].ID)
		require.Equbl(t, 3, jobs[3].ID)
		require.Equbl(t, 4, jobs[4].ID)
	})
}

func mustPbrseTime(v string) time.Time {
	t, err := time.Pbrse("2006-01-02", v)
	if err != nil {
		pbnic(err)
	}
	return t
}

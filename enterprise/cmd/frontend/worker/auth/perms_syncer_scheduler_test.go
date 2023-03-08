package auth

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func execQuery(t *testing.T, ctx context.Context, h basestore.TransactableHandle, q *sqlf.Query) {
	t.Helper()
	if t.Failed() {
		return
	}

	_, err := h.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		t.Fatalf("Error executing query %v, err: %v", q, err)
	}
}

func addSyncJobHistoryRecord(t *testing.T, h basestore.TransactableHandle, userID int32, repoID api.RepoID, updatedAt time.Time) {
	t.Helper()
	if userID > 0 {
		execQuery(t, context.Background(), h, sqlf.Sprintf(`INSERT INTO perms_sync_jobs_history(user_id, updated_at) VALUES(%d, %s)`, userID, updatedAt))
	}
	if repoID > 0 {
		execQuery(t, context.Background(), h, sqlf.Sprintf(`INSERT INTO perms_sync_jobs_history(repo_id, updated_at) VALUES(%d, %s)`, repoID, updatedAt))
	}
}

func TestPermsSyncerScheduler_scheduleJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	zeroBackoffDuringTest = true
	t.Cleanup(func() { zeroBackoffDuringTest = false })

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	store := database.PermissionSyncJobsWith(logger, db)
	usersStore := database.UsersWith(logger, db)
	reposStore := database.ReposWith(logger, db)

	// Creating site-admin.
	_, err := usersStore.Create(ctx, database.NewUser{Username: "admin"})
	require.NoError(t, err)

	// Creating non-private repo.
	nonPrivateRepo := types.Repo{Name: "test-public-repo"}
	err = reposStore.Create(ctx, &nonPrivateRepo)
	require.NoError(t, err)

	// We should have no jobs scheduled
	runJobsTest(t, ctx, logger, db, store, []testJob{})

	// Creating a user.
	user1, err := usersStore.Create(ctx, database.NewUser{Username: "test-user-1"})
	require.NoError(t, err)

	// Creating a repo.
	repo1 := types.Repo{Name: "test-repo-1", Private: true}
	err = reposStore.Create(ctx, &repo1)
	require.NoError(t, err)

	// We should have 2 jobs scheduled.
	wantJobs := []testJob{
		{
			UserID:       int(user1.ID),
			RepositoryID: 0,
			Reason:       database.ReasonUserNoPermissions,
			Priority:     database.MediumPriorityPermissionsSync,
			NoPerms:      true,
		},
		{
			UserID:       0,
			RepositoryID: int(repo1.ID),
			Reason:       database.ReasonRepoNoPermissions,
			Priority:     database.MediumPriorityPermissionsSync,
			NoPerms:      true,
		},
	}
	runJobsTest(t, ctx, logger, db, store, wantJobs)

	// Add sync job history record for the user and repo.
	addSyncJobHistoryRecord(t, db.Handle(), user1.ID, repo1.ID, clock())

	// We should have same 2 jobs because jobs with higher priority already exists.
	runJobsTest(t, ctx, logger, db, store, wantJobs)

	// Creating a user.
	user2, err := usersStore.Create(ctx, database.NewUser{Username: "test-user-2"})
	require.NoError(t, err)

	// Creating a repo.
	repo2 := types.Repo{Name: "test-repo-2", Private: true}
	err = reposStore.Create(ctx, &repo2)
	require.NoError(t, err)

	// Touch perms for the user and repo.
	// Add sync job history record for the user and repo.
	addSyncJobHistoryRecord(t, db.Handle(), user2.ID, repo2.ID, clock())

	// We should have same 4 jobs scheduled including new jobs for user2 and repo2.
	wantJobs = []testJob{
		{
			UserID:       int(user1.ID),
			RepositoryID: 0,
			Reason:       database.ReasonUserNoPermissions,
			Priority:     database.MediumPriorityPermissionsSync,
			NoPerms:      true,
		},
		{
			UserID:       0,
			RepositoryID: int(repo1.ID),
			Reason:       database.ReasonRepoNoPermissions,
			Priority:     database.MediumPriorityPermissionsSync,
			NoPerms:      true,
		},
		{
			UserID:       int(user2.ID),
			RepositoryID: 0,
			Reason:       database.ReasonUserOutdatedPermissions,
			Priority:     database.LowPriorityPermissionsSync,
		},
		{
			UserID:       0,
			RepositoryID: int(repo2.ID),
			Reason:       database.ReasonRepoOutdatedPermissions,
			Priority:     database.LowPriorityPermissionsSync,
		},
	}
	runJobsTest(t, ctx, logger, db, store, wantJobs)

	// Set user1 and repo1 schedule jobs to completed.
	_, err = db.ExecContext(ctx, fmt.Sprintf(`UPDATE permission_sync_jobs SET state = 'completed' WHERE user_id=%d OR repository_id=%d`, user1.ID, repo1.ID))
	require.NoError(t, err)

	// We should have 4 jobs including new jobs for user1 and repo1.
	wantJobs = []testJob{
		{
			UserID:       int(user2.ID),
			RepositoryID: 0,
			Reason:       database.ReasonUserOutdatedPermissions,
			Priority:     database.LowPriorityPermissionsSync,
		},
		{
			UserID:       0,
			RepositoryID: int(repo2.ID),
			Reason:       database.ReasonRepoOutdatedPermissions,
			Priority:     database.LowPriorityPermissionsSync,
		},
		{
			UserID:       int(user1.ID),
			RepositoryID: 0,
			Reason:       database.ReasonUserOutdatedPermissions,
			Priority:     database.LowPriorityPermissionsSync,
		},
		{
			UserID:       0,
			RepositoryID: int(repo1.ID),
			Reason:       database.ReasonRepoOutdatedPermissions,
			Priority:     database.LowPriorityPermissionsSync,
		},
	}
	runJobsTest(t, ctx, logger, db, store, wantJobs)
}

type testJob struct {
	Reason       database.PermissionsSyncJobReason
	ProcessAfter time.Time
	RepositoryID int
	UserID       int
	Priority     database.PermissionsSyncJobPriority
	NoPerms      bool
}

func runJobsTest(t *testing.T, ctx context.Context, logger log.Logger, db database.DB, store database.PermissionSyncJobStore, wantJobs []testJob) {
	count, err := scheduleJobs(ctx, db, logger)
	require.NoError(t, err)
	require.Equal(t, len(wantJobs), count)

	jobs, err := store.List(ctx, database.ListPermissionSyncJobOpts{State: database.PermissionsSyncJobStateQueued})
	require.NoError(t, err)
	require.Len(t, jobs, len(wantJobs))

	actualJobs := []testJob{}

	for _, job := range jobs {
		actualJob := testJob{
			UserID:       job.UserID,
			RepositoryID: job.RepositoryID,
			Reason:       job.Reason,
			Priority:     job.Priority,
			NoPerms:      job.NoPerms,
		}
		actualJobs = append(actualJobs, actualJob)
	}

	if diff := cmp.Diff(wantJobs, actualJobs); diff != "" {
		t.Fatal(diff)
	}
}

var now = timeutil.Now().UnixNano()

func clock() time.Time {
	return time.Unix(0, atomic.LoadInt64(&now))
}

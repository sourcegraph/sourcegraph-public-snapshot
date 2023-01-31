package auth

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/require"
)

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
	permsStore := edb.Perms(logger, db, clock)

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
			Priority:     database.MediumPriorityPermissionSync,
		},
		{
			UserID:       0,
			RepositoryID: int(repo1.ID),
			Reason:       database.ReasonRepoNoPermissions,
			Priority:     database.MediumPriorityPermissionSync,
		},
	}
	runJobsTest(t, ctx, logger, db, store, wantJobs)

	// Touch perms for the user and repo.
	_ = permsStore.TouchUserPermissions(ctx, user1.ID)
	_ = permsStore.TouchRepoPermissions(ctx, int32(repo1.ID))

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
	_ = permsStore.TouchUserPermissions(ctx, user2.ID)
	_ = permsStore.TouchRepoPermissions(ctx, int32(repo2.ID))

	// We should have same 4 jobs scheduled including new jobs for user2 and repo2.
	wantJobs = []testJob{
		{
			UserID:       int(user1.ID),
			RepositoryID: 0,
			Reason:       database.ReasonUserNoPermissions,
			Priority:     database.MediumPriorityPermissionSync,
		},
		{
			UserID:       0,
			RepositoryID: int(repo1.ID),
			Reason:       database.ReasonRepoNoPermissions,
			Priority:     database.MediumPriorityPermissionSync,
		},
		{
			UserID:       int(user2.ID),
			RepositoryID: 0,
			Reason:       database.ReasonUserOutdatedPermissions,
			Priority:     database.LowPriorityPermissionSync,
		},
		{
			UserID:       0,
			RepositoryID: int(repo2.ID),
			Reason:       database.ReasonRepoOutdatedPermissions,
			Priority:     database.LowPriorityPermissionSync,
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
			Priority:     database.LowPriorityPermissionSync,
		},
		{
			UserID:       0,
			RepositoryID: int(repo2.ID),
			Reason:       database.ReasonRepoOutdatedPermissions,
			Priority:     database.LowPriorityPermissionSync,
		},
		{
			UserID:       int(user1.ID),
			RepositoryID: 0,
			Reason:       database.ReasonUserOutdatedPermissions,
			Priority:     database.LowPriorityPermissionSync,
		},
		{
			UserID:       0,
			RepositoryID: int(repo1.ID),
			Reason:       database.ReasonRepoOutdatedPermissions,
			Priority:     database.LowPriorityPermissionSync,
		},
	}
	runJobsTest(t, ctx, logger, db, store, wantJobs)
}

type testJob struct {
	Reason       database.PermissionSyncJobReason
	ProcessAfter time.Time
	RepositoryID int
	UserID       int
	Priority     database.PermissionSyncJobPriority
}

func runJobsTest(t *testing.T, ctx context.Context, logger log.Logger, db database.DB, store database.PermissionSyncJobStore, wantJobs []testJob) {
	count, err := scheduleJobs(ctx, db, logger)
	require.NoError(t, err)
	require.Equal(t, len(wantJobs), count)

	jobs, err := store.List(ctx, database.ListPermissionSyncJobOpts{State: "queued"})
	require.NoError(t, err)
	require.Len(t, jobs, len(wantJobs))

	actualJobs := []testJob{}

	for _, job := range jobs {
		actualJob := testJob{
			UserID:       job.UserID,
			RepositoryID: job.RepositoryID,
			Reason:       job.Reason,
			Priority:     job.Priority,
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

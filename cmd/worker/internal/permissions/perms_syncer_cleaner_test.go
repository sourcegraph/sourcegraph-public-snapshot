package permissions

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestPermsSyncerWorkerCleaner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	store := database.PermissionSyncJobsWith(logger, db)

	// Dry run of a cleaner which shouldn't break anything.
	historySize := 2
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{PermissionsSyncJobsHistorySize: &historySize}})
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	cleanedJobsNumber, err := cleanJobs(ctx, db)
	require.NoError(t, err)
	require.Equal(t, int64(0), cleanedJobsNumber)

	// Creating a user.
	user, err := db.Users().Create(ctx, database.NewUser{Username: "horse"})
	require.NoError(t, err)

	// Create repos.
	repo1 := types.Repo{Name: "test-repo-1", ID: 101}
	err = db.Repos().Create(ctx, &repo1)
	require.NoError(t, err)

	repo2 := types.Repo{Name: "test-repo-2", ID: 102}
	err = db.Repos().Create(ctx, &repo2)
	require.NoError(t, err)

	repo3 := types.Repo{Name: "test-repo-3", ID: 103}
	err = db.Repos().Create(ctx, &repo3)
	require.NoError(t, err)

	// Adding some jobs for user and repos.
	addSyncJobs(t, ctx, db, "user_id", int(user.ID))
	addSyncJobs(t, ctx, db, "repository_id", int(repo1.ID))
	addSyncJobs(t, ctx, db, "repository_id", int(repo2.ID))
	addSyncJobs(t, ctx, db, "repository_id", int(repo3.ID))

	// We should have 20 jobs now.
	jobs, err := store.List(ctx, database.ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	require.Len(t, jobs, 20)

	// Now let's run cleaner function and preserve a history of last 2 items per
	// user/repo. Queued and processing items aren't considered to be history. We
	// should end up with 1 deleted job per repo/user which gives us a total of 4
	// deleted jobs (all "errored" jobs, effectively, as we are deleting the oldest ones first).
	cleanedJobsNumber, err = cleanJobs(ctx, db)
	require.NoError(t, err)
	require.Equal(t, int64(4), cleanedJobsNumber)
	assertThereAreNoJobsWithState(t, ctx, store, database.PermissionsSyncJobStateErrored)

	// Now let's make the history even shorter.
	historySize = 0
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{PermissionsSyncJobsHistorySize: &historySize}})
	cleanedJobsNumber, err = cleanJobs(ctx, db)
	require.NoError(t, err)
	require.Equal(t, int64(8), cleanedJobsNumber)
	assertThereAreNoJobsWithState(t, ctx, store, database.PermissionsSyncJobStateFailed)
	assertThereAreNoJobsWithState(t, ctx, store, database.PermissionsSyncJobStateCompleted)

	// This way we should only have "queued" and "processing" jobs, let's check the
	// number, we should have 8 now.
	jobs, err = store.List(ctx, database.ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	require.Len(t, jobs, 8)

	// If we try to clear the history again, no jobs should be deleted as only
	// "queued" and "processing" are left.
	cleanedJobsNumber, err = cleanJobs(ctx, db)
	require.NoError(t, err)
	require.Equal(t, int64(0), cleanedJobsNumber)
}

var states = []database.PermissionsSyncJobState{
	database.PermissionsSyncJobStateQueued,
	database.PermissionsSyncJobStateProcessing,
	database.PermissionsSyncJobStateErrored,
	database.PermissionsSyncJobStateFailed,
	database.PermissionsSyncJobStateCompleted,
}

func addSyncJobs(t *testing.T, ctx context.Context, db database.DB, repoOrUser string, id int) {
	t.Helper()
	for _, state := range states {
		insertQuery := "INSERT INTO permission_sync_jobs(reason, state, finished_at, %s) VALUES('', '%s', %s, %d)"
		_, err := db.ExecContext(ctx, fmt.Sprintf(insertQuery, repoOrUser, state, getFinishedAt(state), id))
		require.NoError(t, err)
	}
}

// getFinishedAt returns `finished_at` column for inserting test jobs.
//
// Time is mapped to status, from oldest to newest: errored->failed->completed.
//
// Queued and processing jobs doesn't have a `finished_at` value, hence NULL.
func getFinishedAt(state database.PermissionsSyncJobState) string {
	switch state {
	case database.PermissionsSyncJobStateErrored:
		return "NOW() - INTERVAL '5 HOURS'"
	case database.PermissionsSyncJobStateFailed:
		return "NOW() - INTERVAL '2 HOURS'"
	case database.PermissionsSyncJobStateCompleted:
		return "NOW() - INTERVAL '1 HOUR'"
	default:
		return "NULL"
	}
}

func assertThereAreNoJobsWithState(t *testing.T, ctx context.Context, store database.PermissionSyncJobStore, state database.PermissionsSyncJobState) {
	t.Helper()
	allSyncJobs, err := store.List(ctx, database.ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	for _, job := range allSyncJobs {
		if job.State == state {
			t.Fatalf("permissions sync job with state %q should have been deleted", state)
		}
	}
}

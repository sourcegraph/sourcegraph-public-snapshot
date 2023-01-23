package authz

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

func TestPermsSyncerWorker_Handle(t *testing.T) {
	ctx := context.Background()

	dummySyncer := &dummyPermsSyncer{}

	worker := MakePermsSyncerWorker(ctx, &observation.TestContext, dummySyncer)

	t.Run("user sync request", func(t *testing.T) {
		_ = worker.Handle(ctx, logtest.Scoped(t), &database.PermissionSyncJob{
			ID:               99,
			UserID:           1234,
			InvalidateCaches: true,
			HighPriority:     true,
		})

		wantMeta := &requestMeta{
			ID:       1234,
			Priority: priorityHigh,
			Type:     requestTypeUser,
			Options: authz.FetchPermsOptions{
				InvalidateCaches: true,
			},
		}

		if diff := cmp.Diff(dummySyncer.request.requestMeta, wantMeta); diff != "" {
			t.Fatalf("wrong sync request: %s", diff)
		}
	})

	t.Run("repo sync request", func(t *testing.T) {
		_ = worker.Handle(ctx, logtest.Scoped(t), &database.PermissionSyncJob{
			ID:               777,
			RepositoryID:     4567,
			InvalidateCaches: false,
			HighPriority:     false,
		})

		wantMeta := &requestMeta{
			ID:       4567,
			Priority: priorityLow,
			Type:     requestTypeRepo,
			Options: authz.FetchPermsOptions{
				InvalidateCaches: false,
			},
		}

		if diff := cmp.Diff(dummySyncer.request.requestMeta, wantMeta); diff != "" {
			t.Fatalf("wrong sync request: %s", diff)
		}
	})
}

func TestPermsSyncerWorkerCleaner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	store := database.PermissionSyncJobsWith(logger, db)

	// Dry run of a cleaner which shouldn't break anything.
	cleanedJobsNumber, err := cleanJobs(ctx, db, 0)
	require.NoError(t, err)
	require.Equal(t, float64(0), cleanedJobsNumber)

	// Creating a user.
	user, err := db.Users().Create(ctx, database.NewUser{Username: "horse"})
	require.NoError(t, err)

	// Adding some jobs for user and repos.
	addSyncJobs(t, ctx, db, "user_id", int(user.ID))
	addSyncJobs(t, ctx, db, "repository_id", 1)
	addSyncJobs(t, ctx, db, "repository_id", 2)
	addSyncJobs(t, ctx, db, "repository_id", 3)

	// We should have 20 jobs now.
	jobs, err := store.List(ctx, database.ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	require.Len(t, jobs, 20)

	// Now let's run cleaner function and preserve a history of last 2 items per
	// user/repo. Queued and processing items aren't considered to be history. We
	// should end up with 1 deleted job per repo/user which gives us a total of 4
	// deleted jobs (all "completed" jobs, effectively).
	cleanedJobsNumber, err = cleanJobs(ctx, db, 2)
	require.NoError(t, err)
	require.Equal(t, float64(4), cleanedJobsNumber)
	assertThereAreNoJobsWithState(t, ctx, store, "completed")

	// Now let's make the history even shorter.
	cleanedJobsNumber, err = cleanJobs(ctx, db, 0)
	require.NoError(t, err)
	require.Equal(t, float64(8), cleanedJobsNumber)
	assertThereAreNoJobsWithState(t, ctx, store, "failed")
	assertThereAreNoJobsWithState(t, ctx, store, "errored")

	// This way we should only have "queued" and "processing" jobs, let's check the
	// number, we should have 8 now.
	jobs, err = store.List(ctx, database.ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	require.Len(t, jobs, 8)

	// If we try to clear the history again, no jobs should be deleted as only
	// "queued" and "processing" are left.
	cleanedJobsNumber, err = cleanJobs(ctx, db, 0)
	require.NoError(t, err)
	require.Equal(t, float64(0), cleanedJobsNumber)
}

var states = []string{"queued", "processing", "errored", "failed", "completed"}

func addSyncJobs(t *testing.T, ctx context.Context, db database.DB, repoOrUser string, id int) {
	t.Helper()
	for _, state := range states {
		insertQuery := "INSERT INTO permission_sync_jobs(reason, state, %s) VALUES('', '%s', %d)"
		_, err := db.ExecContext(ctx, fmt.Sprintf(insertQuery, repoOrUser, state, id))
		require.NoError(t, err)
	}
}

func assertThereAreNoJobsWithState(t *testing.T, ctx context.Context, store database.PermissionSyncJobStore, state string) {
	t.Helper()
	allSyncJobs, err := store.List(ctx, database.ListPermissionSyncJobOpts{})
	require.NoError(t, err)
	for _, job := range allSyncJobs {
		if job.State == state {
			t.Fatalf("permissions sync job with state %q should have been deleted", state)
		}
	}
}

type dummyPermsSyncer struct {
	request *syncRequest
}

func (d *dummyPermsSyncer) syncPerms(_ context.Context, _ map[requestType]group.ContextGroup, request *syncRequest) {
	d.request = request
}

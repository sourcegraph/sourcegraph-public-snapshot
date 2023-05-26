package repo

import (
	"context"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	api "github.com/sourcegraph/sourcegraph/internal/api"
	database "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func setJobState(t *testing.T, ctx context.Context, store RepoEmbeddingJobsStore, jobID int, state string) {
	t.Helper()
	err := store.Exec(ctx, sqlf.Sprintf("UPDATE repo_embedding_jobs SET state = %s WHERE id = %s", state, jobID))
	if err != nil {
		t.Fatalf("failed to set repo embedding job state: %s", err)
	}
}

func TestRepoEmbeddingJobsStore(t *testing.T) {
	t.Parallel()

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	ctx := context.Background()

	createdRepo := &types.Repo{Name: "github.com/sourcegraph/sourcegraph", URI: "github.com/sourcegraph/sourcegraph", ExternalRepo: api.ExternalRepoSpec{}}
	err := repoStore.Create(ctx, createdRepo)
	require.NoError(t, err)

	store := NewRepoEmbeddingJobsStore(db)

	// no job exists
	exists, err := repoStore.RepoEmbeddingExists(ctx, createdRepo.ID)
	require.NoError(t, err)
	require.Equal(t, exists, false)

	// Create two repo embedding jobs.
	id1, err := store.CreateRepoEmbeddingJob(ctx, createdRepo.ID, "deadbeef")
	require.NoError(t, err)

	id2, err := store.CreateRepoEmbeddingJob(ctx, createdRepo.ID, "coffee")
	require.NoError(t, err)

	count, err := store.CountRepoEmbeddingJobs(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, count)

	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, &database.PaginationArgs{First: &first, OrderBy: database.OrderBy{{Field: "id"}}, Ascending: true})
	require.NoError(t, err)

	// only queued job exists
	exists, err = repoStore.RepoEmbeddingExists(ctx, createdRepo.ID)
	require.NoError(t, err)
	require.Equal(t, exists, false)

	// Expect to get the two repo embedding jobs in the list.
	require.Equal(t, 2, len(jobs))
	require.Equal(t, id1, jobs[0].ID)
	require.Equal(t, id2, jobs[1].ID)

	// Check that we get the correct repo embedding job for repo and revision.
	lastEmbeddingJobForRevision, err := store.GetLastRepoEmbeddingJobForRevision(ctx, createdRepo.ID, "deadbeef")
	require.NoError(t, err)

	require.Equal(t, id1, lastEmbeddingJobForRevision.ID)

	// Complete the second job and check if we get it back when calling GetLastCompletedRepoEmbeddingJob.
	setJobState(t, ctx, store, id2, "completed")
	lastCompletedJob, err := store.GetLastCompletedRepoEmbeddingJob(ctx, createdRepo.ID)
	require.NoError(t, err)

	require.Equal(t, id2, lastCompletedJob.ID)

	// completed job present
	exists, err = repoStore.RepoEmbeddingExists(ctx, createdRepo.ID)
	require.NoError(t, err)
	require.Equal(t, exists, true)
}

func TestCancelRepoEmbeddingJob(t *testing.T) {
	t.Parallel()

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	ctx := context.Background()

	createdRepo := &types.Repo{Name: "github.com/sourcegraph/sourcegraph", URI: "github.com/sourcegraph/sourcegraph", ExternalRepo: api.ExternalRepoSpec{}}
	err := repoStore.Create(ctx, createdRepo)
	require.NoError(t, err)

	store := NewRepoEmbeddingJobsStore(db)

	// Create two repo embedding jobs.
	id1, err := store.CreateRepoEmbeddingJob(ctx, createdRepo.ID, "deadbeef")
	require.NoError(t, err)

	id2, err := store.CreateRepoEmbeddingJob(ctx, createdRepo.ID, "coffee")
	require.NoError(t, err)

	// Cancel the first one.
	err = store.CancelRepoEmbeddingJob(ctx, id1)
	require.NoError(t, err)

	// Move the second job to 'processing' state and cancel it too
	setJobState(t, ctx, store, id2, "processing")
	err = store.CancelRepoEmbeddingJob(ctx, id2)
	require.NoError(t, err)

	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, &database.PaginationArgs{First: &first, OrderBy: database.OrderBy{{Field: "id"}}, Ascending: true})
	require.NoError(t, err)

	// Expect to get the two repo embedding jobs in the list.
	require.Equal(t, 2, len(jobs))
	require.Equal(t, id1, jobs[0].ID)
	require.Equal(t, true, jobs[0].Cancel)
	require.Equal(t, "canceled", jobs[0].State)
	require.Equal(t, id2, jobs[1].ID)
	require.Equal(t, true, jobs[1].Cancel)

	// Attempting to cancel a non-existent job should fail
	err = store.CancelRepoEmbeddingJob(ctx, id1+42)
	require.Error(t, err)

	// Attempting to cancel a completed job should fail
	id3, err := store.CreateRepoEmbeddingJob(ctx, createdRepo.ID, "avocado")
	require.NoError(t, err)

	setJobState(t, ctx, store, id3, "completed")
	err = store.CancelRepoEmbeddingJob(ctx, id3)
	require.Error(t, err)
}

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

func completeJob(t *testing.T, ctx context.Context, store RepoEmbeddingJobsStore, jobID int) {
	t.Helper()
	err := store.Exec(ctx, sqlf.Sprintf("UPDATE repo_embedding_jobs SET state = %s WHERE id = %s", "completed", jobID))
	if err != nil {
		t.Fatalf("failed to set repo embedding job state: %s", err)
	}
}

func TestRepoEmbeddingJobsStore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Background()

	createdRepo := &types.Repo{Name: "github.com/soucegraph/sourcegraph", URI: "github.com/soucegraph/sourcegraph", ExternalRepo: api.ExternalRepoSpec{}}
	err := db.Repos().Create(ctx, createdRepo)
	require.NoError(t, err)

	store := NewRepoEmbeddingJobsStore(db)
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

	// Expect to get the two repo embedding jobs in the list.
	require.Equal(t, 2, len(jobs))
	require.Equal(t, id1, jobs[0].ID)
	require.Equal(t, id2, jobs[1].ID)

	// Check that we get the correct repo embedding job for repo and revision.
	lastEmbeddingJobForRevision, err := store.GetLastRepoEmbeddingJobForRevision(ctx, createdRepo.ID, "deadbeef")
	require.NoError(t, err)

	require.Equal(t, id1, lastEmbeddingJobForRevision.ID)

	// Complete the second job and check if we get it back when calling GetLastCompletedRepoEmbeddingJob.
	completeJob(t, ctx, store, id2)
	lastCompletedJob, err := store.GetLastCompletedRepoEmbeddingJob(ctx, createdRepo.ID)
	require.NoError(t, err)

	require.Equal(t, id2, lastCompletedJob.ID)
}

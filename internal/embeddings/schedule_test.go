package embeddings

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestScheduleRepositoriesForEmbedding(t *testing.T) {
	t.Parallel()

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	createdRepo := &types.Repo{Name: "github.com/sourcegraph/sourcegraph", URI: "github.com/sourcegraph/sourcegraph", ExternalRepo: api.ExternalRepoSpec{}}
	err := repoStore.Create(ctx, createdRepo)
	require.NoError(t, err)

	// Create a repo embedding job.
	store := repo.NewRepoEmbeddingJobsStore(db)
	_, err = store.CreateRepoEmbeddingJob(ctx, createdRepo.ID, "coffee")
	require.NoError(t, err)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.GetDefaultBranchFunc.SetDefaultReturn("main", "coffee", nil)

	// By default, we shouldn't schedule a new job for the same revision
	repoNames := []api.RepoName{"github.com/sourcegraph/sourcegraph"}
	err = ScheduleRepositoriesForEmbedding(ctx, repoNames, false, db, store, gitserverClient)
	require.NoError(t, err)
	count, err := store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// With the 'force' argument, a new job will be scheduled anyways
	err = ScheduleRepositoriesForEmbedding(ctx, repoNames, true, db, store, gitserverClient)
	require.NoError(t, err)
	count, err = store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

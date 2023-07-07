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
	_, err = store.CreateRepoEmbeddingJob(ctx, createdRepo.ID, "coffee", "queued")
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

func TestScheduleRepositoriesForEmbeddingFailedState(t *testing.T) {
	t.Parallel()

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	createdRepo0 := &types.Repo{Name: "github.com/sourcegraph/sourcegraph", URI: "github.com/sourcegraph/sourcegraph", ExternalRepo: api.ExternalRepoSpec{}}
	err := repoStore.Create(ctx, createdRepo0)
	require.NoError(t, err)

	createdRepo1 := &types.Repo{Name: "github.com/sourcegraph/zoekt", URI: "github.com/sourcegraph/zoekt", ExternalRepo: api.ExternalRepoSpec{}}
	err = repoStore.Create(ctx, createdRepo1)
	require.NoError(t, err)

	// Create a repo embedding job.
	store := repo.NewRepoEmbeddingJobsStore(db)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.GetDefaultBranchFunc.PushReturn("", "sgrevision", nil)
	gitserverClient.GetDefaultBranchFunc.PushReturn("main", "zoektrevision", nil)

	repoNames := []api.RepoName{"github.com/sourcegraph/sourcegraph", "github.com/sourcegraph/zoekt"}
	err = ScheduleRepositoriesForEmbedding(ctx, repoNames, false, db, store, gitserverClient)
	count, err := store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equal(t, 2, count)

	pattern := "github.com/sourcegraph/sourcegraph"
	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, repo.ListOpts{PaginationArgs: &database.PaginationArgs{First: &first, OrderBy: database.OrderBy{{Field: "id"}}, Ascending: true}, Query: &pattern})
	require.Equal(t, "failed", jobs[0].State)

	pattern = "github.com/sourcegraph/zoekt"
	jobs, err = store.ListRepoEmbeddingJobs(ctx, repo.ListOpts{PaginationArgs: &database.PaginationArgs{First: &first, OrderBy: database.OrderBy{{Field: "id"}}, Ascending: true}, Query: &pattern})
	require.Equal(t, "queued", jobs[0].State)
}

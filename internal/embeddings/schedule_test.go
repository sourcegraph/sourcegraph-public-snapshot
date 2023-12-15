package embeddings

import (
	"context"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestScheduleRepositories(t *testing.T) {
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
	err = ScheduleRepositories(ctx, repoNames, false, db, store, gitserverClient)
	require.NoError(t, err)
	count, err := store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// With the 'force' argument, a new job will be scheduled anyways
	err = ScheduleRepositories(ctx, repoNames, true, db, store, gitserverClient)
	require.NoError(t, err)
	count, err = store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestScheduleMultipleReposForPolicy(t *testing.T) {
	t.Parallel()

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	repoStore := db.Repos()

	createdRepo := &types.Repo{Name: "github.com/sourcegraph/sourcegraph", URI: "github.com/sourcegraph/sourcegraph", ExternalRepo: api.ExternalRepoSpec{}}
	createdRepo2 := &types.Repo{Name: "github.com/sourcegraph/zoekt", URI: "github.com/sourcegraph/zoekt", ExternalRepo: api.ExternalRepoSpec{}}
	err := repoStore.Create(ctx, createdRepo, createdRepo2)
	require.NoError(t, err)

	// Create a repo embedding job.
	store := repo.NewRepoEmbeddingJobsStore(db)
	_, err = store.CreateRepoEmbeddingJob(ctx, createdRepo.ID, "coffee")
	require.NoError(t, err)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.GetDefaultBranchFunc.SetDefaultReturn("main", "coffee", nil)

	// We should only schedule a new job for the 'zoekt' repo
	repoIDs := []api.RepoID{createdRepo.ID, createdRepo2.ID}
	err = ScheduleRepositoriesForPolicy(ctx, repoIDs, db, store, gitserverClient)
	require.NoError(t, err)

	count, err := store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestScheduleRepositoriesRepoNotFound(t *testing.T) {
	t.Parallel()

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	createdRepo0 := &types.Repo{Name: "github.com/sourcegraph/sourcegraph", URI: "github.com/sourcegraph/sourcegraph", ExternalRepo: api.ExternalRepoSpec{}}
	err := repoStore.Create(ctx, createdRepo0)
	require.NoError(t, err)

	// Create a repo embedding job.
	store := repo.NewRepoEmbeddingJobsStore(db)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.GetDefaultBranchFunc.PushReturn("main", "sgrevision", nil)

	repoNames := []api.RepoName{"github.com/sourcegraph/sourcegraph", "github.com/repo/notfound"}
	err = ScheduleRepositories(ctx, repoNames, false, db, store, gitserverClient)
	require.Error(t, err, database.RepoNotFoundErr{})

	pattern := "github.com/sourcegraph/sourcegraph"
	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, repo.ListOpts{PaginationArgs: &database.PaginationArgs{First: &first, OrderBy: database.OrderBy{{Field: "id"}}, Ascending: true}, Query: &pattern})
	require.NoError(t, err)
	require.Nil(t, jobs)
}

func TestScheduleRepositoriesForPolicyRepoNotFound(t *testing.T) {
	t.Parallel()

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(t))
	repoStore := db.Repos()

	createdRepo0 := &types.Repo{Name: "github.com/sourcegraph/sourcegraph", URI: "github.com/sourcegraph/sourcegraph", ExternalRepo: api.ExternalRepoSpec{}}
	err := repoStore.Create(ctx, createdRepo0)
	require.NoError(t, err)

	// Create a repo embedding job.
	store := repo.NewRepoEmbeddingJobsStore(db)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.GetDefaultBranchFunc.PushReturn("main", "sgrevision", nil)

	repoIDs := []api.RepoID{api.RepoID(234), createdRepo0.ID} // Include one non-existent ID
	err = ScheduleRepositoriesForPolicy(ctx, repoIDs, db, store, gitserverClient)
	require.NoError(t, err)
	count, err := store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equal(t, 1, count)

	pattern := "github.com/sourcegraph/sourcegraph"
	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, repo.ListOpts{PaginationArgs: &database.PaginationArgs{First: &first, OrderBy: database.OrderBy{{Field: "id"}}, Ascending: true}, Query: &pattern})
	require.NoError(t, err)
	require.Equal(t, "queued", jobs[0].State)
}

func TestScheduleRepositoriesInvalidDefaultBranch(t *testing.T) {
	t.Parallel()

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	createdRepo0 := &types.Repo{Name: "github.com/sourcegraph/sourcegraph", URI: "github.com/sourcegraph/sourcegraph", ExternalRepo: api.ExternalRepoSpec{}}
	err := repoStore.Create(ctx, createdRepo0)
	require.NoError(t, err)

	// Create a repo embedding job.
	store := repo.NewRepoEmbeddingJobsStore(db)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.GetDefaultBranchFunc.PushReturn("", "sgrevision", nil)

	repoIDs := []api.RepoID{createdRepo0.ID}
	err = ScheduleRepositoriesForPolicy(ctx, repoIDs, db, store, gitserverClient)
	require.NoError(t, err)
	count, err := store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equal(t, 1, count)

	pattern := "github.com/sourcegraph/sourcegraph"
	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, repo.ListOpts{PaginationArgs: &database.PaginationArgs{First: &first, OrderBy: database.OrderBy{{Field: "id"}}, Ascending: true}, Query: &pattern})
	require.NoError(t, err)
	require.Equal(t, "queued", jobs[0].State)
}

func TestScheduleRepositoriesForPolicyFailed(t *testing.T) {
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
	gitserverClient.GetDefaultBranchFunc.PushReturn("branch", "sgrevision", nil)
	gitserverClient.GetDefaultBranchFunc.PushReturn("main", "zoektrevision", nil)

	repoIDs := []api.RepoID{createdRepo0.ID, createdRepo1.ID}
	err = ScheduleRepositoriesForPolicy(ctx, repoIDs, db, store, gitserverClient)
	require.NoError(t, err)
	count, err := store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	require.Equal(t, 2, count)

	pattern := "github.com/sourcegraph/sourcegraph"
	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, repo.ListOpts{PaginationArgs: &database.PaginationArgs{First: &first, OrderBy: database.OrderBy{{Field: "id"}}, Ascending: true}, Query: &pattern})
	require.NoError(t, err)
	require.Equal(t, "queued", jobs[0].State)

	sgJobID := jobs[0].ID

	pattern = "github.com/sourcegraph/zoekt"
	jobs, err = store.ListRepoEmbeddingJobs(ctx, repo.ListOpts{PaginationArgs: &database.PaginationArgs{First: &first, OrderBy: database.OrderBy{{Field: "id"}}, Ascending: true}, Query: &pattern})
	require.NoError(t, err)
	require.Equal(t, "queued", jobs[0].State)

	zoektJobID := jobs[0].ID

	// Set jobs to "failed" state
	setJobState(t, ctx, store, sgJobID, "failed")
	setJobState(t, ctx, store, zoektJobID, "failed")

	// Reschedule
	gitserverClient.GetDefaultBranchFunc.PushReturn("main", "sgrevision", nil)
	gitserverClient.GetDefaultBranchFunc.PushReturn("main", "zoektrevision", nil)

	err = ScheduleRepositoriesForPolicy(ctx, repoIDs, db, store, gitserverClient)
	require.NoError(t, err)
	count, err = store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	// No jobs should be rescheduled, as there is already an attempted job for these revisions
	require.Equal(t, 2, count)

	// Update one repo's revision and reschedule repo
	gitserverClient.GetDefaultBranchFunc.PushReturn("main", "sgrevision-updated", nil)
	gitserverClient.GetDefaultBranchFunc.PushReturn("main", "zoektrevision", nil)

	err = ScheduleRepositoriesForPolicy(ctx, repoIDs, db, store, gitserverClient)
	require.NoError(t, err)
	count, err = store.CountRepoEmbeddingJobs(ctx, repo.ListOpts{})
	require.NoError(t, err)
	// Repo with previous failure is rescheduled since the revision changed
	require.Equal(t, 3, count)
}

func setJobState(t *testing.T, ctx context.Context, store repo.RepoEmbeddingJobsStore, jobID int, state string) {
	t.Helper()
	err := store.Exec(ctx, sqlf.Sprintf("UPDATE repo_embedding_jobs SET state = %s, finished_at = now() WHERE id = %s", state, jobID))
	if err != nil {
		t.Fatalf("failed to set repo embedding job state: %s", err)
	}
}

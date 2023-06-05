package repo

import (
	"context"
	"testing"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRepoEmbeddingJobsStore(t *testing.T) {
	t.Parallel()

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()

	ctx := context.Background()

	createdRepo := &types.Repo{Name: "github.com/sourcegraph/sourcegraph", URI: "github.com/sourcegraph/sourcegraph", ExternalRepo: api.ExternalRepoSpec{}}
	err := repoStore.Create(ctx, createdRepo)
	require.NoError(t, err)

	createdRepo2 := &types.Repo{Name: "github.com/sourcegraph/zoekt", URI: "github.com/sourcegraph/zoekt", ExternalRepo: api.ExternalRepoSpec{}}
	err = repoStore.Create(ctx, createdRepo2)
	require.NoError(t, err)

	store := NewRepoEmbeddingJobsStore(db)

	// no job exists
	exists, err := repoStore.RepoEmbeddingExists(ctx, createdRepo.ID)
	require.NoError(t, err)
	require.Equal(t, exists, false)

	// Create three repo embedding jobs.
	id1, err := store.CreateRepoEmbeddingJob(ctx, createdRepo.ID, "deadbeef")
	require.NoError(t, err)

	id2, err := store.CreateRepoEmbeddingJob(ctx, createdRepo.ID, "coffee")
	require.NoError(t, err)

	id3, err := store.CreateRepoEmbeddingJob(ctx, createdRepo2.ID, "tea")
	require.NoError(t, err)

	count, err := store.CountRepoEmbeddingJobs(ctx, ListOpts{})
	require.NoError(t, err)
	require.Equal(t, 3, count)

	pattern := "oek" // matching zoekt
	count, err = store.CountRepoEmbeddingJobs(ctx, ListOpts{Query: &pattern})
	require.NoError(t, err)
	require.Equal(t, 1, count)

	pattern = "unknown"
	count, err = store.CountRepoEmbeddingJobs(ctx, ListOpts{Query: &pattern})
	require.NoError(t, err)
	require.Equal(t, 0, count)

	first := 10
	jobs, err := store.ListRepoEmbeddingJobs(ctx, ListOpts{PaginationArgs: &database.PaginationArgs{First: &first, OrderBy: database.OrderBy{{Field: "id"}}, Ascending: true}})
	require.NoError(t, err)

	// only queued job exists
	exists, err = repoStore.RepoEmbeddingExists(ctx, createdRepo.ID)
	require.NoError(t, err)
	require.Equal(t, exists, false)

	// Expect to get the three repo embedding jobs in the list.
	require.Equal(t, 3, len(jobs))
	require.Equal(t, id1, jobs[0].ID)
	require.Equal(t, id2, jobs[1].ID)
	require.Equal(t, id3, jobs[2].ID)

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
	jobs, err := store.ListRepoEmbeddingJobs(ctx, ListOpts{PaginationArgs: &database.PaginationArgs{First: &first, OrderBy: database.OrderBy{{Field: "id"}}, Ascending: true}})
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

func TestGetEmbeddableRepos(t *testing.T) {
	t.Parallel()

	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	repoStore := db.Repos()
	ctx := context.Background()

	// Create two repositories
	firstRepo := &types.Repo{Name: "github.com/sourcegraph/sourcegraph", URI: "github.com/sourcegraph/sourcegraph", ExternalRepo: api.ExternalRepoSpec{}}
	err := repoStore.Create(ctx, firstRepo)
	require.NoError(t, err)

	secondRepo := &types.Repo{Name: "github.com/sourcegraph/zoekt", URI: "github.com/sourcegraph/zoekt", ExternalRepo: api.ExternalRepoSpec{}}
	err = repoStore.Create(ctx, secondRepo)
	require.NoError(t, err)

	// Clone the repos
	gitserverStore := db.GitserverRepos()
	err = gitserverStore.SetCloneStatus(ctx, firstRepo.Name, types.CloneStatusCloned, "test")
	require.NoError(t, err)

	err = gitserverStore.SetCloneStatus(ctx, secondRepo.Name, types.CloneStatusCloned, "test")
	require.NoError(t, err)

	// Create a embeddings policy that applies to all repos
	store := NewRepoEmbeddingJobsStore(db)
	err = createGlobalPolicy(ctx, store)
	require.NoError(t, err)

	// At first, both repos should be embeddable.
	repos, err := store.GetEmbeddableRepos(ctx, EmbeddableRepoOpts{MinimumInterval: 1 * time.Hour})
	require.NoError(t, err)
	require.Equal(t, 2, len(repos))

	// Create and queue an embedding job for the first repo.
	_, err = store.CreateRepoEmbeddingJob(ctx, firstRepo.ID, "coffee")
	require.NoError(t, err)

	// Only the second repo should be embeddable, since the first was recently queued
	repos, err = store.GetEmbeddableRepos(ctx, EmbeddableRepoOpts{MinimumInterval: 1 * time.Hour})
	require.NoError(t, err)
	require.Equal(t, 1, len(repos))
}

func setJobState(t *testing.T, ctx context.Context, store RepoEmbeddingJobsStore, jobID int, state string) {
	t.Helper()
	err := store.Exec(ctx, sqlf.Sprintf("UPDATE repo_embedding_jobs SET state = %s, finished_at = now() WHERE id = %s", state, jobID))
	if err != nil {
		t.Fatalf("failed to set repo embedding job state: %s", err)
	}
}

const insertGlobalPolicyStr = `
INSERT INTO lsif_configuration_policies (
	repository_id,
	repository_patterns,
	name,
	type,
	pattern,
	retention_enabled,
	retention_duration_hours,
	retain_intermediate_commits,
	indexing_enabled,
	index_commit_max_age_hours,
	index_intermediate_commits,
	embeddings_enabled
) VALUES  (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
      `

func createGlobalPolicy(ctx context.Context, store RepoEmbeddingJobsStore) error {
	q := sqlf.Sprintf(insertGlobalPolicyStr,
		nil,
		nil,
		"global",
		string(shared.GitObjectTypeCommit),
		"HEAD",
		false,
		nil,
		false,
		false,
		nil,
		false,
		true, // Embeddings enabled
	)
	return store.Exec(ctx, q)
}

package shared

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestGetCachedRepoEmbeddingIndex(t *testing.T) {
	mockRepoEmbeddingJobsStore := repo.NewMockRepoEmbeddingJobsStore()
	mockRepoStore := database.NewMockRepoStore()

	mockRepoStore.GetByNameFunc.SetDefaultHook(func(ctx context.Context, name api.RepoName) (*types.Repo, error) { return &types.Repo{ID: 1}, nil })

	finishedAt := time.Now()
	mockRepoEmbeddingJobsStore.GetLastCompletedRepoEmbeddingJobFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (*repo.RepoEmbeddingJob, error) {
		return &repo.RepoEmbeddingJob{FinishedAt: &finishedAt}, nil
	})

	hasDownloadedRepoEmbeddingIndex := false
	getRepoEmbeddingIndex, err := getCachedRepoEmbeddingIndex(mockRepoStore, mockRepoEmbeddingJobsStore, func(ctx context.Context, repoEmbeddingIndexName embeddings.RepoEmbeddingIndexName) (*embeddings.RepoEmbeddingIndex, error) {
		hasDownloadedRepoEmbeddingIndex = true
		return &embeddings.RepoEmbeddingIndex{}, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Initial request should download and cache the index.
	_, err = getRepoEmbeddingIndex(ctx, api.RepoName("a"))
	if err != nil {
		t.Fatal(err)
	}
	if !hasDownloadedRepoEmbeddingIndex {
		t.Fatal("expected to download the index on initial request")
	}

	// Subsequent requests should read from the cache.
	hasDownloadedRepoEmbeddingIndex = false
	_, err = getRepoEmbeddingIndex(ctx, api.RepoName("a"))
	if err != nil {
		t.Fatal(err)
	}
	if hasDownloadedRepoEmbeddingIndex {
		t.Fatal("expected to not download the index on subsequent request")
	}

	// Simulate a newer completed repo embedding job.
	finishedAt = finishedAt.Add(time.Hour)
	_, err = getRepoEmbeddingIndex(ctx, api.RepoName("a"))
	if err != nil {
		t.Fatal(err)
	}
	if !hasDownloadedRepoEmbeddingIndex {
		t.Fatal("expected to download the index after a newer embedding job is completed")
	}
}

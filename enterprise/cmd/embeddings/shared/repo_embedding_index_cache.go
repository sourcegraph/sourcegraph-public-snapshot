package shared

import (
	"context"
	"encoding/json"
	"io"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
)

const REPO_EMBEDDING_INDEX_CACHE_MAX_ENTRIES = 5

type downloadRepoEmbeddingIndexFn func(ctx context.Context, repoEmbeddingIndexName embeddings.RepoEmbeddingIndexName) (*embeddings.RepoEmbeddingIndex, error)

func downloadRepoEmbeddingIndex(ctx context.Context, repoEmbeddingIndexName embeddings.RepoEmbeddingIndexName, uploadStore uploadstore.Store) (*embeddings.RepoEmbeddingIndex, error) {
	repoEmbeddingIndexFile, err := uploadStore.Get(ctx, string(repoEmbeddingIndexName))
	if err != nil {
		return nil, err
	}

	repoEmbeddingIndexFileBytes, err := io.ReadAll(repoEmbeddingIndexFile)
	if err != nil {
		return nil, err
	}

	var embeddingIndex embeddings.RepoEmbeddingIndex
	err = json.Unmarshal(repoEmbeddingIndexFileBytes, &embeddingIndex)
	if err != nil {
		return nil, err
	}
	return &embeddingIndex, nil
}

type repoEmbeddingIndexCacheEntry struct {
	index      *embeddings.RepoEmbeddingIndex
	finishedAt time.Time
}

func getCachedRepoEmbeddingIndexFn(
	repoStore database.RepoStore,
	repoEmbeddingJobsStore repo.RepoEmbeddingJobsStore,
	downloadRepoEmbeddingIndex downloadRepoEmbeddingIndexFn,
) (getRepoEmbeddingIndexFn, error) {
	cache, err := lru.New(REPO_EMBEDDING_INDEX_CACHE_MAX_ENTRIES)
	if err != nil {
		return nil, errors.Wrap(err, "creating repo embedding index cache")
	}

	getAndCacheIndex := func(ctx context.Context, repoEmbeddingIndexName embeddings.RepoEmbeddingIndexName, finishedAt *time.Time) (*embeddings.RepoEmbeddingIndex, error) {
		embeddingIndex, err := downloadRepoEmbeddingIndex(ctx, repoEmbeddingIndexName)
		if err != nil {
			return nil, err
		}
		cache.Add(repoEmbeddingIndexName, repoEmbeddingIndexCacheEntry{index: embeddingIndex, finishedAt: *finishedAt})
		return embeddingIndex, nil
	}

	return func(ctx context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
		repo, err := repoStore.GetByName(ctx, repoName)
		if err != nil {
			return nil, err
		}

		lastFinishedRepoEmbeddingJob, err := repoEmbeddingJobsStore.GetLastCompletedRepoEmbeddingJob(ctx, repo.ID)
		if err != nil {
			return nil, err
		}

		repoEmbeddingIndexName := embeddings.GetRepoEmbeddingIndexName(repoName)

		cacheEntry, ok := cache.Get(repoEmbeddingIndexName)
		// Check if the index is in the cache.
		if ok {
			// Check if we have a newer finished embedding job. If so, download the new index, and cache it instead.
			repoEmbeddingIndexCacheEntry := cacheEntry.(repoEmbeddingIndexCacheEntry)
			if lastFinishedRepoEmbeddingJob.FinishedAt.After(repoEmbeddingIndexCacheEntry.finishedAt) {
				return getAndCacheIndex(ctx, repoEmbeddingIndexName, lastFinishedRepoEmbeddingJob.FinishedAt)
			}
			// Otherwise, return the cached index.
			return repoEmbeddingIndexCacheEntry.index, nil
		}
		// We do not have the index in the cache. Download and cache it.
		return getAndCacheIndex(ctx, repoEmbeddingIndexName, lastFinishedRepoEmbeddingJob.FinishedAt)
	}, nil
}

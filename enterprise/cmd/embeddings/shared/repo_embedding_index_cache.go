package shared

import (
	"context"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var cacheSize = env.MustGetInt("EMBEDDINGS_REPO_INDEX_CACHE_SIZE", 5, "Number of repository embedding indexes to cache in memory.")

type downloadRepoEmbeddingIndexFn func(ctx context.Context, repoEmbeddingIndexName embeddings.RepoEmbeddingIndexName) (*embeddings.RepoEmbeddingIndex, error)

type repoEmbeddingIndexCacheEntry struct {
	index      *embeddings.RepoEmbeddingIndex
	finishedAt time.Time
}

// repoMutexMap tracks a mutex for each repository, creating one if it does not yet exist for a given repository ID.
// This allows concurrent access to the cache for different repositories, while serializing access for a given repository.
type repoMutexMap struct {
	// Mutex prevents concurrent access to the repoMutexes map below.
	sync.Mutex
	repoMutexes map[api.RepoID]*sync.Mutex
}

func (r *repoMutexMap) LockRepo(repoID api.RepoID) {
	r.Lock()
	// IMPORTANT: do not use `defer r.Unlock()` because it creates a deadlock.
	// With defer and 2 concurrent requests, request 1 locks the map, and locks the mutex for repo A.
	// Request 2 comes in, locks the map, and tries to lock the mutex for repo A, but it has to wait (while still locking the map mutex).
	// Request 1 finishes processing, tries to unlock the mutex for repo A, but it can't because request 2 locked the access to the map, creating a deadlock.
	repoMutex, ok := r.repoMutexes[repoID]
	if !ok {
		repoMutex = &sync.Mutex{}
		r.repoMutexes[repoID] = repoMutex
	}
	r.Unlock()

	repoMutex.Lock()
}

func (r *repoMutexMap) UnlockRepo(repoID api.RepoID) {
	r.Lock()
	repoMutex := r.repoMutexes[repoID]
	r.Unlock()

	repoMutex.Unlock()
}

func getCachedRepoEmbeddingIndex(
	repoStore database.RepoStore,
	repoEmbeddingJobsStore repo.RepoEmbeddingJobsStore,
	downloadRepoEmbeddingIndex downloadRepoEmbeddingIndexFn,
) (getRepoEmbeddingIndexFn, error) {
	cache, err := lru.New[embeddings.RepoEmbeddingIndexName, repoEmbeddingIndexCacheEntry](cacheSize)
	if err != nil {
		return nil, errors.Wrap(err, "creating repo embedding index cache")
	}

	repoMutexMap := &repoMutexMap{repoMutexes: make(map[api.RepoID]*sync.Mutex)}

	getAndCacheIndex := func(ctx context.Context, repoEmbeddingIndexName embeddings.RepoEmbeddingIndexName, finishedAt *time.Time) (*embeddings.RepoEmbeddingIndex, error) {
		embeddingIndex, err := downloadRepoEmbeddingIndex(ctx, repoEmbeddingIndexName)
		if err != nil {
			return nil, errors.Wrap(err, "downloading repo embedding index")
		}
		cache.Add(repoEmbeddingIndexName, repoEmbeddingIndexCacheEntry{index: embeddingIndex, finishedAt: *finishedAt})
		return embeddingIndex, nil
	}

	return func(ctx context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
		repo, err := repoStore.GetByName(ctx, repoName)
		if err != nil {
			return nil, err
		}

		// Acquire a mutex for the given repository to serialize access.
		// This avoids multiple routines concurrently downloading the same embedding index
		// when the cache is empty, which can lead to an out-of-memory error.
		repoMutexMap.LockRepo(repo.ID)
		defer repoMutexMap.UnlockRepo(repo.ID)

		lastFinishedRepoEmbeddingJob, err := repoEmbeddingJobsStore.GetLastCompletedRepoEmbeddingJob(ctx, repo.ID)
		if err != nil {
			return nil, err
		}

		repoEmbeddingIndexName := embeddings.GetRepoEmbeddingIndexName(repoName)

		cacheEntry, ok := cache.Get(repoEmbeddingIndexName)
		// Check if the index is in the cache.
		if ok {
			// Check if we have a newer finished embedding job. If so, download the new index, cache it, and return it instead.
			if lastFinishedRepoEmbeddingJob.FinishedAt.After(cacheEntry.finishedAt) {
				return getAndCacheIndex(ctx, repoEmbeddingIndexName, lastFinishedRepoEmbeddingJob.FinishedAt)
			}
			// Otherwise, return the cached index.
			return cacheEntry.index, nil
		}
		// We do not have the index in the cache. Download and cache it.
		return getAndCacheIndex(ctx, repoEmbeddingIndexName, lastFinishedRepoEmbeddingJob.FinishedAt)
	}, nil
}

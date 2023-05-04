package shared

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type downloadRepoEmbeddingIndexFn func(ctx context.Context, repoEmbeddingIndexName embeddings.RepoEmbeddingIndexName) (*embeddings.RepoEmbeddingIndex, error)

type repoEmbeddingIndexCacheEntry struct {
	index      *embeddings.RepoEmbeddingIndex
	finishedAt time.Time
}

// repoMutexMap tracks a mutex for each repository, creating one if it does not yet exist for a given repository ID.
// This allows concurrent access to the cache for different repositories, while serializing access for a given repository.
type repoMutexMap struct {
	init        sync.Once
	mu          sync.RWMutex
	repoMutexes map[api.RepoID]*sync.Mutex
}

func (m *repoMutexMap) GetLock(repoID api.RepoID) *sync.Mutex {
	m.init.Do(func() { m.repoMutexes = make(map[api.RepoID]*sync.Mutex) })

	m.mu.RLock()
	lock, ok := m.repoMutexes[repoID]
	m.mu.RUnlock()

	if ok {
		return lock
	}

	m.mu.Lock()
	lock, ok = m.repoMutexes[repoID]
	if !ok {
		lock = &sync.Mutex{}
		m.repoMutexes[repoID] = lock
	}
	m.mu.Unlock()

	return lock
}

// embeddingsIndexCache is a thin wrapper around an LRU cache that is
// memory-bounded, which is useful for embeddings indexes because they can have
// dramatically different sizes.
//
// Note that this is just an LRU cache with a bounded in-memory size.
// A query that hits a large number of repos will fill the cache with
// those repos, which may not be desirable if we are doing many global
// searches.
type embeddingsIndexCache struct {
	mu                 sync.Mutex
	cache              *lru.Cache[embeddings.RepoEmbeddingIndexName, repoEmbeddingIndexCacheEntry]
	maxSizeBytes       int64
	remainingSizeBytes int64
}

// newEmbeddingsIndexCache creates a cache with reasonable settings for an embeddings cache
func newEmbeddingsIndexCache(maxSizeBytes int64) (_ *embeddingsIndexCache, err error) {
	c := &embeddingsIndexCache{
		maxSizeBytes:       maxSizeBytes,
		remainingSizeBytes: maxSizeBytes,
	}

	// arbitrarily large LRU cache because we want to evict based on size, not count
	c.cache, err = lru.NewWithEvict(999_999_999, c.onEvict)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *embeddingsIndexCache) Get(repo embeddings.RepoEmbeddingIndexName) (repoEmbeddingIndexCacheEntry, bool) {
	return c.cache.Get(repo)
}

func (c *embeddingsIndexCache) Add(repo embeddings.RepoEmbeddingIndexName, value repoEmbeddingIndexCacheEntry) {
	size := value.index.EstimateSize()
	if size > c.maxSizeBytes {
		// Return early if the index could never fit in the cache.
		// We don't want to dump the cache just to not be able to fit it.
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict entries until there is space
	for c.remainingSizeBytes < size {
		_, _, ok := c.cache.RemoveOldest()
		if !ok {
			// Since we already checked that the entry can fit in the cache,
			// this should never happen since the cache should never be empty
			// and not fit the entry.
			return
		}
	}

	// Reserve space for the entry and add it to the cache
	c.remainingSizeBytes -= size
	c.cache.Add(repo, value)
}

// onEvict must only be called while the index mutex is held
func (c *embeddingsIndexCache) onEvict(_ embeddings.RepoEmbeddingIndexName, value repoEmbeddingIndexCacheEntry) {
	c.remainingSizeBytes += value.index.EstimateSize()
}

func getCachedRepoEmbeddingIndex(
	repoStore database.RepoStore,
	repoEmbeddingJobsStore repo.RepoEmbeddingJobsStore,
	downloadRepoEmbeddingIndex downloadRepoEmbeddingIndexFn,
	cacheSizeBytes int64,
) (getRepoEmbeddingIndexFn, error) {
	cache, err := newEmbeddingsIndexCache(cacheSizeBytes)
	if err != nil {
		return nil, errors.Wrap(err, "creating repo embedding index cache")
	}

	repoMutexMap := &repoMutexMap{}

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
		lock := repoMutexMap.GetLock(repo.ID)
		lock.Lock()
		defer lock.Unlock()

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

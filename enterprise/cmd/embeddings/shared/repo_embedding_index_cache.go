package shared

import (
	"context"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"

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

// embeddingIndexCache is a thin wrapper around a ristretto cache
// that adds well-typed Get and Set methods to the cache.
type embeddingIndexCache struct {
	cache *ristretto.Cache
}

// newEmbeddingsIndexCache creates a cache with reasonable settings for an embeddings cache
func newEmbeddingsIndexCache(cacheSizeBytes int64) (embeddingIndexCache, error) {
	rc, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     cacheSizeBytes,
		BufferItems: 64,
		Metrics:     false, // we have no way to collect metrics yet
	})
	if err != nil {
		return embeddingIndexCache{}, err
	}

	return embeddingIndexCache{rc}, nil
}

func (c *embeddingIndexCache) Get(repo embeddings.RepoEmbeddingIndexName) (repoEmbeddingIndexCacheEntry, bool) {
	v, ok := c.cache.Get(string(repo))
	if !ok {
		return repoEmbeddingIndexCacheEntry{}, false
	}
	return v.(repoEmbeddingIndexCacheEntry), true
}

func (c *embeddingIndexCache) Set(repo embeddings.RepoEmbeddingIndexName, value repoEmbeddingIndexCacheEntry) {
	c.cache.Set(string(repo), value, value.index.EstimateSize())

	// Since this isn't a performance hotspot, always wait for the
	// cached value to propagate. This makes testing easier and
	// ensures we never refetch unnecessarily.
	c.cache.Wait()
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
		cache.Set(repoEmbeddingIndexName, repoEmbeddingIndexCacheEntry{index: embeddingIndex, finishedAt: *finishedAt})
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

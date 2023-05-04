package shared

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2"
	"golang.org/x/sync/singleflight"

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

func NewCachedEmbeddingIndexGetter(
	repoStore database.RepoStore,
	repoEmbeddingJobStore repo.RepoEmbeddingJobsStore,
	downloadRepoEmbeddingIndex downloadRepoEmbeddingIndexFn,
	cacheSizeBytes int64,
) (*CachedEmbeddingIndexGetter, error) {
	cache, err := newEmbeddingsIndexCache(cacheSizeBytes)
	if err != nil {
		return nil, err
	}
	return &CachedEmbeddingIndexGetter{
		repoStore:                  repoStore,
		repoEmbeddingJobsStore:     repoEmbeddingJobStore,
		downloadRepoEmbeddingIndex: downloadRepoEmbeddingIndex,
		cache:                      cache,
	}, nil
}

type CachedEmbeddingIndexGetter struct {
	repoStore                  database.RepoStore
	repoEmbeddingJobsStore     repo.RepoEmbeddingJobsStore
	downloadRepoEmbeddingIndex downloadRepoEmbeddingIndexFn

	cache *embeddingsIndexCache
	sf    singleflight.Group
}

func (c *CachedEmbeddingIndexGetter) Get(ctx context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
	// Run the fetch request through a singleflight to keep from fetching the
	// same index multiple times concurrently
	v, err, _ := c.sf.Do(string(repoName), func() (interface{}, error) {
		return c.get(ctx, repoName)
	})
	return v.(*embeddings.RepoEmbeddingIndex), err
}

func (c *CachedEmbeddingIndexGetter) get(ctx context.Context, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
	repo, err := c.repoStore.GetByName(ctx, repoName)
	if err != nil {
		return nil, err
	}

	lastFinishedRepoEmbeddingJob, err := c.repoEmbeddingJobsStore.GetLastCompletedRepoEmbeddingJob(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	repoEmbeddingIndexName := embeddings.GetRepoEmbeddingIndexName(repoName)

	cacheEntry, ok := c.cache.Get(repoEmbeddingIndexName)
	if !ok {
		// We do not have the index in the cache. Download and cache it.
		return c.getAndCacheIndex(ctx, repoEmbeddingIndexName, lastFinishedRepoEmbeddingJob.FinishedAt)
	} else if lastFinishedRepoEmbeddingJob.FinishedAt.After(cacheEntry.finishedAt) {
		// Check if we have a newer finished embedding job. If so, download the new index, cache it, and return it instead.
		return c.getAndCacheIndex(ctx, repoEmbeddingIndexName, lastFinishedRepoEmbeddingJob.FinishedAt)
	}

	// Otherwise, return the cached index.
	return cacheEntry.index, nil
}

func (c *CachedEmbeddingIndexGetter) getAndCacheIndex(ctx context.Context, repoEmbeddingIndexName embeddings.RepoEmbeddingIndexName, finishedAt *time.Time) (*embeddings.RepoEmbeddingIndex, error) {
	embeddingIndex, err := c.downloadRepoEmbeddingIndex(ctx, repoEmbeddingIndexName)
	if err != nil {
		return nil, errors.Wrap(err, "downloading repo embedding index")
	}
	c.cache.Add(repoEmbeddingIndexName, repoEmbeddingIndexCacheEntry{index: embeddingIndex, finishedAt: *finishedAt})
	return embeddingIndex, nil
}

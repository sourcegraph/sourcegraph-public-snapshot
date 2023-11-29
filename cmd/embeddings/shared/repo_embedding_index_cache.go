package shared

import (
	"context"
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/singleflight"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type downloadRepoEmbeddingIndexFn func(ctx context.Context, repoID api.RepoID, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error)

type repoEmbeddingIndexCacheEntry struct {
	index      *embeddings.RepoEmbeddingIndex
	finishedAt time.Time
}

var (
	embeddingsCacheHitCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "embeddings_cache_hit_count",
	})
	embeddingsCacheMissCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "embeddings_cache_miss_count",
	})
	embeddingsCacheMissBytes = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "embeddings_cache_miss_bytes",
	})
	embeddingsCacheEvictedCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "embeddings_cache_evicted_count",
	})
)

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
	maxSizeBytes       uint64
	remainingSizeBytes uint64
}

// newEmbeddingsIndexCache creates a cache with reasonable settings for an embeddings cache
func newEmbeddingsIndexCache(maxSizeBytes uint64) (_ *embeddingsIndexCache, err error) {
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
	v, ok := c.cache.Get(repo)
	if ok {
		embeddingsCacheHitCount.Inc()
	} else {
		embeddingsCacheMissCount.Inc()
	}
	return v, ok
}

func (c *embeddingsIndexCache) Add(repo embeddings.RepoEmbeddingIndexName, value repoEmbeddingIndexCacheEntry) {
	size := value.index.EstimateSize()
	embeddingsCacheMissBytes.Add(float64(size))
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
	embeddingsCacheEvictedCount.Inc()
}

func NewCachedEmbeddingIndexGetter(
	repoStore database.RepoStore,
	repoEmbeddingJobStore repo.RepoEmbeddingJobsStore,
	downloadRepoEmbeddingIndex downloadRepoEmbeddingIndexFn,
	cacheSizeBytes uint64,
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

func (c *CachedEmbeddingIndexGetter) Get(ctx context.Context, repoID api.RepoID, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
	var (
		done = make(chan struct{})
		v    interface{}
		err  error
	)
	// Run the fetch in the background, but outside the singleflight so context
	// errors are not shared.
	go func() {
		detachedCtx := context.WithoutCancel(ctx)
		// Run the fetch request through a singleflight to keep from fetching the
		// same index multiple times concurrently
		v, err, _ = c.sf.Do(fmt.Sprintf("%d", repoID), func() (interface{}, error) {
			return c.get(detachedCtx, repoID, repoName)
		})
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		return v.(*embeddings.RepoEmbeddingIndex), err
	}
}

func (c *CachedEmbeddingIndexGetter) get(ctx context.Context, repoID api.RepoID, repoName api.RepoName) (*embeddings.RepoEmbeddingIndex, error) {
	lastFinishedRepoEmbeddingJob, err := c.repoEmbeddingJobsStore.GetLastCompletedRepoEmbeddingJob(ctx, repoID)
	if err != nil {
		return nil, err
	}

	repoEmbeddingIndexName := embeddings.GetRepoEmbeddingIndexName(repoID)

	cacheEntry, ok := c.cache.Get(repoEmbeddingIndexName)
	trace.FromContext(ctx).AddEvent("checked embedding index cache", attribute.Bool("hit", ok))
	if !ok {
		// We do not have the index in the cache. Download and cache it.
		return c.getAndCacheIndex(ctx, repoID, repoName, lastFinishedRepoEmbeddingJob.FinishedAt)
	} else if lastFinishedRepoEmbeddingJob.FinishedAt.After(cacheEntry.finishedAt) {
		// Check if we have a newer finished embedding job. If so, download the new index, cache it, and return it instead.
		return c.getAndCacheIndex(ctx, repoID, repoName, lastFinishedRepoEmbeddingJob.FinishedAt)
	}

	// Otherwise, return the cached index.
	return cacheEntry.index, nil
}

func (c *CachedEmbeddingIndexGetter) getAndCacheIndex(ctx context.Context, repoID api.RepoID, repoName api.RepoName, finishedAt *time.Time) (*embeddings.RepoEmbeddingIndex, error) {
	embeddingIndex, err := c.downloadRepoEmbeddingIndex(ctx, repoID, repoName)
	if err != nil {
		return nil, errors.Wrap(err, "downloading repo embedding index")
	}
	c.cache.Add(embeddings.GetRepoEmbeddingIndexName(repoID), repoEmbeddingIndexCacheEntry{index: embeddingIndex, finishedAt: *finishedAt})
	return embeddingIndex, nil
}

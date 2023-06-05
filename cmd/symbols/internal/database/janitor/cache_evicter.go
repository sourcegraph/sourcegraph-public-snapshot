package janitor

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type cacheEvicter struct {
	// cache is the disk backed cache.
	cache diskcache.Store

	// maxCacheSizeBytes is the maximum size of the cache in bytes. Note that we can
	// be larger than maxCacheSizeBytes temporarily between runs of this handler.
	// When we go over maxCacheSizeBytes we trigger delete files until we get below
	// maxCacheSizeBytes.
	maxCacheSizeBytes int64

	metrics *Metrics
}

var (
	_ goroutine.Handler      = &cacheEvicter{}
	_ goroutine.ErrorHandler = &cacheEvicter{}
)

func NewCacheEvicter(interval time.Duration, cache diskcache.Store, maxCacheSizeBytes int64, metrics *Metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		&cacheEvicter{
			cache:             cache,
			maxCacheSizeBytes: maxCacheSizeBytes,
			metrics:           metrics,
		},
		goroutine.WithName("codeintel.symbols-cache-evictor"),
		goroutine.WithDescription("evicts entries from the symbols cache"),
		goroutine.WithInterval(interval),
	)
}

// Handle periodically checks the size of the cache and evicts/deletes items.
func (e *cacheEvicter) Handle(ctx context.Context) error {
	if e.maxCacheSizeBytes == 0 {
		return nil
	}

	stats, err := e.cache.Evict(e.maxCacheSizeBytes)
	if err != nil {
		return errors.Wrap(err, "cache.Evict")
	}

	e.metrics.cacheSizeBytes.Set(float64(stats.CacheSize))
	e.metrics.evictions.Add(float64(stats.Evicted))
	return nil
}

func (e *cacheEvicter) HandleError(err error) {
	e.metrics.errors.Inc()
	log15.Error("Failed to evict items from cache", "error", err)
}

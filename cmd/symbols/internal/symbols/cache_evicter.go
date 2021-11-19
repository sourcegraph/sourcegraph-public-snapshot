package symbols

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type cacheEvicter struct {
	// cache is the disk backed cache.
	cache *diskcache.Store

	// maxCacheSizeBytes is the maximum size of the cache in bytes. Note that we can
	// be larger than maxCacheSizeBytes temporarily between runs of this handler.
	// When we go over maxCacheSizeBytes we trigger delete files until we get below
	// maxCacheSizeBytes.
	maxCacheSizeBytes int64
}

var _ goroutine.Handler = &cacheEvicter{}
var _ goroutine.ErrorHandler = &cacheEvicter{}

func NewCacheEvicter(interval time.Duration, cache *diskcache.Store, maxCacheSizeBytes int64) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, &cacheEvicter{
		cache:             cache,
		maxCacheSizeBytes: maxCacheSizeBytes,
	})
}

// Handle periodically checks the size of the cache and evicts/deletes items.
func (e *cacheEvicter) Handle(ctx context.Context) error {
	if e.maxCacheSizeBytes == 0 {
		return nil
	}

	stats, err := e.cache.Evict(e.maxCacheSizeBytes)
	if err != nil {
		return err
	}

	cacheSizeBytes.Set(float64(stats.CacheSize))
	evictions.Add(float64(stats.Evicted))
	return nil
}

func (e *cacheEvicter) HandleError(err error) {
	// 	// TODO - add metric, logs
}

var (
	cacheSizeBytes = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "symbols_store_cache_size_bytes",
		Help: "The total size of items in the on disk cache.",
	})
	evictions = promauto.NewCounter(prometheus.CounterOpts{
		Name: "symbols_store_evictions",
		Help: "The total number of items evicted from the cache.",
	})
)

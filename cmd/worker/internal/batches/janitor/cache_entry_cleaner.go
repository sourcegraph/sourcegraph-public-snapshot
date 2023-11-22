package janitor

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

var maxCacheEntriesSize = env.MustGetInt(
	"SRC_BATCH_CHANGES_MAX_CACHE_SIZE_MB",
	5000,
	"Maximum size of the batch_spec_execution_cache_entries.value column. Value is megabytes.",
)

const cacheCleanInterval = 1 * time.Hour

func NewCacheEntryCleaner(ctx context.Context, s *store.Store) goroutine.BackgroundRoutine {
	maxSizeByte := int64(maxCacheEntriesSize * 1024 * 1024)

	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return s.CleanBatchSpecExecutionCacheEntries(ctx, maxSizeByte)
		}),
		goroutine.WithName("batchchanges.cache-cleaner"),
		goroutine.WithDescription("cleaning up LRU batch spec execution cache entries"),
		goroutine.WithInterval(cacheCleanInterval),
	)
}

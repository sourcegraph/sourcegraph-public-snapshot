package background

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

const maxCacheEntriesTableSize = 500 * 1024 * 1024 // 500mb

func newCacheEntryCleanerJob(ctx context.Context, s *store.Store) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		ctx,
		specExpireInteral,
		goroutine.NewHandlerWithErrorMessage("cleaning up LRU batch spec execution cache entries", func(ctx context.Context) error {
			return s.DeleteExpiredChangesetSpecs(ctx)
		}),
	)
}

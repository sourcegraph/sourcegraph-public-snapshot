package janitor

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

const changesetCleanInterval = 24 * time.Hour

func NewChangesetDetachedCleaner(ctx context.Context, s *store.Store) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		ctx,
		changesetCleanInterval,
		goroutine.NewHandlerWithErrorMessage("cleaning detached changeset entries", func(ctx context.Context) error {
			//return s.CleanBatchSpecExecutionCacheEntries(ctx, maxSizeByte)
			return nil
		}),
	)
}

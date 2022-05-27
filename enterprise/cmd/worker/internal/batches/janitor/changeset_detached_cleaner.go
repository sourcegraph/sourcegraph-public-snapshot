package janitor

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

const changesetCleanInterval = 24 * time.Hour

// NewChangesetDetachedCleaner creates a new goroutine.PeriodicGoroutine that deletes Changesets that have been
// detached for a period of time.
func NewChangesetDetachedCleaner(ctx context.Context, s *store.Store) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		ctx,
		changesetCleanInterval,
		goroutine.NewHandlerWithErrorMessage("cleaning detached changeset entries", func(ctx context.Context) error {
			// delete detached changesets that are 2 weeks old
			return s.CleanDetachedChangesets(ctx, 14)
		}),
	)
}

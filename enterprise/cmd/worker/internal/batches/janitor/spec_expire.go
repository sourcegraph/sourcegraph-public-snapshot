package janitor

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const specExpireInteral = 60 * time.Minute

func NewSpecExpirer(ctx context.Context, bstore *store.Store) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		ctx,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			// Delete all unattached changeset specs...
			if err := bstore.DeleteUnattachedExpiredChangesetSpecs(ctx); err != nil {
				return errors.Wrap(err, "DeleteExpiredChangesetSpecs")
			}
			// ... and all expired changeset specs...
			if err := bstore.DeleteExpiredChangesetSpecs(ctx); err != nil {
				return errors.Wrap(err, "DeleteExpiredChangesetSpecs")
			}
			// ... and then the BatchSpecs, that are expired.
			if err := bstore.DeleteExpiredBatchSpecs(ctx); err != nil {
				return errors.Wrap(err, "DeleteExpiredBatchSpecs")
			}
			return nil
		}),
		goroutine.WithName("batchchanges.spec-expirer"),
		goroutine.WithDescription("expire batch changes specs"),
		goroutine.WithInterval(specExpireInteral),
	)
}

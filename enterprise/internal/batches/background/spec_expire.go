package background

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

const specExpireInteral = 2 * time.Minute

func newSpecExpireJob(ctx context.Context, cstore *store.Store) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		ctx,
		specExpireInteral,
		goroutine.NewHandlerWithErrorMessage("expire batch changes specs", func(ctx context.Context) error {
			// Delete all unattached, expired ChangesetSpecs...
			if err := cstore.DeleteExpiredChangesetSpecs(ctx); err != nil {
				return errors.Wrap(err, "DeleteExpiredChangesetSpecs")
			}
			// ... and then the BatchSpecs, that are expired.
			if err := cstore.DeleteExpiredBatchSpecs(ctx); err != nil {
				return errors.Wrap(err, "DeleteExpiredBatchSpecs")
			}
			return nil
		}),
	)
}

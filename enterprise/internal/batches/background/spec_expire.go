package background

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func newSpecExpireWorker(ctx context.Context, cstore *store.Store) goroutine.BackgroundRoutine {
	expireSpecs := goroutine.NewHandlerWithErrorMessage("expire batch changes specs", func(ctx context.Context) error {
		// We first need to delete expired ChangesetSpecs...
		if err := cstore.DeleteExpiredChangesetSpecs(ctx); err != nil {
			return errors.Wrap(err, "DeleteExpiredChangesetSpecs")
		}
		// ... and then the BatchSpecs, due to the batch_spec_id
		// foreign key on changeset_specs.
		return errors.Wrap(cstore.DeleteExpiredBatchSpecs(ctx), "DeleteExpiredBatchSpecs")
	})
	return goroutine.NewPeriodicGoroutine(ctx, 2*time.Minute, expireSpecs)
}

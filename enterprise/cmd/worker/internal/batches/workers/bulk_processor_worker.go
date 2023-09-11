package workers

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/batches/processor"
	"github.com/sourcegraph/sourcegraph/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// NewBulkOperationWorker creates a dbworker.Worker that fetches enqueued changeset_jobs
// from the database and passes them to the bulk executor for processing.
func NewBulkOperationWorker(
	ctx context.Context,
	observationCtx *observation.Context,
	s *store.Store,
	workerStore dbworkerstore.Store[*btypes.ChangesetJob],
	sourcer sources.Sourcer,
) *workerutil.Worker[*btypes.ChangesetJob] {
	r := &bulkProcessorWorker{sourcer: sourcer, store: s}

	options := workerutil.WorkerOptions{
		Name:              "batches_bulk_processor",
		Description:       "executes the bulk operations in the background",
		NumHandlers:       5,
		HeartbeatInterval: 15 * time.Second,
		Interval:          5 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "batch_changes_bulk_processor"),
	}

	worker := dbworker.NewWorker[*btypes.ChangesetJob](ctx, workerStore, r.HandlerFunc(), options)
	return worker
}

// bulkProcessorWorker is a wrapper for the workerutil handlerfunc to create a
// bulkProcessor with a source and store.
type bulkProcessorWorker struct {
	store   *store.Store
	sourcer sources.Sourcer
}

func (b *bulkProcessorWorker) HandlerFunc() workerutil.HandlerFunc[*btypes.ChangesetJob] {
	return func(ctx context.Context, logger log.Logger, job *btypes.ChangesetJob) (err error) {
		tx, err := b.store.Transact(ctx)
		if err != nil {
			return err
		}

		p := processor.New(logger, tx, b.sourcer)
		afterDone, err := p.Process(ctx, job)

		defer func() {
			err = tx.Done(err)
			// If afterDone is provided, it is enqueuing a new webhook. We call afterDone
			// regardless of whether or not the transaction succeeds because the webhook
			// should represent the interaction with the code host, not the database
			// transaction. The worst case is that the transaction actually did fail and
			// thus the changeset in the webhook payload is out-of-date. But we will still
			// have enqueued the appropriate webhook.
			if afterDone != nil {
				afterDone(b.store)
			}
		}()

		return err
	}
}

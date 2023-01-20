package workers

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/processor"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
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
		defer func() { err = tx.Done(err) }()

		p := processor.New(logger, tx, b.sourcer)

		return p.Process(ctx, job)
	}
}

package workers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/processor"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

// NewBulkOperationWorker creates a dbworker.Worker that fetches enqueued changeset_jobs
// from the database and passes them to the bulk executor for processing.
func NewBulkOperationWorker(
	ctx context.Context,
	s *store.Store,
	workerStore dbworkerstore.Store,
	sourcer sources.Sourcer,
	observationContext *observation.Context,
) *workerutil.Worker {
	r := &bulkProcessorWorker{sourcer: sourcer, store: s}

	options := workerutil.WorkerOptions{
		Name:              "batches_bulk_processor",
		NumHandlers:       5,
		HeartbeatInterval: 15 * time.Second,
		Interval:          5 * time.Second,
		Metrics:           workerutil.NewMetrics(observationContext, "batch_changes_bulk_processor"),
	}

	worker := dbworker.NewWorker(ctx, workerStore, r.HandlerFunc(), options)
	return worker
}

// bulkProcessorWorker is a wrapper for the workerutil handlerfunc to create a
// bulkProcessor with a source and store.
type bulkProcessorWorker struct {
	store   *store.Store
	sourcer sources.Sourcer
}

func (b *bulkProcessorWorker) HandlerFunc() workerutil.HandlerFunc {
	return func(ctx context.Context, logger log.Logger, record workerutil.Record) (err error) {
		job := record.(*btypes.ChangesetJob)

		tx, err := b.store.Transact(ctx)
		if err != nil {
			return err
		}
		defer func() { err = tx.Done(err) }()

		p := processor.New(tx, b.sourcer)

		return p.Process(ctx, job)
	}
}

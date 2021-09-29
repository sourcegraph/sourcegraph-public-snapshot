package background

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/processor"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// bulkProcessorMaxNumRetries is the maximum number of attempts the bulkProcessor
// makes to process a changeset job when it fails.
const bulkProcessorMaxNumRetries = 10

// bulkProcessorMaxNumResets is the maximum number of attempts the bulkProcessor
// makes to process a changeset job when it stalls (process crashes, etc.).
const bulkProcessorMaxNumResets = 60

// newBulkOperationWorker creates a dbworker.Worker that fetches enqueued changeset_jobs
// from the database and passes them to the bulk executor for processing.
func newBulkOperationWorker(
	ctx context.Context,
	s *store.Store,
	workerStore dbworkerstore.Store,
	sourcer sources.Sourcer,
	metrics batchChangesMetrics,
) *workerutil.Worker {
	r := &bulkProcessorWorker{sourcer: sourcer, store: s}

	options := workerutil.WorkerOptions{
		Name:              "batches_bulk_processor",
		NumHandlers:       5,
		HeartbeatInterval: 15 * time.Second,
		Interval:          5 * time.Second,
		Metrics:           metrics.bulkProcessorWorkerMetrics,
	}

	worker := dbworker.NewWorker(ctx, workerStore, r.HandlerFunc(), options)
	return worker
}

// newBulkOperationWorkerResetter creates a dbworker.Resetter that reenqueues lost jobs
// for processing.
func newBulkOperationWorkerResetter(workerStore dbworkerstore.Store, metrics batchChangesMetrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batches_bulk_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.bulkProcessorWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}

func NewBulkOperationDBWorkerStore(handle *basestore.TransactableHandle, observationContext *observation.Context) dbworkerstore.Store {
	options := dbworkerstore.Options{
		Name:              "batches_bulk_worker_store",
		TableName:         "changeset_jobs",
		ColumnExpressions: store.ChangesetJobColumns.ToSqlf(),
		Scan:              scanFirstChangesetJobRecord,

		OrderByExpression: sqlf.Sprintf("changeset_jobs.state = 'errored', changeset_jobs.updated_at DESC"),

		StalledMaxAge: 60 * time.Second,
		MaxNumResets:  bulkProcessorMaxNumResets,

		RetryAfter:    5 * time.Second,
		MaxNumRetries: bulkProcessorMaxNumRetries,
	}

	return dbworkerstore.NewWithMetrics(handle, options, observationContext)
}

// scanFirstChangesetJobRecord wraps store.ScanFirstChangesetJob to return a
// generic workerutil.Record.
func scanFirstChangesetJobRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return store.ScanFirstChangesetJob(rows, err)
}

// bulkProcessorWorker is a wrapper for the workerutil handlerfunc to create a
// bulkProcessor with a source and store.
type bulkProcessorWorker struct {
	store   *store.Store
	sourcer sources.Sourcer
}

func (b *bulkProcessorWorker) HandlerFunc() workerutil.HandlerFunc {
	return func(ctx context.Context, record workerutil.Record) (err error) {
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

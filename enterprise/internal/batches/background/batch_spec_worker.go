package background

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

const batchspecMaxNumRetries = 60
const batchspecMaxNumResets = 60

// newBatchSpecWorker creates a dbworker.newWorker that fetches enqueued batch
// specs and passes them to the batchSpecWorkspaceCreator.
func newBatchSpecWorker(
	ctx context.Context,
	s *store.Store,
	workerStore dbworkerstore.Store,
	metrics batchChangesMetrics,
) *workerutil.Worker {
	e := &batchSpecWorkspaceCreator{store: s}

	options := workerutil.WorkerOptions{
		Name:              "batches_batchspec_worker",
		NumHandlers:       5,
		Interval:          5 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           metrics.reconcilerWorkerMetrics,
	}

	worker := dbworker.NewWorker(ctx, workerStore, e.HandlerFunc(), options)
	return worker
}

func newBatchSpecWorkerResetter(workerStore dbworkerstore.Store, metrics batchChangesMetrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batches_batch_spec_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.batchSpecWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}

func scanFirstBatchSpecRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return store.ScanFirstBatchSpec(rows, err)
}

func NewBatchSpecDBWorkerStore(handle *basestore.TransactableHandle, observationContext *observation.Context) dbworkerstore.Store {
	options := dbworkerstore.Options{
		Name:              "batches_batch_spec_worker_store",
		TableName:         "batch_specs",
		ColumnExpressions: store.BatchSpecColumns,
		Scan:              scanFirstBatchSpecRecord,

		OrderByExpression: sqlf.Sprintf("batch_specs.state = 'errored', batch_specs.updated_at DESC"),

		StalledMaxAge: 60 * time.Second,
		MaxNumResets:  batchspecMaxNumResets,

		RetryAfter:    5 * time.Second,
		MaxNumRetries: batchspecMaxNumRetries,
	}

	return dbworkerstore.NewWithMetrics(handle, options, observationContext)
}

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

// batchSpecResolutionMaxNumRetries sets the number of retries for batch spec
// resolutions to 0. We don't want to retry automatically and instead wait for
// user input
const batchSpecResolutionMaxNumRetries = 0
const batchSpecResolutionMaxNumResets = 60

// newBatchSpecResolutionWorker creates a dbworker.newWorker that fetches BatchSpecResolutionJobs
// specs and passes them to the batchSpecWorkspaceCreator.
func newBatchSpecResolutionWorker(
	ctx context.Context,
	s *store.Store,
	workerStore dbworkerstore.Store,
	metrics batchChangesMetrics,
) *workerutil.Worker {
	e := &batchSpecWorkspaceCreator{store: s}

	options := workerutil.WorkerOptions{
		Name:              "batch_changes_batch_spec_resolution_worker",
		NumHandlers:       5,
		Interval:          5 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           metrics.batchSpecResolutionWorkerMetrics,
	}

	worker := dbworker.NewWorker(ctx, workerStore, e.HandlerFunc(), options)
	return worker
}

func newBatchSpecResolutionWorkerResetter(workerStore dbworkerstore.Store, metrics batchChangesMetrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batch_changes_batch_spec_resolution_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.batchSpecResolutionWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}

func scanFirstBatchSpecResolutionJobRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return store.ScanFirstBatchSpecResolutionJob(rows, err)
}

func newBatchSpecResolutionWorkerStore(handle *basestore.TransactableHandle, observationContext *observation.Context) dbworkerstore.Store {
	options := dbworkerstore.Options{
		Name:              "batch_changes_batch_spec_resolution_worker_store",
		TableName:         "batch_spec_resolution_jobs",
		ColumnExpressions: store.BatchSpecResolutionJobColums.ToSqlf(),
		Scan:              scanFirstBatchSpecResolutionJobRecord,

		OrderByExpression: sqlf.Sprintf("batch_spec_resolution_jobs.state = 'errored', batch_spec_resolution_jobs.updated_at DESC"),

		StalledMaxAge: 60 * time.Second,
		MaxNumResets:  batchSpecResolutionMaxNumResets,

		RetryAfter:    5 * time.Second,
		MaxNumRetries: batchSpecResolutionMaxNumRetries,
	}

	return dbworkerstore.NewWithMetrics(handle, options, observationContext)
}

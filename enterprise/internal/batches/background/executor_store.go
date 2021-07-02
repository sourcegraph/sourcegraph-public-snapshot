package background

import (
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// executorStalledJobMaximumAge is the maximum allowable duration between updating the state of a
// job as "processing" and locking the record during processing. An unlocked row that is
// marked as processing likely indicates that the executor that dequeued the job has died.
// There should be a nearly-zero delay between these states during normal operation.
const executorStalledJobMaximumAge = time.Second * 5

// executorMaximumNumResets is the maximum number of times a job can be reset. If a job's failed
// attempts counter reaches this threshold, it will be moved into "errored" rather than
// "queued" on its next reset.
const executorMaximumNumResets = 3

var executorWorkerStoreOptions = dbworkerstore.Options{
	Name:              "batch_spec_executor_worker_store",
	TableName:         "batch_spec_executions",
	ColumnExpressions: store.BatchSpecExecutionColumns,
	Scan:              scanFirstExecutionRecord,
	OrderByExpression: sqlf.Sprintf("batch_spec_executions.created_at, batch_spec_executions.id"),
	StalledMaxAge:     executorStalledJobMaximumAge,
	MaxNumResets:      executorMaximumNumResets,
	// Explicitly disable retries.
	MaxNumRetries: 0,
}

// NewExecutorStore creates a dbworker store that wraps the batch_spec_executions
// table.
func NewExecutorStore(s basestore.ShareableStore, observationContext *observation.Context) dbworkerstore.Store {
	return dbworkerstore.NewWithMetrics(s.Handle(), executorWorkerStoreOptions, observationContext)
}

// scanFirstExecutionRecord scans a slice of batch change executions and returns the first.
func scanFirstExecutionRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return store.ScanFirstBatchSpecExecution(rows, err)
}

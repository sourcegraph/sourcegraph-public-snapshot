package batches

import (
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	apiserver "github.com/sourcegraph/sourcegraph/enterprise/cmd/executor-queue/internal/server"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// stalledJobMaximumAge is the maximum allowable duration between updating the state of a
// job as "processing" and locking the record during processing. An unlocked row that is
// marked as processing likely indicates that the executor that dequeued the job has died.
// There should be a nearly-zero delay between these states during normal operation.
const stalledJobMaximumAge = time.Second * 5

// maximumNumResets is the maximum number of times a job can be reset. If a job's failed
// attempts counter reaches this threshold, it will be moved into "errored" rather than
// "queued" on its next reset.
const maximumNumResets = 3

func QueueOptions(db dbutil.DB, config *Config, observationContext *observation.Context) apiserver.QueueOptions {
	recordTransformer := func(record workerutil.Record) (apiclient.Job, error) {
		return transformRecord(record.(*btypes.BatchSpecExecution), config)
	}

	return apiserver.QueueOptions{
		Store:             newWorkerStore(db, observationContext),
		RecordTransformer: recordTransformer,
	}
}

// newWorkerStore creates a dbworker store that wraps the batch_spec_executions
// table.
func newWorkerStore(db dbutil.DB, observationContext *observation.Context) dbworkerstore.Store {
	handle := basestore.NewHandleWithDB(db, sql.TxOptions{})
	options := dbworkerstore.Options{
		Name:              "batch_spec_execution_worker_store",
		TableName:         "batch_spec_executions",
		ColumnExpressions: store.BatchSpecExecutionColumns,
		Scan:              scanFirstRecord,
		OrderByExpression: sqlf.Sprintf("batch_spec_executions.created_at, batch_spec_executions.id"),
		StalledMaxAge:     stalledJobMaximumAge,
		MaxNumResets:      maximumNumResets,
		// Explicitly disable retries.
		MaxNumRetries: 0,
	}

	return dbworkerstore.NewWithMetrics(handle, options, observationContext)
}

// scanFirstRecord scans a slice of batch change executions and returns the first.
func scanFirstRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return store.ScanFirstBatchSpecExecution(rows, err)
}

package store

import (
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// bulkProcessorMaxNumRetries is the maximum number of attempts the bulkProcessor
// makes to process a changeset job when it fails.
const bulkProcessorMaxNumRetries = 10

// bulkProcessorMaxNumResets is the maximum number of attempts the bulkProcessor
// makes to process a changeset job when it stalls (process crashes, etc.).
const bulkProcessorMaxNumResets = 60

var bulkOperationWorkerStoreOpts = dbworkerstore.Options{
	Name:              "batches_bulk_worker_store",
	TableName:         "changeset_jobs",
	ColumnExpressions: changesetJobColumns.ToSqlf(),

	Scan: func(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
		return scanFirstChangesetJob(rows, err)
	},

	OrderByExpression: sqlf.Sprintf("changeset_jobs.state = 'errored', changeset_jobs.updated_at DESC"),

	StalledMaxAge: 60 * time.Second,
	MaxNumResets:  bulkProcessorMaxNumResets,

	RetryAfter:    5 * time.Second,
	MaxNumRetries: bulkProcessorMaxNumRetries,
}

func NewBulkOperationWorkerStore(handle *basestore.TransactableHandle, observationContext *observation.Context) dbworkerstore.Store {
	return dbworkerstore.NewWithMetrics(handle, bulkOperationWorkerStoreOpts, observationContext)
}

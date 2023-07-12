package store

import (
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// batchSpecResolutionMaxNumRetries sets the number of retries for batch spec
// resolutions to 0. We don't want to retry automatically and instead wait for
// user input
const batchSpecResolutionMaxNumRetries = 0
const batchSpecResolutionMaxNumResets = 60

var batchSpecResolutionWorkerOpts = dbworkerstore.Options[*types.BatchSpecResolutionJob]{
	Name:              "batch_changes_batch_spec_resolution_worker_store",
	TableName:         "batch_spec_resolution_jobs",
	ColumnExpressions: batchSpecResolutionJobColums.ToSqlf(),

	Scan: dbworkerstore.BuildWorkerScan(buildRecordScanner(scanBatchSpecResolutionJob)),

	OrderByExpression: sqlf.Sprintf("batch_spec_resolution_jobs.state = 'errored', batch_spec_resolution_jobs.updated_at DESC"),

	StalledMaxAge: 60 * time.Second,
	MaxNumResets:  batchSpecResolutionMaxNumResets,

	RetryAfter:    5 * time.Second,
	MaxNumRetries: batchSpecResolutionMaxNumRetries,
}

func NewBatchSpecResolutionWorkerStore(observationCtx *observation.Context, handle basestore.TransactableHandle) dbworkerstore.Store[*types.BatchSpecResolutionJob] {
	return dbworkerstore.New(observationCtx, handle, batchSpecResolutionWorkerOpts)
}

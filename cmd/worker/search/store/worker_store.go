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

// maxNumRetries is the maximum number of attempts the search indexer makes to index
// a repo at a revision.
const maxNumRetries = 10

// maxNumResets is the maximum number of attempts the search indexer makes to index
// a repo at a revision when it stalls (process crashes, etc.).
const maxNumResets = 60

var workerStoreOpts = dbworkerstore.Options{
	Name:              "search_index_worker_store",
	TableName:         "search_index_jobs",
	ColumnExpressions: searchIndexJobColumns,

	Scan: func(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
		if err != nil {
			return nil, false, err
		}
		return scanFirstSearchIndexJob(rows)
	},

	OrderByExpression: sqlf.Sprintf("search_index_jobs.state = 'errored', search_index_jobs.updated_at DESC"),

	StalledMaxAge: 60 * time.Second,
	MaxNumResets:  maxNumResets,

	RetryAfter:    5 * time.Second,
	MaxNumRetries: maxNumRetries,
}

func NewSearchIndexWorkerStore(handle *basestore.TransactableHandle, observationContext *observation.Context) dbworkerstore.Store {
	return dbworkerstore.NewWithMetrics(handle, workerStoreOpts, observationContext)
}

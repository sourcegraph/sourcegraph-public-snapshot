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

// reconcilerMaxNumRetries is the maximum number of attempts the reconciler
// makes to process a changeset when it fails.
const reconcilerMaxNumRetries = 60

// reconcilerMaxNumResets is the maximum number of attempts the reconciler
// makes to process a changeset when it stalls (process crashes, etc.).
const reconcilerMaxNumResets = 60

var reconcilerWorkerStoreOpts = dbworkerstore.Options{
	Name:                 "batches_reconciler_worker_store",
	TableName:            "changesets",
	ViewName:             "reconciler_changesets changesets",
	AlternateColumnNames: map[string]string{"state": "reconciler_state"},
	ColumnExpressions:    changesetColumns,

	Scan: func(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
		return scanFirstChangeset(rows, err)
	},

	// Order changesets by state, so that freshly enqueued changesets have
	// higher priority.
	// If state is equal, prefer the newer ones.
	OrderByExpression: sqlf.Sprintf("changesets.reconciler_state = 'errored', changesets.updated_at DESC"),

	StalledMaxAge: 60 * time.Second,
	MaxNumResets:  reconcilerMaxNumResets,

	RetryAfter:    5 * time.Second,
	MaxNumRetries: reconcilerMaxNumRetries,
}

func NewReconcilerWorkerStore(handle *basestore.TransactableHandle, observationContext *observation.Context) dbworkerstore.Store {
	return dbworkerstore.NewWithMetrics(handle, reconcilerWorkerStoreOpts, observationContext)
}

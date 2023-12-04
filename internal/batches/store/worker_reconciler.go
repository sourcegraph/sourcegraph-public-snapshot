package store

import (
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// reconcilerMaxNumRetries is the maximum number of attempts the reconciler
// makes to process a changeset when it fails.
const reconcilerMaxNumRetries = 10

// reconcilerMaxNumResets is the maximum number of attempts the reconciler
// makes to process a changeset when it stalls (process crashes, etc.).
const reconcilerMaxNumResets = 10

var reconcilerWorkerStoreOpts = dbworkerstore.Options[*types.Changeset]{
	Name:                 "batches_reconciler_worker_store",
	TableName:            "changesets",
	ViewName:             "reconciler_changesets changesets",
	AlternateColumnNames: map[string]string{"state": "reconciler_state"},
	ColumnExpressions:    ChangesetColumns,

	Scan: dbworkerstore.BuildWorkerScan(buildRecordScanner(ScanChangeset)),

	// Order changesets by state, so that freshly enqueued changesets have
	// higher priority.
	// If state is equal, prefer the newer ones.
	OrderByExpression: sqlf.Sprintf("changesets.reconciler_state = 'errored', changesets.updated_at DESC"),

	StalledMaxAge: 60 * time.Second,
	MaxNumResets:  reconcilerMaxNumResets,

	RetryAfter:    5 * time.Second,
	MaxNumRetries: reconcilerMaxNumRetries,
}

func NewReconcilerWorkerStore(observationCtx *observation.Context, handle basestore.TransactableHandle) dbworkerstore.Store[*types.Changeset] {
	return dbworkerstore.New(observationCtx, handle, reconcilerWorkerStoreOpts)
}

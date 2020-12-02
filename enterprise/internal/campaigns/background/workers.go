package background

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// reconcilerMaxNumRetries is the maximum number of attempts the reconciler
// makes to process a changeset when it fails.
const reconcilerMaxNumRetries = 60

// reconcilerMaxNumResets is the maximum number of attempts the reconciler
// makes to process a changeset when it stalls (process crashes, etc.).
const reconcilerMaxNumResets = 60

// newWorker creates a dbworker.newWorker that fetches enqueued changesets
// from the database and passes them to the changeset reconciler for
// processing.
func newWorker(
	ctx context.Context,
	s *campaigns.Store,
	gitClient campaigns.GitserverClient,
	sourcer repos.Sourcer,
	metrics campaignsMetrics,
) *workerutil.Worker {
	r := &campaigns.Reconciler{GitserverClient: gitClient, Sourcer: sourcer, Store: s}

	options := workerutil.WorkerOptions{
		NumHandlers: 5,
		Interval:    5 * time.Second,
		Metrics: workerutil.WorkerMetrics{
			HandleOperation: metrics.handleOperation,
		},
	}

	workerStore := createDBWorkerStore(s)

	worker := dbworker.NewWorker(ctx, workerStore, r.HandlerFunc(), options)
	return worker
}

func newWorkerResetter(s *campaigns.Store, metrics campaignsMetrics) *dbworker.Resetter {
	workerStore := createDBWorkerStore(s)

	options := dbworker.ResetterOptions{
		Name:     "campaigns_reconciler_resetter",
		Interval: 1 * time.Minute,
		Metrics: dbworker.ResetterMetrics{
			Errors:              metrics.errors,
			RecordResetFailures: metrics.resetFailures,
			RecordResets:        metrics.resets,
		},
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}

func scanFirstChangesetRecord(rows *sql.Rows, err error) (workerutil.Record, bool, error) {
	return campaigns.ScanFirstChangeset(rows, err)
}

func createDBWorkerStore(s *campaigns.Store) dbworkerstore.Store {
	return dbworkerstore.New(s.Handle(), dbworkerstore.Options{
		TableName:            "changesets",
		AlternateColumnNames: map[string]string{"state": "reconciler_state"},
		ColumnExpressions:    campaigns.ChangesetColumns,
		Scan:                 scanFirstChangesetRecord,

		// Order changesets by state, so that freshly enqueued changesets have
		// higher priority.
		// If state is equal, prefer the newer ones.
		OrderByExpression: sqlf.Sprintf("reconciler_state = 'errored', changesets.updated_at DESC"),

		StalledMaxAge: 60 * time.Second,
		MaxNumResets:  reconcilerMaxNumResets,

		RetryAfter:    5 * time.Second,
		MaxNumRetries: reconcilerMaxNumRetries,
	})
}

package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/reconciler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// newReconcilerWorker creates a dbworker.newWorker that fetches enqueued changesets
// from the database and passes them to the changeset reconciler for
// processing.
func newReconcilerWorker(
	ctx context.Context,
	s *store.Store,
	workerStore dbworkerstore.Store,
	gitClient reconciler.GitserverClient,
	sourcer sources.Sourcer,
	metrics batchChangesMetrics,
) *workerutil.Worker {
	r := reconciler.New(gitClient, sourcer, s)

	options := workerutil.WorkerOptions{
		Name:              "batches_reconciler_worker",
		NumHandlers:       5,
		Interval:          5 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           metrics.reconcilerWorkerMetrics,
	}

	worker := dbworker.NewWorker(ctx, workerStore, r.HandlerFunc(), options)
	return worker
}

func newReconcilerWorkerResetter(workerStore dbworkerstore.Store, metrics batchChangesMetrics) *dbworker.Resetter {
	options := dbworker.ResetterOptions{
		Name:     "batches_reconciler_worker_resetter",
		Interval: 1 * time.Minute,
		Metrics:  metrics.reconcilerWorkerResetterMetrics,
	}

	resetter := dbworker.NewResetter(workerStore, options)
	return resetter
}

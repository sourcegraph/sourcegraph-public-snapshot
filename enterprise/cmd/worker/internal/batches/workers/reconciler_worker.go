package workers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/reconciler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// NewReconcilerWorker creates a dbworker.newWorker that fetches enqueued changesets
// from the database and passes them to the changeset reconciler for
// processing.
func NewReconcilerWorker(
	ctx context.Context,
	s *store.Store,
	workerStore dbworkerstore.Store,
	gitClient reconciler.GitserverClient,
	sourcer sources.Sourcer,
	observationContext *observation.Context,
) *workerutil.Worker {
	r := reconciler.New(gitClient, sourcer, s)

	options := workerutil.WorkerOptions{
		Name:              "batches_reconciler_worker",
		NumHandlers:       5,
		Interval:          5 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observationContext, "batch_changes_reconciler"),
	}

	worker := dbworker.NewWorker(ctx, workerStore, r.HandlerFunc(), options)
	return worker
}

package workers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/batches/reconciler"
	"github.com/sourcegraph/sourcegraph/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
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
	observationCtx *observation.Context,
	s *store.Store,
	workerStore dbworkerstore.Store[*btypes.Changeset],
	gitClient gitserver.Client,
	sourcer sources.Sourcer,
) *workerutil.Worker[*btypes.Changeset] {
	r := reconciler.New(gitClient, sourcer, s)

	options := workerutil.WorkerOptions{
		Name:              "batches_reconciler_worker",
		Description:       "changeset reconciler that publishes, modifies and closes changesets on the code host",
		NumHandlers:       5,
		Interval:          5 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		Metrics:           workerutil.NewMetrics(observationCtx, "batch_changes_reconciler"),
	}

	worker := dbworker.NewWorker[*btypes.Changeset](ctx, workerStore, r.HandlerFunc(), options)
	return worker
}

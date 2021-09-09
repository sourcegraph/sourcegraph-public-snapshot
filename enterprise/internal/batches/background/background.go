package background

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/scheduler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Routines(ctx context.Context, batchesStore *store.Store, cf *httpcli.Factory, observationContext *observation.Context) []goroutine.BackgroundRoutine {
	sourcer := sources.NewSourcer(cf)
	metrics := newMetrics(observationContext)

	reconcilerWorkerStore := NewReconcilerDBWorkerStore(batchesStore.Handle(), observationContext)
	bulkProcessorWorkerStore := NewBulkOperationDBWorkerStore(batchesStore.Handle(), observationContext)
	specExecutionWorkerStore := NewExecutorStore(batchesStore.Handle(), observationContext)

	batchSpecWorkspaceExecutionWorkerStore := NewBatchSpecWorkspaceExecutionWorkerStore(batchesStore.Handle(), observationContext)
	batchSpecResolutionWorkerStore := newBatchSpecResolutionWorkerStore(batchesStore.Handle(), observationContext)

	routines := []goroutine.BackgroundRoutine{
		newReconcilerWorker(ctx, batchesStore, reconcilerWorkerStore, gitserver.DefaultClient, sourcer, metrics),
		newReconcilerWorkerResetter(reconcilerWorkerStore, metrics),

		newSpecExpireJob(ctx, batchesStore),

		scheduler.NewScheduler(ctx, batchesStore),

		newBulkOperationWorker(ctx, batchesStore, bulkProcessorWorkerStore, sourcer, metrics),
		newBulkOperationWorkerResetter(bulkProcessorWorkerStore, metrics),

		newBatchSpecExecutionResetter(specExecutionWorkerStore, metrics),

		newBatchSpecResolutionWorker(ctx, batchesStore, batchSpecResolutionWorkerStore, metrics),
		newBatchSpecResolutionWorkerResetter(batchSpecResolutionWorkerStore, metrics),

		newBatchSpecWorkspaceExecutionWorkerResetter(batchSpecWorkspaceExecutionWorkerStore, metrics),
	}
	return routines
}

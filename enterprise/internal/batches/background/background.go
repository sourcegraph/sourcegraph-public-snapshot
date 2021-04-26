package background

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/scheduler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

func Routines(ctx context.Context, batchesStore *store.Store, cf *httpcli.Factory) []goroutine.BackgroundRoutine {
	sourcer := sources.NewSourcer(cf)
	metrics := newMetrics()

	routines := []goroutine.BackgroundRoutine{
		newReconcilerWorker(ctx, batchesStore, gitserver.DefaultClient, sourcer, metrics),
		newReconcilerWorkerResetter(batchesStore, metrics),

		newSpecExpireWorker(ctx, batchesStore),

		scheduler.NewScheduler(ctx, batchesStore),

		newBulkJobWorker(ctx, batchesStore, sourcer, metrics),
		newBulkJobWorkerResetter(batchesStore, metrics),
	}
	return routines
}

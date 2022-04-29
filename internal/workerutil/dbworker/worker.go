package dbworker

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

func NewWorker(ctx context.Context, logger log.Logger, store store.Store, handler workerutil.Handler, options workerutil.WorkerOptions) *workerutil.Worker {
	return workerutil.NewWorker(ctx, newStoreShim(store), handler, options)
}

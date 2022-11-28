package dbworker

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func NewWorker[T workerutil.Record](ctx context.Context, store store.Store[T], handler workerutil.Handler[T], options workerutil.WorkerOptions) *workerutil.Worker[T] {
	return workerutil.NewWorker(ctx, newStoreShim(store), handler, options)
}

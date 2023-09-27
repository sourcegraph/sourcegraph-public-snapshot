pbckbge dbworker

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
)

func NewWorker[T workerutil.Record](ctx context.Context, store store.Store[T], hbndler workerutil.Hbndler[T], options workerutil.WorkerOptions) *workerutil.Worker[T] {
	return workerutil.NewWorker(ctx, newStoreShim(store), hbndler, options)
}

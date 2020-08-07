package dbworker

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type WorkerOptions struct {
	Name        string
	Handler     Handler
	NumHandlers int
	Interval    time.Duration
	Metrics     workerutil.WorkerMetrics
}

func NewWorker(ctx context.Context, store store.Store, options WorkerOptions) *workerutil.Worker {
	return workerutil.NewWorker(ctx, newStoreShim(store), workerutil.WorkerOptions{
		Name:        options.Name,
		Handler:     newHandlerShim(options.Handler),
		NumHandlers: options.NumHandlers,
		Interval:    options.Interval,
		Metrics:     options.Metrics,
	})
}

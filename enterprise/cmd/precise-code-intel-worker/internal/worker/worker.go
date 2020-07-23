package worker

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func NewWorker(
	s store.Store,
	bundleManagerClient bundles.BundleManagerClient,
	gitserverClient gitserver.Client,
	pollInterval time.Duration,
	numProcessorRoutines int,
	budgetMax int64,
	metrics metrics.WorkerMetrics,
) *workerutil.Worker {
	rootContext := actor.WithActor(context.Background(), &actor.Actor{Internal: true})

	processor := &processor{
		bundleManagerClient: bundleManagerClient,
		gitserverClient:     gitserverClient,
		metrics:             metrics,
	}

	handler := &handler{
		store:           s,
		processor:       processor,
		enableBudget:    budgetMax > 0,
		budgetRemaining: budgetMax,
	}

	workerMetrics := workerutil.WorkerMetrics{
		HandleOperation: metrics.ProcessOperation,
	}

	options := workerutil.WorkerOptions{
		Handler:     handler,
		NumHandlers: numProcessorRoutines,
		Interval:    pollInterval,
		Metrics:     workerMetrics,
	}

	return workerutil.NewWorker(rootContext, store.WorkerutilUploadStore(s), options)
}

type handler struct {
	store           store.Store
	processor       *processor
	enableBudget    bool
	budgetRemaining int64
}

func (h *handler) Handle(ctx context.Context, tx workerutil.Store, record workerutil.Record) error {
	_, err := h.processor.Process(ctx, h.store.With(tx), record.(store.Upload))
	return err
}

func (h *handler) PreDequeue(ctx context.Context) (bool, []*sqlf.Query, error) {
	if !h.enableBudget {
		return true, nil, nil
	}

	budgetRemaining := atomic.LoadInt64(&h.budgetRemaining)
	if budgetRemaining <= 0 {
		return false, nil, nil
	}

	return true, []*sqlf.Query{sqlf.Sprintf("(upload_size IS NULL OR upload_size <= %s)", budgetRemaining)}, nil
}

func (h *handler) PreHandle(ctx context.Context, record workerutil.Record) {
	atomic.AddInt64(&h.budgetRemaining, -h.getSize(record))
}

func (h *handler) PostHandle(ctx context.Context, record workerutil.Record) {
	atomic.AddInt64(&h.budgetRemaining, +h.getSize(record))
}

func (h *handler) getSize(record workerutil.Record) int64 {
	if size := record.(store.Upload).UploadSize; size != nil {
		return *size
	}

	return 0
}

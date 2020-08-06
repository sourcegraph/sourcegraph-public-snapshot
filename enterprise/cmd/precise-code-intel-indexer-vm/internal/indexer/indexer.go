package indexer

import (
	"context"
	"time"

	indexmanager "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/index_manager"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/queue/client"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type IndexerOptions struct {
	NumIndexers    int
	Interval       time.Duration
	Metrics        IndexerMetrics
	HandlerOptions HandlerOptions
}

func NewIndexer(ctx context.Context, queueClient client.Client, indexManager *indexmanager.Manager, options IndexerOptions) *workerutil.Worker {
	handler := &Handler{
		queueClient:  queueClient,
		indexManager: indexManager,
		commander:    DefaultCommander,
		options:      options.HandlerOptions,
	}

	workerMetrics := workerutil.WorkerMetrics{
		HandleOperation: options.Metrics.ProcessOperation,
	}

	return workerutil.NewWorker(ctx, &storeShim{queueClient}, workerutil.WorkerOptions{
		Handler:     handler,
		NumHandlers: options.NumIndexers,
		Interval:    options.Interval,
		Metrics:     workerMetrics,
	})
}

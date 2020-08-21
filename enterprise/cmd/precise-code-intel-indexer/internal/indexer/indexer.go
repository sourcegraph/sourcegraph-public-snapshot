package indexer

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func NewIndexer(
	s store.Store,
	gitserverClient gitserver.Client,
	frontendURL string,
	pollInterval time.Duration,
	metrics IndexerMetrics,
) goroutine.BackgroundRoutine {
	rootContext := actor.WithActor(context.Background(), &actor.Actor{Internal: true})

	processor := &processor{
		store:           s,
		gitserverClient: gitserverClient,
		frontendURL:     frontendURL,
	}

	handler := dbworker.HandlerFunc(func(ctx context.Context, tx dbworkerstore.Store, record workerutil.Record) error {
		return processor.Process(ctx, record.(store.Index))
	})

	workerMetrics := workerutil.WorkerMetrics{
		HandleOperation: metrics.ProcessOperation,
	}

	options := dbworker.WorkerOptions{
		Handler:     handler,
		NumHandlers: 1,
		Interval:    pollInterval,
		Metrics:     workerMetrics,
	}

	return dbworker.NewWorker(rootContext, store.WorkerutilIndexStore(s), options)
}

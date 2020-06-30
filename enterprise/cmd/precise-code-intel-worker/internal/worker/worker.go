package worker

import (
	"context"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

type Worker struct {
	store        store.Store
	processor    Processor
	pollInterval time.Duration
	metrics      metrics.WorkerMetrics
	done         chan struct{}
	once         sync.Once
}

func NewWorker(
	store store.Store,
	bundleManagerClient bundles.BundleManagerClient,
	gitserverClient gitserver.Client,
	pollInterval time.Duration,
	metrics metrics.WorkerMetrics,
) *Worker {
	processor := &processor{
		bundleManagerClient: bundleManagerClient,
		gitserverClient:     gitserverClient,
		metrics:             metrics,
	}

	return &Worker{
		store:        store,
		processor:    processor,
		pollInterval: pollInterval,
		metrics:      metrics,
		done:         make(chan struct{}),
	}
}

func (w *Worker) Start() {
	ctx := actor.WithActor(context.Background(), &actor.Actor{Internal: true})

	for {
		if ok, _ := w.dequeueAndProcess(ctx); !ok {
			select {
			case <-time.After(w.pollInterval):
			case <-w.done:
				return
			}
		} else {
			select {
			case <-w.done:
				return
			default:
			}
		}
	}
}

func (w *Worker) Stop() {
	w.once.Do(func() {
		close(w.done)
	})
}

// TODO(efritz) - use cancellable context

// dequeueAndProcess pulls a job from the queue and processes it. If there
// were no jobs ready to process, this method returns a false-valued flag.
func (w *Worker) dequeueAndProcess(ctx context.Context) (_ bool, err error) {
	upload, store, ok, err := w.store.Dequeue(ctx)
	if err != nil || !ok {
		return false, errors.Wrap(err, "store.Dequeue")
	}

	// Enable tracing on this context
	ctx = ot.WithShouldTrace(ctx, true)

	// Trace the remainder of the operation including the transaction commit call in
	// the following defered function.
	ctx, endOperation := w.metrics.ProcessOperation.With(ctx, &err, observation.Args{})

	defer func() {
		err = store.Done(err)
		endOperation(1, observation.Args{})
	}()

	log15.Info("Dequeued upload for processing", "id", upload.ID)

	// TODO - same for janitors/resetters
	if requeued, processErr := w.processor.Process(ctx, store, upload); processErr == nil {
		if requeued {
			log15.Info("Requeueing upload", "id", upload.ID)
		} else {
			log15.Info("Processed upload", "id", upload.ID)
		}
	} else {
		// TODO(efritz) - distinguish between correlation and system errors
		log15.Warn("Failed to process upload", "id", upload.ID, "err", processErr)

		if markErr := store.MarkErrored(ctx, upload.ID, processErr.Error()); markErr != nil {
			return true, errors.Wrap(markErr, "store.MarkErrored")
		}
	}

	return true, nil
}

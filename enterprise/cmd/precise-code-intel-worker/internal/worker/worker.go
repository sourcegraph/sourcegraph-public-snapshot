package worker

import (
	"context"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type Worker struct {
	store        store.Store
	processor    Processor
	pollInterval time.Duration
	metrics      WorkerMetrics
	done         chan struct{}
	once         sync.Once
}

func NewWorker(
	store store.Store,
	bundleManagerClient bundles.BundleManagerClient,
	gitserverClient gitserver.Client,
	pollInterval time.Duration,
	metrics WorkerMetrics,
) *Worker {
	processor := &processor{
		bundleManagerClient: bundleManagerClient,
		gitserverClient:     gitserverClient,
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
	for {
		if ok, _ := w.dequeueAndProcess(context.Background()); !ok {
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
	start := time.Now()

	upload, store, ok, err := w.store.Dequeue(ctx)
	if err != nil || !ok {
		return false, errors.Wrap(err, "store.Dequeue")
	}
	defer func() {
		err = store.Done(err)

		// TODO(efritz) - set error if correlation failed
		w.metrics.Processor.Observe(time.Since(start).Seconds(), 1, &err)
	}()

	log15.Info("Dequeued upload for processing", "id", upload.ID)

	if processErr := w.processor.Process(ctx, store, upload); processErr == nil {
		log15.Info("Processed upload", "id", upload.ID)
	} else {
		// TODO(efritz) - distinguish between correlation and system errors
		log15.Warn("Failed to process upload", "id", upload.ID, "err", processErr)

		if markErr := store.MarkErrored(ctx, upload.ID, processErr.Error()); markErr != nil {
			return true, errors.Wrap(markErr, "store.MarkErrored")
		}
	}

	return true, nil
}

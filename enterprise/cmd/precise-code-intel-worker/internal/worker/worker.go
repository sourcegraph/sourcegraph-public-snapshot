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
	store           store.Store
	processor       Processor
	pollInterval    time.Duration
	metrics         metrics.WorkerMetrics
	ctx             context.Context
	done            chan<- struct{}
	once            sync.Once
	semaphore       chan struct{}
	budgetMax       int
	budgetRemaining int
}

func NewWorker(
	store store.Store,
	bundleManagerClient bundles.BundleManagerClient,
	gitserverClient gitserver.Client,
	pollInterval time.Duration,
	numProcessorRoutines int,
	budgetMax int,
	metrics metrics.WorkerMetrics,
) *Worker {
	processor := &processor{
		bundleManagerClient: bundleManagerClient,
		gitserverClient:     gitserverClient,
		metrics:             metrics,
	}

	return newWorker(
		store,
		processor,
		pollInterval,
		numProcessorRoutines,
		budgetMax,
		metrics,
	)
}

func newWorker(
	store store.Store,
	processor Processor,
	pollInterval time.Duration,
	numProcessorRoutines int,
	budgetMax int,
	metrics metrics.WorkerMetrics,
) *Worker {
	ctx, done := makeContext()

	semaphore := make(chan struct{}, numProcessorRoutines)
	for i := 0; i < numProcessorRoutines; i++ {
		semaphore <- struct{}{}
	}

	return &Worker{
		store:           store,
		processor:       processor,
		pollInterval:    pollInterval,
		metrics:         metrics,
		ctx:             ctx,
		done:            done,
		semaphore:       semaphore,
		budgetMax:       budgetMax,
		budgetRemaining: budgetMax,
	}
}

func (w *Worker) Start() {
	ctx := w.ctx

	for {
		ok, err := w.dequeueAndProcess(ctx)
		if err != nil {
			log15.Error("Failed to dequeue upload", "err", err)
		}

		delay := w.pollInterval
		if ok {
			// Don't wait between successful dequeues
			delay = 0
		}

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) Stop() {
	w.once.Do(func() {
		close(w.done)
	})
}

// dequeueAndProcess selects a queued upload record to process. This method returns false
// if no such method can be dequeued and returns an error only on failure to dequeue a new
// record (no processor errors are bubbled up past this point).
func (w *Worker) dequeueAndProcess(ctx context.Context) (dequeued bool, err error) {
	if !w.reserveProcessorRoutine(ctx) {
		return false, nil
	}
	defer func() {
		if !dequeued {
			// Ensure we release the processor routine back to the
			// pool if we did not start a new one.
			w.releaseProcessorRoutine()
		}
	}()

	maxSize := 0
	if w.budgetMax > 0 {
		if w.budgetRemaining <= 0 {
			return false, nil
		}

		maxSize = w.budgetRemaining
	}

	// Select a queued upload to process and the transaction that holds it
	upload, store, ok, err := w.store.Dequeue(ctx, maxSize)
	if err != nil {
		return false, errors.Wrap(err, "store.Dequeue")
	}
	if !ok {
		return false, nil
	}

	size := 0
	if upload.UploadSize != nil {
		size = *upload.UploadSize
	}

	w.budgetRemaining -= size

	go func() {
		defer func() {
			w.budgetRemaining += size
			w.releaseProcessorRoutine()
		}()

		if err := w.handle(ctx, store, upload); err != nil {
			log15.Error("Failed to finalize upload record", "err", err)
		}
	}()

	return true, nil
}

// handle processes the given upload record. This method returns an error only if there
// is an issue committing the transaction (no processor errors are bubbled up past this point).
func (w *Worker) handle(ctx context.Context, store store.Store, upload store.Upload) (err error) {
	// Enable tracing on this context
	ctx = ot.WithShouldTrace(ctx, true)

	// Trace the remainder of the operation including the transaction commit call in
	// the following deferred function.
	ctx, endOperation := w.metrics.ProcessOperation.With(ctx, &err, observation.Args{})

	defer func() {
		err = store.Done(err)
		endOperation(1, observation.Args{})
	}()

	log15.Info("Dequeued upload for processing", "id", upload.ID)

	requeued, processErr := w.processor.Process(ctx, store, upload)
	if processErr != nil {
		log15.Warn("Failed to process upload", "id", upload.ID, "err", processErr)

		if markErr := store.MarkErrored(ctx, upload.ID, processErr.Error()); markErr != nil {
			return errors.Wrap(markErr, "store.MarkErrored")
		}

		return nil
	}

	if requeued {
		log15.Info("Requeueing upload", "id", upload.ID)
	} else {
		log15.Info("Processed upload", "id", upload.ID)
	}

	return nil
}

// reserveProcessorRoutine blocks until there is room for another processor routine
// to start. This method returns false if the context is canceled before a blocking
// processor has finished. If this method returns true, releaseProcessorRoutine must
// be called at the end of the processor function, or the worker will leak capacity.
func (w *Worker) reserveProcessorRoutine(ctx context.Context) bool {
	select {
	case <-w.semaphore:
		return true
	case <-ctx.Done():
		return false
	}
}

// releaseProcessOrRoutine signals that a processor routine has finished.
func (w *Worker) releaseProcessorRoutine() {
	w.semaphore <- struct{}{}
}

// makeContext returns an internal context and a write-only channel. The context will
// be closed when the user closes the channel.
func makeContext() (context.Context, chan<- struct{}) {
	ctx, cancel := context.WithCancel(actor.WithActor(context.Background(), &actor.Actor{Internal: true}))

	done := make(chan struct{})
	go func() {
		<-done
		defer cancel()
	}()

	return ctx, done
}

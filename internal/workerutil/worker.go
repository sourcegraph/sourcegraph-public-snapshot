package workerutil

import (
	"context"
	"sync"
	"time"

	"github.com/efritz/glock"
	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// Worker is a generic consumer of records from the workerutil store.
type Worker struct {
	store            Store
	options          WorkerOptions
	clock            glock.Clock
	handlerSemaphore chan struct{}   // tracks available handler slots
	ctx              context.Context // root context passed to the handler
	cancel           func()          // cancels the root context
	wg               sync.WaitGroup  // tracks active handler routines
	finished         chan struct{}   // signals that Start has finished
}

type WorkerOptions struct {
	Name        string
	Handler     Handler
	NumHandlers int
	Interval    time.Duration
	Metrics     WorkerMetrics
}

// Handler is the configurable consumer within a worker. Types that conform to this
// interface may also optionally conform to the PreDequeuer, PreHandler, and PostHandler
// interfaces to further configure the behavior of the worker routine.
type Handler interface {
	// Handle processes a single record. The store provided by this method a store backed
	// by the transaction that is locking the given record. If use of a database is necessary
	// within this handler, other stores should take the underlying handler to keep work
	// within the same transaction.
	//
	//     func (h *handler) Handle(ctx context.Context, tx workerutil.Store, record workerutil.Record) error {
	//         myStore := h.myStore.With(tx) // combine store handles
	//         myRecord := record.(MyType)   // convert type of record
	//         // do processing ...
	//         return nil
	//     }
	Handle(ctx context.Context, store Store, record Record) error
}

// HandlerWithPreDequeue is an extension of the Handler interface.
type HandlerWithPreDequeue interface {
	Handler

	// PreDequeue is called, if implemented, directly before a call to the store's Dequeue method.
	// If this method returns false, then the current worker iteration is skipped and the next iteration
	// will begin after waiting for the configured polling interval. Any SQL queries returned by this
	// method will be supplied as additional conditions to the store's Dequeue method.
	PreDequeue(ctx context.Context) (bool, []*sqlf.Query, error)
}

// HandlerWithHooks is an extension of the Handler interface.
//
// Example use case:
// The processor for LSIF uploads has a maximum budget based on input size. PreHandle will subtract
// the input size (atomically) from the budget and PostHandle will restore the input size back to the
// budget. The PreDequeue hook is also implemented to supply additional SQL conditions that ensures no
// record with a larger input sizes than the current budget will be dequeued by the worker process.
type HandlerWithHooks interface {
	Handler

	// PreHandle is called, if implemented, directly before a invoking the handler with the given
	// record. This method is invoked before starting a handler goroutine - therefore, any expensive
	// operations in this method will block the dequeue loop from proceeding.
	PreHandle(ctx context.Context, record Record)

	// PostHandle is called, if implemented, directly after the handler for the given record has
	// completed. This method is invoked inside the handler goroutine. Note that if PreHandle and
	// PostHandle both operate on shared data, that they will be operating on the data from different
	// goroutines and it is up to the caller to properly synchronize access to it.
	PostHandle(ctx context.Context, record Record)
}

type HandlerFunc func(ctx context.Context, store Store, record Record) error

func (f HandlerFunc) Handle(ctx context.Context, store Store, record Record) error {
	return f(ctx, store, record)
}

type WorkerMetrics struct {
	HandleOperation *observation.Operation
}

func NewWorker(ctx context.Context, store Store, options WorkerOptions) *Worker {
	return newWorker(ctx, store, options, glock.NewRealClock())
}

func newWorker(ctx context.Context, store Store, options WorkerOptions, clock glock.Clock) *Worker {
	ctx, cancel := context.WithCancel(ctx)

	handlerSemaphore := make(chan struct{}, options.NumHandlers)
	for i := 0; i < options.NumHandlers; i++ {
		handlerSemaphore <- struct{}{}
	}

	return &Worker{
		store:            store,
		options:          options,
		clock:            clock,
		handlerSemaphore: handlerSemaphore,
		ctx:              ctx,
		cancel:           cancel,
		finished:         make(chan struct{}),
	}
}

// Start begins polling for work from the underlying store and processing records.
func (w *Worker) Start() {
	defer close(w.finished)

loop:
	for {
		ok, err := w.dequeueAndHandle()
		if err != nil {
			for ex := err; ex != nil; ex = errors.Unwrap(ex) {
				if err == w.ctx.Err() {
					break loop
				}
			}

			log15.Error("Failed to dequeue and handle record", "name", w.options.Name, "err", err)
		}

		delay := w.options.Interval
		if ok {
			// If we had a successful dequeue, do not wait the poll interval.
			// Just attempt to get another handler routine and process the next
			// unit of work immediately.
			delay = 0
		}

		select {
		case <-w.clock.After(delay):
		case <-w.ctx.Done():
			break loop
		}
	}

	w.wg.Wait()
}

// Stop will cause the worker loop to exit after the current iteration. This is done by canceling the
// context passed to the database and the handler functions (which may cause the currently processing
// unit of work to fail). This method blocks until all handler goroutines have exited.
func (w *Worker) Stop() {
	w.cancel()
	<-w.finished
}

// dequeueAndHandle selects a queued record to process. This method returns false if no such record
// can be dequeued and returns an error only on failure to dequeue a new record - no handler errors
// will bubble up.
func (w *Worker) dequeueAndHandle() (dequeued bool, err error) {
	select {
	// If we block here we are waiting for a handler to exit so that we
	// do not exceed our configured concurrency limit.
	case <-w.handlerSemaphore:
	case <-w.ctx.Done():
		return false, w.ctx.Err()
	}
	defer func() {
		if !dequeued {
			// Ensure that if we do not dequeue a record successfully we do not
			// leak from the semaphore. This will happen if the pre dequeue hook
			// fails, if the dequeue call fails, or if there are no records to
			// process.
			w.handlerSemaphore <- struct{}{}
		}
	}()

	dequeueable, conditions, err := w.preDequeueHook()
	if err != nil {
		return false, errors.Wrap(err, "Handler.PreDequeueHook")
	}
	if !dequeueable {
		// Hook declined to dequeue a record
		return false, nil
	}

	// Select a queued record to process and the transaction that holds it
	record, tx, dequeued, err := w.store.Dequeue(w.ctx, conditions)
	if err != nil {
		return false, errors.Wrap(err, "store.Dequeue")
	}
	if !dequeued {
		// Nothing to process
		return false, nil
	}

	log15.Info("Dequeued record for processing", "name", w.options.Name, "id", record.RecordID())

	if hook, ok := w.options.Handler.(HandlerWithHooks); ok {
		hook.PreHandle(w.ctx, record)
	}

	w.wg.Add(1)

	go func() {
		defer func() {
			if hook, ok := w.options.Handler.(HandlerWithHooks); ok {
				hook.PostHandle(w.ctx, record)
			}

			w.handlerSemaphore <- struct{}{}
			w.wg.Done()
		}()

		if err := w.handle(tx, record); err != nil {
			log15.Error("Failed to finalize record", "name", w.options.Name, "err", err)
		}
	}()

	return true, nil
}

// handle processes the given record locked by the given transaction. This method returns an
// error only if there is an issue committing the transaction - no handler errors will bubble
// up.
func (w *Worker) handle(tx Store, record Record) (err error) {
	// Enable tracing on the context and trace the remainder of the operation including the
	// transaction commit call in the following deferred function.
	ctx, endOperation := w.options.Metrics.HandleOperation.With(ot.WithShouldTrace(w.ctx, true), &err, observation.Args{})
	defer func() {
		endOperation(1, observation.Args{})
	}()

	defer func() {
		// Notice that we will commit the transaction even on error from the handler
		// function. We will only rollback the transaction if we fail to mark the job
		// as completed or errored.
		err = tx.Done(err)
	}()

	if handleErr := w.options.Handler.Handle(ctx, tx, record); handleErr != nil {
		if marked, markErr := tx.MarkErrored(ctx, record.RecordID(), handleErr.Error()); markErr != nil {
			return errors.Wrap(markErr, "store.MarkErrored")
		} else if marked {
			log15.Warn("Marked record as errored", "name", w.options.Name, "id", record.RecordID(), "err", handleErr)
		}
	} else {
		if marked, markErr := tx.MarkComplete(ctx, record.RecordID()); markErr != nil {
			return errors.Wrap(markErr, "store.MarkComplete")
		} else if marked {
			log15.Info("Marked record as complete", "name", w.options.Name, "id", record.RecordID())
		}
	}

	log15.Info("Handled record", "name", w.options.Name, "id", record.RecordID())
	return nil
}

// preDequeueHook invokes the handler's pre-dequeue hook if it exists.
func (w *Worker) preDequeueHook() (bool, []*sqlf.Query, error) {
	if o, ok := w.options.Handler.(HandlerWithPreDequeue); ok {
		return o.PreDequeue(w.ctx)
	}

	return true, nil, nil
}

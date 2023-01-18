package goroutine

import (
	"context"
	"time"

	"github.com/derision-test/glock"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/goroutine/recorder"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// PeriodicGoroutine represents a goroutine whose main behavior is reinvoked periodically.
//
// See
// https://docs.sourcegraph.com/dev/background-information/backgroundroutine
// for more information and a step-by-step guide on how to implement a
// PeriodicBackgroundRoutine.
type PeriodicGoroutine struct {
	name        string
	routineType recorder.RoutineType
	description string
	jobName     string
	recorder    *recorder.Recorder
	interval    time.Duration
	handler     unifiedHandler
	operation   *observation.Operation
	clock       glock.Clock
	ctx         context.Context    // root context passed to the handler
	cancel      context.CancelFunc // cancels the root context
	finished    chan struct{}      // signals that Start has finished
}

var _ recorder.Recordable = &PeriodicGoroutine{}

type unifiedHandler interface {
	Handler
	ErrorHandler
}

// Handler represents the main behavior of a PeriodicGoroutine.
type Handler interface {
	// Handle performs an action with the given context.
	Handle(ctx context.Context) error
}

// ErrorHandler is an optional extension of the Handler interface.
type ErrorHandler interface {
	// HandleError is called with error values returned from Handle. This will not
	// be called with error values due to a context cancellation during a graceful
	// shutdown.
	HandleError(err error)
}

// Finalizer is an optional extension of the Handler interface.
type Finalizer interface {
	// OnShutdown is called after the last call to Handle during a graceful shutdown.
	OnShutdown()
}

// HandlerFunc wraps a function, so it can be used as a Handler.
type HandlerFunc func(ctx context.Context) error

func (f HandlerFunc) Handle(ctx context.Context) error {
	return f(ctx)
}

type simpleHandler struct {
	name  string
	scope log.Logger
	Handler
}

var (
	_ unifiedHandler = (*simpleHandler)(nil)
	_ Finalizer      = (*simpleHandler)(nil)
)

func (h *simpleHandler) Handle(ctx context.Context) error {
	return h.Handler.Handle(ctx)
}

func (h *simpleHandler) HandleError(err error) {
	h.scope.Error("An error occurred in a background task", log.String("handler", h.name), log.Error(err))
}

func (h *simpleHandler) OnShutdown() {
	if finalizer, ok := h.Handler.(Finalizer); ok {
		finalizer.OnShutdown()
	}
}

// NewPeriodicGoroutine creates a new PeriodicGoroutine with the given handler. The context provided will propagate into
// the executing goroutine and will terminate the goroutine if cancelled.
func NewPeriodicGoroutine(ctx context.Context, name, description string, interval time.Duration, handler Handler) *PeriodicGoroutine {
	return NewPeriodicGoroutineWithMetrics(ctx, name, description, interval, handler, nil)
}

// NewPeriodicGoroutineWithMetrics creates a new PeriodicGoroutine with the given handler. The context provided will propagate into
// the executing goroutine and will terminate the goroutine if cancelled.
func NewPeriodicGoroutineWithMetrics(ctx context.Context, name, description string, interval time.Duration, handler Handler, operation *observation.Operation) *PeriodicGoroutine {
	return newPeriodicGoroutine(ctx, name, description, interval, handler, operation, glock.NewRealClock())
}

func newPeriodicGoroutine(ctx context.Context, name, description string, interval time.Duration, handler Handler, operation *observation.Operation, clock glock.Clock) *PeriodicGoroutine {
	ctx, cancel := context.WithCancel(ctx)

	var h unifiedHandler
	if uh, ok := handler.(unifiedHandler); ok {
		h = uh
	} else {
		h = &simpleHandler{
			name:    name,
			scope:   log.Scoped(name, description),
			Handler: handler,
		}
	}

	return &PeriodicGoroutine{
		name:        name,
		description: description,
		handler:     h,
		interval:    interval,
		operation:   operation,
		clock:       clock,
		ctx:         ctx,
		cancel:      cancel,
		finished:    make(chan struct{}),
	}
}

// Start begins the process of calling the registered handler in a loop. This process will
// wait the interval supplied at construction between invocations.
func (r *PeriodicGoroutine) Start() {
	if r.recorder != nil {
		go r.recorder.LogStart(r)
	}
	defer close(r.finished)

loop:
	for {
		start := time.Now()
		shutdown, err := runPeriodicHandler(r.ctx, r.handler, r.operation)
		duration := time.Since(start)
		if r.recorder != nil {
			go r.recorder.LogRun(r, duration, err)
		}

		if shutdown {
			break
		} else if h, ok := r.handler.(ErrorHandler); ok && err != nil {
			h.HandleError(err)
		}

		select {
		case <-r.clock.After(r.interval):
		case <-r.ctx.Done():
			break loop
		}
	}

	if h, ok := r.handler.(Finalizer); ok {
		h.OnShutdown()
	}
}

// Stop will cancel the context passed to the handler function to stop the current
// iteration of work, then break the loop in the Start method so that no new work
// is accepted. This method blocks until Start has returned.
func (r *PeriodicGoroutine) Stop() {
	if r.recorder != nil {
		go r.recorder.LogStop(r)
	}
	r.cancel()
	<-r.finished
}

func (r *PeriodicGoroutine) Name() string {
	return r.name
}

func (r *PeriodicGoroutine) Type() recorder.RoutineType {
	if r.operation != nil {
		return recorder.PeriodicWithMetrics
	} else {
		return recorder.PeriodicRoutine
	}
}

func (r *PeriodicGoroutine) Description() string {
	return r.description
}

func (r *PeriodicGoroutine) Interval() time.Duration {
	return r.interval
}

func (r *PeriodicGoroutine) JobName() string {
	return r.jobName
}

func (r *PeriodicGoroutine) SetJobName(jobName string) {
	r.jobName = jobName
}

func (r *PeriodicGoroutine) RegisterRecorder(recorder *recorder.Recorder) {
	r.recorder = recorder
}

func runPeriodicHandler(ctx context.Context, handler Handler, operation *observation.Operation) (_ bool, err error) {
	if operation != nil {
		tmpCtx, _, endObservation := operation.With(ctx, &err, observation.Args{})
		defer endObservation(1, observation.Args{})
		ctx = tmpCtx
	}

	err = handler.Handle(ctx)
	if err != nil {
		if ctx.Err() != nil && errors.Is(err, ctx.Err()) {
			// If the error is due to the loop being shut down, break
			// from the run loop in the calling function
			return true, nil
		}
	}

	return false, err
}

package group

import (
	"context"
	"sync"
)

// NewWithStreaming creates a new StreamGroup
func NewWithStreaming[T any]() StreamGroup[T] {
	return &streamGroup[T]{
		os: newOrderedStreamer[T](),
	}
}

// StreamGroup is a group that processes an ordered stream in parallel.
type StreamGroup[T any] interface {
	// Go starts a task in a goroutine then passes its result the provided callback.
	// This interface guarantees that the callbacks are called in the same order
	// that the tasks are submitted. Additionally, it guarantees that the submitted
	// callbacks will all be called from a single goroutine.
	Go(task func() T, callback func(T))

	// Wait blocks until all started goroutines have completed and all callbacks
	// have been called and have completed.
	Wait()

	// Configuration methods. See interface definitions for details.
	Errorable[ErrorStreamGroup[T]]
	Contextable[ContextStreamGroup[T]]
	Limitable[StreamGroup[T]]
}

// ErrorStreamGroup is a group that processes an ordered stream in parallel with
// tasks that might return an error.
type ErrorStreamGroup[T any] interface {
	// Go starts a task in a goroutine then passes its result the provided callback.
	// This interface guarantees that the callbacks are called in the same order
	// that the tasks are submitted. Additionally, it guarantees that the submitted
	// callbacks will all be called from a single goroutine.
	//
	// Note that, unlike Group and ResultGroup, the nil-ness of the error does not
	// change behavior.
	Go(task func() (T, error), callback func(T, error))

	// Wait blocks until all started goroutines have completed and all callbacks
	// have been called and have completed.
	Wait()

	// Configuration methods. See interface definitions for details.
	Contextable[ContextStreamGroup[T]]
	Limitable[ErrorStreamGroup[T]]
}

// ContextStreamGroup is a group that processes an ordered stream in parallel with
// tasks that require a context and might return an error.
type ContextStreamGroup[T any] interface {
	// Go starts a task in a goroutine then passes its result the provided callback.
	// This interface guarantees that the callbacks are called in the same order
	// that the tasks are submitted. Additionally, it guarantees that the submitted
	// callbacks will all be called from a single goroutine.
	//
	// Note that, unlike Group and ResultGroup, the nil-ness of the error does not
	// change behavior.
	Go(task func(context.Context) (T, error), callback func(context.Context, T, error))

	// Wait blocks until all started goroutines have completed and all callbacks
	// have been called and have completed.
	Wait()

	// Configuration methods. See interface definitions for details.
	Limitable[ContextStreamGroup[T]]
}

// streamGroup is a group that allows streaming task results with a callback
type streamGroup[T any] struct {
	os *orderedStreamer[T]
}

func (g *streamGroup[T]) Go(task func() T, callback func(T)) {
	// acquire will not error unless the context is canceled
	_, release, _ := g.os.acquire(context.Background())
	taskWithRelease := func() T {
		defer release()
		return task()
	}
	g.os.start(funcPair[T]{taskWithRelease, callback})
}

func (g *streamGroup[T]) Wait() {
	g.os.wait()
}

func (g *streamGroup[T]) WithErrors() ErrorStreamGroup[T] {
	esg := &errorStreamGroup[T]{
		os: newOrderedStreamer[resultAndError[T]](),
	}
	esg.os.group = g.os.group // copy the group with its settings
	return esg
}

func (g *streamGroup[T]) WithContext(ctx context.Context) ContextStreamGroup[T] {
	csg := &contextStreamGroup[T]{
		os:  newOrderedStreamer[resultAndError[T]](),
		ctx: ctx,
	}
	csg.os.group = g.os.group // copy the group with its settings
	return csg
}

func (g *streamGroup[T]) WithMaxConcurrency(limit int) StreamGroup[T] {
	g.os.group.limiter = NewBasicLimiter(limit)
	return g
}

func (g *streamGroup[T]) WithConcurrencyLimiter(limiter Limiter) StreamGroup[T] {
	g.os.group.limiter = limiter
	return g
}

// errorStreamGroup is a stream group for tasks that return a result and an error.
type errorStreamGroup[T any] struct {
	os *orderedStreamer[resultAndError[T]]
}

// a utility any+error tuple
type resultAndError[T any] struct {
	res T
	err error
}

func (g *errorStreamGroup[T]) Go(task func() (T, error), callback func(T, error)) {
	// acquire will not fail unless context is canceled
	_, release, _ := g.os.acquire(context.Background())

	// Create functions that merge the input/output return value and error
	// into a single type so we can use orderedStreamer.
	pairedTask := func() resultAndError[T] {
		defer release()
		res, err := task()
		return resultAndError[T]{res, err}
	}

	pairedCallback := func(pair resultAndError[T]) {
		callback(pair.res, pair.err)
	}

	g.os.start(funcPair[resultAndError[T]]{pairedTask, pairedCallback})
}

func (g *errorStreamGroup[T]) Wait() {
	g.os.wait()
}

func (g *errorStreamGroup[T]) WithContext(ctx context.Context) ContextStreamGroup[T] {
	return &contextStreamGroup[T]{
		os:  g.os,
		ctx: ctx,
	}
}

func (g *errorStreamGroup[T]) WithMaxConcurrency(limit int) ErrorStreamGroup[T] {
	g.os.group.limiter = NewBasicLimiter(limit)
	return g
}

func (g *errorStreamGroup[T]) WithConcurrencyLimiter(limiter Limiter) ErrorStreamGroup[T] {
	g.os.group.limiter = limiter
	return g
}

// errorStreamGroup is a stream group for tasks that require a context and return a result and an error.
type contextStreamGroup[T any] struct {
	os  *orderedStreamer[resultAndError[T]]
	ctx context.Context
}

func (g *contextStreamGroup[T]) Go(task func(context.Context) (T, error), callback func(context.Context, T, error)) {
	ctx, release, err := g.os.acquire(g.ctx)

	// Create functions that merge the input/output return value and error
	// into a single type so we can use orderedStreamer.
	pairedTask := func() resultAndError[T] {
		defer release()

		// If acquiring failed, return its error immediately without running the task.
		// We do this here so it can plug into the error handling code.
		if err != nil {
			var t T
			return resultAndError[T]{t, err}
		}

		res, err := task(ctx)
		return resultAndError[T]{res, err}
	}

	pairedCallback := func(pair resultAndError[T]) {
		callback(ctx, pair.res, pair.err)
	}

	g.os.start(funcPair[resultAndError[T]]{pairedTask, pairedCallback})
}

func (g *contextStreamGroup[T]) Wait() {
	g.os.wait()
}

func (g *contextStreamGroup[T]) WithMaxConcurrency(limit int) ContextStreamGroup[T] {
	g.os.group.limiter = NewBasicLimiter(limit)
	return g
}

func (g *contextStreamGroup[T]) WithConcurrencyLimiter(limiter Limiter) ContextStreamGroup[T] {
	g.os.group.limiter = limiter
	return g
}

func newOrderedStreamer[T any]() *orderedStreamer[T] {
	return &orderedStreamer[T]{
		group: &group{},
		// Set reasonably high default limit on the output channel by default.
		// This doesn't limit the max goroutines, it just limits the number of
		// goroutines waiting for their results to be handled.
		resChans: make(chan chan resultAndCallback[T], 32),
	}
}

type orderedStreamer[T any] struct {
	group *group

	// A queue of channels that will receive the results of running tasks
	resChans chan chan resultAndCallback[T]

	handlerOnce sync.Once
	handlerWg   sync.WaitGroup
}

func (o *orderedStreamer[T]) acquire(ctx context.Context) (context.Context, context.CancelFunc, error) {
	return o.group.acquire(ctx)
}

// A utility type that represents a completed task and
// a callback that will be called with the task's result.
type resultAndCallback[T any] struct {
	res      T
	callback func(T)
}

// A utility type that represents a pair of functions where the
// return type of the first is the argument type of the second.
type funcPair[T any] struct {
	task     func() T
	callback func(T)
}

func (o *orderedStreamer[T]) start(funcs funcPair[T]) {
	o.initOnce()

	// Create a channel that we will receive the return value of
	// the task. Send this channel to resChans so that the callbacks
	// happen in the same order that tasks were queued.
	resChan := make(chan resultAndCallback[T], 1)
	o.resChans <- resChan

	// Start the task, and send its result and its callback to the handler
	o.group.start(func() {
		resChan <- resultAndCallback[T]{funcs.task(), funcs.callback}
	})
}

func (o *orderedStreamer[T]) initOnce() {
	// start the callback handler
	o.handlerOnce.Do(func() {
		o.handlerWg.Add(1)
		go func() {
			defer o.handlerWg.Done()

			// For each task, wait for it to complete and call its callback
			// with the return value.
			for resChan := range o.resChans {
				event := <-resChan
				event.callback(event.res)
			}
		}()
	})
}

func (g *orderedStreamer[T]) wait() {
	close(g.resChans)
	g.group.Wait()
	g.handlerWg.Wait()
}

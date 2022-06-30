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

type streamGroup[T any] struct {
	os orderedStreamer[T]
}

func (g *streamGroup[T]) Go(task func() T, callback func(T)) {
	// acquire will not error unless the context is canceled
	_, release, _ := g.os.acquire(context.Background())
	g.os.start(funcPair[T]{task, callback}, release)
}

func (g *streamGroup[T]) Wait() {
	g.os.wait()
}

func (g *streamGroup[T]) WithErrors() ErrorStreamGroup[T] {
	esg := &errorStreamGroup[T]{
		os: newOrderedStreamer[resultPair[T]](),
	}
	esg.os.group = g.os.group
	return esg
}

func (g *streamGroup[T]) WithContext(ctx context.Context) ContextStreamGroup[T] {
	csg := &contextStreamGroup[T]{
		os:  newOrderedStreamer[resultPair[T]](),
		ctx: ctx,
	}
	csg.os.group = g.os.group
	return csg
}

func (g *streamGroup[T]) WithLimit(limit int) StreamGroup[T] {
	g.os.group.limiter = NewBasicLimiter(limit)
	return g
}

func (g *streamGroup[T]) WithLimiter(limiter Limiter) StreamGroup[T] {
	g.os.group.limiter = limiter
	return g
}

type errorStreamGroup[T any] struct {
	os orderedStreamer[resultPair[T]]
}

type resultPair[T any] struct {
	res T
	err error
}

func (g *errorStreamGroup[T]) Go(task func() (T, error), callback func(T, error)) {
	pairedTask := func() resultPair[T] {
		res, err := task()
		return resultPair[T]{res, err}
	}

	pairedCallback := func(pair resultPair[T]) {
		callback(pair.res, pair.err)
	}

	_, release, _ := g.os.acquire(context.Background())
	g.os.start(funcPair[resultPair[T]]{pairedTask, pairedCallback}, release)
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

func (g *errorStreamGroup[T]) WithLimit(limit int) ErrorStreamGroup[T] {
	g.os.group.limiter = NewBasicLimiter(limit)
	return g
}

func (g *errorStreamGroup[T]) WithLimiter(limiter Limiter) ErrorStreamGroup[T] {
	g.os.group.limiter = limiter
	return g
}

type contextStreamGroup[T any] struct {
	os  orderedStreamer[resultPair[T]]
	ctx context.Context
}

func (g *contextStreamGroup[T]) Go(task func(context.Context) (T, error), callback func(context.Context, T, error)) {
	ctx, release, err := g.os.acquire(g.ctx)
	pairedTask := func() resultPair[T] {
		// If acquiring failed, return its error immediately without running the task.
		if err != nil {
			var t T
			return resultPair[T]{t, err}
		}
		res, err := task(ctx)
		return resultPair[T]{res, err}
	}

	pairedCallback := func(pair resultPair[T]) {
		callback(ctx, pair.res, pair.err)
	}

	g.os.start(funcPair[resultPair[T]]{pairedTask, pairedCallback}, release)
}

func (g *contextStreamGroup[T]) Wait() {
	g.os.wait()
}

func (g *contextStreamGroup[T]) WithLimit(limit int) ContextStreamGroup[T] {
	g.os.group.limiter = NewBasicLimiter(limit)
	return g
}

func (g *contextStreamGroup[T]) WithLimiter(limiter Limiter) ContextStreamGroup[T] {
	g.os.group.limiter = limiter
	return g
}

func newOrderedStreamer[T any]() orderedStreamer[T] {
	return orderedStreamer[T]{
		group: &group{},
		// Set reasonably high default limit on the output channel by default.
		// This doesn't limit the max goroutines, it just limits the number of
		// goroutines waiting for their results to be handled.
		resChans:    make(chan chan streamEvent[T], 32),
		handlerDone: make(chan struct{}),
	}
}

type orderedStreamer[T any] struct {
	group    *group
	resChans chan chan streamEvent[T]

	handlerOnce sync.Once
	handlerDone chan struct{}
}

// A utility type that represents a completed task and
// a callback that will be called with the task's result.
type streamEvent[T any] struct {
	res      T
	callback func(T)
}

// A utility type that represents a pair of functions where the
// return type of the first is the argument type of the second.
type funcPair[T any] struct {
	task     func() T
	callback func(T)
}

func (o *orderedStreamer[T]) acquire(ctx context.Context) (context.Context, context.CancelFunc, error) {
	return o.group.acquire(ctx)
}

func (o *orderedStreamer[T]) start(funcs funcPair[T], release func()) {
	o.initOnce()

	resChan := make(chan streamEvent[T], 1)
	o.resChans <- resChan

	o.group.start(func() {
		resChan <- streamEvent[T]{funcs.task(), funcs.callback}
		release()
	})
}

func (o *orderedStreamer[T]) initOnce() {
	// start the callback handler
	o.handlerOnce.Do(func() {
		go func() {
			defer close(o.handlerDone)

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
	<-g.handlerDone
}

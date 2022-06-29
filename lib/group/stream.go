package group

import (
	"sync"
)

func NewWithStreaming[T any]() StreamGroup[T] {
	return &streamGroup[T]{
		os: newOrderedStreamer[T](),
	}
}

type StreamGroup[T any] interface {
	Go(first func() T, then func(T))
	Wait()

	Errorable[ErrorStreamGroup[T]]
	Limitable[StreamGroup[T]]
}

type ErrorStreamGroup[T any] interface {
	Go(first func() (T, error), then func(T, error))
	Wait()

	Limitable[ErrorStreamGroup[T]]
}

type streamGroup[T any] struct {
	os orderedStreamer[T]
}

func (g *streamGroup[T]) Go(first func() T, then func(T)) {
	g.os.submit(funcPair[T]{first, then})
}

func (g *streamGroup[T]) Wait() {
	g.os.wait()
}

func (g *streamGroup[T]) WithErrors() ErrorStreamGroup[T] {
	return &errorStreamGroup[T]{
		os: newOrderedStreamer[resultPair[T]](),
	}
}

func (g *streamGroup[T]) WithLimit(limit int) StreamGroup[T] {
	g.os.group = g.os.group.WithLimit(limit)
	return g
}

func (g *streamGroup[T]) WithLimiter(limiter Limiter) StreamGroup[T] {
	g.os.group = g.os.group.WithLimiter(limiter)
	return g
}

type errorStreamGroup[T any] struct {
	os orderedStreamer[resultPair[T]]
}

type resultPair[T any] struct {
	res T
	err error
}

func (g *errorStreamGroup[T]) Go(first func() (T, error), then func(T, error)) {
	pairedFirst := func() resultPair[T] {
		res, err := first()
		return resultPair[T]{res, err}
	}

	pairedThen := func(pair resultPair[T]) {
		then(pair.res, pair.err)
	}

	g.os.submit(funcPair[resultPair[T]]{pairedFirst, pairedThen})
}

func (g *errorStreamGroup[T]) Wait() {
	g.os.wait()
}

func (g *errorStreamGroup[T]) WithLimit(limit int) ErrorStreamGroup[T] {
	g.os.group = g.os.group.WithLimit(limit)
	return g
}

func (g *errorStreamGroup[T]) WithLimiter(limiter Limiter) ErrorStreamGroup[T] {
	g.os.group = g.os.group.WithLimiter(limiter)
	return g
}

func newOrderedStreamer[T any]() orderedStreamer[T] {
	return orderedStreamer[T]{
		group: New(),
		// Set reasonably high default limit on the output channel by default.
		// This doesn't limit the max goroutines, it just limits the number of
		// goroutines waiting for their results to be handled.
		resChans:    make(chan chan streamEvent[T], 32),
		handlerDone: make(chan struct{}),
	}
}

type orderedStreamer[T any] struct {
	group    Group
	resChans chan chan streamEvent[T]

	handlerOnce sync.Once
	handlerDone chan struct{}
}

type streamEvent[T any] struct {
	res  T
	then func(T)
}

type funcPair[T any] struct {
	first func() T
	then  func(T)
}

func (o *orderedStreamer[T]) submit(funcs funcPair[T]) {
	o.initOnce()

	resChan := make(chan streamEvent[T], 1)
	o.resChans <- resChan

	o.group.Go(func() {
		resChan <- streamEvent[T]{funcs.first(), funcs.then}
	})
}

func (o *orderedStreamer[T]) initOnce() {
	o.handlerOnce.Do(func() {
		go func() {
			defer close(o.handlerDone)

			for resChan := range o.resChans {
				event := <-resChan
				event.then(event.res)
			}
		}()
	})
}

func (g *orderedStreamer[T]) wait() {
	close(g.resChans)
	g.group.Wait()
	<-g.handlerDone
}

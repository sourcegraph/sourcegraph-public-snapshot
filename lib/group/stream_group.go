package group

import (
	"sync"
)

func NewWithStreaming[T any]() StreamGroup[T] {
	return &streamGroup[T]{}
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
	orderedStreamer[T]
}

func (g *streamGroup[T]) Go(first func() T, then func(T)) {
	g.orderedStreamer.submit(funcPair[T]{first, then})
}

func (g *streamGroup[T]) WithErrors() ErrorStreamGroup[T] {
	return &errorStreamGroup[T]{}
}

func (g *streamGroup[T]) WithLimit(limit int) StreamGroup[T] {
	g.Group = g.Group.WithLimit(limit)
	g.orderedStreamer.resChans = make(chan chan streamEvent[T])
	return g
}

func (g *streamGroup[T]) WithLimiter(limiter Limiter) StreamGroup[T] {
	g.Group = g.Group.WithLimiter(limiter)
	return g
}

type errorStreamGroup[T any] struct {
	orderedStreamer[resultPair[T]]
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

	g.orderedStreamer.submit(funcPair[resultPair[T]]{pairedFirst, pairedThen})
}

func (g *errorStreamGroup[T]) WithLimit(limit int) ErrorStreamGroup[T] {
	g.Group = g.Group.WithLimit(limit)
	g.orderedStreamer.resChans = make(chan chan streamEvent[resultPair[T]])
	return g
}

func (g *errorStreamGroup[T]) WithLimiter(limiter Limiter) ErrorStreamGroup[T] {
	g.Group = g.Group.WithLimiter(limiter)
	return g
}

type orderedStreamer[T any] struct {
	Group
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
	o.init()

	resChan := make(chan streamEvent[T], 1)
	o.resChans <- resChan

	o.Group.Go(func() {
		resChan <- streamEvent[T]{funcs.first(), funcs.then}
	})
}

func (o *orderedStreamer[T]) init() {
	o.handlerOnce.Do(func() {
		if o.resChans == nil { // may be set by limit
			// Set reasonably high limit on the output channel by default.
			// Since callbacks shoudn't take as much time as the parallelized
			// funcs, this should be fine.
			o.resChans = make(chan chan streamEvent[T], 32)
		}
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
	g.Group.Wait()
	<-g.handlerDone
}

package group

import (
	"context"
	"sync"
)

func NewWithResults[T any]() ResultGroup[T] {
	return &resultGroup[T]{
		Group: New(),
	}
}

type ResultGroup[T any] interface {
	Go(func() T)
	Wait() []T

	Contextable[ResultContextErrorGroup[T]]
	Errorable[ResultErrorGroup[T]]
	Limitable[ResultGroup[T]]
}

type ResultErrorGroup[T any] interface {
	Go(func() (T, error))
	Wait() ([]T, error)

	Contextable[ResultContextErrorGroup[T]]
	Limitable[ResultErrorGroup[T]]
}

type ResultContextErrorGroup[T any] interface {
	Go(func(context.Context) (T, error))
	Wait() ([]T, error)

	Limitable[ResultContextErrorGroup[T]]
}

type resultAggregator[T any] struct {
	mu      sync.Mutex
	results []T
}

func (r *resultAggregator[T]) add(res T) {
	r.mu.Lock()
	r.results = append(r.results, res)
	r.mu.Unlock()
}

type resultGroup[T any] struct {
	Group
	resultAggregator[T]
}

func (g *resultGroup[T]) Go(f func() T) {
	g.Group.Go(func() {
		res := f()
		g.add(res)
	})
}

func (g *resultGroup[T]) Wait() []T {
	g.Group.Wait()
	return g.results
}

func (g *resultGroup[T]) WithErrors() ResultErrorGroup[T] {
	return &resultErrorGroup[T]{
		ErrorGroup: g.Group.WithErrors(),
	}
}

func (g *resultGroup[T]) WithContext(ctx context.Context) ResultContextErrorGroup[T] {
	return &resultContextErrorGroup[T]{
		ContextErrorGroup: g.Group.WithContext(ctx),
	}
}

func (g *resultGroup[T]) WithLimit(limit int) ResultGroup[T] {
	g.Group = g.Group.WithLimit(limit)
	return g
}

func (g *resultGroup[T]) WithLimiter(limiter Limiter) ResultGroup[T] {
	g.Group = g.Group.WithLimiter(limiter)
	return g
}

type resultErrorGroup[T any] struct {
	ErrorGroup
	resultAggregator[T]
}

func (g *resultErrorGroup[T]) Go(f func() (T, error)) {
	g.ErrorGroup.Go(func() error {
		res, err := f()
		if err == nil {
			g.add(res)
		}
		return err
	})
}

func (g *resultErrorGroup[T]) Wait() ([]T, error) {
	err := g.ErrorGroup.Wait()
	return g.results, err
}

func (g *resultErrorGroup[T]) WithLimit(limit int) ResultErrorGroup[T] {
	g.ErrorGroup = g.ErrorGroup.WithLimit(limit)
	return g
}

func (g *resultErrorGroup[T]) WithLimiter(limiter Limiter) ResultErrorGroup[T] {
	g.ErrorGroup = g.ErrorGroup.WithLimiter(limiter)
	return g
}

func (g *resultErrorGroup[T]) WithContext(ctx context.Context) ResultContextErrorGroup[T] {
	return &resultContextErrorGroup[T]{
		ContextErrorGroup: g.ErrorGroup.WithContext(ctx),
	}
}

type resultContextErrorGroup[T any] struct {
	ContextErrorGroup
	resultAggregator[T]
}

func (g *resultContextErrorGroup[T]) Go(f func(context.Context) (T, error)) {
	g.ContextErrorGroup.Go(func(ctx context.Context) error {
		res, err := f(ctx)
		if err == nil {
			g.add(res)
		}
		return err
	})
}

func (g *resultContextErrorGroup[T]) Wait() ([]T, error) {
	err := g.ContextErrorGroup.Wait()
	return g.results, err
}

func (g *resultContextErrorGroup[T]) WithLimit(limit int) ResultContextErrorGroup[T] {
	g.ContextErrorGroup = g.ContextErrorGroup.WithLimit(limit)
	return g
}

func (g *resultContextErrorGroup[T]) WithLimiter(limiter Limiter) ResultContextErrorGroup[T] {
	g.ContextErrorGroup = g.ContextErrorGroup.WithLimiter(limiter)
	return g
}

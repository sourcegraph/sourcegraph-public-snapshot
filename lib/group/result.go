package group

import (
	"context"
	"sync"
)

// NewWithResults creates a new group that aggregates the return values
// of the functions passed to its `Go` method.
func NewWithResults[T any]() ResultGroup[T] {
	return &resultGroup[T]{
		group: New(),
	}
}

// ResultGroup is a group that runs tasks that return a value.
type ResultGroup[T any] interface {
	// Go starts a task in a goroutine and collects its result. It will
	// not return until the goroutine has been started.
	Go(func() T)

	// Wait blocks until all goroutines started with Go() have completed.
	// It returns the collection of return values from the started tasks. There
	// are no guarantees about the order of the slice.
	Wait() []T

	// Configuration methods. See interface definitions for details.
	Contextable[ResultContextGroup[T]]
	Errorable[ResultErrorGroup[T]]
	Limitable[ResultGroup[T]]
}

// ResultErrorGroup is a group that runs tasks that return a value and an error.
type ResultErrorGroup[T any] interface {
	// Go starts a task in a goroutine and collects its result. It will
	// not return until the goroutine has been started.
	Go(func() (T, error))

	// Wait blocks until all goroutines started with Go() have completed.
	// It returns the collection of return values from the started tasks. There
	// are no guarantees about the order of the slice. Additionally, it returns
	// a combined error composed of any non-nil errors returned from the tasks.
	Wait() ([]T, error)

	// WithCollectErrored configures the group to collect results even from
	// tasks that errored. By default, the return values from errored tasks are
	// dropped.
	WithCollectErrored() ResultErrorGroup[T]

	// WithFirstError will configure the group to only retain the first error,
	// ignoring any subsequent errors.
	WithFirstError() ResultErrorGroup[T]

	// Configuration methods. See interface definitions for details.
	Contextable[ResultContextGroup[T]]
	Limitable[ResultErrorGroup[T]]
}

// ResultErrorGroup is a group that runs tasks that require a context and
// return a value and an error.
type ResultContextGroup[T any] interface {
	// Go starts a task in a goroutine and collects its result. It will
	// not return until the goroutine has been started.
	Go(func(context.Context) (T, error))

	// Wait blocks until all goroutines started with Go() have completed.
	// It returns the collection of return values from the started tasks. There
	// are no guarantees about the order of the slice. Additionally, it returns
	// a combined error composed of any non-nil errors returned from the tasks.
	Wait() ([]T, error)

	// WithCollectErrored configures the group to collect results even from
	// tasks that errored. By default, the return values from errored tasks are
	// dropped.
	WithCollectErrored() ResultContextGroup[T]

	// WithCancelOnError will cancel the group's context whenever any of the
	// functions started with Go() return an error.
	WithCancelOnError() ResultContextGroup[T]

	// WithFirstError will configure the group to only retain the first error,
	// ignoring any subsequent errors.
	WithFirstError() ResultContextGroup[T]

	// Configuration methods. See interface definitions for details.
	Limitable[ResultContextGroup[T]]
}

// resultAggregator is a utility type that lets us safely append from multiple
// goroutines. The zero value is valid and ready to use.
type resultAggregator[T any] struct {
	mu      sync.Mutex
	results []T
}

func (r *resultAggregator[T]) add(res T) {
	r.mu.Lock()
	r.results = append(r.results, res)
	r.mu.Unlock()
}

// resultGroup wraps a Group and a resultAggregator to collect
// the return values of tasks run with the group.
type resultGroup[T any] struct {
	group Group
	agg   resultAggregator[T]
}

func (g *resultGroup[T]) Go(f func() T) {
	g.group.Go(func() {
		g.agg.add(f())
	})
}

func (g *resultGroup[T]) Wait() []T {
	g.group.Wait()
	return g.agg.results
}

func (g *resultGroup[T]) WithErrors() ResultErrorGroup[T] {
	return &resultErrorGroup[T]{
		errorGroup: g.group.WithErrors(),
	}
}

func (g *resultGroup[T]) WithContext(ctx context.Context) ResultContextGroup[T] {
	return &resultContextGroup[T]{
		contextGroup: g.group.WithContext(ctx),
	}
}

func (g *resultGroup[T]) WithMaxConcurrency(limit int) ResultGroup[T] {
	g.group = g.group.WithMaxConcurrency(limit)
	return g
}

func (g *resultGroup[T]) WithConcurrencyLimiter(limiter Limiter) ResultGroup[T] {
	g.group = g.group.WithConcurrencyLimiter(limiter)
	return g
}

// resultErrorGroup wraps an ErrorGroup and a resultAggregator to collect
// the results and errors of tasks run with the group.
type resultErrorGroup[T any] struct {
	errorGroup     ErrorGroup
	agg            resultAggregator[T]
	collectErrored bool
}

func (g *resultErrorGroup[T]) Go(f func() (T, error)) {
	g.errorGroup.Go(func() error {
		res, err := f()
		if err == nil || g.collectErrored {
			g.agg.add(res)
		}
		return err
	})
}

func (g *resultErrorGroup[T]) Wait() ([]T, error) {
	err := g.errorGroup.Wait()
	return g.agg.results, err
}

func (g *resultErrorGroup[T]) WithCollectErrored() ResultErrorGroup[T] {
	g.collectErrored = true
	return g
}

func (g *resultErrorGroup[T]) WithFirstError() ResultErrorGroup[T] {
	g.errorGroup = g.errorGroup.WithFirstError()
	return g
}

func (g *resultErrorGroup[T]) WithMaxConcurrency(limit int) ResultErrorGroup[T] {
	g.errorGroup = g.errorGroup.WithMaxConcurrency(limit)
	return g
}

func (g *resultErrorGroup[T]) WithConcurrencyLimiter(limiter Limiter) ResultErrorGroup[T] {
	g.errorGroup = g.errorGroup.WithConcurrencyLimiter(limiter)
	return g
}

func (g *resultErrorGroup[T]) WithContext(ctx context.Context) ResultContextGroup[T] {
	return &resultContextGroup[T]{
		contextGroup: g.errorGroup.WithContext(ctx),
	}
}

// resultContextGroup wraps a ContextGroup and a resultAggregator to collect
// the return values and errors of tasks that require a context.
type resultContextGroup[T any] struct {
	contextGroup   ContextGroup
	agg            resultAggregator[T]
	collectErrored bool
}

func (g *resultContextGroup[T]) Go(f func(context.Context) (T, error)) {
	g.contextGroup.Go(func(ctx context.Context) error {
		res, err := f(ctx)
		if err == nil || g.collectErrored {
			g.agg.add(res)
		}
		return err
	})
}

func (g *resultContextGroup[T]) Wait() ([]T, error) {
	err := g.contextGroup.Wait()
	return g.agg.results, err
}

func (g *resultContextGroup[T]) WithCollectErrored() ResultContextGroup[T] {
	g.collectErrored = true
	return g
}

func (g *resultContextGroup[T]) WithMaxConcurrency(limit int) ResultContextGroup[T] {
	g.contextGroup = g.contextGroup.WithMaxConcurrency(limit)
	return g
}

func (g *resultContextGroup[T]) WithConcurrencyLimiter(limiter Limiter) ResultContextGroup[T] {
	g.contextGroup = g.contextGroup.WithConcurrencyLimiter(limiter)
	return g
}

func (g *resultContextGroup[T]) WithCancelOnError() ResultContextGroup[T] {
	g.contextGroup = g.contextGroup.WithCancelOnError()
	return g
}

func (g *resultContextGroup[T]) WithFirstError() ResultContextGroup[T] {
	g.contextGroup = g.contextGroup.WithFirstError()
	return g
}

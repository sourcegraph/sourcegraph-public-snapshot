// Package group provides utilities for working with groups of goroutines.
// The types exported by the package make it easy to handle common patterns
// like limiting parallelism, collecting errors, and inheriting contexts.
package group

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// New creates a new goroutine group. It can be used directly,
// or it can be used a starting point to construct more specific
// group types (for more information, see the With* methods).
func New() Group {
	return &group{
		limiter: &unlimitedLimiter{},
	}
}

// Group is the most basic group type. It starts goroutines
// with Go(), and waits for them to finish with Wait().
type Group interface {
	// Go starts a background goroutine. It will not return
	// until the goroutine has started.
	Go(func())

	// Wait blocks until all goroutines started with Go() have completed
	Wait()

	// Configuration methods. See interface definitions for details.
	Contextable[ContextErrorGroup]
	Errorable[ErrorGroup]
	Limitable[Group]
}

// ErrorGroup is a group that handles functions that might return errors.
// Any non-nil errors will be collected and returned by the Wait() method.
type ErrorGroup interface {
	// Go starts a background goroutine, collecting any returned errors.
	// It will not return until the goroutine has started.
	Go(func() error)

	// Wait blocks until all goroutines started with Go() have completed,
	// returning a combined error with any non-nil errors returned from
	// the submitted functions.
	Wait() error

	// Configuration methods. See interface definitions for details.
	Contextable[ContextErrorGroup]
	Limitable[ErrorGroup]
}

type ContextErrorGroup interface {
	// Go starts a background goroutine, calling the provided function with the
	// group's context and collecting any returned errors. It will not return
	// until the goroutine has started.
	Go(func(context.Context) error)

	// Wait blocks until all goroutines started with Go() have completed,
	// returning a combined error with any non-nil errors returned from
	// the submitted functions.
	Wait() error

	// Configuration methods. See interfaces for details.
	Limitable[ContextErrorGroup]

	// WithCancelOnError will cancel the group's context whenever any of the
	// functions started with Go() return an error. All further errors will
	// be ignored.
	WithCancelOnError() ContextErrorGroup

	// WithFirstError will configure the group to only retain the first error,
	// ignoring any subsequent errors.
	WithFirstError() ContextErrorGroup
}

// Limitable is a group that can be configured to limit the number of live
// goroutines. By default, groups can start an unlimited number of concurrent
// goroutines. Its type parameter is the return type after limiting.
type Limitable[T any] interface {
	// WithLimit will set the maximum number concurrent goroutines running
	// as part of this group.
	WithLimit(int) T

	// WithLimiter will set the limiter of this group. This is useful if
	// you want to share a limiter between multiple groups.
	WithLimiter(Limiter) T
}

// Contextable is a group that can be configured to operate with a context.
// By default, groups are not context-aware. Its type parameter is the return
// type after configuring a group with a context.
type Contextable[T any] interface {
	// WithContext creates a new group with the given context. Note
	// that WithContext implies WithErrors because it is difficult to
	// use a context group correctly without error support.
	WithContext(context.Context) T
}

// Errorable is a group that can be configured to run functions that return
// errors. By default, groups are not error-aware. Its type parameter
// is the return type after configuring a group to use error-returning
// functions.
type Errorable[T any] interface {
	WithErrors() T
}

type group struct {
	wg      sync.WaitGroup
	limiter Limiter
}

func (g *group) Go(f func()) {
	// g.goCtx will never error if the context is not canceled
	_ = g.goCtx(context.Background(), f)
}

// goCtx starts a goroutine, using the provided context to
// wait for the limiter.
func (g *group) goCtx(ctx context.Context, f func()) error {
	_, release, err := g.limiter.Acquire(ctx)
	if err != nil {
		return errors.Wrap(err, "acquire limiter")
	}

	g.wg.Add(1)
	go func() {
		defer release()
		defer g.wg.Done()
		// TODO add panic handlers
		f()
	}()

	return nil
}

func (g *group) Wait() {
	g.wg.Wait()
}

func (g *group) WithLimit(limit int) Group {
	g.limiter = newBasicLimiter(limit)
	return g
}

func (g *group) WithLimiter(limiter Limiter) Group {
	g.limiter = limiter
	return g
}

func (g *group) WithErrors() ErrorGroup {
	return &errorGroup{group: g}
}

func (g *group) WithContext(ctx context.Context) ContextErrorGroup {
	return &contextErrorGroup{
		ctx: ctx,
		errorGroup: &errorGroup{
			group: g,
		},
	}
}

type errorGroup struct {
	*group

	onlyFirst bool // if true, only keep the first error

	mu   sync.Mutex
	errs error
}

func (g *errorGroup) Go(f func() error) {
	g.goCtx(context.Background(), f)
}

// goCtx starts the goroutine, capturing any errors from the limiter
// in the set of errors.
func (g *errorGroup) goCtx(ctx context.Context, f func() error) {
	err := g.group.goCtx(ctx, func() {
		g.addErr(f())
	})
	g.addErr(err)
}

func (g *errorGroup) addErr(err error) {
	if err != nil {
		g.mu.Lock()
		if g.onlyFirst {
			if g.errs != nil {
				g.errs = err
			}
		} else {
			g.errs = errors.Append(g.errs, err)
		}
		g.mu.Unlock()
	}
}

func (g *errorGroup) Wait() error {
	g.group.Wait()
	return g.errs
}

func (g *errorGroup) WithLimit(limit int) ErrorGroup {
	g.group.limiter = newBasicLimiter(limit)
	return g
}

func (g *errorGroup) WithLimiter(limiter Limiter) ErrorGroup {
	g.group.limiter = limiter
	return g
}

func (g *errorGroup) WithContext(ctx context.Context) ContextErrorGroup {
	return &contextErrorGroup{
		ctx:        ctx,
		errorGroup: g,
	}
}

func (g *errorGroup) WithFirstError() ErrorGroup {
	g.onlyFirst = true
	return g
}

type contextErrorGroup struct {
	*errorGroup

	ctx    context.Context
	cancel context.CancelFunc // nil unless WithCancelOnError
}

func (g *contextErrorGroup) Go(f func(context.Context) error) {
	g.errorGroup.goCtx(g.ctx, func() error {
		err := f(g.ctx)
		if err != nil && g.cancel != nil {
			g.cancel()
		}
		return err
	})
}

func (g *contextErrorGroup) WithCancelOnError() ContextErrorGroup {
	g.ctx, g.cancel = context.WithCancel(g.ctx)
	return g
}

func (g *contextErrorGroup) WithLimit(limit int) ContextErrorGroup {
	g.limiter = newBasicLimiter(limit)
	return g
}

func (g *contextErrorGroup) WithLimiter(limiter Limiter) ContextErrorGroup {
	g.limiter = limiter
	return g
}

func (g *contextErrorGroup) WithFirstError() ContextErrorGroup {
	g.onlyFirst = true
	return g
}

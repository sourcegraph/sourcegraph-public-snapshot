// Package group provides utilities for working with groups of goroutines.
// The types exported by the package make it easy to handle common patterns
// like limiting parallelism, collecting errors, and inheriting contexts.
package group

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// New creates a new goroutine group. It can be used directly,
// or it can be used a starting point to construct more specific
// group types (for more information, see the With* methods).
func New() Group {
	return &group{}
}

// Group is the most basic group type. It starts goroutines
// with Go(), and waits for them to finish with Wait().
type Group interface {
	// Go starts a background goroutine. It will not return
	// until the goroutine has been started.
	Go(func())

	// Wait blocks until all goroutines started with Go() have completed.
	Wait()

	// Configuration methods. See interface definitions for details.
	Contextable[ContextGroup]
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

	// WithFirstError will configure the group to only retain the first error,
	// ignoring any subsequent errors.
	WithFirstError() ErrorGroup

	// Configuration methods. See interface definitions for details.
	Contextable[ContextGroup]
	Limitable[ErrorGroup]
}

type ContextGroup interface {
	// Go starts a background goroutine, calling the provided function with the
	// group's context and collecting any returned errors. It will not return
	// until the goroutine has started.
	Go(func(context.Context) error)

	// Wait blocks until all goroutines started with Go() have completed,
	// returning a combined error with any non-nil errors returned from
	// the submitted functions.
	Wait() error

	// WithCancelOnError will cancel the group's context whenever any of the
	// functions started with Go() return an error. All further errors will
	// be ignored.
	WithCancelOnError() ContextGroup

	// WithFirstError will configure the group to only retain the first error,
	// ignoring any subsequent errors.
	WithFirstError() ContextGroup

	// Configuration methods. See interfaces for details.
	Limitable[ContextGroup]
}

// Limitable is a group that can be configured to limit the number of live
// goroutines. By default, groups can start an unlimited number of concurrent
// goroutines. Its type parameter is the return type after limiting.
type Limitable[T any] interface {
	// WithMaxConcurrency will set the maximum number concurrent goroutines running
	// as part of this group.
	WithMaxConcurrency(int) T

	// WithConcurrencyLimiter will set the limiter of this group. This is useful if
	// you want to share a limiter between multiple groups.
	WithConcurrencyLimiter(Limiter) T
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
	limiter Limiter // nil limiter means unlimited (default)

	recoverMux   sync.Mutex
	recoveredErr error
}

func (g *group) Go(f func()) {
	// acquire will never error if the context is not canceled
	_, release, _ := g.acquire(context.Background())
	g.start(func() {
		f()
		release()
	})
}

// acquire a slot from the limiter. Will only return an error if the context expires.
func (g *group) acquire(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if g.limiter == nil {
		// nil limiter means unlimited
		return ctx, func() {}, nil
	}
	ctx, release, err := g.limiter.Acquire(ctx)
	return ctx, release, errors.Wrap(err, "acquire limiter")
}

// start a goroutine with the given function
func (g *group) start(f func()) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		defer g.recoverPanic()

		f()
	}()
}

func (g *group) Wait() {
	g.wg.Wait()

	// Propagate panic from child goroutine
	if g.recoveredErr != nil {
		panic(g.recoveredErr)
	}
}

func (g *group) WithMaxConcurrency(limit int) Group {
	g.limiter = NewBasicLimiter(limit)
	return g
}

func (g *group) WithConcurrencyLimiter(limiter Limiter) Group {
	g.limiter = limiter
	return g
}

func (g *group) WithErrors() ErrorGroup {
	return &errorGroup{group: g}
}

func (g *group) WithContext(ctx context.Context) ContextGroup {
	return &contextGroup{
		ctx: ctx,
		errorGroup: &errorGroup{
			group: g,
		},
	}
}

func (g *group) recoverPanic() {
	if val := recover(); val != nil {
		g.recoverMux.Lock()
		defer g.recoverMux.Unlock()

		var err error
		if valErr, ok := val.(error); ok {
			err = valErr
		} else {
			err = errors.Errorf("%#v", val)
		}

		g.recoveredErr = errors.Wrapf(err, "recovered from panic in child goroutine with stacktrace:\n%s", debug.Stack())
	}
}

// errorGroup wraps a *group with error collection
type errorGroup struct {
	group *group

	onlyFirst bool // if true, only keep the first error

	mu   sync.Mutex
	errs error
}

func (g *errorGroup) Go(f func() error) {
	// acquire will not error unless the context has been canceled
	_, release, _ := g.acquire(context.Background())
	g.start(func() error {
		defer release()
		return f()
	})
}

func (g *errorGroup) acquire(ctx context.Context) (context.Context, context.CancelFunc, bool) {
	ctx, release, err := g.group.acquire(ctx)
	// Collect the error, then return whether an error occured
	// so callers know whether the goroutine was started
	g.addErr(err)
	return ctx, release, err == nil
}

// goCtx starts the goroutine, capturing any errors from the limiter
// in the set of errors.
func (g *errorGroup) start(f func() error) {
	g.group.start(func() {
		g.addErr(f())
	})
}

func (g *errorGroup) addErr(err error) {
	if err != nil {
		g.mu.Lock()
		if g.onlyFirst {
			if g.errs == nil {
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

func (g *errorGroup) WithMaxConcurrency(limit int) ErrorGroup {
	g.group.limiter = NewBasicLimiter(limit)
	return g
}

func (g *errorGroup) WithConcurrencyLimiter(limiter Limiter) ErrorGroup {
	g.group.limiter = limiter
	return g
}

func (g *errorGroup) WithContext(ctx context.Context) ContextGroup {
	return &contextGroup{
		ctx:        ctx,
		errorGroup: g,
	}
}

func (g *errorGroup) WithFirstError() ErrorGroup {
	g.onlyFirst = true
	return g
}

type contextGroup struct {
	errorGroup *errorGroup

	ctx    context.Context
	cancel context.CancelFunc // nil unless WithCancelOnError
}

func (g *contextGroup) Go(f func(context.Context) error) {
	ctx, release, ok := g.errorGroup.acquire(g.ctx)
	if !ok {
		// If acquire fails, this means the context was canceled,
		// so there is no reason to re-cancel here.
		return
	}
	g.errorGroup.start(func() error {
		defer release()

		err := f(ctx)
		if err != nil && g.cancel != nil {
			// Add the error directly because otherwise, canceling could cause
			// another goroutine to exit and return an error before this error
			// was added, which breaks the expectations of WithFirstError().
			g.errorGroup.addErr(err)
			g.cancel()
			return nil
		}
		return err
	})
}

func (g *contextGroup) Wait() error {
	return g.errorGroup.Wait()
}

func (g *contextGroup) WithCancelOnError() ContextGroup {
	g.ctx, g.cancel = context.WithCancel(g.ctx)
	return g
}

func (g *contextGroup) WithMaxConcurrency(limit int) ContextGroup {
	g.errorGroup.group.limiter = NewBasicLimiter(limit)
	return g
}

func (g *contextGroup) WithConcurrencyLimiter(limiter Limiter) ContextGroup {
	g.errorGroup.group.limiter = limiter
	return g
}

func (g *contextGroup) WithFirstError() ContextGroup {
	g.errorGroup.onlyFirst = true
	return g
}

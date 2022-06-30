// Package group provides utilities for working with groups of goroutines.
// The types exported by the package make it easy to handle common patterns
// like limiting parallelism, collecting errors, and inheriting contexts.
package group

import (
	"context"
	"fmt"
	"sync"

	"github.com/sourcegraph/log"

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

	// Wait blocks until all goroutines started with Go() have completed
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
	limiter Limiter // nil limiter means unlimited (default)
}

func (g *group) Go(f func()) {
	// acquire will never error if the context is not canceled
	_, release, _ := g.acquire(context.Background())
	g.start(f, release)
}

// acquire a slot from the limiter
func (g *group) acquire(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if g.limiter == nil {
		// nil limiter means unlimited
		return ctx, func() {}, nil
	}
	ctx, release, err := g.limiter.Acquire(ctx)
	return ctx, release, errors.Wrap(err, "acquire limiter")
}

// start a goroutine
func (g *group) start(f, release func()) {
	g.wg.Add(1)
	go func() {
		defer release()
		defer g.wg.Done()
		defer recoverPanic()

		f()
	}()
}

func (g *group) Wait() {
	g.wg.Wait()
}

func (g *group) WithLimit(limit int) Group {
	g.limiter = NewBasicLimiter(limit)
	return g
}

func (g *group) WithLimiter(limiter Limiter) Group {
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

type errorGroup struct {
	*group

	onlyFirst bool // if true, only keep the first error

	mu   sync.Mutex
	errs error
}

func (g *errorGroup) Go(f func() error) {
	_, release, _ := g.acquire(context.Background())
	g.start(f, release)
}

func (g *errorGroup) acquire(ctx context.Context) (context.Context, context.CancelFunc, bool) {
	ctx, release, err := g.group.acquire(ctx)
	g.addErr(err)
	return ctx, release, err == nil
}

// goCtx starts the goroutine, capturing any errors from the limiter
// in the set of errors.
func (g *errorGroup) start(f func() error, release func()) {
	g.group.start(func() {
		g.addErr(f())
	}, release)
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

func (g *errorGroup) WithLimit(limit int) ErrorGroup {
	g.limiter = NewBasicLimiter(limit)
	return g
}

func (g *errorGroup) WithLimiter(limiter Limiter) ErrorGroup {
	g.limiter = limiter
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
	*errorGroup

	ctx    context.Context
	cancel context.CancelFunc // nil unless WithCancelOnError
}

func (g *contextGroup) Go(f func(context.Context) error) {
	ctx, release, ok := g.errorGroup.acquire(g.ctx)
	if !ok {
		// acquire will only fail if the context is canceled
		return
	}
	g.errorGroup.start(func() error {
		err := f(ctx)
		if err != nil && g.cancel != nil {
			g.cancel()
		}
		return err
	}, release)
}

func (g *contextGroup) WithCancelOnError() ContextGroup {
	g.ctx, g.cancel = context.WithCancel(g.ctx)
	return g
}

func (g *contextGroup) WithLimit(limit int) ContextGroup {
	g.limiter = NewBasicLimiter(limit)
	return g
}

func (g *contextGroup) WithLimiter(limiter Limiter) ContextGroup {
	g.limiter = limiter
	return g
}

func (g *contextGroup) WithFirstError() ContextGroup {
	g.onlyFirst = true
	return g
}

func recoverPanic() {
	if val := recover(); val != nil {
		if err, ok := val.(error); ok {
			log.Scoped("internal", "group").Error("recovered from panic", log.Error(err))
		} else {
			log.Scoped("internal", "group").Error("recovered from panic", log.Error(errors.New(fmt.Sprintf("%#v", val))))
		}
	}
}

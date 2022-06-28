package group

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Group interface {
	Go(func())

	Contextable[ContextGroup]
	Errorable[ErrorGroup]
	Limitable[Group]
}

type ErrorGroup interface {
	Go(func() error)
	Wait() error

	Contextable[ContextErrorGroup]
	Limitable[ErrorGroup]
}

type ContextGroup interface {
	Go(func(context.Context))
	Wait()

	Errorable[ContextErrorGroup]
	Limitable[ContextGroup]
}

type ContextErrorGroup interface {
	Go(func(context.Context) error)
	Wait() error

	WithCancelOnError() ContextErrorGroup
	Limitable[ContextErrorGroup]
}

type Limitable[T any] interface {
	WithLimit(int) T
	WithLimiter(Limiter) T
}

type Contextable[T any] interface {
	WithContext(context.Context) T
}

type Errorable[T any] interface {
	WithErrors() T
}

type Limiter interface {
	Acquire(context.Context) (context.Context, context.CancelFunc, error)
}

type unlimitedLimiter struct{}

func (l *unlimitedLimiter) Acquire(ctx context.Context) (context.Context, context.CancelFunc, error) {
	return ctx, func() {}, nil
}

type basicLimiter chan struct{}

func (l basicLimiter) Acquire(ctx context.Context) (context.Context, context.CancelFunc, error) {
	select {
	case l <- struct{}{}:
		return ctx, func() { <-l }, nil
	case <-ctx.Done():
		return ctx, func() {}, ctx.Err()
	}
}

func New() Group {
	return &group{
		limiter: &unlimitedLimiter{},
	}
}

type group struct {
	wg      sync.WaitGroup
	limiter Limiter
}

func (g *group) Go(f func()) {
	// g.go will never error if the context is not canceled
	_ = g.go_(context.Background(), f)
}

func (g *group) go_(ctx context.Context, f func()) error {
	// this will block until available, but should never error
	_, release, err := g.limiter.Acquire(ctx)
	if err != nil {
		return err
	}

	g.wg.Add(1)
	go func() {
		// TODO add panic handlers
		f()
		g.wg.Done()
		release()
	}()

	return nil
}

func (g *group) Wait() {
	g.wg.Wait()
}

func (g *group) WithLimit(limit int) Group {
	g.limiter = make(basicLimiter, limit)
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
	return &contextGroup{ctx: ctx, group: g}
}

type contextGroup struct {
	ctx context.Context
	*group
}

func (g *contextGroup) Go(f func(context.Context)) {
	// ignore error, just cancel
	g.group.go_(g.ctx, func() {
		f(g.ctx)
	})
}

func (g *contextGroup) WithLimit(limit int) ContextGroup {
	g.group.limiter = make(basicLimiter, limit)
	return g
}

func (g *contextGroup) WithLimiter(limiter Limiter) ContextGroup {
	g.group.limiter = limiter
	return g
}

func (g *contextGroup) WithErrors() ContextErrorGroup {
	return &contextErrorGroup{
		group: g.group,
		ctx:   g.ctx,
	}
}

type errorGroup struct {
	*group

	mu   sync.Mutex
	errs error
}

func (g *errorGroup) Go(f func() error) {
	g.group.Go(func() {
		err := f()
		if err != nil {
			g.mu.Lock()
			g.errs = errors.Append(g.errs, err)
			g.mu.Unlock()
		}
	})
}

func (g *errorGroup) Wait() (err error) {
	g.group.Wait()
	g.mu.Lock()
	err = g.errs
	g.mu.Unlock()
	return err
}

func (g *errorGroup) WithLimit(limit int) ErrorGroup {
	g.group.limiter = make(basicLimiter, limit)
	return g
}

func (g *errorGroup) WithLimiter(limiter Limiter) ErrorGroup {
	g.group.limiter = limiter
	return g
}

func (g *errorGroup) WithContext(ctx context.Context) ContextErrorGroup {
	return &contextErrorGroup{
		group: g.group,
		ctx:   ctx,
	}
}

type contextErrorGroup struct {
	*group

	ctx    context.Context
	cancel context.CancelFunc

	mu   sync.Mutex
	errs error
}

func (g *contextErrorGroup) Go(f func(context.Context) error) {
	err := g.group.go_(g.ctx, func() {
		err := f(g.ctx)
		if err != nil {
			g.mu.Lock()
			g.errs = errors.Append(g.errs, err)
			if g.cancel != nil {
				g.cancel()
			}
			g.mu.Unlock()
		}
	})
	if err != nil {
		g.mu.Lock()
		g.errs = errors.Append(g.errs, errors.Wrap(err, "acquire limiter"))
		g.mu.Unlock()
	}
}

func (g *contextErrorGroup) Wait() (err error) {
	g.group.Wait()
	g.mu.Lock()
	err = g.errs
	g.mu.Unlock()
	return err
}

func (g *contextErrorGroup) WithCancelOnError() ContextErrorGroup {
	g.ctx, g.cancel = context.WithCancel(g.ctx)
	return g
}

func (g *contextErrorGroup) WithLimit(limit int) ContextErrorGroup {
	g.group.limiter = make(basicLimiter, limit)
	return g
}

func (g *contextErrorGroup) WithLimiter(limiter Limiter) ContextErrorGroup {
	g.group.limiter = limiter
	return g
}

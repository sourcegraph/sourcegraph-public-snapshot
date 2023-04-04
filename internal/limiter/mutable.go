package limiter

import (
	"container/list"
	"context"
)

// MutableLimiter is a semaphore which supports having its limit (capacity)
// adjusted. It integrates with context.Context to handle adjusting the limit
// down.
//
// Note: Each MutableLimiter has an associated goroutine managing the semaphore
// state. We do not expose a way to stop this goroutine, so ensure the number
// of Limiters created is bounded.
type MutableLimiter struct {
	adjustLimit chan int
	acquire     chan acquireRequest
	getLimit    chan struct{ cap, len int }
}

type acquireResponse struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type acquireRequest struct {
	ctx  context.Context
	resp chan<- acquireResponse
}

// NewMutable returns a new Limiter (Semaphore).
func NewMutable(limit int) *MutableLimiter {
	l := &MutableLimiter{
		adjustLimit: make(chan int),
		getLimit:    make(chan struct{ cap, len int }),
		acquire:     make(chan acquireRequest),
	}
	go l.do(limit)
	return l
}

// SetLimit adjusts the limit. If we currently have more than limit context
// acquired, then contexts are canceled until we are within limit. Contexts
// are canceled such that the older contexts are canceled.
func (l *MutableLimiter) SetLimit(limit int) {
	l.adjustLimit <- limit
}

// GetLimit reports the current state of the limiter, returning the
// capacity and length (maximum and currently-in-use).
func (l MutableLimiter) GetLimit() (cap, len int) {
	s := <-l.getLimit
	return s.cap, s.len
}

// Acquire tries to acquire a context. On success a child context of ctx is
// returned. The cancel function must be called to release the acquired
// context. Cancel will also cancel the child context and is safe to call more
// than once (idempotent).
//
// If ctx is Done before we can acquire, then the context error is returned.
func (l *MutableLimiter) Acquire(ctx context.Context) (context.Context, context.CancelFunc, error) {
	respC := make(chan acquireResponse)
	req := acquireRequest{
		ctx:  ctx,
		resp: respC,
	}

	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case l.acquire <- req:
	}

	// We managed to send our acquire request. We now _must_ read the response
	// or we will block Limiter.do
	resp := <-respC
	return resp.ctx, resp.cancel, nil
}

func (l *MutableLimiter) do(limit int) {
	cancelFuncs := list.New()
	release := make(chan *list.Element)
	hidden := make(chan acquireRequest)

	for {
		// Use our acquire channel if we are not at limit, otherwise use a
		// channel which is never written to (to avoid acquiring).
		acquire := l.acquire
		if cancelFuncs.Len() == limit {
			acquire = hidden
		}

		select {
		case limit = <-l.adjustLimit:
			// If we adjust the limit down we need to release until we are
			// within limit.
			for limit >= 0 && cancelFuncs.Len() > limit {
				el := cancelFuncs.Front()
				cancelFuncs.Remove(el)
				el.Value.(context.CancelFunc)()
			}

		case el := <-release:
			// We may get the same element more than once. This is fine since
			// Remove ensures el is still part of the list and
			// context.CancelFuncs are idempotent.
			cancelFuncs.Remove(el)
			el.Value.(context.CancelFunc)()

		case l.getLimit <- struct{ cap, len int }{cap: limit, len: cancelFuncs.Len()}:
			// nothing to do, this is just so GetLimit() works
		case req := <-acquire:
			ctx, cancel := context.WithCancel(req.ctx)
			el := cancelFuncs.PushBack(cancel)
			req.resp <- acquireResponse{
				ctx: ctx,
				cancel: func() {
					release <- el
				},
			}
		}
	}
}

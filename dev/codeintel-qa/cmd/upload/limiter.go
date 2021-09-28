package main

import "context"

// limiter implements a counting semaphore.
type limiter struct {
	ch chan struct{}
}

// newLimiter creates a new limiter with the given maximum concurrency.
func newLimiter(concurrency int) *limiter {
	ch := make(chan struct{}, concurrency)
	for i := 0; i < concurrency; i++ {
		ch <- struct{}{}
	}

	return &limiter{ch: ch}
}

// acquire blocks until it can acquire a value from the inner channel.
func (l *limiter) acquire(ctx context.Context) error {
	select {
	case <-l.ch:
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

// release adds a value back to the limiter, unblocking one waiter.
func (l *limiter) release() {
	l.ch <- struct{}{}
}

// close closes the underlying channel.
func (l *limiter) close() {
	close(l.ch)
}

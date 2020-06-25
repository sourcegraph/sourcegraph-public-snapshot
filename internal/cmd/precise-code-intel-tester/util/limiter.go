package util

import "context"

// Limiter implements a counting semaphore.
type Limiter struct {
	ch chan struct{}
}

// NewLimiter creates a new limiter with the given maximum concurrency.
func NewLimiter(concurrency int) *Limiter {
	ch := make(chan struct{}, concurrency)
	for i := 0; i < concurrency; i++ {
		ch <- struct{}{}
	}

	return &Limiter{ch: ch}
}

// Acquire blocks until it can acquire a value from the inner channel.
func (l *Limiter) Acquire(ctx context.Context) error {
	select {
	case <-l.ch:
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release adds a value back to the limiter, unblocking one waiter.
func (l *Limiter) Release() {
	l.ch <- struct{}{}
}

// Close closes the underlying channel.
func (l *Limiter) Close() {
	close(l.ch)
}

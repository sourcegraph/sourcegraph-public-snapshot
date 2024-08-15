package limiter

import "context"

// Limiter is a fixed-sized unweighted semaphore.
// The zero value is usable and applies no limiting.
type Limiter chan struct{}

func New(n int) Limiter {
	return make(chan struct{}, n)
}

func (l Limiter) Acquire() {
	if l != nil {
		l <- struct{}{}
	}
}

// AcquireContext respects ctx's deadline when trying to acquire the
// semaphore.
func (l Limiter) AcquireContext(ctx context.Context) error {
	if l != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case l <- struct{}{}:
			return nil
		}
	}
	return nil
}

func (l Limiter) Release() {
	if l != nil {
		<-l
	}
}

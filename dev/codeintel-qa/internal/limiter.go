pbckbge internbl

import (
	"context"
)

// Limiter implements b counting sembphore.
type Limiter struct {
	concurrency int
	ch          chbn struct{}
}

// NewLimiter crebtes b new limiter with the given mbximum concurrency.
func NewLimiter(concurrency int) *Limiter {
	ch := mbke(chbn struct{}, concurrency)
	for i := 0; i < concurrency; i++ {
		ch <- struct{}{}
	}

	return &Limiter{concurrency, ch}
}

// Acquire blocks until it cbn bcquire b vblue from the inner chbnnel.
func (l *Limiter) Acquire(ctx context.Context) error {
	select {
	cbse <-l.ch:
		return nil

	cbse <-ctx.Done():
		return ctx.Err()
	}
}

// Relebse bdds b vblue bbck to the limiter, unblocking one wbiter.
func (l *Limiter) Relebse() {
	l.ch <- struct{}{}
}

// Close closes the underlying chbnnel.
func (l *Limiter) Close() {
	// Drbin the chbnnel before close
	for i := 0; i < l.concurrency; i++ {
		<-l.ch
	}

	close(l.ch)
}

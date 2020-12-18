package apiclient

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type CircuitBreaker struct {
	base            float64
	unit            time.Duration
	max             time.Duration
	waiter          chan struct{}
	bumper          chan bool
	penaltyModifier uint32
	once            sync.Once
}

func makeCircuitBreaker(base float64, unit, max time.Duration) *CircuitBreaker {
	cb := &CircuitBreaker{
		base:   base,
		unit:   unit,
		max:    max,
		waiter: make(chan struct{}),
		bumper: make(chan bool),
	}

	go cb.produce()
	return cb
}

func (cb *CircuitBreaker) Stop() {
	cb.once.Do(func() {
		close(cb.bumper)
	})
}

func (cb *CircuitBreaker) Wait(ctx context.Context, privileged bool) error {
	if !privileged {
		return cb.waitUnprivileged(ctx)
	}

	return cb.waitPrivileged(ctx)
}

func (cb *CircuitBreaker) Bump(success bool) {
	cb.bumper <- success
}

func (cb *CircuitBreaker) produce() {
	defer close(cb.waiter)

	for {
		select {
		case cb.waiter <- struct{}{}:
		case success, ok := <-cb.bumper:
			if !ok {
				return
			}

			for !success {
				if success, ok = <-cb.bumper; !ok {
					return
				}

				atomic.AddUint32(&cb.penaltyModifier, 1)
			}

			atomic.StoreUint32(&cb.penaltyModifier, 0)
		}
	}
}

func (cb *CircuitBreaker) waitUnprivileged(ctx context.Context) error {
	select {
	case <-cb.waiter:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (cb *CircuitBreaker) waitPrivileged(ctx context.Context) error {
	if delay := cb.delay(); delay > 0 {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (cb *CircuitBreaker) delay() time.Duration {
	penaltyModifier := atomic.LoadUint32(&cb.penaltyModifier)
	if penaltyModifier == 0 {
		return 0
	}

	delay := time.Duration(math.Pow(cb.base, float64(penaltyModifier))) * cb.unit
	if delay < cb.max {
		return delay
	}

	return cb.max
}

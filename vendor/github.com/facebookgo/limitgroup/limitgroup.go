// Package limitgroup provides a sync.WaitGroup equivalent with a
// configurable upper bound. This is implemented by having Add() block as well
// if necessary.
package limitgroup

import (
	"sync"
)

var sentinel = struct{}{}

// LimitGroup provides a WaitGroup with a limited upper bound. Once the limit
// is hit, Add will block until sufficient deltas are returned.
type LimitGroup struct {
	sync.WaitGroup
	slots chan struct{}
}

// NewLimitGroup creates a new LimitGroup with the configured limit.
func NewLimitGroup(limit uint) *LimitGroup {
	if limit == 0 {
		panic("zero is not a valid limit")
	}
	slots := make(chan struct{}, limit)
	return &LimitGroup{slots: slots}
}

// Add adds delta, which may be negative. It will block if we have hit the
// limit, and will unblock as Done is called.
func (l *LimitGroup) Add(delta int) {
	if delta > cap(l.slots) {
		panic("delta greater than limit")
	}
	if delta == 0 {
		return
	}

	// If we're adding, we need to accquire slots, if we're subtracting, we need
	// to return slots.
	if delta > 0 {
		l.WaitGroup.Add(delta)
		for i := 0; i < delta; i++ {
			l.slots <- sentinel
		}
	} else {
		for i := 0; i > delta; i-- {
			select {
			case <-l.slots:
			default:
				panic("trying to return more slots than acquired")
			}
		}
		l.WaitGroup.Add(delta)
	}
}

// Done decrements the counter.
func (l *LimitGroup) Done() {
	select {
	case <-l.slots:
	default:
		panic("trying to return more slots than acquired")
	}
	l.WaitGroup.Done()
}

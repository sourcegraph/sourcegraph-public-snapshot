package glock

import (
	"sync"
	"time"
)

type Advanceable interface {
	Advance(duration time.Duration)
	BlockingAdvance(duration time.Duration)
	SetCurrent(now time.Time)
}

// advanceable is the base type for mock clocks and tickers. This struct is
// written to be used as as a mixin, where the containing struct can mutate
// its internals (assuming correct coordination is used).
//
// An "advanceable" struct has a current time set of subscribers that may
// change over time. The current time can be moved explicitly by the user.
type advanceable struct {
	now         time.Time
	subscribers []subscriber
	m           *sync.Mutex
	cond        *sync.Cond
}

type subscriber interface {
	// signal performs some behavior if the given current time is after a
	// deadline registered previously. This method should not block. If the
	// subscriber is still interested in being updated with the current
	// time, it should return true; the clock or timer instance will drop
	// a reference to this subscriber otherwise.
	signal(now time.Time) (requeue bool)
}

// newAdvanceableAt returns a new advanceable struct with the given current time.
func newAdvanceableAt(now time.Time) *advanceable {
	m := &sync.Mutex{}

	return &advanceable{
		now:  now,
		m:    m,
		cond: sync.NewCond(m),
	}
}

// Advance will advance the clock's internal time by the given duration.
func (a *advanceable) Advance(duration time.Duration) {
	a.m.Lock()
	a.setCurrent(a.now.Add(duration))
	a.m.Unlock()
}

// SetCurrent sets the clock's internal time to the given time.
func (a *advanceable) SetCurrent(now time.Time) {
	a.m.Lock()
	a.setCurrent(now)
	a.m.Unlock()
}

// setCurrent sets the new current time and invokes and filters the list of
// subscribers.
func (a *advanceable) setCurrent(now time.Time) {
	filtered := a.subscribers[:0]
	for _, e := range a.subscribers {
		if e.signal(now) {
			filtered = append(filtered, e)
		}
	}

	a.now = now
	a.subscribers = filtered
	a.cond.Broadcast()
}

// register marks a subscriber to be updated when the current time changes.
func (a *advanceable) register(subscriber subscriber) {
	a.subscribers = append(a.subscribers, subscriber)
	a.cond.Broadcast()
}

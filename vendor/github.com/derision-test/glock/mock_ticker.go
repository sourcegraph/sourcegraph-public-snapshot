package glock

import "time"

// Ticker is a wrapper around a time.Ticker, which allows interface access to the
// underlying channel (instead of bare access like the time.Ticker struct allows).
type Ticker interface {
	// Chan returns the underlying ticker channel.
	Chan() <-chan time.Time

	// Stop stops the ticker.
	Stop()
}

// MockTicker is an implementation of Ticker that can be moved forward in time
// in increments for testing code that relies on timeouts or other time-sensitive
// constructs.
type MockTicker struct {
	*advanceable
	duration time.Duration
	deadline time.Time
	ch       chan time.Time
	stopped  bool
}

var _ Ticker = &MockTicker{}
var _ Advanceable = &MockTicker{}

// NewTicker creates a new Ticker tied to the internal MockClock time that ticks
// at intervals similar to time.NewTicker().  It will also skip or drop ticks
// for slow readers similar to time.NewTicker() as well.
func (c *MockClock) NewTicker(duration time.Duration) Ticker {
	c.m.Lock()
	defer c.m.Unlock()

	c.tickerArgs = append(c.tickerArgs, duration)

	return newMockTickerAt(c.advanceable, duration)
}

// NewMockTicker creates a new MockTicker with the internal time set to time.Now().
func NewMockTicker(duration time.Duration) *MockTicker {
	return NewMockTickerAt(time.Now(), duration)
}

// NewMockTickerAt creates a new MockTicker with the internal time set to the given time.
func NewMockTickerAt(now time.Time, duration time.Duration) *MockTicker {
	return newMockTickerAt(newAdvanceableAt(now), duration)
}

func newMockTickerAt(advanceable *advanceable, duration time.Duration) *MockTicker {
	if duration == 0 {
		panic("duration cannot be 0")
	}

	t := &MockTicker{
		advanceable: advanceable,
		duration:    duration,
		deadline:    advanceable.now.Add(duration),
		ch:          make(chan time.Time),
	}

	go t.process()
	advanceable.register(t)
	return t
}

// Chan returns a channel which will receive the tickers's internal time at the
// interval given when creating the ticker.
func (t *MockTicker) Chan() <-chan time.Time {
	return t.ch
}

// Stop will stop the ticker from ticking.
func (t *MockTicker) Stop() {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	t.stopped = true
	t.cond.Broadcast()
}

// BlockingAdvance will bump the ticker's internal time by the given duration. If
// If the new internal time passes the next tick threshold, a signal will be sent.
// This method will not return until the signal is read by a consumer of the
// ticker.
func (t *MockTicker) BlockingAdvance(duration time.Duration) {
	t.m.Lock()
	defer t.m.Unlock()

	t.now = t.now.Add(duration)

	if !t.now.Before(t.deadline) {
		t.ch <- t.deadline
		t.deadline = t.deadline.Add(t.duration)
	}
}

func (t *MockTicker) process() {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	for !t.stopped {
		if !t.now.Before(t.deadline) {
			t.ch <- t.deadline

			for !t.now.Before(t.deadline) {
				t.deadline = t.deadline.Add(t.duration)
			}
		}

		t.cond.Wait()
	}
}

// signal conforms to the subscriber interface.
func (t *MockTicker) signal(now time.Time) (requeue bool) {
	return !t.stopped
}

package glock

import (
	"time"
)

var closedChan = make(chan time.Time)

func init() {
	close(closedChan)
}

// Clock is a wrapper around common functions in the time package. This interface
// is designed to allow easy mocking of time functions.
type Clock interface {
	// Now returns the current time.
	Now() time.Time

	// After returns a channel which receives the current time after
	// the given duration elapses.
	After(duration time.Duration) <-chan time.Time

	// Sleep blocks until the given duration elapses.
	Sleep(duration time.Duration)

	// Since returns the time elapsed since t.
	Since(t time.Time) time.Duration

	// Until returns the duration until t.
	Until(t time.Time) time.Duration

	// NewTicker will construct a ticker which will continually fire,
	// pausing for the given duration in between invocations.
	NewTicker(duration time.Duration) Ticker
}

// MockClock is an implementation of Clock that can be moved forward in time
// in increments for testing code that relies on timeouts or other time-sensitive
// constructs.
type MockClock struct {
	*advanceable
	afterArgs  []time.Duration
	tickerArgs []time.Duration
}

var _ Clock = &MockClock{}
var _ Advanceable = &MockClock{}

// NewMockClock creates a new MockClock with the internal time set to time.Now().
func NewMockClock() *MockClock {
	return NewMockClockAt(time.Now())
}

// NewMockClockAt creates a new MockClick with the internal time set to the given time.
func NewMockClockAt(now time.Time) *MockClock {
	return &MockClock{advanceable: newAdvanceableAt(now)}
}

// Now returns the clock's internal time.
func (c *MockClock) Now() time.Time {
	c.m.Lock()
	defer c.m.Unlock()

	return c.now
}

// After returns a channel that will be sent the clock's internal time once the
// clock's internal time is at or past the supplied duration.
func (c *MockClock) After(duration time.Duration) <-chan time.Time {
	c.m.Lock()
	defer c.m.Unlock()

	c.afterArgs = append(c.afterArgs, duration)

	if duration <= 0 {
		return closedChan
	}

	ch := make(chan time.Time, 1)
	deadline := c.now.Add(duration)
	c.register(&afterSubscriber{ch: ch, deadline: deadline})
	return ch
}

// BlockedOnAfter returns the number of calls to After that are blocked waiting for
// a call to Advance to trigger them.
func (c *MockClock) BlockedOnAfter() int {
	c.m.Lock()
	defer c.m.Unlock()
	return len(c.subscribers)
}

// Sleep will block until the clock's internal time is at or past the given duration.
func (c *MockClock) Sleep(duration time.Duration) {
	<-c.After(duration)
}

// Since returns the time elapsed since t.
func (c *MockClock) Since(t time.Time) time.Duration {
	return c.Now().Sub(t)
}

// Until returns the duration until t.
func (c *MockClock) Until(t time.Time) time.Duration {
	return t.Sub(c.Now())
}

// BlockingAdvance will call Advance but only after there is another goroutine
// with a reference to a new channel returned by the After method.
func (c *MockClock) BlockingAdvance(duration time.Duration) {
	c.m.Lock()
	defer c.m.Unlock()

	for len(c.subscribers) == 0 {
		c.cond.Wait()
	}

	c.setCurrent(c.now.Add(duration))
}

// GetAfterArgs returns the duration of each call to After in the
// same order as they were called. The list is cleared each time
// GetAfterArgs is called.
func (c *MockClock) GetAfterArgs() []time.Duration {
	c.m.Lock()
	defer c.m.Unlock()

	args := c.afterArgs
	c.afterArgs = c.afterArgs[:0]
	return args
}

// GetTickerArgs returns the duration of each call to create a new
// ticker in the same order as they were called. The list is cleared
// each time GetTickerArgs is called.
func (c *MockClock) GetTickerArgs() []time.Duration {
	c.m.Lock()
	defer c.m.Unlock()

	args := c.tickerArgs
	c.tickerArgs = c.tickerArgs[:0]
	return args
}

type afterSubscriber struct {
	ch       chan time.Time
	deadline time.Time
}

// signal conforms to the subscriber interface.
func (s *afterSubscriber) signal(now time.Time) (requeue bool) {
	if now.Before(s.deadline) {
		// not enough time elapsed
		// try again on the next signal
		return true
	}

	s.ch <- s.deadline // inform user
	return false       // unsubscribe
}

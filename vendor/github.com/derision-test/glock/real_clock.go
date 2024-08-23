package glock

import "time"

type realClock struct{}

var _ Clock = &realClock{}

// NewRealClock returns a Clock whose implementation falls back to the
// methods available in the time package.
func NewRealClock() Clock {
	return &realClock{}
}

func (c *realClock) Now() time.Time {
	return time.Now()
}

func (c *realClock) After(duration time.Duration) <-chan time.Time {
	return time.After(duration)
}

func (c *realClock) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

func (c *realClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

func (c *realClock) Until(t time.Time) time.Duration {
	return time.Until(t)
}

func (c *realClock) NewTicker(duration time.Duration) Ticker {
	return NewRealTicker(duration)
}

func NewRealTicker(duration time.Duration) Ticker {
	return &realTicker{
		ticker: time.NewTicker(duration),
	}
}

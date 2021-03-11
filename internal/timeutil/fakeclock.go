package timeutil

import (
	"time"
)

// FakeClock provides a controllable clock for use in tests.
type FakeClock struct {
	epoch time.Time
	step  time.Duration
	steps int
}

// NewFakeClock returns a FakeClock instance that starts telling time at the given epoch.
// Every invocation of Now adds step amount of time to the clock.
func NewFakeClock(epoch time.Time, step time.Duration) FakeClock {
	return FakeClock{epoch: epoch, step: step}
}

// Now returns the current fake time and advances the clock "step" amount of time.
func (c *FakeClock) Now() time.Time {
	c.steps++
	return c.Time(c.steps)
}

// Time tells the time at the given step from the provided epoch.
func (c FakeClock) Time(step int) time.Time {
	// We truncate to microsecond precision because Postgres' timestamptz type
	// doesn't handle nanoseconds.
	return c.epoch.Add(time.Duration(step) * c.step).UTC().Truncate(time.Microsecond)
}

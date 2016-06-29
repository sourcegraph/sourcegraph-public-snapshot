package clock

import "time"

// Clock is an interface that provides an abstraction over time.Now() to
// help you control it in certain contexts, such as testing.
type Clock interface {
	Now() time.Time
}

type prodClock struct{}

func (c prodClock) Now() time.Time {
	return time.Now()
}

// NewProductionClock creates a Clock that simply mirrors
// time.Now()'s original behavior, with no additional logic.
func NewProductionClock() Clock {
	return prodClock{}
}

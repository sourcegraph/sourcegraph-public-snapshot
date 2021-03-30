package window

import (
	"time"

	"go.uber.org/ratelimit"
)

type Schedule interface {
	Take() time.Time
	ValidUntil() time.Time

	total() int
}

type schedule struct {
	ratelimit.Limiter

	// until really needs to contain a monotonic time, which means that care
	// must be taken to construct the baseSchedule without a time zone in
	// production use. (Testing doesn't really matter.) time.Now() is OK.
	until time.Time

	// Fields we need to keep around for total calculation.

	duration time.Duration
	rate     rate
}

var _ Schedule = &schedule{}

func newSchedule(base time.Time, d time.Duration, rate rate) Schedule {
	var limiter ratelimit.Limiter
	if rate.IsUnlimited() {
		limiter = ratelimit.NewUnlimited()
	} else {
		limiter = ratelimit.New(rate.n, ratelimit.Per(rate.unit.AsDuration()))
	}

	return &schedule{
		duration: d,
		Limiter:  limiter,
		rate:     rate,
		until:    base.Add(d),
	}
}

func (s *schedule) ValidUntil() time.Time {
	return s.until
}

func (s *schedule) total() int {
	if s.rate.IsUnlimited() {
		return -1
	}
	return int((float64(s.rate.n) * float64(s.rate.unit)) * (float64(s.duration) / float64(s.rate.unit)))
}

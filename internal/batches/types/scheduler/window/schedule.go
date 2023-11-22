package window

import (
	"time"

	"go.uber.org/ratelimit"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ErrZeroSchedule indicates a Schedule that has a zero rate limit, and for
// which Take() will never succeed.
var ErrZeroSchedule = errors.New("schedule will never yield")

// Schedule represents a single Schedule in time: for a certain amount of time,
// this particular rate limit will be in enforced.
type Schedule struct {
	limiter ratelimit.Limiter

	// until really needs to contain a monotonic time, which means that care
	// must be taken to construct the schedule without a time zone in production
	// use. (Testing doesn't really matter.) time.Now() is OK.
	until time.Time

	// Fields we need to keep around for total calculation.
	duration time.Duration
	rate     rate
}

func newSchedule(base time.Time, d time.Duration, rate rate) *Schedule {
	var limiter ratelimit.Limiter
	if rate.IsUnlimited() {
		limiter = ratelimit.NewUnlimited()
	} else if rate.n > 0 {
		limiter = ratelimit.New(rate.n, ratelimit.Per(rate.unit.AsDuration()))
	}

	return &Schedule{
		duration: d,
		limiter:  limiter,
		rate:     rate,
		until:    base.Add(d),
	}
}

// Take blocks until a scheduling event can occur, and returns the time the
// event occurred.
func (s *Schedule) Take() (time.Time, error) {
	if s.limiter == nil {
		return time.Time{}, ErrZeroSchedule
	}
	return s.limiter.Take(), nil
}

// ValidUntil returns the time the schedule is valid until. After that time, a
// new Schedule must be created and used.
func (s *Schedule) ValidUntil() time.Time {
	return s.until
}

// total returns the total number of events the schedule expects to be able to
// handle while valid. If the schedule does not apply any rate limiting, then
// this will be -1.
func (s *Schedule) total() int {
	if s.limiter == nil {
		return 0
	}
	if s.rate.IsUnlimited() {
		return -1
	}

	// How many events would occur in an hour?
	//
	// We use an hour here because that's the maximum unit value a rate can
	// have, and therefore we can always calculate an exact integer value out of
	// this.
	perHour := s.rate.n * int(time.Hour/s.rate.unit.AsDuration())

	// What fraction of an hour is this schedule valid for?
	inAnHour := float64(s.duration) / float64(time.Hour)

	// Technically, this will truncate the floating point value, but since we're
	// only ever using this to estimate times for the user, this should be fine:
	// if it's plus or minus a single notch in the rate limit, nobody is likely
	// to notice, and our estimates can't be perfect anyway given code host rate
	// limits.
	return int(inAnHour * float64(perHour))
}

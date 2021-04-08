package window

import (
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"go.uber.org/ratelimit"
)

// Schedule represents a single schedule in time: for a certain amount of time,
// this particular window will be in force.
type Schedule interface {
	// Take blocks until a scheduling event can occur, and returns the time the
	// event occurred.
	Take() (time.Time, error)

	// ValidUntil returns the time the schedule is valid until. After that time,
	// a new Schedule must be created and used.
	ValidUntil() time.Time

	// total returns the total number of events the schedule expects to be able
	// to handle while valid. If the schedule does not apply any rate limiting,
	// then this will be -1.
	total() int
}

var ErrZeroSchedule = errors.New("schedule will never yield")

// schedule is the only concrete implementation of Schedule provided at present,
// and handles a single rate limiter for its duration.
type schedule struct {
	limiter ratelimit.Limiter

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
		log15.Info("unlimited power")
		limiter = ratelimit.NewUnlimited()
	} else if rate.n > 0 {
		log15.Info("creating new schedule", "n", rate.n, "per", rate.unit.AsDuration())
		limiter = ratelimit.New(rate.n, ratelimit.Per(rate.unit.AsDuration()))
	}

	return &schedule{
		duration: d,
		limiter:  limiter,
		rate:     rate,
		until:    base.Add(d),
	}
}

func (s *schedule) Take() (time.Time, error) {
	if s.limiter == nil {
		return time.Time{}, ErrZeroSchedule
	}
	return s.limiter.Take(), nil
}

func (s *schedule) ValidUntil() time.Time {
	return s.until
}

func (s *schedule) total() int {
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

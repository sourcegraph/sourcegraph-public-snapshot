package base

import (
	"strconv"
	"time"

	"github.com/resonancelabs/go-pub/base/imath"
)

// We use "Micros" to represent timestamps. The units are microseconds since
// the unix epoch.
//
// Note: Go's UnixNano() ensures the times are UTC rather than relative to the
// local time.
//
// Given a time.Time struct `t`, one could call `t.UnixNano()/1000` to convert
// to Micros, but experience has shown that certain programmers (e.g., Ben)
// have a nasty habit of forgetting the `/1000`. We thus lean on the compiler
// to help us not-screw this up.
type Micros int64

const (
	MICROS_PER_MILLIS = 1000
	MICROS_PER_SECOND = MICROS_PER_MILLIS * 1000
	MICROS_PER_MINUTE = MICROS_PER_SECOND * 60
	MICROS_PER_HOUR   = MICROS_PER_MINUTE * 60
	MICROS_PER_DAY    = MICROS_PER_HOUR * 24
	MICROS_PER_WEEK   = MICROS_PER_DAY * 7
	NS_PER_MICRO      = 1000
	EPOCH_MICROS      = Micros(0)
)

// Note that ToMicros() and ToTime() are not inverses of each other, as
// ToMicros will drop the last 3 decimal places of nanosecond timestamps.
func ToMicros(t time.Time) Micros {
	return Micros(t.UnixNano() / NS_PER_MICRO)
}

func DurationInMicros(d time.Duration) Micros {
	return Micros(d / NS_PER_MICRO)
}

func NowMicros() Micros {
	return ToMicros(time.Now())
}

func (m Micros) Int64() int64 {
	return int64(m)
}

// Returns the maximum of the two Micros
func (m Micros) Max(other Micros) Micros {
	return Micros(imath.Max64(int64(m), int64(other)))
}

// Returns the minimum of the two Micros
func (m Micros) Min(other Micros) Micros {
	return Micros(imath.Min64(int64(m), int64(other)))
}

func (m Micros) String() string {
	return strconv.FormatInt(int64(m), 10)
}

// Returns -1 on failure.
func ParseMicros(s string) Micros {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		i = -1
	}
	return Micros(i)
}

func (m Micros) ToTime() time.Time {
	return time.Unix(int64(m)/MICROS_PER_SECOND, NS_PER_MICRO*(int64(m)%MICROS_PER_SECOND))
}

func (m Micros) ToDuration() time.Duration {
	return time.Duration(int64(m) * NS_PER_MICRO)
}

func (m Micros) ToPtr() *Micros {
	return &m
}

func (m Micros) ToMillis() float64 {
	return float64(m) / float64(MICROS_PER_MILLIS)
}

// Returns the next time that will be a multiple of the specified duration, relative to the Unix Epoch.
// Do not try to use this method to return a specific time of day, use NextTimeOfDay.
func NextDurationMultiple(period time.Duration) time.Time {
	now := time.Now()
	return now.Add(period - (time.Duration(now.UnixNano()) % period))
}

// Returns the next time that the time-of-day will be as specified.
func NextTimeOfDay(hour, minute, second int) time.Time {
	// Next time defaults to pacific time
	location := LoadLocation("America/Los_Angeles")
	now := time.Now()
	y, m, d := now.Date()
	then := time.Date(y, m, d, hour, minute, second, 0, location)
	if then.Before(now) {
		// Oy, off by an hour when we change to daylight savings time or back
		then = then.Add(24 * time.Hour)
	}
	return then
}

// Returns the earliest of the given times.
func Earliest(date time.Time, dates ...time.Time) time.Time {
	earliest := date
	for _, t := range dates {
		if t.Before(earliest) {
			earliest = t
		}
	}
	return earliest
}

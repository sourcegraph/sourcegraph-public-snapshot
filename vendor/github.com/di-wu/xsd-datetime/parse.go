package xsd_datetime

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	datetime = "2006-01-02T15:04:05"
	timezone = "-07:00"
)

// Parse parses a formatted string and returns the time value it represents.
//
// Layout used:
//	[-]YYYY-MM-DDThh:mm:ss[.fffffffff][Z|(+|-)hh:mm]
// In the absence of a time zone indicator, Parse returns a time in UTC.
func Parse(value string) (time.Time, error) {

	// Time being constructed.
	var (
		date time.Time      // YYYY-MM-DDThh:mm:ss
		tz   *time.Location // Z|(+|-)hh:mm
		frac int            // .fffffffff
	)

	// [-]
	var neg bool
	if strings.HasPrefix(value, "-") {
		neg = true
		value = value[1:]
	}

	// YYYY-MM-DDThh:mm:ss
	if len(value) < 19 {
		return time.Time{}, errors.New("time: value too short")
	}
	p, value := value[:19], value[19:]
	date, err := time.Parse(datetime, p)
	if err != nil {
		return time.Time{}, err
	}

	// [.fffffffff]
	var digits int64
	if strings.HasPrefix(value, ".") {
		digits, value, err = leadingInt(value[1:])
		if err != nil {
			return time.Time{}, err
		}
		if len(strconv.FormatInt(digits, 10)) > 9 {
			return time.Time{}, errors.New("time: too many fractional seconds")
		}
		frac = int(digits) * int(math.Pow10(9-len(strconv.FormatInt(digits, 10))))
	}

	// [zzzzzz]
	switch {
	case len(value) == 0 || (len(value) == 1 && value[0] == 'Z'):
		tz = time.UTC
	case strings.HasPrefix(value, "+") || strings.HasPrefix(value, "-"):
		if len(value) < 6 {
			return time.Time{}, errors.New("time: value too short")
		}
		if value[1:] == "00:00" { // same as Z
			tz = time.UTC
			break
		}

		t, err := time.Parse(timezone, value)
		if err != nil {
			return time.Time{}, err
		}
		tz = t.Location()
	default:
		return time.Time{}, errors.New("time: value too long")
	}

	year := date.Year()
	if neg {
		year *= -1
	}

	datetime := time.Date(
		year, date.Month(), date.Day(),
		date.Hour(), date.Minute(), date.Second(), frac,
		tz,
	)

	return datetime, nil
}

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (int64, string, error) {
	var x int64
	var i int
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<63-1)/10 {
			// overflow
			return 0, "", errors.New("time: invalid number")
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", errors.New("time: invalid number")
		}
	}
	return x, s[i:], nil
}

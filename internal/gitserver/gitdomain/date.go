package gitdomain

import (
	"strconv"
	"strings"
	"time"

	"github.com/tj/go-naturaldate"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ParseGitDate implements date parsing for before/after arguments.
// The intent is to replicate the behavior of git CLI's date parsing as documented here:
// https://github.com/git/git/blob/master/Documentation/date-formats.txt
func ParseGitDate(s string, now func() time.Time) (time.Time, error) {
	// Git internal format
	if t, err := parseGitInternalFormat(s); err == nil {
		return t, nil
	}

	// RFC 3339
	{
		// Only date
		if t, err := time.Parse("2006-01-02", s); err == nil {
			return t, nil
		}

		// With timezone
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t, nil
		}

		// Without timezone
		if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
			return t, nil
		}

		// With timezone and space
		if t, err := time.Parse("2006-01-02 15:04:05Z07:00", s); err == nil {
			return t, nil
		}

		// Without timezone and space
		if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
			return t, nil
		}
	}

	// RFC 2822
	if t, err := time.Parse("Thu, 02 Jan 2006 15:04:05 -0700", s); err == nil {
		return t, nil
	}

	// YYYY.MM.DD
	if t, err := time.Parse("2006.01.02", s); err == nil {
		return t, nil
	}

	// MM/DD/YYYY
	if t, err := time.Parse("1/2/2006", s); err == nil {
		return t, nil
	}

	// DD.MM.YYYY
	if t, err := time.Parse("2.1.2006", s); err == nil {
		return t, nil
	}

	// 1 november 2020 or november 1 2020
	if t, err := parseSimpleDate(s); err == nil {
		return t, nil
	}

	// Human date
	n := now()
	if t, err := naturaldate.Parse(s, n, naturaldate.WithDirection(naturaldate.Past)); err == nil && t != n {
		// We test that t != n because naturaldate won't necessarily error
		// if it doesn't find any time values in the string
		return t, nil
	}

	return time.Time{}, errInvalidDate
}

// Seconds since unix epoch plus an optional time zone offset
// As documented here: https://github.com/git/git/blob/master/Documentation/date-formats.txt
var gitInternalTimestampRegexp = lazyregexp.New(`^(?P<epoch_seconds>\d{5,})( (?P<zone_offset>(?P<pm>\+|\-)(?P<hours>\d{2})(?P<minutes>\d{2})))?$`)

var errInvalidDate = errors.New("invalid date format")

func parseGitInternalFormat(s string) (time.Time, error) {
	re := gitInternalTimestampRegexp
	match := re.FindStringSubmatch(s)
	if match == nil {
		return time.Time{}, errInvalidDate
	}

	locationName := match[re.SubexpIndex("zone_offset")]

	epochSeconds, err := strconv.Atoi(match[re.SubexpIndex("epoch_seconds")])
	if err != nil {
		return time.Time{}, errInvalidDate
	}

	// If a time zone offset is set, respect it
	offsetSeconds := 0
	if locationName != "" {
		hours, err := strconv.Atoi(match[re.SubexpIndex("hours")])
		if err != nil {
			return time.Time{}, errInvalidDate
		}

		minutes, err := strconv.Atoi(match[re.SubexpIndex("minutes")])
		if err != nil {
			return time.Time{}, errInvalidDate
		}

		offsetSeconds = hours*60*60 + minutes*60
		if match[re.SubexpIndex("pm")] == "-" {
			offsetSeconds *= -1
		}
	}

	// This looks weird because there is no way to force the location of a time.Time.
	// time.Unix() defaults to local time, but we need to set the time zone, and (*Time).setLoc() is private.
	// Instead, we parse the unix timestamp into a time.Time in UTC, then use that to create a new  time
	// with our desired time zone.
	t := time.Unix(int64(epochSeconds), 0).In(time.UTC)
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.FixedZone(locationName, offsetSeconds)), nil
}

var (
	simpleDateRe1 = lazyregexp.New(`(?P<month>[A-Za-z]{3,9})\s+(?P<day>\d{1,2}),?\s+(?P<year>\d{4})`)
	simpleDateRe2 = lazyregexp.New(`(?P<day>\d{1,2})\s+(?P<month>[A-Za-z]{3,9}),?\s+(?P<year>\d{4})`)
	monthNums     = map[string]time.Month{
		"january":   time.January,
		"jan":       time.January,
		"february":  time.February,
		"feb":       time.February,
		"march":     time.March,
		"mar":       time.March,
		"april":     time.April,
		"apr":       time.April,
		"may":       time.May,
		"june":      time.June,
		"jun":       time.June,
		"july":      time.July,
		"jul":       time.July,
		"august":    time.August,
		"aug":       time.August,
		"september": time.September,
		"sep":       time.September,
		"october":   time.October,
		"oct":       time.October,
		"november":  time.November,
		"nov":       time.November,
		"december":  time.December,
		"dec":       time.December,
	}
)

// parseSimpleDate parses dates of the form "1 january 1996" or "january 1 1996"
func parseSimpleDate(s string) (time.Time, error) {
	re := simpleDateRe1
	match := re.FindStringSubmatch(s)
	if match == nil {
		re = simpleDateRe2
		match = re.FindStringSubmatch(s)
		if match == nil {
			return time.Time{}, errInvalidDate
		}
	}

	month := strings.ToLower(match[re.SubexpIndex("month")])
	monthNum, ok := monthNums[month]
	if !ok {
		return time.Time{}, errInvalidDate
	}

	day, err := strconv.Atoi(match[re.SubexpIndex("day")])
	if err != nil {
		return time.Time{}, errInvalidDate
	}

	year, err := strconv.Atoi(match[re.SubexpIndex("year")])
	if err != nil {
		return time.Time{}, errInvalidDate
	}

	return time.Date(year, monthNum, day, 0, 0, 0, 0, time.UTC), nil
}

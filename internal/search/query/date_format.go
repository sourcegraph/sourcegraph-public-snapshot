package query

import (
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/tj/go-naturaldate"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
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

	// Human date
	n := now()
	if t, err := naturaldate.Parse(s, n); err == nil && t != n {
		// We test that t != n because naturaldate won't necessarily error
		// if it doesn't find any time values in the string
		return t, nil
	}

	return time.Time{}, errInvalidDate
}

var gitInternalTimestampRegexp = lazyregexp.New(`^(?P<epoch_seconds>\d+) (?P<zone_offset>(?P<pm>\+|\-)(?P<hours>\d{2})(?P<minutes>\d{2}))$`)

var errInvalidDate = errors.New("invalid date format")

func parseGitInternalFormat(s string) (time.Time, error) {
	re := gitInternalTimestampRegexp
	match := re.FindStringSubmatch(s)
	if match == nil {
		return time.Time{}, errInvalidDate
	}

	epochSeconds, err := strconv.Atoi(match[re.SubexpIndex("epoch_seconds")])
	if err != nil {
		return time.Time{}, errInvalidDate
	}

	hours, err := strconv.Atoi(match[re.SubexpIndex("hours")])
	if err != nil {
		return time.Time{}, errInvalidDate
	}

	minutes, err := strconv.Atoi(match[re.SubexpIndex("minutes")])
	if err != nil {
		return time.Time{}, errInvalidDate
	}

	locationName := match[re.SubexpIndex("zone_offset")]

	offsetSeconds := hours*60*60 + minutes*60
	if match[re.SubexpIndex("pm")] == "-" {
		offsetSeconds *= -1
	}

	return time.Unix(int64(epochSeconds), 0).In(time.FixedZone(locationName, offsetSeconds)), nil
}

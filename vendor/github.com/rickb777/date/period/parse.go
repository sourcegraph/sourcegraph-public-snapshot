// Copyright 2015 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package period

import (
	"fmt"
	"strconv"
	"strings"
)

// MustParse is as per Parse except that it panics if the string cannot be parsed.
// This is intended for setup code; don't use it for user inputs.
// By default, the value is normalised.
// Normalisation can be disabled using the optional flag.
func MustParse(value string, normalise ...bool) Period {
	d, err := Parse(value, normalise...)
	if err != nil {
		panic(err)
	}
	return d
}

// Parse parses strings that specify periods using ISO-8601 rules.
//
// In addition, a plus or minus sign can precede the period, e.g. "-P10D"
//
// By default, the value is normalised, e.g. multiple of 12 months become years
// so "P24M" is the same as "P2Y". However, this is done without loss of precision,
// so for example whole numbers of days do not contribute to the months tally
// because the number of days per month is variable.
//
// Normalisation can be disabled using the optional flag.
//
// The zero value can be represented in several ways: all of the following
// are equivalent: "P0Y", "P0M", "P0W", "P0D", "PT0H", PT0M", PT0S", and "P0".
// The canonical zero is "P0D".
func Parse(period string, normalise ...bool) (Period, error) {
	return ParseWithNormalise(period, len(normalise) == 0 || normalise[0])
}

// ParseWithNormalise parses strings that specify periods using ISO-8601 rules
// with an option to specify whether to normalise parsed period components.
//
// This method is deprecated and should not be used. It may be removed in a
// future version.
func ParseWithNormalise(period string, normalise bool) (Period, error) {
	if period == "" || period == "-" || period == "+" {
		return Period{}, fmt.Errorf("cannot parse a blank string as a period")
	}

	if period == "P0" {
		return Period{}, nil
	}

	p64, err := parse(period, normalise)
	if err != nil {
		return Period{}, err
	}
	return p64.toPeriod()
}

func parse(period string, normalise bool) (*period64, error) {
	neg := false
	remaining := period
	if remaining[0] == '-' {
		neg = true
		remaining = remaining[1:]
	} else if remaining[0] == '+' {
		remaining = remaining[1:]
	}

	if remaining[0] != 'P' {
		return nil, fmt.Errorf("%s: expected 'P' period mark at the start", period)
	}
	remaining = remaining[1:]

	var number, weekValue, prevFraction int64
	result := &period64{input: period, neg: neg}
	var years, months, weeks, days, hours, minutes, seconds itemState
	var designator, prevDesignator byte
	var err error
	nComponents := 0

	years, months, weeks, days = Armed, Armed, Armed, Armed

	isHMS := false
	for len(remaining) > 0 {
		if remaining[0] == 'T' {
			if isHMS {
				return nil, fmt.Errorf("%s: 'T' designator cannot occur more than once", period)
			}
			isHMS = true

			years, months, weeks, days = Unready, Unready, Unready, Unready
			hours, minutes, seconds = Armed, Armed, Armed

			remaining = remaining[1:]

		} else {
			number, designator, remaining, err = parseNextField(remaining, period)
			if err != nil {
				return nil, err
			}

			fraction := number % 10
			if prevFraction != 0 && fraction != 0 {
				return nil, fmt.Errorf("%s: '%c' & '%c' only the last field can have a fraction", period, prevDesignator, designator)
			}

			switch designator {
			case 'Y':
				years, err = years.testAndSet(number, 'Y', result, &result.years)
			case 'W':
				weeks, err = weeks.testAndSet(number, 'W', result, &weekValue)
			case 'D':
				days, err = days.testAndSet(number, 'D', result, &result.days)
			case 'H':
				hours, err = hours.testAndSet(number, 'H', result, &result.hours)
			case 'S':
				seconds, err = seconds.testAndSet(number, 'S', result, &result.seconds)
			case 'M':
				if isHMS {
					minutes, err = minutes.testAndSet(number, 'M', result, &result.minutes)
				} else {
					months, err = months.testAndSet(number, 'M', result, &result.months)
				}
			default:
				return nil, fmt.Errorf("%s: expected a number not '%c'", period, designator)
			}
			nComponents++

			if err != nil {
				return nil, err
			}

			prevFraction = fraction
			prevDesignator = designator
		}
	}

	if nComponents == 0 {
		return nil, fmt.Errorf("%s: expected 'Y', 'M', 'W', 'D', 'H', 'M', or 'S' designator", period)
	}

	result.days += weekValue * 7

	if normalise {
		result = result.normalise64(true)
	}

	return result, nil
}

//-------------------------------------------------------------------------------------------------

type itemState int

const (
	Unready itemState = iota
	Armed
	Set
)

func (i itemState) testAndSet(number int64, designator byte, result *period64, value *int64) (itemState, error) {
	switch i {
	case Unready:
		return i, fmt.Errorf("%s: '%c' designator cannot occur here", result.input, designator)
	case Set:
		return i, fmt.Errorf("%s: '%c' designator cannot occur more than once", result.input, designator)
	}

	*value = number
	return Set, nil
}

//-------------------------------------------------------------------------------------------------

func parseNextField(str, original string) (int64, byte, string, error) {
	i := scanDigits(str)
	if i < 0 {
		return 0, 0, "", fmt.Errorf("%s: missing designator at the end", original)
	}

	des := str[i]
	number, err := parseDecimalNumber(str[:i], original, des)
	return number, des, str[i+1:], err
}

// Fixed-point one decimal place
func parseDecimalNumber(number, original string, des byte) (int64, error) {
	dec := strings.IndexByte(number, '.')
	if dec < 0 {
		dec = strings.IndexByte(number, ',')
	}

	var integer, fraction int64
	var err error
	if dec >= 0 {
		integer, err = strconv.ParseInt(number[:dec], 10, 64)
		if err == nil {
			number = number[dec+1:]
			switch len(number) {
			case 0: // skip
			case 1:
				fraction, err = strconv.ParseInt(number, 10, 64)
			default:
				fraction, err = strconv.ParseInt(number[:1], 10, 64)
			}
		}
	} else {
		integer, err = strconv.ParseInt(number, 10, 64)
	}

	if err != nil {
		return 0, fmt.Errorf("%s: expected a number but found '%c'", original, des)
	}

	return integer*10 + fraction, err
}

// scanDigits finds the first non-digit byte after a given starting point.
// Note that it does not care about runes or UTF-8 encoding; it assumes that
// a period string is always valid ASCII as well as UTF-8.
func scanDigits(s string) int {
	for i, c := range s {
		if !isDigit(c) {
			return i
		}
	}
	return -1
}

func isDigit(c rune) bool {
	return ('0' <= c && c <= '9') || c == '.' || c == ','
}

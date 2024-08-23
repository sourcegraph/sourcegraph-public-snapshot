// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bigquery

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// IntervalValue is a go type for representing BigQuery INTERVAL values.
// Intervals are represented using three distinct parts:
// * Years and Months
// * Days
// * Time (Hours/Mins/Seconds/Fractional Seconds).
//
// More information about BigQuery INTERVAL types can be found at:
// https://cloud.google.com/bigquery/docs/reference/standard-sql/data-types#interval_type
//
// IntervalValue is EXPERIMENTAL and subject to change or removal without notice.
type IntervalValue struct {
	// In canonical form, Years and Months share a consistent sign and reduced
	// to avoid large month values.
	Years  int32
	Months int32

	// In canonical form, Days are independent of the other parts and can have it's
	// own sign.  There is no attempt to reduce larger Day values into the Y-M part.
	Days int32

	// In canonical form, the time parts all share a consistent sign and are reduced.
	Hours   int32
	Minutes int32
	Seconds int32
	// This represents the fractional seconds as nanoseconds.
	SubSecondNanos int32
}

// String returns string representation of the interval value using the canonical format.
// The canonical format is as follows:
//
// [sign]Y-M [sign]D [sign]H:M:S[.F]
func (iv *IntervalValue) String() string {
	// Don't canonicalize the current value.  Instead, if it's not canonical,
	// compute the canonical form and use that.
	src := iv
	if !iv.IsCanonical() {
		src = iv.Canonicalize()
	}
	out := fmt.Sprintf("%d-%d %d %d:%d:%d", src.Years, int32abs(src.Months), src.Days, src.Hours, int32abs(src.Minutes), int32abs(src.Seconds))
	if src.SubSecondNanos != 0 {
		mantStr := fmt.Sprintf("%09d", src.SubSecondNanos)
		for len(mantStr) > 0 && mantStr[len(mantStr)-1:] == "0" {
			mantStr = mantStr[0 : len(mantStr)-1]
		}
		out = fmt.Sprintf("%s.%s", out, mantStr)
	}
	return out
}

// intervalPart is used for parsing string representations.
type intervalPart int

const (
	yearsPart = iota
	monthsPart
	daysPart
	hoursPart
	minutesPart
	secondsPart
	subsecsPart
)

func (i intervalPart) String() string {
	knownParts := []string{"YEARS", "MONTHS", "DAYS", "HOURS", "MINUTES", "SECONDS", "SUBSECONDS"}
	if i < 0 || int(i) > len(knownParts) {
		return fmt.Sprintf("UNKNOWN(%d)", i)
	}
	return knownParts[i]
}

// canonicalParts indicates the parse order for canonical format.
var canonicalParts = []intervalPart{yearsPart, monthsPart, daysPart, hoursPart, minutesPart, secondsPart, subsecsPart}

// ParseInterval parses an interval in canonical string format and returns the IntervalValue it represents.
func ParseInterval(value string) (*IntervalValue, error) {
	iVal := &IntervalValue{}
	for _, part := range canonicalParts {
		remaining, v, err := getPartValue(part, value)
		if err != nil {
			return nil, err
		}
		switch part {
		case yearsPart:
			iVal.Years = v
		case monthsPart:
			iVal.Months = v
			if iVal.Years < 0 {
				iVal.Months = -v
			}
		case daysPart:
			iVal.Days = v
		case hoursPart:
			iVal.Hours = v
		case minutesPart:
			iVal.Minutes = v
			if iVal.Hours < 0 {
				iVal.Minutes = -v
			}
		case secondsPart:
			iVal.Seconds = v
			if iVal.Hours < 0 {
				iVal.Seconds = -v
			}
		case subsecsPart:
			iVal.SubSecondNanos = v
			if iVal.Hours < 0 {
				iVal.SubSecondNanos = -v
			}
		default:
			return nil, fmt.Errorf("encountered invalid part %s during parse", part)
		}
		value = remaining
	}
	return iVal, nil
}

func getPartValue(part intervalPart, s string) (string, int32, error) {
	s = trimPrefix(part, s)
	return getNumVal(part, s)
}

// trimPrefix removes formatting prefix relevant to the given type.
func trimPrefix(part intervalPart, s string) string {
	var trimByte byte
	switch part {
	case yearsPart, daysPart, hoursPart:
		trimByte = byte(' ')
	case monthsPart:
		trimByte = byte('-')
	case minutesPart, secondsPart:
		trimByte = byte(':')
	case subsecsPart:
		trimByte = byte('.')
	}
	for len(s) > 0 && s[0] == trimByte {
		s = s[1:]
	}
	return s
}

func getNumVal(part intervalPart, s string) (string, int32, error) {

	allowedVals := []byte("0123456789")
	var allowedSign bool
	captured := ""
	switch part {
	case yearsPart, daysPart, hoursPart:
		allowedSign = true
	}
	// capture sign prefix +/-
	if len(s) > 0 && allowedSign {
		switch s[0] {
		case '-':
			captured = "-"
			s = s[1:]
		case '+':
			s = s[1:]
		}
	}
	for len(s) > 0 && bytes.IndexByte(allowedVals, s[0]) >= 0 {
		captured = captured + string(s[0])
		s = s[1:]
	}

	if len(captured) == 0 {
		if part == subsecsPart {
			return s, 0, nil
		}
		return "", 0, fmt.Errorf("no value parsed for part %s", part.String())
	}
	// special case: subsecs is a mantissa, convert it to nanos
	if part == subsecsPart {
		parsed, err := strconv.ParseFloat(fmt.Sprintf("0.%s", captured), 64)
		if err != nil {
			return "", 0, fmt.Errorf("couldn't parse %s as %s", captured, part.String())
		}
		return s, int32(parsed * 1e9), nil
	}
	parsed, err := strconv.ParseInt(captured, 10, 32)
	if err != nil {
		return "", 0, fmt.Errorf("error parsing value %s for %s: %w", captured, part.String(), err)
	}
	return s, int32(parsed), nil
}

// IntervalValueFromDuration converts a time.Duration to an IntervalType representation.
//
// The converted duration only leverages the hours/minutes/seconds part of the interval,
// the other parts representing days, months, and years are not used.
func IntervalValueFromDuration(in time.Duration) *IntervalValue {
	nanos := in.Nanoseconds()
	out := &IntervalValue{}
	out.Hours = int32(nanos / 3600 / 1e9)
	nanos = nanos - (int64(out.Hours) * 3600 * 1e9)
	out.Minutes = int32(nanos / 60 / 1e9)
	nanos = nanos - (int64(out.Minutes) * 60 * 1e9)
	out.Seconds = int32(nanos / 1e9)
	nanos = nanos - (int64(out.Seconds) * 1e9)
	out.SubSecondNanos = int32(nanos)
	return out
}

// ToDuration converts an interval to a time.Duration value.
//
// For the purposes of conversion:
// Years are normalized to 12 months.
// Months are normalized to 30 days.
// Days are normalized to 24 hours.
func (iv *IntervalValue) ToDuration() time.Duration {
	var accum int64
	accum = 12*int64(iv.Years) + int64(iv.Months)
	// widen to days
	accum = accum*30 + int64(iv.Days)
	// hours
	accum = accum*24 + int64(iv.Hours)
	// minutes
	accum = accum*60 + int64(iv.Minutes)
	// seconds
	accum = accum*60 + int64(iv.Seconds)
	// subsecs
	accum = accum*1e9 + int64(iv.SubSecondNanos*1e9)
	return time.Duration(accum)
}

// Canonicalize returns an IntervalValue where signs for elements in the
// Y-M and H:M:S.F are consistent and values are normalized/reduced.
//
// Canonical form enables more consistent comparison of the encoded
// interval.  For example, encoding an interval with 12 months is equivalent
// to an interval of 1 year.
func (iv *IntervalValue) Canonicalize() *IntervalValue {
	newIV := &IntervalValue{iv.Years, iv.Months, iv.Days, iv.Hours, iv.Minutes, iv.Seconds, iv.SubSecondNanos}
	// canonicalize Y-M part
	totalMonths := iv.Years*12 + iv.Months
	newIV.Years = totalMonths / 12
	totalMonths = totalMonths - (newIV.Years * 12)
	newIV.Months = totalMonths % 12

	// No canonicalization for the Days part.

	// canonicalize time part by switching to Nanos.
	totalNanos := int64(iv.Hours)*3600*1e9 +
		int64(iv.Minutes)*60*1e9 +
		int64(iv.Seconds)*1e9 +
		int64(iv.SubSecondNanos)

	// Reduce to parts.
	newIV.Hours = int32(totalNanos / 60 / 60 / 1e9)
	totalNanos = totalNanos - (int64(newIV.Hours) * 3600 * 1e9)
	newIV.Minutes = int32(totalNanos / 60 / 1e9)
	totalNanos = totalNanos - (int64(newIV.Minutes) * 60 * 1e9)
	newIV.Seconds = int32(totalNanos / 1e9)
	totalNanos = totalNanos - (int64(newIV.Seconds) * 1e9)
	newIV.SubSecondNanos = int32(totalNanos)
	return newIV
}

// IsCanonical evaluates whether the current representation is in canonical
// form.
func (iv *IntervalValue) IsCanonical() bool {
	if !sameSign(iv.Years, iv.Months) ||
		!sameSign(iv.Hours, iv.Minutes) {
		return false
	}
	// We allow large days and hours values, because they are within different parts.
	if int32abs(iv.Months) > 12 ||
		int32abs(iv.Minutes) > 60 ||
		int32abs(iv.Seconds) > 60 ||
		int32abs(iv.SubSecondNanos) > 1e9 {
		return false
	}
	// TODO: We don't currently validate that each part represents value smaller than 10k years.
	return true
}

func int32abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

func sameSign(nums ...int32) bool {
	var pos, neg int
	for _, n := range nums {
		if n > 0 {
			pos = pos + 1
		}
		if n < 0 {
			neg = neg + 1
		}
	}
	if pos > 0 && neg > 0 {
		return false
	}
	return true
}

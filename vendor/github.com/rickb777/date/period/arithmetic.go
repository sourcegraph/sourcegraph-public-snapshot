// Copyright 2015 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package period

import (
	"time"
)

// Add adds two periods together. Use this method along with Negate in order to subtract periods.
//
// The result is not normalised and may overflow arithmetically (to make this unlikely, use Normalise on
// the inputs before adding them).
func (period Period) Add(that Period) Period {
	return Period{
		period.years + that.years,
		period.months + that.months,
		period.days + that.days,
		period.hours + that.hours,
		period.minutes + that.minutes,
		period.seconds + that.seconds,
	}
}

//-------------------------------------------------------------------------------------------------

// AddTo adds the period to a time, returning the result.
// A flag is also returned that is true when the conversion was precise and false otherwise.
//
// When the period specifies hours, minutes and seconds only, the result is precise.
// Also, when the period specifies whole years, months and days (i.e. without fractions), the
// result is precise. However, when years, months or days contains fractions, the result
// is only an approximation (it assumes that all days are 24 hours and every year is 365.2425
// days, as per Gregorian calendar rules).
func (period Period) AddTo(t time.Time) (time.Time, bool) {
	wholeYears := (period.years % 10) == 0
	wholeMonths := (period.months % 10) == 0
	wholeDays := (period.days % 10) == 0

	if wholeYears && wholeMonths && wholeDays {
		// in this case, time.AddDate provides an exact solution
		stE3 := totalSecondsE3(period)
		t1 := t.AddDate(int(period.years/10), int(period.months/10), int(period.days/10))
		return t1.Add(stE3 * time.Millisecond), true
	}

	d, precise := period.Duration()
	return t.Add(d), precise
}

//-------------------------------------------------------------------------------------------------

// Scale a period by a multiplication factor. Obviously, this can both enlarge and shrink it,
// and change the sign if negative. The result is normalised, but integer overflows are silently
// ignored.
//
// Bear in mind that the internal representation is limited by fixed-point arithmetic with two
// decimal places; each field is only int16.
//
// Known issue: scaling by a large reduction factor (i.e. much less than one) doesn't work properly.
func (period Period) Scale(factor float32) Period {
	result, _ := period.ScaleWithOverflowCheck(factor)
	return result
}

// ScaleWithOverflowCheck a period by a multiplication factor. Obviously, this can both enlarge and shrink it,
// and change the sign if negative. The result is normalised. An error is returned if integer overflow
// happened.
//
// Bear in mind that the internal representation is limited by fixed-point arithmetic with one
// decimal place; each field is only int16.
//
// Known issue: scaling by a large reduction factor (i.e. much less than one) doesn't work properly.
func (period Period) ScaleWithOverflowCheck(factor float32) (Period, error) {
	ap, neg := period.absNeg()

	if -0.5 < factor && factor < 0.5 {
		d, pr1 := ap.Duration()
		mul := float64(d) * float64(factor)
		p2, pr2 := NewOf(time.Duration(mul))
		return p2.Normalise(pr1 && pr2), nil
	}

	y := int64(float32(ap.years) * factor)
	m := int64(float32(ap.months) * factor)
	d := int64(float32(ap.days) * factor)
	hh := int64(float32(ap.hours) * factor)
	mm := int64(float32(ap.minutes) * factor)
	ss := int64(float32(ap.seconds) * factor)

	p64 := &period64{years: y, months: m, days: d, hours: hh, minutes: mm, seconds: ss, neg: neg}
	return p64.normalise64(true).toPeriod()
}

// RationalScale scales a period by a rational multiplication factor. Obviously, this can both enlarge and shrink it,
// and change the sign if negative. The result is normalised. An error is returned if integer overflow
// happened.
//
// If the divisor is zero, a panic will arise.
//
// Bear in mind that the internal representation is limited by fixed-point arithmetic with two
// decimal places; each field is only int16.
//func (period Period) RationalScale(multiplier, divisor int) (Period, error) {
//	return period.rationalScale64(int64(multiplier), int64(divisor))
//}

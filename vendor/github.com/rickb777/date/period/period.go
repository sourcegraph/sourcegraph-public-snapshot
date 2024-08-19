// Copyright 2015 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package period

import (
	"fmt"
	"time"
)

const daysPerYearE4 = 3652425   // 365.2425 days by the Gregorian rule
const daysPerMonthE4 = 304369   // 30.4369 days per month
const daysPerMonthE6 = 30436875 // 30.436875 days per month

const oneE4 = 10000
const oneE5 = 100000
const oneE6 = 1000000
const oneE7 = 10000000

const hundredMs = 100 * time.Millisecond

// reminder: int64 overflow is after 9,223,372,036,854,775,807 (math.MaxInt64)

// Period holds a period of time and provides conversion to/from ISO-8601 representations.
// Therefore there are six fields: years, months, days, hours, minutes, and seconds.
//
// In the ISO representation, decimal fractions are supported, although only the last non-zero
// component is allowed to have a fraction according to the Standard. For example "P2.5Y"
// is 2.5 years.
//
// However, in this implementation, the precision is limited to one decimal place only, by
// means of integers with fixed point arithmetic. (This avoids using float32 in the struct,
// so there are no problems testing equality using ==.)
//
// The implementation limits the range of possible values to ± 2^16 / 10 in each field.
// Note in particular that the range of years is limited to approximately ± 3276.
//
// The concept of weeks exists in string representations of periods, but otherwise weeks
// are unimportant. The period contains a number of days from which the number of weeks can
// be calculated when needed.
//
// Note that although fractional weeks can be parsed, they will never be returned via String().
// This is because the number of weeks is always inferred from the number of days.
//
type Period struct {
	years, months, days, hours, minutes, seconds int16
}

// NewYMD creates a simple period without any fractional parts. The fields are initialised verbatim
// without any normalisation; e.g. 12 months will not become 1 year. Use the Normalise method if you
// need to.
//
// All the parameters must have the same sign (otherwise a panic occurs).
// Because this implementation uses int16 internally, the paramters must
// be within the range ± 2^16 / 10.
func NewYMD(years, months, days int) Period {
	return New(years, months, days, 0, 0, 0)
}

// NewHMS creates a simple period without any fractional parts. The fields are initialised verbatim
// without any normalisation; e.g. 120 seconds will not become 2 minutes. Use the Normalise method
// if you need to.
//
// All the parameters must have the same sign (otherwise a panic occurs).
// Because this implementation uses int16 internally, the paramters must
// be within the range ± 2^16 / 10.
func NewHMS(hours, minutes, seconds int) Period {
	return New(0, 0, 0, hours, minutes, seconds)
}

// New creates a simple period without any fractional parts. The fields are initialised verbatim
// without any normalisation; e.g. 120 seconds will not become 2 minutes. Use the Normalise method
// if you need to.
//
// All the parameters must have the same sign (otherwise a panic occurs).
func New(years, months, days, hours, minutes, seconds int) Period {
	if (years >= 0 && months >= 0 && days >= 0 && hours >= 0 && minutes >= 0 && seconds >= 0) ||
		(years <= 0 && months <= 0 && days <= 0 && hours <= 0 && minutes <= 0 && seconds <= 0) {
		return Period{
			int16(years) * 10, int16(months) * 10, int16(days) * 10,
			int16(hours) * 10, int16(minutes) * 10, int16(seconds) * 10,
		}
	}
	panic(fmt.Sprintf("Periods must have homogeneous signs; got P%dY%dM%dDT%dH%dM%dS",
		years, months, days, hours, minutes, seconds))
}

// NewOf converts a time duration to a Period, and also indicates whether the conversion is precise.
// Any time duration that spans more than ± 3276 hours will be approximated by assuming that there
// are 24 hours per day, 365.2425 days per year (as per Gregorian calendar rules), and a month
// being 1/12 of that (approximately 30.4369 days).
//
// The result is not always fully normalised; for time differences less than 3276 hours (about 4.5 months),
// it will contain zero in the years, months and days fields but the number of days may be up to 3275; this
// reduces errors arising from the variable lengths of months. For larger time differences, greater than
// 3276 hours, the days, months and years fields are used as well.
func NewOf(duration time.Duration) (p Period, precise bool) {
	var sign int16 = 1
	d := duration
	if duration < 0 {
		sign = -1
		d = -duration
	}

	sign10 := sign * 10

	totalHours := int64(d / time.Hour)

	// check for 16-bit overflow - occurs near the 4.5 month mark
	if totalHours < 3277 {
		// simple HMS case
		minutes := d % time.Hour / time.Minute
		seconds := d % time.Minute / hundredMs
		return Period{0, 0, 0, sign10 * int16(totalHours), sign10 * int16(minutes), sign * int16(seconds)}, true
	}

	totalDays := totalHours / 24 // ignoring daylight savings adjustments

	if totalDays < 3277 {
		hours := totalHours - totalDays*24
		minutes := d % time.Hour / time.Minute
		seconds := d % time.Minute / hundredMs
		return Period{0, 0, sign10 * int16(totalDays), sign10 * int16(hours), sign10 * int16(minutes), sign * int16(seconds)}, false
	}

	// TODO it is uncertain whether this is too imprecise and should be improved
	years := (oneE4 * totalDays) / daysPerYearE4
	months := ((oneE4 * totalDays) / daysPerMonthE4) - (12 * years)
	hours := totalHours - totalDays*24
	totalDays = ((totalDays * oneE4) - (daysPerMonthE4 * months) - (daysPerYearE4 * years)) / oneE4
	return Period{sign10 * int16(years), sign10 * int16(months), sign10 * int16(totalDays), sign10 * int16(hours), 0, 0}, false
}

// Between converts the span between two times to a period. Based on the Gregorian conversion
// algorithms of `time.Time`, the resultant period is precise.
//
// To improve precision, result is not always fully normalised; for time differences less than 3276 hours
// (about 4.5 months), it will contain zero in the years, months and days fields but the number of hours
// may be up to 3275; this reduces errors arising from the variable lengths of months. For larger time
// differences (greater than 3276 hours) the days, months and years fields are used as well.
//
// Remember that the resultant period does not retain any knowledge of the calendar, so any subsequent
// computations applied to the period can only be precise if they concern either the date (year, month,
// day) part, or the clock (hour, minute, second) part, but not both.
func Between(t1, t2 time.Time) (p Period) {
	sign := 1
	if t2.Before(t1) {
		t1, t2, sign = t2, t1, -1
	}

	if t1.Location() != t2.Location() {
		t2 = t2.In(t1.Location())
	}

	year, month, day, hour, min, sec, hundredth := daysDiff(t1, t2)

	if sign < 0 {
		p = New(-year, -month, -day, -hour, -min, -sec)
		p.seconds -= int16(hundredth)
	} else {
		p = New(year, month, day, hour, min, sec)
		p.seconds += int16(hundredth)
	}
	return
}

func daysDiff(t1, t2 time.Time) (year, month, day, hour, min, sec, hundredth int) {
	duration := t2.Sub(t1)

	hh1, mm1, ss1 := t1.Clock()
	hh2, mm2, ss2 := t2.Clock()

	day = int(duration / (24 * time.Hour))

	hour = hh2 - hh1
	min = mm2 - mm1
	sec = ss2 - ss1
	hundredth = (t2.Nanosecond() - t1.Nanosecond()) / 100000000

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}

	if min < 0 {
		min += 60
		hour--
	}

	if hour < 0 {
		hour += 24
		// no need to reduce day - it's calculated differently.
	}

	// test 16bit storage limit (with 1 fixed decimal place)
	if day > 3276 {
		y1, m1, d1 := t1.Date()
		y2, m2, d2 := t2.Date()
		year = y2 - y1
		month = int(m2 - m1)
		day = d2 - d1
	}

	return
}

// IsZero returns true if applied to a zero-length period.
func (period Period) IsZero() bool {
	return period == Period{}
}

// IsPositive returns true if any field is greater than zero. By design, this also implies that
// all the other fields are greater than or equal to zero.
func (period Period) IsPositive() bool {
	return period.years > 0 || period.months > 0 || period.days > 0 ||
		period.hours > 0 || period.minutes > 0 || period.seconds > 0
}

// IsNegative returns true if any field is negative. By design, this also implies that
// all the other fields are negative or zero.
func (period Period) IsNegative() bool {
	return period.years < 0 || period.months < 0 || period.days < 0 ||
		period.hours < 0 || period.minutes < 0 || period.seconds < 0
}

// Sign returns +1 for positive periods and -1 for negative periods. If the period is zero, it returns zero.
func (period Period) Sign() int {
	if period.IsZero() {
		return 0
	}
	if period.IsNegative() {
		return -1
	}
	return 1
}

// OnlyYMD returns a new Period with only the year, month and day fields. The hour,
// minute and second fields are zeroed.
func (period Period) OnlyYMD() Period {
	return Period{period.years, period.months, period.days, 0, 0, 0}
}

// OnlyHMS returns a new Period with only the hour, minute and second fields. The year,
// month and day fields are zeroed.
func (period Period) OnlyHMS() Period {
	return Period{0, 0, 0, period.hours, period.minutes, period.seconds}
}

// Abs converts a negative period to a positive one.
func (period Period) Abs() Period {
	a, _ := period.absNeg()
	return a
}

func (period Period) absNeg() (Period, bool) {
	if period.IsNegative() {
		return period.Negate(), true
	}
	return period, false
}

func (period Period) condNegate(neg bool) Period {
	if neg {
		return period.Negate()
	}
	return period
}

// Negate changes the sign of the period.
func (period Period) Negate() Period {
	return Period{-period.years, -period.months, -period.days, -period.hours, -period.minutes, -period.seconds}
}

func absInt16(v int16) int16 {
	if v < 0 {
		return -v
	}
	return v
}

// Years gets the whole number of years in the period.
// The result is the number of years and does not include any other field.
func (period Period) Years() int {
	return int(period.YearsFloat())
}

// YearsFloat gets the number of years in the period, including a fraction if any is present.
// The result is the number of years and does not include any other field.
func (period Period) YearsFloat() float32 {
	return float32(period.years) / 10
}

// Months gets the whole number of months in the period.
// The result is the number of months and does not include any other field.
//
// Note that after normalisation, whole multiple of 12 months are added to
// the number of years, so the number of months will be reduced correspondingly.
func (period Period) Months() int {
	return int(period.MonthsFloat())
}

// MonthsFloat gets the number of months in the period.
// The result is the number of months and does not include any other field.
//
// Note that after normalisation, whole multiple of 12 months are added to
// the number of years, so the number of months will be reduced correspondingly.
func (period Period) MonthsFloat() float32 {
	return float32(period.months) / 10
}

// Days gets the whole number of days in the period. This includes the implied
// number of weeks but does not include any other field.
func (period Period) Days() int {
	return int(period.DaysFloat())
}

// DaysFloat gets the number of days in the period. This includes the implied
// number of weeks but does not include any other field.
func (period Period) DaysFloat() float32 {
	return float32(period.days) / 10
}

// Weeks calculates the number of whole weeks from the number of days. If the result
// would contain a fraction, it is truncated.
// The result is the number of weeks and does not include any other field.
//
// Note that weeks are synthetic: they are internally represented using days.
// See ModuloDays(), which returns the number of days excluding whole weeks.
func (period Period) Weeks() int {
	return int(period.days) / 70
}

// WeeksFloat calculates the number of weeks from the number of days.
// The result is the number of weeks and does not include any other field.
func (period Period) WeeksFloat() float32 {
	return float32(period.days) / 70
}

// ModuloDays calculates the whole number of days remaining after the whole number of weeks
// has been excluded.
func (period Period) ModuloDays() int {
	days := absInt16(period.days) % 70
	f := int(days / 10)
	if period.days < 0 {
		return -f
	}
	return f
}

// Hours gets the whole number of hours in the period.
// The result is the number of hours and does not include any other field.
func (period Period) Hours() int {
	return int(period.HoursFloat())
}

// HoursFloat gets the number of hours in the period.
// The result is the number of hours and does not include any other field.
func (period Period) HoursFloat() float32 {
	return float32(period.hours) / 10
}

// Minutes gets the whole number of minutes in the period.
// The result is the number of minutes and does not include any other field.
//
// Note that after normalisation, whole multiple of 60 minutes are added to
// the number of hours, so the number of minutes will be reduced correspondingly.
func (period Period) Minutes() int {
	return int(period.MinutesFloat())
}

// MinutesFloat gets the number of minutes in the period.
// The result is the number of minutes and does not include any other field.
//
// Note that after normalisation, whole multiple of 60 minutes are added to
// the number of hours, so the number of minutes will be reduced correspondingly.
func (period Period) MinutesFloat() float32 {
	return float32(period.minutes) / 10
}

// Seconds gets the whole number of seconds in the period.
// The result is the number of seconds and does not include any other field.
//
// Note that after normalisation, whole multiple of 60 seconds are added to
// the number of minutes, so the number of seconds will be reduced correspondingly.
func (period Period) Seconds() int {
	return int(period.SecondsFloat())
}

// SecondsFloat gets the number of seconds in the period.
// The result is the number of seconds and does not include any other field.
//
// Note that after normalisation, whole multiple of 60 seconds are added to
// the number of minutes, so the number of seconds will be reduced correspondingly.
func (period Period) SecondsFloat() float32 {
	return float32(period.seconds) / 10
}

//-------------------------------------------------------------------------------------------------

// DurationApprox converts a period to the equivalent duration in nanoseconds.
// When the period specifies hours, minutes and seconds only, the result is precise.
// however, when the period specifies years, months and days, it is impossible to be precise
// because the result may depend on knowing date and timezone information, so the duration
// is estimated on the basis of a year being 365.2425 days (as per Gregorian calendar rules)
// and a month being 1/12 of a that; days are all assumed to be 24 hours long.
func (period Period) DurationApprox() time.Duration {
	d, _ := period.Duration()
	return d
}

// Duration converts a period to the equivalent duration in nanoseconds.
// A flag is also returned that is true when the conversion was precise and false otherwise.
//
// When the period specifies hours, minutes and seconds only, the result is precise.
// however, when the period specifies years, months and days, it is impossible to be precise
// because the result may depend on knowing date and timezone information, so the duration
// is estimated on the basis of a year being 365.2425 days as per Gregorian calendar rules)
// and a month being 1/12 of a that; days are all assumed to be 24 hours long.
func (period Period) Duration() (time.Duration, bool) {
	// remember that the fields are all fixed-point 1E1
	tdE6 := time.Duration(totalDaysApproxE7(period) * 8640)
	stE3 := totalSecondsE3(period)
	return tdE6*time.Microsecond + stE3*time.Millisecond, tdE6 == 0
}

func totalSecondsE3(period Period) time.Duration {
	// remember that the fields are all fixed-point 1E1
	// and these are divided by 1E1
	hhE3 := time.Duration(period.hours) * 360000
	mmE3 := time.Duration(period.minutes) * 6000
	ssE3 := time.Duration(period.seconds) * 100
	return hhE3 + mmE3 + ssE3
}

func totalDaysApproxE7(period Period) int64 {
	// remember that the fields are all fixed-point 1E1
	ydE6 := int64(period.years) * (daysPerYearE4 * 100)
	mdE6 := int64(period.months) * daysPerMonthE6
	ddE6 := int64(period.days) * oneE6
	return ydE6 + mdE6 + ddE6
}

// TotalDaysApprox gets the approximate total number of days in the period. The approximation assumes
// a year is 365.2425 days as per Gregorian calendar rules) and a month is 1/12 of that. Whole
// multiples of 24 hours are also included in the calculation.
func (period Period) TotalDaysApprox() int {
	pn := period.Normalise(false)
	tdE6 := totalDaysApproxE7(pn)
	hE6 := (int64(pn.hours) * oneE6) / 24
	return int((tdE6 + hE6) / oneE7)
}

// TotalMonthsApprox gets the approximate total number of months in the period. The days component
// is included by approximation, assuming a year is 365.2425 days (as per Gregorian calendar rules)
// and a month is 1/12 of that. Whole multiples of 24 hours are also included in the calculation.
func (period Period) TotalMonthsApprox() int {
	pn := period.Normalise(false)
	mE1 := int64(pn.years)*12 + int64(pn.months)
	hE1 := int64(pn.hours) / 24
	dE1 := ((int64(pn.days) + hE1) * oneE6) / daysPerMonthE6
	return int((mE1 + dE1) / 10)
}

// Normalise attempts to simplify the fields. It operates in either precise or imprecise mode.
//
// Because the number of hours per day is imprecise (due to daylight savings etc), and because
// the number of days per month is variable in the Gregorian calendar, there is a reluctance
// to transfer time to or from the days element, or to transfer days to or from the months
// element. To give control over this, there are two modes.
//
// In precise mode:
// Multiples of 60 seconds become minutes.
// Multiples of 60 minutes become hours.
// Multiples of 12 months become years.
//
// Additionally, in imprecise mode:
// Multiples of 24 hours become days.
// Multiples of approx. 30.4 days become months.
//
// Note that leap seconds are disregarded: every minute is assumed to have 60 seconds.
func (period Period) Normalise(precise bool) Period {
	n, _ := period.toPeriod64("").normalise64(precise).toPeriod()
	return n
}

// Simplify applies some heuristic simplifications with the objective of reducing the number
// of non-zero fields and thus making the rendered form simpler. It should be applied to
// a normalised period, otherwise the results may be unpredictable.
//
// Note that months and days are never combined, due to the variability of month lengths.
// Days and hours are only combined when imprecise behaviour is selected; this is due to
// daylight savings transitions, during which there are more than or fewer than 24 hours
// per day.
//
// The following transformation rules are applied in order:
//
// * P1YnM becomes 12+n months for 0 < n <= 6
// * P1DTnH becomes 24+n hours for 0 < n <= 6 (unless precise is true)
// * PT1HnM becomes 60+n minutes for 0 < n <= 10
// * PT1MnS becomes 60+n seconds for 0 < n <= 10
//
// At each step, if a fraction exists and would affect the calculation, the transformations
// stop. Also, when not precise,
//
// * for periods of at least ten years, month proper fractions are discarded
// * for periods of at least a year, day proper fractions are discarded
// * for periods of at least a month, hour proper fractions are discarded
// * for periods of at least a day, minute proper fractions are discarded
// * for periods of at least an hour, second proper fractions are discarded
//
// The thresholds can be set using the varargs th parameter. By default, the thresholds a,
// b, c, d are 6 months, 6 hours, 10 minutes, 10 seconds respectively as listed in the rules
// above.
//
// * No thresholds is equivalent to 6, 6, 10, 10.
// * A single threshold a is equivalent to a, a, a, a.
// * Two thresholds a, b are equivalent to a, a, b, b.
// * Three thresholds a, b, c are equivalent to a, b, c, c.
// * Four thresholds a, b, c, d are used as provided.
//
func (period Period) Simplify(precise bool, th ...int) Period {
	switch len(th) {
	case 0:
		return period.doSimplify(precise, 60, 60, 100, 100)
	case 1:
		return period.doSimplify(precise, int16(th[0]*10), int16(th[0]*10), int16(th[0]*10), int16(th[0]*10))
	case 2:
		return period.doSimplify(precise, int16(th[0]*10), int16(th[0]*10), int16(th[1]*10), int16(th[1]*10))
	case 3:
		return period.doSimplify(precise, int16(th[0]*10), int16(th[1]*10), int16(th[2]*10), int16(th[2]*10))
	default:
		return period.doSimplify(precise, int16(th[0]*10), int16(th[1]*10), int16(th[2]*10), int16(th[3]*10))
	}
}

func (period Period) doSimplify(precise bool, a, b, c, d int16) Period {
	if period.years%10 != 0 {
		return period
	}

	ap, neg := period.absNeg()

	// single year is dropped if there are some months
	if ap.years == 10 &&
		0 < ap.months && ap.months <= a &&
		ap.days == 0 {
		ap.months += 120
		ap.years = 0
	}

	if ap.months%10 != 0 {
		// month fraction is dropped for periods of at least ten years (1:120)
		months := ap.months / 10
		if !precise && ap.years >= 100 && months == 0 {
			ap.months = 0
		}
		return ap.condNegate(neg)
	}

	if ap.days%10 != 0 {
		// day fraction is dropped for periods of at least a year (1:365)
		days := ap.days / 10
		if !precise && (ap.years > 0 || ap.months >= 120) && days == 0 {
			ap.days = 0
		}
		return ap.condNegate(neg)
	}

	if !precise && ap.days == 10 &&
		ap.years == 0 &&
		ap.months == 0 &&
		0 < ap.hours && ap.hours <= b {
		ap.hours += 240
		ap.days = 0
	}

	if ap.hours%10 != 0 {
		// hour fraction is dropped for periods of at least a month (1:720)
		hours := ap.hours / 10
		if !precise && (ap.years > 0 || ap.months > 0 || ap.days >= 300) && hours == 0 {
			ap.hours = 0
		}
		return ap.condNegate(neg)
	}

	if ap.hours == 10 &&
		0 < ap.minutes && ap.minutes <= c {
		ap.minutes += 600
		ap.hours = 0
	}

	if ap.minutes%10 != 0 {
		// minute fraction is dropped for periods of at least a day (1:1440)
		minutes := ap.minutes / 10
		if !precise && (ap.years > 0 || ap.months > 0 || ap.days > 0 || ap.hours >= 240) && minutes == 0 {
			ap.minutes = 0
		}
		return ap.condNegate(neg)
	}

	if ap.minutes == 10 &&
		ap.hours == 0 &&
		0 < ap.seconds && ap.seconds <= d {
		ap.seconds += 600
		ap.minutes = 0
	}

	if ap.seconds%10 != 0 {
		// second fraction is dropped for periods of at least an hour (1:3600)
		seconds := ap.seconds / 10
		if !precise && (ap.years > 0 || ap.months > 0 || ap.days > 0 || ap.hours > 0 || ap.minutes >= 600) && seconds == 0 {
			ap.seconds = 0
		}
	}

	return ap.condNegate(neg)
}

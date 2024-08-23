// Copyright 2015 Rick Beton. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package period

import (
	"fmt"
	"io"
	"strings"

	"github.com/rickb777/plural"
)

// Format converts the period to human-readable form using the default localisation.
// Multiples of 7 days are shown as weeks.
func (period Period) Format() string {
	return period.FormatWithPeriodNames(PeriodYearNames, PeriodMonthNames, PeriodWeekNames, PeriodDayNames, PeriodHourNames, PeriodMinuteNames, PeriodSecondNames)
}

// FormatWithoutWeeks converts the period to human-readable form using the default localisation.
// Multiples of 7 days are not shown as weeks.
func (period Period) FormatWithoutWeeks() string {
	return period.FormatWithPeriodNames(PeriodYearNames, PeriodMonthNames, plural.Plurals{}, PeriodDayNames, PeriodHourNames, PeriodMinuteNames, PeriodSecondNames)
}

// FormatWithPeriodNames converts the period to human-readable form in a localisable way.
func (period Period) FormatWithPeriodNames(yearNames, monthNames, weekNames, dayNames, hourNames, minNames, secNames plural.Plurals) string {
	period = period.Abs()

	parts := make([]string, 0)
	parts = appendNonBlank(parts, yearNames.FormatFloat(float10(period.years)))
	parts = appendNonBlank(parts, monthNames.FormatFloat(float10(period.months)))

	if period.days > 0 || (period.IsZero()) {
		if len(weekNames) > 0 {
			weeks := period.days / 70
			mdays := period.days % 70
			//fmt.Printf("%v %#v - %d %d\n", period, period, weeks, mdays)
			if weeks > 0 {
				parts = appendNonBlank(parts, weekNames.FormatInt(int(weeks)))
			}
			if mdays > 0 || weeks == 0 {
				parts = appendNonBlank(parts, dayNames.FormatFloat(float10(mdays)))
			}
		} else {
			parts = appendNonBlank(parts, dayNames.FormatFloat(float10(period.days)))
		}
	}
	parts = appendNonBlank(parts, hourNames.FormatFloat(float10(period.hours)))
	parts = appendNonBlank(parts, minNames.FormatFloat(float10(period.minutes)))
	parts = appendNonBlank(parts, secNames.FormatFloat(float10(period.seconds)))

	return strings.Join(parts, ", ")
}

func appendNonBlank(parts []string, s string) []string {
	if s == "" {
		return parts
	}
	return append(parts, s)
}

// PeriodDayNames provides the English default format names for the days part of the period.
// This is a sequence of plurals where the first match is used, otherwise the last one is used.
// The last one must include a "%v" placeholder for the number.
var PeriodDayNames = plural.FromZero("%v days", "%v day", "%v days")

// PeriodWeekNames is as for PeriodDayNames but for weeks.
var PeriodWeekNames = plural.FromZero("", "%v week", "%v weeks")

// PeriodMonthNames is as for PeriodDayNames but for months.
var PeriodMonthNames = plural.FromZero("", "%v month", "%v months")

// PeriodYearNames is as for PeriodDayNames but for years.
var PeriodYearNames = plural.FromZero("", "%v year", "%v years")

// PeriodHourNames is as for PeriodDayNames but for hours.
var PeriodHourNames = plural.FromZero("", "%v hour", "%v hours")

// PeriodMinuteNames is as for PeriodDayNames but for minutes.
var PeriodMinuteNames = plural.FromZero("", "%v minute", "%v minutes")

// PeriodSecondNames is as for PeriodDayNames but for seconds.
var PeriodSecondNames = plural.FromZero("", "%v second", "%v seconds")

// String converts the period to ISO-8601 form.
func (period Period) String() string {
	return period.toPeriod64("").String()
}

func (p64 period64) String() string {
	if p64 == (period64{}) {
		return "P0D"
	}

	buf := &strings.Builder{}
	if p64.neg {
		buf.WriteByte('-')
	}

	buf.WriteByte('P')

	writeField64(buf, p64.years, byte(Year))
	writeField64(buf, p64.months, byte(Month))

	if p64.days != 0 {
		if p64.days%70 == 0 {
			writeField64(buf, p64.days/7, byte(Week))
		} else {
			writeField64(buf, p64.days, byte(Day))
		}
	}

	if p64.hours != 0 || p64.minutes != 0 || p64.seconds != 0 {
		buf.WriteByte('T')
	}

	writeField64(buf, p64.hours, byte(Hour))
	writeField64(buf, p64.minutes, byte(Minute))
	writeField64(buf, p64.seconds, byte(Second))

	return buf.String()
}

func writeField64(w io.Writer, field int64, designator byte) {
	if field != 0 {
		if field%10 != 0 {
			fmt.Fprintf(w, "%g", float32(field)/10)
		} else {
			fmt.Fprintf(w, "%d", field/10)
		}
		w.(io.ByteWriter).WriteByte(designator)
	}
}

func float10(v int16) float32 {
	return float32(v) / 10
}

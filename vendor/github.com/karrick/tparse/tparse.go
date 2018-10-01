package tparse

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Parse will return the time value corresponding to the specified layout and value.  It also parses
// floating point and integer epoch values.
func Parse(layout, value string) (time.Time, error) {
	return ParseWithMap(layout, value, nil)
}

// ParseNow will return the time value corresponding to the specified layout and value.  It also
// parses floating point and integer epoch values.  It recognizes the special string `now` and
// replaces that with the time ParseNow is called.  This allows a suffix adding or subtracting
// various values from the base time.  For instance, ParseNow(time.ANSIC, "now+1d") will return a
// time corresponding to 24 hours from the moment the function is invoked.
//
// In addition to the duration abbreviations recognized by time.ParseDuration, it recognizes various
// tokens for days, weeks, months, and years.
//
//	package main
//
//	import (
//		"fmt"
//		"os"
//		"time"
//
//		"github.com/karrick/tparse"
//	)
//
//	func main() {
//		actual, err := tparse.ParseNow(time.RFC3339, "now+1d3w4mo7y6h4m")
//		if err != nil {
//			fmt.Fprintf(os.Stderr, "error: %s\n", err)
//			os.Exit(1)
//		}
//
//		fmt.Printf("time is: %s\n", actual)
//	}
func ParseNow(layout, value string) (time.Time, error) {
	if strings.HasPrefix(value, "now") {
		return AddDuration(time.Now(), value[3:])
	}
	return ParseWithMap(layout, value, nil)
}

// ParseWithMap will return the time value corresponding to the specified layout and value.  It also
// parses floating point and integer epoch values.  It accepts a map of strings to time.Time values,
// and if the value string starts with one of the keys in the map, it replaces the string with the
// corresponding time.Time value.
//
//     package main
//
//     import (
//         "fmt"
//         "os"
//         "time"
//         "github.com/karrick/tparse"
//     )
//
//     func main() {
//         m := make(map[string]time.Time)
//         m["end"] = time.Now()
//
//         start, err := tparse.ParseWithMap(time.RFC3339, "end-12h", m)
//         if err != nil {
//             fmt.Fprintf(os.Stderr, "error: %s\n", err)
//             os.Exit(1)
//         }
//
//         fmt.Printf("start: %s; end: %s\n", start, end)
//     }
func ParseWithMap(layout, value string, dict map[string]time.Time) (time.Time, error) {
	// find longest matching key in dict
	var matchKey string
	for k := range dict {
		if strings.HasPrefix(value, k) && len(k) > len(matchKey) {
			matchKey = k
		}
	}
	if len(matchKey) > 0 {
		return AddDuration(dict[matchKey], value[len(matchKey):])
	}

	// takes about 90ns even if fails
	if epoch, err := strconv.ParseFloat(value, 64); err == nil && epoch >= 0 {
		trunc := math.Trunc(epoch)
		nanos := fractionToNanos(epoch - trunc)
		return time.Unix(int64(trunc), int64(nanos)), nil
	}

	return time.Parse(layout, value)
}

func fractionToNanos(fraction float64) int64 {
	return int64(fraction * float64(time.Second/time.Nanosecond))
}

var unitMap = map[string]float64{
	"ns":      float64(time.Nanosecond),
	"us":      float64(time.Microsecond),
	"µs":      float64(time.Microsecond), // U+00B5 = micro symbol
	"μs":      float64(time.Microsecond), // U+03BC = Greek letter mu
	"ms":      float64(time.Millisecond),
	"s":       float64(time.Second),
	"sec":     float64(time.Second),
	"second":  float64(time.Second),
	"seconds": float64(time.Second),
	"m":       float64(time.Minute),
	"min":     float64(time.Minute),
	"minute":  float64(time.Minute),
	"minutes": float64(time.Minute),
	"h":       float64(time.Hour),
	"hr":      float64(time.Hour),
	"hour":    float64(time.Hour),
	"hours":   float64(time.Hour),
	"d":       float64(time.Hour * 24),
	"day":     float64(time.Hour * 24),
	"days":    float64(time.Hour * 24),
	"w":       float64(time.Hour * 24 * 7),
	"week":    float64(time.Hour * 24 * 7),
	"weeks":   float64(time.Hour * 24 * 7),
	"wk":      float64(time.Hour * 24 * 7),
}

// AddDuration parses the duration string, and adds the calculated duration value to the provided
// base time. On error, it returns the base time and the error.
//
// Like `time.ParseDuration`, this accepts multiple fractional scalars, so "now+1.5days-3.21hours"
// is evaluated properly.
//
// The following tokens may be used to specify the respective unit of time:
//
// * Nanosecond: ns
// * Microsecond: us, µs (U+00B5 = micro symbol), μs (U+03BC = Greek letter mu)
// * Millisecond: ms
// * Second: s, sec, second, seconds
// * Minute: m, min, minute, minutes
// * Hour: h, hr, hour, hours
// * Day: d, day, days
// * Week: w, wk, week, weeks
// * Month: mo, mon, month, months
// * Year: y, yr, year, years
//
//	package main
//
//	import (
//		"fmt"
//		"os"
//		"time"
//
//		"github.com/karrick/tparse"
//	)
//
//	func main() {
//              now := time.Now()
//		another, err := tparse.AddDuration(now, "now+1d3w4mo-7y6h4m")
//		if err != nil {
//			fmt.Fprintf(os.Stderr, "error: %s\n", err)
//			os.Exit(1)
//		}
//
//		fmt.Printf("time is: %s\n", another)
//	}
func AddDuration(base time.Time, s string) (time.Time, error) {
	if len(s) == 0 {
		return base, nil
	}
	var isNegative bool
	var exp, whole, fraction int64
	var number, totalYears, totalMonths, totalDays, totalDuration float64

	for s != "" {
		// consume possible sign
		if s[0] == '+' {
			if len(s) == 1 {
				return base, fmt.Errorf("cannot parse sign without digits: '+'")
			}
			isNegative = false
			s = s[1:]
		} else if s[0] == '-' {
			if len(s) == 1 {
				return base, fmt.Errorf("cannot parse sign without digits: '-'")
			}
			isNegative = true
			s = s[1:]
		}
		// consume digits
		var done bool
		for !done {
			c := s[0]
			switch {
			case c >= '0' && c <= '9':
				d := int64(c - '0')
				if exp > 0 {
					exp++
					fraction = 10*fraction + d
				} else {
					whole = 10*whole + d
				}
				s = s[1:]
			case c == '.':
				if exp > 0 {
					return base, fmt.Errorf("invalid floating point number format: two decimal points found")
				}
				exp = 1
				fraction = 0
				s = s[1:]
			default:
				done = true
			}
		}
		// adjust number
		number = float64(whole)
		if exp > 0 {
			number += float64(fraction) * math.Pow(10, float64(1-exp))
		}
		if isNegative {
			number *= -1
		}
		// find end of unit
		var i int
		for ; i < len(s) && s[i] != '+' && s[i] != '-' && (s[i] < '0' || s[i] > '9'); i++ {
			// identifier bytes: no-op
		}
		unit := s[:i]
		// fmt.Printf("number: %f; unit: %q\n", number, unit)
		if duration, ok := unitMap[unit]; ok {
			totalDuration += number * duration
		} else {
			switch unit {
			case "mo", "mon", "month", "months":
				totalMonths += number
			case "y", "yr", "year", "years":
				totalYears += number
			default:
				return base, fmt.Errorf("unknown unit in duration: %q", unit)
			}
		}

		s = s[i:]
		whole = 0
	}
	if totalYears != 0 {
		whole := math.Trunc(totalYears)
		fraction := totalYears - whole
		totalYears = whole
		totalMonths += 12 * fraction
	}
	if totalMonths != 0 {
		whole := math.Trunc(totalMonths)
		fraction := totalMonths - whole
		totalMonths = whole
		totalDays += 30 * fraction
	}
	if totalDays != 0 {
		whole := math.Trunc(totalDays)
		fraction := totalDays - whole
		totalDays = whole
		totalDuration += (fraction * 24.0 * float64(time.Hour))
	}
	if totalYears != 0 || totalMonths != 0 || totalDays != 0 {
		base = base.AddDate(int(totalYears), int(totalMonths), int(totalDays))
	}
	if totalDuration != 0 {
		base = base.Add(time.Duration(totalDuration))
	}
	return base, nil
}

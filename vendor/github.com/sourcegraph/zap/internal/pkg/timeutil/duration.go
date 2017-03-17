package timeutil

import (
	"fmt"
	"math"
	"time"
)

// FormatDuration returns a human-readable description of the
// duration.
func FormatDuration(duration time.Duration) string {
	if duration.Hours() < 24 {
		switch {
		case duration.Hours() >= 1:
			return pluralize(0, duration.Hours(), "hour")
		case duration.Minutes() >= 1:
			return pluralize(0, duration.Minutes(), "minute")
		case duration.Seconds() >= 1:
			return pluralize(0, duration.Seconds(), "second")
		default:
			return pluralize(0, duration.Seconds()*1000, "millisecond")
		}
	} else {
		switch {
		case duration.Hours() >= 8760:
			return pluralize(int(duration.Hours()/8760), 0, "year")
		case duration.Hours() >= 730:
			return pluralize(int(duration.Hours()/730), 0, "month")
		case duration.Hours() >= 168:
			return pluralize(int(duration.Hours()/168), 0, "week")
		default:
			return pluralize(int(duration.Hours()/24), 0, "day")
		}
	}
}

func pluralize(v int, f float64, noun string) string {
	if f != 0 {
		v = int(math.Floor(f + 0.5)) // round
	}
	if v == 0 || v >= 2 {
		return fmt.Sprintf("%d %ss", v, noun) // plural
	}
	return fmt.Sprintf("%d %s", v, noun) // no plural
}

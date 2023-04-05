package timeutil

import (
	"strconv"
	"strings"
	"time"
)

func FormatDuration(duration time.Duration) string {
	totalSeconds := int(duration.Seconds())
	seconds := totalSeconds % 60
	minutes := totalSeconds / 60 % 60
	hours := totalSeconds / 60 / 60 % 24
	days := totalSeconds / 60 / 60 / 24

	formatted := ""
	if days > 0 {
		formatted += strconv.Itoa(int(days)) + "d "
	}
	if hours > 0 {
		formatted += strconv.Itoa(int(hours)) + "h "
	}
	if minutes > 0 {
		formatted += strconv.Itoa(int(minutes)) + "m "
	}
	if seconds > 0 {
		formatted += strconv.Itoa(int(seconds)) + "s "
	}
	return strings.Trim(formatted, " ")
}

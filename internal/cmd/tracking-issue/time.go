package main

import (
	"time"
)

// now returns the current time for relative formatting. This is overwritten
// during tests to ensure that our output can be byte-for-byte compared against
// the golden output file.
var now = time.Now

// formatTime will return a string containing the number of days since the
// given time.
func formatTime(t time.Time) string {
	return t.Format(time.DateOnly)
}

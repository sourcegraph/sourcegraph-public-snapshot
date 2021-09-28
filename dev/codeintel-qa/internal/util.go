package internal

import "time"

// TImeSince returns the time since the given duration rounded down to the nearest second.
func TimeSince(start time.Time) time.Duration {
	return time.Since(start) / time.Second * time.Second
}

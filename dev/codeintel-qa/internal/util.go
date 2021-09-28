package internal

import "time"

func TimeSince(start time.Time) time.Duration {
	return time.Since(start) / time.Second * time.Second
}

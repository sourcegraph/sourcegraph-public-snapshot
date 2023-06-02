package codygateway

import (
	"time"
)

// ConcurrentLimitConfig is the configuration for the concurrent requests limit.
type ConcurrentLimitConfig struct {
	// Percentage is the percentage of the daily rate limit to be used to compute the
	// concurrent limit.
	Percentage float32
	// Interval is the time interval of the limit bucket.
	Interval time.Duration
}

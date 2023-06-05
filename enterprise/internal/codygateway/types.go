package codygateway

import (
	"time"
)

// ActorConcurrentLimitConfig is the configuration for the concurrent requests
// limit of an actor.
type ActorConcurrentLimitConfig struct {
	// Percentage is the percentage of the daily rate limit to be used to compute the
	// concurrent limit.
	Percentage float32
	// Interval is the time interval of the limit bucket.
	Interval time.Duration
}

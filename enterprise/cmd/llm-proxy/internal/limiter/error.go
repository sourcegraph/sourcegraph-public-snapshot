package limiter

import (
	"fmt"
	"time"
)

type RateLimitExceededError struct {
	Limit      int
	Used       int
	RetryAfter time.Time
}

func (e RateLimitExceededError) Error() string {
	return fmt.Sprintf("you exceeded the rate limit for completions, only %d requests are allowed per day at the moment to ensure the service stays functional. Current usage: %d. Retry after %s", e.Limit, e.Used, e.RetryAfter.Truncate(time.Second))
}

type NoAccessError struct{}

func (e NoAccessError) Error() string {
	return fmt.Sprintf("completions access has not been granted")
}

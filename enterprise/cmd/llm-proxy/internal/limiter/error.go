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
	return fmt.Sprintf("you exceeded the rate limit for completions. Current usage: %d out of %d requests. Retry after %s",
		e.Used, e.Limit, e.RetryAfter.Truncate(time.Second))
}

type NoAccessError struct{}

func (e NoAccessError) Error() string {
	return "completions access has not been granted"
}

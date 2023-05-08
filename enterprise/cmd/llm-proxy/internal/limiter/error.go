package limiter

import (
	"fmt"
	"net/http"
	"strconv"
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

func (e RateLimitExceededError) WriteResponse(w http.ResponseWriter) {
	// Rate limit exceeded, write well known headers and return correct status code.
	w.Header().Set("x-ratelimit-limit", strconv.Itoa(e.Limit))
	w.Header().Set("x-ratelimit-remaining", strconv.Itoa(max(e.Limit-e.Used, 0)))
	w.Header().Set("retry-after", e.RetryAfter.Format(time.RFC3339))
	http.Error(w, e.Error(), http.StatusTooManyRequests)
}

type NoAccessError struct{}

func (e NoAccessError) Error() string {
	return "completions access has not been granted"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

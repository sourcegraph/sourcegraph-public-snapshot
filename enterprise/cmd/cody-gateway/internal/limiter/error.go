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

// Error generates a simple string that is fairly static for use in logging.
// This helps with categorizing errors. For more detailed output use Summary().
func (e RateLimitExceededError) Error() string { return "rate limit exceeded" }

func (e RateLimitExceededError) Summary() string {
	return fmt.Sprintf("You have exceeded your rate limit. Current usage: %d out of %d requests. Retry after %s",
		e.Used, e.Limit, e.RetryAfter.Truncate(time.Second))
}

func (e RateLimitExceededError) WriteResponse(w http.ResponseWriter) {
	// Rate limit exceeded, write well known headers and return correct status code.
	w.Header().Set("x-ratelimit-limit", strconv.Itoa(e.Limit))
	w.Header().Set("x-ratelimit-remaining", strconv.Itoa(max(e.Limit-e.Used, 0)))
	w.Header().Set("retry-after", e.RetryAfter.Format(time.RFC1123))
	// Use Summary instead of Error for more informative text
	http.Error(w, e.Summary(), http.StatusTooManyRequests)
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

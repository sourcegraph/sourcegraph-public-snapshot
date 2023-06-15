package limiter

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type RateLimitExceededError struct {
	Limit      int64
	RetryAfter time.Time
}

// Error generates a simple string that is fairly static for use in logging.
// This helps with categorizing errors. For more detailed output use Summary().
func (e RateLimitExceededError) Error() string { return "rate limit exceeded" }

func (e RateLimitExceededError) Summary() string {
	return fmt.Sprintf("you have exceeded the rate limit of %d requests. Retry after %s",
		e.Limit, e.RetryAfter.Truncate(time.Second))
}

func (e RateLimitExceededError) WriteResponse(w http.ResponseWriter) {
	// Rate limit exceeded, write well known headers and return correct status code.
	w.Header().Set("x-ratelimit-limit", strconv.FormatInt(e.Limit, 10))
	w.Header().Set("x-ratelimit-remaining", "0")
	w.Header().Set("retry-after", e.RetryAfter.Format(time.RFC1123))
	// Use Summary instead of Error for more informative text
	http.Error(w, e.Summary(), http.StatusTooManyRequests)
}

type NoAccessError struct{}

func (e NoAccessError) Error() string {
	return "completions access has not been granted"
}

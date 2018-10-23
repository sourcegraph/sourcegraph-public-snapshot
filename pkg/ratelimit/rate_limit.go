package ratelimit

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Monitor monitors an external service's rate limit based on the X-RateLimit-Remaining or RateLimit-Remaining
// headers. It supports both GitHub's and GitLab's APIs.
//
// It is intended to be embedded in an API client struct.
type Monitor struct {
	HeaderPrefix string // "X-" (GitHub) or "" (GitLab)

	mu        sync.Mutex
	known     bool
	limit     int       // last RateLimit-Limit HTTP response header value
	remaining int       // last RateLimit-Remaining HTTP response header value
	reset     time.Time // last RateLimit-Remaining HTTP response header value
}

// Get reports the client's rate limit status (as of the last API response it received).
func (c *Monitor) Get() (remaining int, reset time.Duration, known bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.known {
		return 0, 0, false
	}
	return c.remaining, time.Until(c.reset), true
}

// RecommendedWaitForBackgroundOp returns the recommended wait time before performing a periodic
// background operation with the given rate limit cost. It takes the rate limit information from the last API
// request into account.
//
// For example, suppose the rate limit resets to 5,000 points in 30 minutes and currently 1,500 points remain. You
// want to perform a cost-500 operation. Only 4 more cost-500 operations are allowed in the next 30 minutes (per
// the rate limit):
//
//                          -500         -500         -500
//         Now   |------------*------------*------------*------------| 30 min from now
//   Remaining  1500         1000         500           0           5000 (reset)
//
// Assuming no other operations are being performed (that count against the rate limit), the recommended wait would
// be 7.5 minutes (30 minutes / 4), so that the operations are evenly spaced out.
//
// A small constant additional wait is added to account for other simultaneous operations and clock
// out-of-synchronization.
//
// See https://developer.github.com/v4/guides/resource-limitations/#rate-limit.
func (c *Monitor) RecommendedWaitForBackgroundOp(cost int) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.known {
		return 0
	}

	// If our rate limit info is out of date, assume it was reset.
	limitRemaining := float64(c.remaining)
	resetAt := c.reset
	if time.Now().After(c.reset) {
		limitRemaining = float64(c.limit)
		resetAt = time.Now().Add(1 * time.Hour)
	}

	// Be conservative.
	limitRemaining = float64(limitRemaining) * 0.8
	timeRemaining := time.Until(resetAt) + 3*time.Minute

	n := limitRemaining / float64(cost) // number of times this op can run before exhausting rate limit
	if n < 1 {
		return timeRemaining
	}
	if n > 500 {
		return 0
	}
	if n > 250 {
		return 200 * time.Millisecond
	}
	// N is limitRemaining / cost. timeRemaining / N is thus
	// timeRemaining / (limitRemaining / cost). However, time.Duration is
	// an integer type, and drops fractions. We get more accurate
	// calculations computing this the other way around:
	return timeRemaining * time.Duration(cost) / time.Duration(limitRemaining)
}

// Update updates the monitor's rate limit information based on the HTTP response headers.
func (c *Monitor) Update(h http.Header) {
	if cached := h.Get("X-From-Cache"); cached != "" {
		// Cached responses have stale RateLimit headers.
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// See https://developer.github.com/v3/#rate-limiting.
	limit, err := strconv.Atoi(h.Get(c.HeaderPrefix + "RateLimit-Limit"))
	if err != nil {
		c.known = false
		return
	}
	remaining, err := strconv.Atoi(h.Get(c.HeaderPrefix + "RateLimit-Remaining"))
	if err != nil {
		c.known = false
		return
	}
	resetAtSeconds, err := strconv.ParseInt(h.Get(c.HeaderPrefix+"RateLimit-Reset"), 10, 64)
	if err != nil {
		c.known = false
		return
	}
	c.known = true
	c.limit = limit
	c.remaining = remaining
	c.reset = time.Unix(resetAtSeconds, 0)
}

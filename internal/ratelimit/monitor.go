package ratelimit

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// DefaultMonitorRegistry is the default global rate limit monitor registry. It will hold rate limit mappings
// for each instance of our services.
var DefaultMonitorRegistry = NewMonitorRegistry()

// NewMonitorRegistry creates a new empty registry.
func NewMonitorRegistry() *MonitorRegistry {
	return &MonitorRegistry{
		monitors: make(map[string]*Monitor),
	}
}

// MonitorRegistry keeps a mapping of external service URL to *Monitor.
type MonitorRegistry struct {
	mu sync.Mutex
	// Monitor per code host / token tuple, keys are the normalized base URL for a
	// code host, plus the token hash.
	monitors map[string]*Monitor
}

// GetOrSet fetches the rate limit monitor associated with the given code host /
// token tuple and an optional resource key. If none has been configured yet, the
// provided monitor will be set.
func (r *MonitorRegistry) GetOrSet(baseURL, authHash, resource string, monitor *Monitor) *Monitor {
	baseURL = normaliseURL(baseURL)
	key := baseURL + ":" + authHash
	if len(resource) > 0 {
		key = key + ":" + resource
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.monitors[key]; !ok {
		r.monitors[key] = monitor
	}
	return r.monitors[key]
}

// Count returns the total number of rate limiters in the registry
func (r *MonitorRegistry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.monitors)
}

// MetricsCollector is used so that we can inject metric collection functions for
// difference monitor configurations.
type MetricsCollector struct {
	Remaining    func(n float64)
	WaitDuration func(n time.Duration)
}

// Monitor monitors an external service's rate limit based on the X-RateLimit-Remaining or RateLimit-Remaining
// headers. It supports both GitHub's and GitLab's APIs.
//
// It is intended to be embedded in an API client struct.
type Monitor struct {
	HeaderPrefix string // "X-" (GitHub and Azure DevOps) or "" (GitLab)

	mu        sync.Mutex
	known     bool
	limit     int               // last RateLimit-Limit HTTP response header value
	remaining int               // last RateLimit-Remaining HTTP response header value
	reset     time.Time         // last RateLimit-Remaining HTTP response header value
	retry     time.Time         // deadline based on Retry-After HTTP response header value
	collector *MetricsCollector // metrics collector

	clock func() time.Time
}

// Get reports the client's rate limit status (as of the last API response it received).
func (c *Monitor) Get() (remaining int, reset, retry time.Duration, known bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	return c.remaining, c.reset.Sub(now), c.retry.Sub(now), c.known
}

// TODO(keegancsmith) Update RecommendedWaitForBackgroundOp to work with other
// rate limits. Such as:
//
// - GitHub search 30req/m
// - GitLab.com 600 req/h

// RecommendedWaitForBackgroundOp returns the recommended wait time before performing a periodic
// background operation with the given rate limit cost. It takes the rate limit information from the last API
// request into account.
//
// For example, suppose the rate limit resets to 5,000 points in 30 minutes and currently 1,500 points remain. You
// want to perform a cost-500 operation. Only 4 more cost-500 operations are allowed in the next 30 minutes (per
// the rate limit):
//
//	                       -500         -500         -500
//	      Now   |------------*------------*------------*------------| 30 min from now
//	Remaining  1500         1000         500           0           5000 (reset)
//
// Assuming no other operations are being performed (that count against the rate limit), the recommended wait would
// be 7.5 minutes (30 minutes / 4), so that the operations are evenly spaced out.
//
// A small constant additional wait is added to account for other simultaneous operations and clock
// out-of-synchronization.
//
// See https://developer.github.com/v4/guides/resource-limitations/#rate-limit.
func (c *Monitor) RecommendedWaitForBackgroundOp(cost int) (timeRemaining time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.collector != nil && c.collector.WaitDuration != nil {
		defer func() {
			c.collector.WaitDuration(timeRemaining)
		}()
	}

	now := c.now()
	if !c.retry.IsZero() {
		if remaining := c.retry.Sub(now); remaining > 0 {
			return remaining
		}
		c.retry = time.Time{}
	}

	if !c.known {
		return 0
	}

	// If our rate limit info is out of date, assume it was reset.
	limitRemaining := float64(c.remaining)
	resetAt := c.reset
	if now.After(c.reset) {
		limitRemaining = float64(c.limit)
		resetAt = now.Add(1 * time.Hour)
	}

	// Be conservative.
	limitRemaining *= 0.8
	timeRemaining = resetAt.Sub(now) + 3*time.Minute

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

func (c *Monitor) calcRateLimitWaitTime(cost int) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.retry.IsZero() {
		if timeRemaining := c.retry.Sub(c.now()); timeRemaining > 0 {
			// Unlock before sleeping
			return timeRemaining
		}
	}

	// If the external rate limit is unknown,
	// or if there are still enough remaining tokens,
	// or if the cost is greater than the actual rate limit (in which case there will never be enough tokens),
	// we don't wait.
	if !c.known || c.remaining >= cost || cost > c.limit {
		return time.Duration(0)
	}

	// If the rate limit reset is still in the future, we wait until the limit is reset.
	// If it is in the past, the rate limit is outdated and we don't need to wait.
	if timeRemaining := c.reset.Sub(c.now()); timeRemaining > 0 {
		// Unlock before sleeping
		return timeRemaining
	}

	return time.Duration(0)
}

// WaitForRateLimit determines whether or not an external rate limit is being applied
// and sleeps an amount of time recommended by the external rate limiter.
// It returns true if rate limiting was applying, and false if not.
// This can be used to determine whether or not a request should be retried.
//
// The cost parameter can be used to check for a minimum number of available rate limit tokens.
// For normal REST requests, this can usually be set to 1. For GraphQL requests, rate limit costs
// can be more expensive and a different cost can be used. If there aren't enough rate limit
// tokens available, then the function will sleep until the tokens reset.
func (c *Monitor) WaitForRateLimit(ctx context.Context, cost int) bool {
	sleepDuration := c.calcRateLimitWaitTime(cost)

	if sleepDuration == 0 {
		return false
	}

	timeutil.SleepWithContext(ctx, sleepDuration)
	return true
}

// Update updates the monitor's rate limit information based on the HTTP response headers.
func (c *Monitor) Update(h http.Header) {
	if cached := h.Get("X-From-Cache"); cached != "" {
		// Cached responses have stale RateLimit headers.
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	retry, _ := strconv.ParseInt(h.Get("Retry-After"), 10, 64)
	if retry > 0 {
		c.retry = c.now().Add(time.Duration(retry) * time.Second)
	}

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

	if c.known && c.collector != nil && c.collector.Remaining != nil {
		c.collector.Remaining(float64(c.remaining))
	}
}

// SetCollector sets the metric collector.
func (c *Monitor) SetCollector(collector *MetricsCollector) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collector = collector
}

func (c *Monitor) now() time.Time {
	if c.clock != nil {
		return c.clock()
	}
	return time.Now()
}

package ratelimit

import (
	"math"
	"sync"

	"golang.org/x/time/rate"
)

// DefaultRegistry is the default global rate limit registry, which holds rate
// limit mappings for each instance of our services.
var DefaultRegistry = NewRegistry()

// NewRegistry creates and returns an empty rate limit registry.
func NewRegistry() *Registry {
	return &Registry{
		rateLimiters: make(map[string]*rate.Limiter),
	}
}

// Registry manages rate limiters for external services.
type Registry struct {
	mu sync.Mutex
	// rateLimiters contains mappings of external service to its *rate.Limiter. The
	// key should be the URN of the external service.
	rateLimiters map[string]*rate.Limiter
}

// Get returns the rate limiter configured for the given URN of an external
// service. It returns an infinite limiter if no rate limiter has been configured
// for the URN.
//
// Modifications to the returned rate limiter takes effect on all call sites.
func (r *Registry) Get(urn string) *rate.Limiter {
	return r.GetOrSet(urn, nil)
}

// GetOrSet returns the rate limiter configured for the given URN of an external
// service, and sets the `fallback` to be the rate limiter if no rate limiter has
// been configured for the URN. A nil `fallback` indicates an infinite limiter.
//
// Modifications to the returned rate limiter takes effect on all call sites.
func (r *Registry) GetOrSet(urn string, fallback *rate.Limiter) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()
	l := r.rateLimiters[urn]
	if l == nil {
		if fallback == nil {
			fallback = rate.NewLimiter(rate.Inf, 1)
		}
		r.rateLimiters[urn] = fallback
		return fallback
	}
	return l
}

// Count returns the total number of rate limiters in the registry.
func (r *Registry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.rateLimiters)
}

type LimitInfo struct {
	// Maximum allowed burst of requests
	Burst int
	// Maximum allowed requests per second. If the limit is infinite, Limit will be
	// zero and Infinite will be true
	Limit float64
	// Infinite is true if Limit is infinite. This is required since infinity cannot
	// be marshalled in JSON.
	Infinite bool
}

// LimitInfo reports how all the existing rate limiters are configured, keyed by
// URN.
func (r *Registry) LimitInfo() map[string]LimitInfo {
	r.mu.Lock()
	defer r.mu.Unlock()

	m := make(map[string]LimitInfo, len(r.rateLimiters))
	for urn, rl := range r.rateLimiters {
		limit := rl.Limit()
		info := LimitInfo{
			Burst: rl.Burst(),
			Limit: float64(limit),
		}
		if math.IsInf(info.Limit, 0) || limit == rate.Inf {
			info.Limit = 0
			info.Infinite = true
		}
		m[urn] = info
	}
	return m
}

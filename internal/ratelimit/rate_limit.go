package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
)

// DefaultRegistry is the default global rate limit registry. It will hold rate limit mappings
// for each instance of our services.
var DefaultRegistry = NewRegistry()

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		rateLimiters: make(map[string]*rate.Limiter),
	}
}

// Registry keeps a mapping of external service URL to *rate.Limiter.
// By default an infinite limiter is returned.
type Registry struct {
	mu sync.Mutex
	// Rate limiter per code host, keys are the normalized base URL for a
	// code host.
	rateLimiters map[string]*rate.Limiter
}

// Get fetches the rate limiter associated with the given code host. If none has been
// configured an infinite limiter is returned.
func (r *Registry) Get(baseURL string) *rate.Limiter {
	return r.GetOrSet(baseURL, nil)
}

// GetOrSet fetches the rate limiter associated with the given code host. If none has been configured
// yet, the provided limiter will be set. A nil limiter will fall back to an infinite limiter.
func (r *Registry) GetOrSet(baseURL string, fallback *rate.Limiter) *rate.Limiter {
	baseURL = normaliseURL(baseURL)
	if fallback == nil {
		fallback = rate.NewLimiter(rate.Inf, 100)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	l := r.rateLimiters[baseURL]
	if l == nil {
		l = fallback
		r.rateLimiters[baseURL] = l
	}
	return l
}

// Count returns the total number of rate limiters in the registry
func (r *Registry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.rateLimiters)
}

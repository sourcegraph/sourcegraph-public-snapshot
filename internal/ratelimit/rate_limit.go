package ratelimit

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DefaultRegistry is the default global rate limit registry, which holds rate
// limit mappings for each instance of our services.
var DefaultRegistry = NewRegistry()

const defaultBurst = 10

// NewRegistry creates and returns an empty rate limit registry. If a global default rate limit is specified a fallback
// rate limiter will be added.
func NewRegistry() *Registry {
	return &Registry{
		rateLimiters: make(map[string]*InstrumentedLimiter),
	}
}

// Registry manages rate limiters for external services.
type Registry struct {
	mu sync.Mutex
	// rateLimiters contains mappings of external service to its *rate.Limiter. The
	// key should be the URN of the external service.
	rateLimiters map[string]*InstrumentedLimiter
}

// Get returns the rate limiter configured for the given URN of an external
// service. If no rate limiter has been configured for the URN, it returns either
// the default rate limiter specified by the config or an infinite limiter if no
// limiter specified in the config.
//
// Modifications to the returned rate limiter takes effect on all call sites.
func (r *Registry) Get(urn string) *InstrumentedLimiter {
	return r.getOrSet(urn, nil)
}

// getOrSet returns the rate limiter configured for the given URN of an external
// service, and sets the `fallback` to be the rate limiter if no rate limiter has
// been configured for the URN. A nil `fallback` indicates an infinite limiter.
//
// Modifications to the returned rate limiter takes effect on all call sites.
func (r *Registry) getOrSet(urn string, fallback *InstrumentedLimiter) *InstrumentedLimiter {
	r.mu.Lock()
	defer r.mu.Unlock()
	l := r.rateLimiters[urn]
	if l != nil {
		return l
	}
	if fallback == nil {
		defaultRateLimit := conf.Get().DefaultRateLimit
		// the rate limit in the config is in requests per hour, whereas rate.Limit is in
		// requests per second.
		fallbackRateLimit := rate.Limit(defaultRateLimit / 3600.0)
		if defaultRateLimit <= 0 {
			fallbackRateLimit = rate.Inf
		}
		fallback = NewInstrumentedLimiter(urn, rate.NewLimiter(fallbackRateLimit, defaultBurst))
	}
	r.rateLimiters[urn] = fallback
	return fallback
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

// InstrumentedLimiter is wraps a *rate.Limiter with instrumentation
type InstrumentedLimiter struct {
	urn string
	*rate.Limiter
}

// NewInstrumentedLimiter creates new InstrumentedLimiter with given URN and rate.Limiter
func NewInstrumentedLimiter(urn string, limiter *rate.Limiter) *InstrumentedLimiter {
	return &InstrumentedLimiter{
		urn:     urn,
		Limiter: limiter,
	}
}

// Wait is shorthand for WaitN(ctx, 1).
func (i *InstrumentedLimiter) Wait(ctx context.Context) error {
	return i.WaitN(ctx, 1)
}

// WaitN blocks until lim permits n events to happen.
// It returns an error if n exceeds the Limiter's burst size, the Context is
// canceled, or the expected wait time exceeds the Context's Deadline.
// The burst limit is ignored if the rate limit is Inf.
func (i *InstrumentedLimiter) WaitN(ctx context.Context, n int) error {
	if i.Limit() == 0 && i.Burst() == 0 {
		// We're not allowing anything through the limiter, return a custom error so that
		// we can handle it correctly.
		return ErrBlockAll
	}

	start := time.Now()
	err := i.Limiter.WaitN(ctx, n)
	d := time.Since(start)
	failedLabel := "false"
	if err != nil {
		failedLabel = "true"
	}

	metricWaitDuration.WithLabelValues(i.urn, failedLabel).Observe(d.Seconds())
	return err
}

// SetBurst is calling SetBurstAt(time.Now(), newBurst) method of the wrapped *rate.Limiter.
func (i *InstrumentedLimiter) SetBurst(newBurst int) {
	i.Limiter.SetBurstAt(time.Now(), newBurst)
}

// SetLimit is calling SetLimitAt(time.Now(), newLimit) method of the wrapped *rate.Limiter.
func (i *InstrumentedLimiter) SetLimit(newLimit rate.Limit) {
	i.Limiter.SetLimitAt(time.Now(), newLimit)
}

// ErrBlockAll indicates that the limiter is set to block all requests
var ErrBlockAll = errors.New("ratelimit: limit and burst are zero")

var metricWaitDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "src_internal_rate_limit_wait_duration",
	Help:    "Time spent waiting for our internal rate limiter",
	Buckets: []float64{0.2, 0.5, 1, 2, 5, 10, 30, 60},
}, []string{"urn", "failed"})

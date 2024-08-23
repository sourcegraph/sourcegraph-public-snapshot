package throttled

import (
	"context"
	"net/http"
	"time"
)

// Quota returns the number of requests allowed and the custom time window.
//
// Deprecated: Use Rate and RateLimiter instead.
func (q Rate) Quota() (int, time.Duration) {
	return q.count, q.period * time.Duration(q.count)
}

// Q represents a custom quota.
//
// Deprecated: Use Rate and RateLimiter instead.
type Q struct {
	Requests int
	Window   time.Duration
}

// Quota returns the number of requests allowed and the custom time window.
//
// Deprecated: Use Rate and RateLimiter instead.
func (q Q) Quota() (int, time.Duration) {
	return q.Requests, q.Window
}

// The Quota interface defines the method to implement to describe
// a time-window quota, as required by the RateLimit throttler.
//
// Deprecated: Use Rate and RateLimiter instead.
type Quota interface {
	// Quota returns a number of requests allowed, and a duration.
	Quota() (int, time.Duration)
}

// Throttler is a backwards-compatible alias for HTTPLimiter.
//
// Deprecated: Use Rate and RateLimiter instead.
type Throttler struct {
	HTTPRateLimiter
}

// Throttle is an alias for HTTPLimiter#Limit
//
// Deprecated: Use Rate and RateLimiter instead.
func (t *Throttler) Throttle(h http.Handler) http.Handler {
	return t.RateLimit(h)
}

// RateLimit creates a Throttler that conforms to the given
// rate limits
//
// Deprecated: Use Rate and RateLimiter instead.
func RateLimit(q Quota, vary *VaryBy, store GCRAStore) *Throttler {
	count, period := q.Quota()

	if count < 1 {
		count = 1
	}
	if period <= 0 {
		period = time.Second
	}

	rate := Rate{period: period / time.Duration(count)}
	limiter, err := NewGCRARateLimiterCtx(WrapStoreWithContext(store), RateQuota{rate, count - 1})

	// This panic in unavoidable because the original interface does
	// not support returning an error.
	if err != nil {
		panic(err)
	}

	return &Throttler{
		HTTPRateLimiter{
			RateLimiter: limiter,
			VaryBy:      vary,
		},
	}
}

// Store is an alias for GCRAStore
//
// Deprecated: Use Rate and RateLimiter instead.
type Store interface {
	GCRAStore
}

// HTTPRateLimiter is an adapter for HTTPRateLimiterCtx to provide backwards
// compatibility.
//
// Deprecated: Use HTTPRateLimiterCtx instead. If the used RateLimiter does
// not implement RateLimiterCtx, wrap it with WrapRateLimiterWithContext().
type HTTPRateLimiter struct {
	// DeniedHandler is called if the request is disallowed. If it is
	// nil, the DefaultDeniedHandler variable is used.
	DeniedHandler http.Handler

	// Error is called if the RateLimiter returns an error. If it is
	// nil, the DefaultErrorFunc is used.
	Error func(w http.ResponseWriter, r *http.Request, err error)

	// Limiter is call for each request to determine whether the
	// request is permitted and update internal state. It must be set.
	RateLimiter RateLimiter

	// VaryBy is called for each request to generate a key for the
	// limiter. If it is nil, all requests use an empty string key.
	VaryBy interface {
		Key(*http.Request) string
	}
}

// RateLimit provides an adapter for HTTPRateLimiterCtx.RateLimit.
//
// Deprecated: Use HTTPRateLimiterCtx instead
func (t *HTTPRateLimiter) RateLimit(h http.Handler) http.Handler {
	l := HTTPRateLimiterCtx{
		DeniedHandler: t.DeniedHandler,
		Error:         t.Error,
		RateLimiter:   WrapRateLimiterWithContext(t.RateLimiter),
		VaryBy:        t.VaryBy,
	}
	return l.RateLimit(h)
}

// GCRAStore is the version of GCRAStoreCtx that is not aware of context.
//
// Deprecated: Implement GCRAStoreCtx instead.
type GCRAStore interface {
	GetWithTime(key string) (int64, time.Time, error)
	SetIfNotExistsWithTTL(key string, value int64, ttl time.Duration) (bool, error)
	CompareAndSwapWithTTL(key string, old, new int64, ttl time.Duration) (bool, error)
}

// NewGCRARateLimiter is a backwards compatible adapter for NewGCRARateLimiterCtx.
//
// Deprecated: Use NewGCRARateLimiterCtx instead. If the used store does
// not implement GCRAStoreCtx, wrap it with WrapStoreWithContext().
func NewGCRARateLimiter(st GCRAStore, quota RateQuota) (*GCRARateLimiterCtx, error) {
	return NewGCRARateLimiterCtx(WrapStoreWithContext(st), quota)
}

// A RateLimiter manages limiting the rate of actions by key.
//
// Deprecated: Use RateLimiterCtx instead.
type RateLimiter interface {
	// RateLimit checks whether a particular key has exceeded a rate
	// limit. It also returns a RateLimitResult to provide additional
	// information about the state of the RateLimiter.
	//
	// If the rate limit has not been exceeded, the underlying storage
	// is updated by the supplied quantity. For example, a quantity of
	// 1 might be used to rate limit a single request while a greater
	// quantity could rate limit based on the size of a file upload in
	// megabytes. If quantity is 0, no update is performed allowing
	// you to "peek" at the state of the RateLimiter for a given key.
	RateLimit(key string, quantity int) (bool, RateLimitResult, error)
}

// RateLimit is provided as a backwards compatible variant of RateLimitCtx.
//
// Deprecated: Use RateLimitCtx instead.
func (g *GCRARateLimiterCtx) RateLimit(key string, quantity int) (bool, RateLimitResult, error) {
	return g.RateLimitCtx(context.Background(), key, quantity)
}

// WrapStoreWithContext can be used to use GCRAStore in a place where a GCRAStoreCtx is required.
func WrapStoreWithContext(store GCRAStore) GCRAStoreCtx {
	return gcraStoreCtxAdapter{
		gcraStore: store,
	}
}

// WrapRateLimiterWithContext can be used to use RateLimiter in a place where a RateLimiterCtx is required.
func WrapRateLimiterWithContext(rateLimier RateLimiter) RateLimiterCtx {
	return rateLimiterCtxAdapter{
		rateLimiter: rateLimier,
	}
}

// gcraStoreCtxAdapter is an adapter that is used to use a GCRAStore where a GCRAStoreCtx is required.
type gcraStoreCtxAdapter struct {
	gcraStore GCRAStore
}

func (g gcraStoreCtxAdapter) GetWithTime(_ context.Context, key string) (int64, time.Time, error) {
	return g.gcraStore.GetWithTime(key)
}

func (g gcraStoreCtxAdapter) SetIfNotExistsWithTTL(_ context.Context, key string, value int64, ttl time.Duration) (bool, error) {
	return g.gcraStore.SetIfNotExistsWithTTL(key, value, ttl)
}

func (g gcraStoreCtxAdapter) CompareAndSwapWithTTL(_ context.Context, key string, old, new int64, ttl time.Duration) (bool, error) {
	return g.gcraStore.CompareAndSwapWithTTL(key, old, new, ttl)
}

// rateLimiterCtxAdapter is an adapter that is used to use a RateLimiter where a RateLimiterCtx is required.
type rateLimiterCtxAdapter struct {
	rateLimiter RateLimiter
}

func (r rateLimiterCtxAdapter) RateLimitCtx(_ context.Context, key string, quantity int) (bool, RateLimitResult, error) {
	return r.rateLimiter.RateLimit(key, quantity)
}

package ratelimit

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Limiter interface {
	WaitN(context.Context, int) error
}

type InspectableLimiter interface {
	Limiter

	Limit() rate.Limit
	Burst() int
}

// InstrumentedLimiter wraps a Limiter with instrumentation.
type InstrumentedLimiter struct {
	Limiter

	urn string
}

// NewInstrumentedLimiter creates new InstrumentedLimiter with given URN and Limiter,
// usually a rate.Limiter.
func NewInstrumentedLimiter(urn string, limiter Limiter) *InstrumentedLimiter {
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
	if il, ok := i.Limiter.(InspectableLimiter); ok {
		if il.Limit() == 0 && il.Burst() == 0 {
			// We're not allowing anything through the limiter, return a custom error so that
			// we can handle it correctly.
			return ErrBlockAll
		}
	}

	start := time.Now()
	err := i.Limiter.WaitN(ctx, n)
	// For GlobalLimiter instances, we return a special error type for BlockAll,
	// since we don't want to make two preflight redis calls to check limit and burst
	// above. We map it back to ErrBlockAll here then.
	if err != nil && errors.HasType[AllBlockedError](err) {
		return ErrBlockAll
	}
	d := time.Since(start)
	failedLabel := "false"
	if err != nil {
		failedLabel = "true"
	}

	metricWaitDuration.WithLabelValues(i.urn, failedLabel).Observe(d.Seconds())
	return err
}

// ErrBlockAll indicates that the limiter is set to block all requests
var ErrBlockAll = errors.New("ratelimit: limit and burst are zero")

var metricWaitDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "src_internal_rate_limit_wait_duration",
	Help:    "Time spent waiting for our internal rate limiter",
	Buckets: []float64{0.2, 0.5, 1, 2, 5, 10, 30, 60},
}, []string{"urn", "failed"})

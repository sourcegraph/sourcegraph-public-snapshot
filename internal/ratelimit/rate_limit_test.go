package ratelimit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestDefaultRateLimiter(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			DefaultRateLimit: 7200,
		},
	})
	defer conf.Mock(nil)

	r := NewRegistry()
	got := r.Get("unknown")
	want := &InstrumentedLimiter{urn: "unknown", Limiter: rate.NewLimiter(rate.Limit(2), defaultBurst)}
	assert.Equal(t, want, got)
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	got := r.Get("404")
	want := &InstrumentedLimiter{urn: "404", Limiter: rate.NewLimiter(rate.Inf, defaultBurst)}
	assert.Equal(t, want, got)

	rl := &InstrumentedLimiter{urn: "extsvc:github:1", Limiter: rate.NewLimiter(10, 10)}
	got = r.getOrSet("extsvc:github:1", rl)
	assert.Equal(t, rl, got)

	got = r.getOrSet("extsvc:github:1", &InstrumentedLimiter{urn: "extsvc:githu:1", Limiter: rate.NewLimiter(1000, 10)})
	assert.Equal(t, rl, got)

	assert.Equal(t, 2, r.Count())
}

func TestLimitInfo(t *testing.T) {
	r := NewRegistry()
	r.getOrSet("extsvc:github:1", &InstrumentedLimiter{urn: "extsvc:github:1", Limiter: rate.NewLimiter(rate.Inf, 1)})
	r.getOrSet("extsvc:github:2", &InstrumentedLimiter{urn: "extsvc:github:2", Limiter: rate.NewLimiter(10, 1)})

	info := r.LimitInfo()

	assert.Equal(t, info["extsvc:github:1"], LimitInfo{
		Limit:    0,
		Burst:    1,
		Infinite: true,
	})
	assert.Equal(t, info["extsvc:github:2"], LimitInfo{
		Limit:    10,
		Burst:    1,
		Infinite: false,
	})
}

package ratelimit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	got := r.Get("404")
	want := rate.NewLimiter(rate.Inf, 1)
	assert.Equal(t, want, got)

	rl := rate.NewLimiter(10, 10)
	got = r.GetOrSet("extsvc:github:1", rl)
	assert.Equal(t, rl, got)

	got = r.GetOrSet("extsvc:github:1", rate.NewLimiter(1000, 10))
	assert.Equal(t, rl, got)

	assert.Equal(t, 2, r.Count())
}

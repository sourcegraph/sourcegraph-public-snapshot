package github

import (
	"testing"
	"time"
)

func TestRecommendedRateLimitWaitForBackgroundOp(t *testing.T) {
	c := &Client{
		rateLimitKnown:     true,
		rateLimit:          5000,
		rateLimitRemaining: 1500,
		rateLimitReset:     time.Now().Add(30 * time.Minute),
	}

	durationsApproxEqual := func(a, b time.Duration) bool {
		d := a - b
		if d < 0 {
			d = -1 * d
		}
		return d < 10*time.Second
	}

	tests := map[int]time.Duration{
		1:    0,
		10:   time.Second,
		100:  99 * time.Second,
		500:  9 * time.Minute,
		3500: 63 * time.Minute,
	}
	for cost, want := range tests {
		if got := c.RecommendedRateLimitWaitForBackgroundOp(cost); !durationsApproxEqual(got, want) {
			t.Errorf("got %s, want %s", got, want)
		}
	}
}

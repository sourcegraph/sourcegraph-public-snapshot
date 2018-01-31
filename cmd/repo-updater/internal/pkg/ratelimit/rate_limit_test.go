package ratelimit

import (
	"testing"
	"time"
)

func TestMonitor_RecommendedWaitForBackgroundOp(t *testing.T) {
	m := &Monitor{
		known:     true,
		limit:     5000,
		remaining: 1500,
		reset:     time.Now().Add(30 * time.Minute),
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
		if got := m.RecommendedWaitForBackgroundOp(cost); !durationsApproxEqual(got, want) {
			t.Errorf("got %s, want %s", got, want)
		}
	}
}

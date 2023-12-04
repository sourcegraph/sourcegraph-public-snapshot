package ratelimit

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		return d < 2*time.Second
	}

	// The conservative handling of rate limiting means that the 1500 remaining
	// will be treated as roughly 1200. For cost smaller than 1200, we should
	// expect a time of (reset + 3 minutes) * cost / 1200. For cost greater than
	// 1200, we should expect exactly reset + 3 minutes, because we won't wait
	// past the reset, as there'd be no point.
	tests := map[int]time.Duration{
		1:    0,
		10:   33 * time.Minute * 10 / 1200,
		100:  33 * time.Minute * 100 / 1200,
		500:  33 * time.Minute * 500 / 1200,
		3500: 33 * time.Minute,
	}
	for cost, want := range tests {
		if got := m.RecommendedWaitForBackgroundOp(cost); !durationsApproxEqual(got, want) {
			t.Errorf("for %d, got %s, want %s", cost, got, want)
		}
	}
	// Verify that we use the full limit, not the remaining limit, if the reset
	// time has passed. This should scale times based on 4,000 items in 63 minutes.
	m.reset = time.Now().Add(-1 * time.Second)
	tests = map[int]time.Duration{
		1:    0,                             // Things you could do >=500 times should just run
		10:   200 * time.Millisecond,        // Things you could do 250-500 times in the limit should get 200ms
		400:  63 * time.Minute * 400 / 4000, // 1/10 of 63 minutes
		9001: 3780 * time.Second,            // The full reset period
	}
	for cost, want := range tests {
		if got := m.RecommendedWaitForBackgroundOp(cost); !durationsApproxEqual(got, want) {
			t.Errorf("with reset: for %d, got %s, want %s", cost, got, want)
		}
	}
}

func TestMonitor_WaitForRateLimit(t *testing.T) {
	t.Run("no wait time if cost is lower than remaining", func(t *testing.T) {
		m := &Monitor{
			known:     true,
			limit:     5000,
			remaining: 10,
			reset:     time.Now().Add(30 * time.Minute),
		}

		sleepDuration := m.calcRateLimitWaitTime(5)

		assert.Equal(t, time.Duration(0), sleepDuration)
	})
	t.Run("no wait time if cost is equal to remaining", func(t *testing.T) {
		m := &Monitor{
			known:     true,
			limit:     5000,
			remaining: 10,
			reset:     time.Now().Add(30 * time.Minute),
		}

		sleepDuration := m.calcRateLimitWaitTime(10)

		assert.Equal(t, time.Duration(0), sleepDuration)
	})
	t.Run("wait if cost is higher than remaining", func(t *testing.T) {
		m := &Monitor{
			known:     true,
			limit:     5000,
			remaining: 10,
			reset:     time.Now().Add(30 * time.Minute),
		}

		sleepDuration := m.calcRateLimitWaitTime(11)

		// Assert that the sleep duration is about 30 minutes (slightly inaccurate, so checking between 29 and 30 minutes)
		assert.True(t, time.Duration(29)*time.Minute < sleepDuration)
		assert.True(t, time.Duration(30)*time.Minute > sleepDuration)
	})
}

func TestMonitor_RecommendedWaitForBackgroundOp_RetryAfter(t *testing.T) {
	now := time.Now()
	for _, tc := range []struct {
		retry time.Time
		now   time.Time
		wait  time.Duration
	}{
		// 30 seconds remaining from now until retry
		{now.Add(30 * time.Second), now, 30 * time.Second},
		// 0 seconds remaining from now until retry
		{now.Add(30 * time.Second), now.Add(30 * time.Second), 0},
		// -30 seconds remaining from now until retry
		{now.Add(30 * time.Second), now.Add(60 * time.Second), 0},
	} {
		m := Monitor{
			retry: tc.retry,
			clock: func() time.Time { return tc.now },
		}

		wait := m.RecommendedWaitForBackgroundOp(1)
		if have, want := wait, tc.wait; have != want {
			t.Errorf("retry: %s, now: %s: wait: have %s, want %s", tc.retry, tc.now, have, want)
		}
	}
}

func TestMonitor_Update(t *testing.T) {
	now := time.Now()
	clock := func() time.Time { return now }

	equal := func(a, b *Monitor) bool {
		return a.HeaderPrefix == b.HeaderPrefix &&
			a.known == b.known &&
			a.limit == b.limit &&
			a.remaining == b.remaining &&
			a.reset.Equal(b.reset) &&
			a.retry.Equal(b.retry)
	}

	for _, tc := range []struct {
		name   string
		before *Monitor
		h      http.Header
		after  *Monitor
	}{
		{
			name:   "Retry-After header sets retry deadline",
			before: &Monitor{clock: clock},
			h:      http.Header{"Retry-After": []string{"30"}},
			after:  &Monitor{retry: now.Add(30 * time.Second)},
		},
		{
			name:   "Empty Retry-After header leaves deadline intact",
			before: &Monitor{clock: clock, retry: now.Add(time.Second)},
			h:      http.Header{},
			after:  &Monitor{retry: now.Add(time.Second)},
		},
		{
			name:   "RateLimit headers must come together",
			before: &Monitor{clock: clock, known: true},
			// Missing the other headers, so nothing gets set and known becomes false
			h:     http.Header{"RateLimit-Limit": []string{"500"}},
			after: &Monitor{known: false},
		},
		{
			name:   "RateLimit headers are set",
			before: &Monitor{HeaderPrefix: "X-", clock: clock},
			h: http.Header{
				"X-RateLimit-Limit":     []string{"500"},
				"X-RateLimit-Remaining": []string{"1"},
				"X-RateLimit-Reset":     []string{strconv.FormatInt(now.Add(time.Minute).Unix(), 10)},
			},
			after: &Monitor{
				HeaderPrefix: "X-",
				known:        true,
				limit:        500,
				remaining:    1,
				reset:        time.Unix(now.Add(time.Minute).Unix(), 0),
			},
		},
		{
			// GitLab uses different casing
			name:   "RateLimit headers are set for GitLab",
			before: &Monitor{clock: clock},
			h: http.Header{
				"Ratelimit-Limit":     []string{"500"},
				"Ratelimit-Remaining": []string{"1"},
				"Ratelimit-Reset":     []string{strconv.FormatInt(now.Add(time.Minute).Unix(), 10)},
			},
			after: &Monitor{
				HeaderPrefix: "",
				known:        true,
				limit:        500,
				remaining:    1,
				reset:        time.Unix(now.Add(time.Minute).Unix(), 0),
			},
		},
		{
			name:   "Responses with X-From-Cache header are ignored",
			before: &Monitor{clock: clock},
			h: http.Header{
				"X-From-Cache":        []string{"1"},
				"RateLimit-Limit":     []string{"500"},
				"RateLimit-Remaining": []string{"1"},
				"RateLimit-Reset":     []string{strconv.FormatInt(now.Add(time.Minute).Unix(), 10)},
			},
			after: &Monitor{},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := make(http.Header, len(tc.h))
			for k, vs := range tc.h {
				for _, v := range vs {
					// So that header keys are made canonical with
					// textproto.CanonicalMIMEHeaderKey
					h.Add(k, v)
				}
			}

			tc.before.Update(h)
			if have, want := tc.before, tc.after; !equal(have, want) {
				t.Errorf("\nhave: %#v\nwant: %#v", have, want)
			}
		})
	}
}

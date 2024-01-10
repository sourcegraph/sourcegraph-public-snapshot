package actor

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestNewRateLimitWithPercentageConcurrency(t *testing.T) {
	concurrencyLimitConfig := codygateway.ActorConcurrencyLimitConfig{
		Percentage: 0.1,
		Interval:   10 * time.Second,
	}
	tests := []struct {
		name                 string
		limit                int64
		interval             time.Duration
		wantConcurrencyLimit int
	}{
		{
			name:                 "feature limit internal is daily",
			limit:                100,
			interval:             24 * time.Hour,
			wantConcurrencyLimit: 10,
		},
		{
			name:                 "feature limit internal is more than a day",
			limit:                210,
			interval:             7 * 24 * time.Hour,
			wantConcurrencyLimit: 3,
		},
		{
			name:                 "feature limit internal is less than a day",
			limit:                10,
			interval:             time.Hour,
			wantConcurrencyLimit: 24,
		},
		{
			name:                 "computed concurrency limit is less than 1",
			limit:                3,
			interval:             24 * time.Hour,
			wantConcurrencyLimit: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewRateLimitWithPercentageConcurrency(test.limit, test.interval, []string{"model"}, concurrencyLimitConfig)
			assert.Equal(t, test.wantConcurrencyLimit, got.ConcurrentRequests)
		})
	}
}

func TestConcurrencyLimiter_TryAcquire(t *testing.T) {
	// Stable time for testing
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	nowFunc := func() time.Time { return now }
	featureLimiter := limiter.StaticLimiter{
		Identifier: "foobar",
		Redis:      limiter.MockRedisStore{},
		Limit:      10,
		Interval:   24 * time.Hour,
	}

	tests := []struct {
		name      string
		limiter   *concurrencyLimiter
		wantErr   autogold.Value
		wantStore autogold.Value
	}{
		{
			name: "new entry",
			limiter: &concurrencyLimiter{
				actor: &Actor{ID: "foobar"},
				redis: limiter.MockRedisStore{},

				concurrentRequests: 2,
				concurrentInterval: 10 * time.Second,

				nextLimiter: featureLimiter,
				nowFunc:     nowFunc,
			},
			wantErr: nil,
			wantStore: autogold.Expect(limiter.MockRedisStore{
				"foobar": limiter.MockRedisEntry{Value: 1},
			}),
		},
		{
			name: "increments existing entry",
			limiter: &concurrencyLimiter{
				actor: &Actor{ID: "foobar"},
				redis: limiter.MockRedisStore{
					"foobar": limiter.MockRedisEntry{Value: 1, TTL: 10},
				},

				concurrentRequests: 2,
				concurrentInterval: 10 * time.Second,

				nextLimiter: featureLimiter,
				nowFunc:     nowFunc,
			},
			wantErr: nil,
			wantStore: autogold.Expect(limiter.MockRedisStore{
				"foobar": limiter.MockRedisEntry{Value: 2, TTL: 10},
			}),
		},
		{
			name: "existing limit's TTL is longer than desired interval",
			limiter: &concurrencyLimiter{
				actor: &Actor{
					ID:          "foobar",
					LastUpdated: &now,
				},
				redis: limiter.MockRedisStore{
					"foobar": limiter.MockRedisEntry{Value: 1, TTL: 999},
				},

				concurrentRequests: 2,
				concurrentInterval: 10 * time.Second,

				nextLimiter: featureLimiter,
				nowFunc:     nowFunc,
			},
			wantErr: nil,
			wantStore: autogold.Expect(limiter.MockRedisStore{
				"foobar": limiter.MockRedisEntry{Value: 2, TTL: 10},
			}),
		},
		{
			name: "rejects request over quota",
			limiter: &concurrencyLimiter{
				actor:   &Actor{ID: "foobar"},
				feature: codygateway.FeatureCodeCompletions,
				redis: limiter.MockRedisStore{
					"foobar": limiter.MockRedisEntry{Value: 2, TTL: 10},
				},

				concurrentRequests: 2,
				concurrentInterval: 10 * time.Second,

				nextLimiter: featureLimiter,
				nowFunc:     nowFunc,
			},
			wantErr: autogold.Expect(`"code_completions": concurrency limit exceeded`),
			wantStore: autogold.Expect(limiter.MockRedisStore{
				"foobar": limiter.MockRedisEntry{Value: 2, TTL: 10},
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.limiter.TryAcquire(context.Background())
			if test.wantErr != nil {
				require.Error(t, err)
				test.wantErr.Equal(t, err.Error())
			} else {
				require.NoError(t, err)
			}
			test.wantStore.Equal(t, test.limiter.redis)
		})
	}
}

func TestAsErrConcurrencyLimitExceeded(t *testing.T) {
	var err error = ErrConcurrencyLimitExceeded{}
	assert.True(t, errors.As(err, &ErrConcurrencyLimitExceeded{}))
	assert.True(t, errors.As(errors.Wrap(err, "foo"), &ErrConcurrencyLimitExceeded{}))
}

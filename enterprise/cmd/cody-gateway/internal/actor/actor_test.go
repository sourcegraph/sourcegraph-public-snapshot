package actor

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
)

func TestActor_Limiter_Concurrency(t *testing.T) {
	concurrencyLimitConfig := codygateway.ActorConcurrencyLimitConfig{
		Percentage: 0.1,
		Interval:   10 * time.Second,
	}
	tests := []struct {
		name                 string
		actor                *Actor
		wantConcurrencyLimit int
	}{
		{
			name: "feature limit internal is daily",
			actor: &Actor{
				RateLimits: map[codygateway.Feature]RateLimit{
					codygateway.FeatureChatCompletions: {
						Limit:         100,
						Interval:      24 * time.Hour,
						AllowedModels: []string{"model"},
					},
				},
			},
			wantConcurrencyLimit: 10,
		},
		{
			name: "feature limit internal is more than a day",
			actor: &Actor{
				RateLimits: map[codygateway.Feature]RateLimit{
					codygateway.FeatureChatCompletions: {
						Limit:         210,
						Interval:      7 * 24 * time.Hour,
						AllowedModels: []string{"model"},
					},
				},
			},
			wantConcurrencyLimit: 3,
		},
		{
			name: "feature limit internal is less than a day",
			actor: &Actor{
				RateLimits: map[codygateway.Feature]RateLimit{
					codygateway.FeatureChatCompletions: {
						Limit:         10,
						Interval:      time.Hour,
						AllowedModels: []string{"model"},
					},
				},
			},
			wantConcurrencyLimit: 24,
		},
		{
			name: "computed concurrency limit is less than 1",
			actor: &Actor{
				RateLimits: map[codygateway.Feature]RateLimit{
					codygateway.FeatureChatCompletions: {
						Limit:         3,
						Interval:      24 * time.Hour,
						AllowedModels: []string{"model"},
					},
				},
			},
			wantConcurrencyLimit: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := test.actor.Limiter(nil, nil, codygateway.FeatureChatCompletions, concurrencyLimitConfig)
			assert.True(t, ok, "should have returned a limiter")
			require.NotNil(t, got)

			gotConcurrencyLimit, ok := got.(*concurrencyLimiter)
			require.True(t, ok, "should be a *concurrencyLimiter")
			assert.Equal(t, test.wantConcurrencyLimit, gotConcurrencyLimit.rateLimit.Limit)
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
				rateLimit: RateLimit{
					Limit:    2,
					Interval: 10 * time.Second,
				},
				featureLimiter: featureLimiter,
				nowFunc:        nowFunc,
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
				rateLimit: RateLimit{
					Limit:    2,
					Interval: 10 * time.Second,
				},
				featureLimiter: featureLimiter,
				nowFunc:        nowFunc,
			},
			wantErr: nil,
			wantStore: autogold.Expect(limiter.MockRedisStore{
				"foobar": limiter.MockRedisEntry{Value: 2, TTL: 10},
			}),
		},
		{
			name: "existing limit's TTL is longer than desired interval but UpdateRateLimitTTL=false",
			limiter: &concurrencyLimiter{
				actor: &Actor{ID: "foobar"},
				redis: limiter.MockRedisStore{
					"foobar": limiter.MockRedisEntry{Value: 1, TTL: 999},
				},
				rateLimit: RateLimit{
					Limit:    2,
					Interval: 10 * time.Second,
				},
				featureLimiter: featureLimiter,
				nowFunc:        nowFunc,
			},
			wantErr: nil,
			wantStore: autogold.Expect(limiter.MockRedisStore{
				"foobar": limiter.MockRedisEntry{Value: 2, TTL: 999},
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
				rateLimit: RateLimit{
					Limit:    2,
					Interval: 10 * time.Second,
				},
				featureLimiter: featureLimiter,
				nowFunc:        nowFunc,
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
				rateLimit: RateLimit{
					Limit:    2,
					Interval: 10 * time.Second,
				},
				featureLimiter: featureLimiter,
				nowFunc:        nowFunc,
			},
			wantErr: autogold.Expect(`you exceeded the concurrency limit of 2 requests for "code_completions". Retry after 2000-01-01 00:00:10 +0000 UTC`),
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
			if test.wantStore != nil {
				test.wantStore.Equal(t, test.limiter.redis)
			}
		})
	}
}

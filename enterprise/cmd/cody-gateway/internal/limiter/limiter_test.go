package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticLimiterTryAcquire(t *testing.T) {
	// Stable time for testing
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name string

		limiter  StaticLimiter
		noCommit bool

		wantErr   autogold.Value
		wantStore autogold.Value
	}{
		{
			name:      "no limits set",
			noCommit:  true, // error scenario
			wantErr:   autogold.Expect("completions access has not been granted"),
			wantStore: autogold.Expect(MockRedisStore{}),
		},
		{
			name: "new entry",
			limiter: StaticLimiter{
				Identifier: "foobar",
				Limit:      10,
				Interval:   24 * time.Hour,
			},
			wantErr: nil,
			wantStore: autogold.Expect(MockRedisStore{"foobar": MockRedisEntry{
				Value: 1,
			}}),
		},
		{
			name: "increments existing entry",
			limiter: StaticLimiter{
				Identifier: "foobar",
				Redis: MockRedisStore{
					"foobar": MockRedisEntry{
						Value: 9,
						TTL:   60,
					},
				},
				Limit:    10,
				Interval: 24 * time.Hour,
			},
			wantErr: nil,
			wantStore: autogold.Expect(MockRedisStore{"foobar": MockRedisEntry{
				Value: 10,
				TTL:   60,
			}}),
		},
		{
			name: "no increment without commit",
			limiter: StaticLimiter{
				Identifier: "foobar",
				Redis: MockRedisStore{
					"foobar": MockRedisEntry{
						Value: 5,
						TTL:   60,
					},
				},
				Limit:    10,
				Interval: 24 * time.Hour,
			},
			noCommit: true, // value should be the same
			wantErr:  nil,
			wantStore: autogold.Expect(MockRedisStore{"foobar": MockRedisEntry{
				Value: 5,
				TTL:   60,
			}}),
		},
		{
			name: "existing limit's TTL is longer than desired interval but UpdateRateLimitTTL=false",
			limiter: StaticLimiter{
				Identifier: "foobar",
				Redis: MockRedisStore{
					"foobar": MockRedisEntry{
						Value: 1,
						TTL:   999,
					},
				},
				Limit:              10,
				Interval:           10 * time.Minute,
				UpdateRateLimitTTL: false,
			},
			wantErr: nil,
			wantStore: autogold.Expect(MockRedisStore{"foobar": MockRedisEntry{
				Value: 2,
				TTL:   999,
			}}),
		},
		{
			name: "existing limit's TTL is longer than desired interval",
			limiter: StaticLimiter{
				Identifier: "foobar",
				Redis: MockRedisStore{
					"foobar": MockRedisEntry{
						Value: 1,
						TTL:   999,
					},
				},
				Limit:              10,
				Interval:           10 * time.Minute,
				UpdateRateLimitTTL: true,
			},
			wantErr: nil,
			wantStore: autogold.Expect(MockRedisStore{"foobar": MockRedisEntry{
				Value: 2,
				TTL:   600,
			}}),
		},
		{
			name: "rejects request over quota",
			limiter: StaticLimiter{
				Identifier: "foobar",
				Redis: MockRedisStore{
					"foobar": MockRedisEntry{
						Value: 10,
						TTL:   60,
					},
				},
				Limit:    10,
				Interval: 24 * time.Hour,
			},
			noCommit: true, // error scenario
			wantErr:  autogold.Expect("rate limit exceeded"),
			wantStore: autogold.Expect(MockRedisStore{"foobar": MockRedisEntry{
				Value: 10,
				TTL:   60,
			}}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.limiter.Redis == nil {
				tc.limiter.Redis = MockRedisStore{}
			}
			tc.limiter.NowFunc = func() time.Time { return now }
			commit, err := tc.limiter.TryAcquire(context.Background())
			if tc.wantErr != nil {
				require.Error(t, err)
				tc.wantErr.Equal(t, err.Error())
			} else {
				require.NoError(t, err)
			}
			if !tc.noCommit {
				assert.NoError(t, commit(context.Background(), 1))
			}
			tc.wantStore.Equal(t, tc.limiter.Redis)
		})
	}
}

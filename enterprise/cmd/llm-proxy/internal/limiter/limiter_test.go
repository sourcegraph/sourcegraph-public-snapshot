package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

func TestStaticLimiterTryAcquire(t *testing.T) {
	// Stable time for testing
	now := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	for _, tc := range []struct {
		name string

		limiter StaticLimiter

		wantErr   autogold.Value
		wantStore autogold.Value
	}{
		{
			name:      "no limits set",
			wantErr:   autogold.Expect("completions access has not been granted"),
			wantStore: autogold.Expect(mockStore{}),
		},
		{
			name: "new entry",
			limiter: StaticLimiter{
				Limit:    10,
				Interval: 24 * time.Hour,
			},
			wantErr: nil,
			wantStore: autogold.Expect(mockStore{"foobar": mockRedisEntry{
				value: 1,
			}}),
		},
		{
			name: "increments existing entry",
			limiter: StaticLimiter{
				Redis: mockStore{
					"foobar": mockRedisEntry{
						value: 9,
						ttl:   60,
					},
				},
				Limit:    10,
				Interval: 24 * time.Hour,
			},
			wantErr: nil,
			wantStore: autogold.Expect(mockStore{"foobar": mockRedisEntry{
				value: 10,
				ttl:   60,
			}}),
		},
		{
			name: "existing limit's TTL is longer than desired interval",
			limiter: StaticLimiter{
				Redis: mockStore{
					"foobar": mockRedisEntry{
						value: 1,
						ttl:   999,
					},
				},
				Limit:    10,
				Interval: 10 * time.Minute,
			},
			wantErr: nil,
			wantStore: autogold.Expect(mockStore{"foobar": mockRedisEntry{
				value: 2,
				ttl:   600,
			}}),
		},
		{
			name: "rejects request over quota",
			limiter: StaticLimiter{
				Redis: mockStore{
					"foobar": mockRedisEntry{
						value: 10,
						ttl:   60,
					},
				},
				Limit:    10,
				Interval: 24 * time.Hour,
			},
			wantErr: autogold.Expect("you exceeded the rate limit for completions. Current usage: 10 out of 10 requests. Retry after 2000-01-01 00:01:00 +0000 UTC"),
			wantStore: autogold.Expect(mockStore{"foobar": mockRedisEntry{
				value: 11,
				ttl:   60,
			}}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.limiter.Redis == nil {
				tc.limiter.Redis = mockStore{}
			}
			tc.limiter.nowFunc = func() time.Time { return now }
			err := tc.limiter.TryAcquire(context.Background(), "foobar")
			if tc.wantErr != nil {
				require.Error(t, err)
				tc.wantErr.Equal(t, err.Error())
			} else {
				require.NoError(t, err)
			}
			if tc.wantStore != nil {
				tc.wantStore.Equal(t, tc.limiter.Redis)
			}
		})
	}
}

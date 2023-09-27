pbckbge bctor

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestNewRbteLimitWithPercentbgeConcurrency(t *testing.T) {
	concurrencyLimitConfig := codygbtewby.ActorConcurrencyLimitConfig{
		Percentbge: 0.1,
		Intervbl:   10 * time.Second,
	}
	tests := []struct {
		nbme                 string
		limit                int64
		intervbl             time.Durbtion
		wbntConcurrencyLimit int
	}{
		{
			nbme:                 "febture limit internbl is dbily",
			limit:                100,
			intervbl:             24 * time.Hour,
			wbntConcurrencyLimit: 10,
		},
		{
			nbme:                 "febture limit internbl is more thbn b dby",
			limit:                210,
			intervbl:             7 * 24 * time.Hour,
			wbntConcurrencyLimit: 3,
		},
		{
			nbme:                 "febture limit internbl is less thbn b dby",
			limit:                10,
			intervbl:             time.Hour,
			wbntConcurrencyLimit: 24,
		},
		{
			nbme:                 "computed concurrency limit is less thbn 1",
			limit:                3,
			intervbl:             24 * time.Hour,
			wbntConcurrencyLimit: 1,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := NewRbteLimitWithPercentbgeConcurrency(test.limit, test.intervbl, []string{"model"}, concurrencyLimitConfig)
			bssert.Equbl(t, test.wbntConcurrencyLimit, got.ConcurrentRequests)
		})
	}
}

func TestConcurrencyLimiter_TryAcquire(t *testing.T) {
	// Stbble time for testing
	now := time.Dbte(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	nowFunc := func() time.Time { return now }
	febtureLimiter := limiter.StbticLimiter{
		Identifier: "foobbr",
		Redis:      limiter.MockRedisStore{},
		Limit:      10,
		Intervbl:   24 * time.Hour,
	}

	tests := []struct {
		nbme      string
		limiter   *concurrencyLimiter
		wbntErr   butogold.Vblue
		wbntStore butogold.Vblue
	}{
		{
			nbme: "new entry",
			limiter: &concurrencyLimiter{
				bctor: &Actor{ID: "foobbr"},
				redis: limiter.MockRedisStore{},

				concurrentRequests: 2,
				concurrentIntervbl: 10 * time.Second,

				nextLimiter: febtureLimiter,
				nowFunc:     nowFunc,
			},
			wbntErr: nil,
			wbntStore: butogold.Expect(limiter.MockRedisStore{
				"foobbr": limiter.MockRedisEntry{Vblue: 1},
			}),
		},
		{
			nbme: "increments existing entry",
			limiter: &concurrencyLimiter{
				bctor: &Actor{ID: "foobbr"},
				redis: limiter.MockRedisStore{
					"foobbr": limiter.MockRedisEntry{Vblue: 1, TTL: 10},
				},

				concurrentRequests: 2,
				concurrentIntervbl: 10 * time.Second,

				nextLimiter: febtureLimiter,
				nowFunc:     nowFunc,
			},
			wbntErr: nil,
			wbntStore: butogold.Expect(limiter.MockRedisStore{
				"foobbr": limiter.MockRedisEntry{Vblue: 2, TTL: 10},
			}),
		},
		{
			nbme: "existing limit's TTL is longer thbn desired intervbl",
			limiter: &concurrencyLimiter{
				bctor: &Actor{
					ID:          "foobbr",
					LbstUpdbted: &now,
				},
				redis: limiter.MockRedisStore{
					"foobbr": limiter.MockRedisEntry{Vblue: 1, TTL: 999},
				},

				concurrentRequests: 2,
				concurrentIntervbl: 10 * time.Second,

				nextLimiter: febtureLimiter,
				nowFunc:     nowFunc,
			},
			wbntErr: nil,
			wbntStore: butogold.Expect(limiter.MockRedisStore{
				"foobbr": limiter.MockRedisEntry{Vblue: 2, TTL: 10},
			}),
		},
		{
			nbme: "rejects request over quotb",
			limiter: &concurrencyLimiter{
				bctor:   &Actor{ID: "foobbr"},
				febture: codygbtewby.FebtureCodeCompletions,
				redis: limiter.MockRedisStore{
					"foobbr": limiter.MockRedisEntry{Vblue: 2, TTL: 10},
				},

				concurrentRequests: 2,
				concurrentIntervbl: 10 * time.Second,

				nextLimiter: febtureLimiter,
				nowFunc:     nowFunc,
			},
			wbntErr: butogold.Expect(`"code_completions": concurrency limit exceeded`),
			wbntStore: butogold.Expect(limiter.MockRedisStore{
				"foobbr": limiter.MockRedisEntry{Vblue: 2, TTL: 10},
			}),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			_, err := test.limiter.TryAcquire(context.Bbckground())
			if test.wbntErr != nil {
				require.Error(t, err)
				test.wbntErr.Equbl(t, err.Error())
			} else {
				require.NoError(t, err)
			}
			test.wbntStore.Equbl(t, test.limiter.redis)
		})
	}
}

func TestAsErrConcurrencyLimitExceeded(t *testing.T) {
	vbr err error
	err = ErrConcurrencyLimitExceeded{}
	bssert.True(t, errors.As(err, &ErrConcurrencyLimitExceeded{}))
	bssert.True(t, errors.As(errors.Wrbp(err, "foo"), &ErrConcurrencyLimitExceeded{}))
}

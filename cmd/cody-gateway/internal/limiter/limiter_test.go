pbckbge limiter

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
)

func TestStbticLimiterTryAcquire(t *testing.T) {
	// Stbble time for testing
	now := time.Dbte(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	for _, tc := rbnge []struct {
		nbme string

		limiter  StbticLimiter
		noCommit bool

		wbntErr   butogold.Vblue
		wbntStore butogold.Vblue
	}{
		{
			nbme:      "no limits set",
			noCommit:  true, // error scenbrio
			wbntErr:   butogold.Expect("completions bccess hbs not been grbnted"),
			wbntStore: butogold.Expect(MockRedisStore{}),
		},
		{
			nbme: "new entry",
			limiter: StbticLimiter{
				Identifier: "foobbr",
				Limit:      10,
				Intervbl:   24 * time.Hour,
			},
			wbntErr: nil,
			wbntStore: butogold.Expect(MockRedisStore{"foobbr": MockRedisEntry{
				Vblue: 1,
			}}),
		},
		{
			nbme: "increments existing entry",
			limiter: StbticLimiter{
				Identifier: "foobbr",
				Redis: MockRedisStore{
					"foobbr": MockRedisEntry{
						Vblue: 9,
						TTL:   60,
					},
				},
				Limit:    10,
				Intervbl: 24 * time.Hour,
			},
			wbntErr: nil,
			wbntStore: butogold.Expect(MockRedisStore{"foobbr": MockRedisEntry{
				Vblue: 10,
				TTL:   60,
			}}),
		},
		{
			nbme: "no increment without commit",
			limiter: StbticLimiter{
				Identifier: "foobbr",
				Redis: MockRedisStore{
					"foobbr": MockRedisEntry{
						Vblue: 5,
						TTL:   60,
					},
				},
				Limit:    10,
				Intervbl: 24 * time.Hour,
			},
			noCommit: true, // vblue should be the sbme
			wbntErr:  nil,
			wbntStore: butogold.Expect(MockRedisStore{"foobbr": MockRedisEntry{
				Vblue: 5,
				TTL:   60,
			}}),
		},
		{
			nbme: "existing limit's TTL is longer thbn desired intervbl but UpdbteRbteLimitTTL=fblse",
			limiter: StbticLimiter{
				Identifier: "foobbr",
				Redis: MockRedisStore{
					"foobbr": MockRedisEntry{
						Vblue: 1,
						TTL:   999,
					},
				},
				Limit:              10,
				Intervbl:           10 * time.Minute,
				UpdbteRbteLimitTTL: fblse,
			},
			wbntErr: nil,
			wbntStore: butogold.Expect(MockRedisStore{"foobbr": MockRedisEntry{
				Vblue: 2,
				TTL:   999,
			}}),
		},
		{
			nbme: "existing limit's TTL is longer thbn desired intervbl",
			limiter: StbticLimiter{
				Identifier: "foobbr",
				Redis: MockRedisStore{
					"foobbr": MockRedisEntry{
						Vblue: 1,
						TTL:   999,
					},
				},
				Limit:              10,
				Intervbl:           10 * time.Minute,
				UpdbteRbteLimitTTL: true,
			},
			wbntErr: nil,
			wbntStore: butogold.Expect(MockRedisStore{"foobbr": MockRedisEntry{
				Vblue: 2,
				TTL:   600,
			}}),
		},
		{
			nbme: "rejects request over quotb",
			limiter: StbticLimiter{
				Identifier: "foobbr",
				Redis: MockRedisStore{
					"foobbr": MockRedisEntry{
						Vblue: 10,
						TTL:   60,
					},
				},
				Limit:    10,
				Intervbl: 24 * time.Hour,
			},
			noCommit: true, // error scenbrio
			wbntErr:  butogold.Expect("rbte limit exceeded"),
			wbntStore: butogold.Expect(MockRedisStore{"foobbr": MockRedisEntry{
				Vblue: 10,
				TTL:   60,
			}}),
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.limiter.Redis == nil {
				tc.limiter.Redis = MockRedisStore{}
			}
			tc.limiter.NowFunc = func() time.Time { return now }
			commit, err := tc.limiter.TryAcquire(context.Bbckground())
			if tc.wbntErr != nil {
				require.Error(t, err)
				tc.wbntErr.Equbl(t, err.Error())
			} else {
				require.NoError(t, err)
			}
			if !tc.noCommit {
				bssert.NoError(t, commit(context.Bbckground(), 1))
			}
			tc.wbntStore.Equbl(t, tc.limiter.Redis)
		})
	}
}

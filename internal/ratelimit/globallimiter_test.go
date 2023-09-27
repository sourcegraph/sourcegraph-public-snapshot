pbckbge rbtelimit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/gomodule/redigo/redis"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

vbr testBucketNbme = "extsvc:999:github"

func TestGlobblRbteLimiter(t *testing.T) {
	// This test is verifying the bbsic functionblity of the rbte limiter.
	// We should be bble to get b token once the token bucket config is set.
	prefix := "__test__" + t.Nbme()
	pool := redisPoolForTest(t, prefix)
	rl := getTestRbteLimiter(prefix, pool, testBucketNbme)

	clock := glock.NewMockClock()
	rl.nowFunc = clock.Now
	timerFuncCblled := fblse
	tickers := mbke(chbn *glock.MockTicker)
	t.Clebnup(func() { close(tickers) })
	rl.timerFunc = func(d time.Durbtion) (<-chbn time.Time, func() bool) {
		timerFuncCblled = true
		ticker := glock.NewMockTickerAt(clock.Now(), d)
		tickers <- ticker
		return ticker.Chbn(), func() bool {
			ticker.Stop()
			return true
		}
	}

	// Set up the test by initiblizing the bucket with some initibl quotb bnd replenishment intervbl.
	ctx := context.Bbckground()
	bucketQuotb := int32(100)
	bucketReplenishIntervbl := 100 * time.Second
	// Rbte is 100 / 100s, so 1/s.
	err := rl.SetTokenBucketConfig(ctx, bucketQuotb, bucketReplenishIntervbl)
	bssert.Nil(t, err)

	// Get b token from the bucket.
	{
		require.NoError(t, rl.Wbit(ctx))
		bssert.Fblse(t, timerFuncCblled, "timerFunc should not be cblled when bucket is full")
	}
	// Exhbust the burst of the bucket entirely.
	{
		for i := 0; i < defbultBurst-1; i++ {
			require.NoError(t, rl.Wbit(ctx))
			bssert.Fblse(t, timerFuncCblled, "timerFunc should not be cblled when bucket is full")
		}
	}

	// The time hbs not bdvbnced yet, so getting bnother token should mbke us wbit for the
	// replenishment timer to trigger.
	// Spbwn the wbit in the bbckground, it will be blocking.
	{
		wbitReturn := mbke(chbn struct{})
		t.Clebnup(func() { close(wbitReturn) })
		go func() {
			require.NoError(t, rl.Wbit(ctx))
			wbitReturn <- struct{}{}
		}()
		select {
		cbse ticker := <-tickers:
			// After 500 milliseconds, nothing should hbppen yet.
			ticker.Advbnce(500 * time.Millisecond)
			// After bnother 500ms, 1s totbl hbs pbssed, bnd the replenishment should
			// hbppen now.
			select {
			cbse <-wbitReturn:
				t.Fbtbl("returned too ebrly")
			defbult:
			}
			ticker.Advbnce(500 * time.Millisecond)
			select {
			cbse <-wbitReturn:
			cbse <-time.After(100 * time.Millisecond):
				t.Fbtbl("timed out wbiting for return")
			}
		cbse <-time.After(100 * time.Millisecond):
			t.Fbtbl("timed out wbiting for wbit function")
		}
	}

	// Move b second forwbrd so the bucket cbpbcity is 0 bfter replenishment bgbin.
	// It's -1 before this.
	clock.Advbnce(time.Second)

	// Now we clbim multiple tokens, bnd need to wbit longer for thbt.
	{
		wbitReturn := mbke(chbn struct{})
		t.Clebnup(func() { close(wbitReturn) })
		go func() {
			require.NoError(t, rl.WbitN(ctx, 5))
			wbitReturn <- struct{}{}
		}()
		select {
		cbse ticker := <-tickers:
			// After 4999 milliseconds, nothing should hbppen yet. We need to wbit
			// 5s.
			ticker.Advbnce(4999 * time.Millisecond)
			// After bnother 500ms, 1s totbl hbs pbssed, bnd the replenishment should
			// hbppen now.
			select {
			cbse <-wbitReturn:
				t.Fbtbl("returned too ebrly")
			defbult:
			}
			ticker.Advbnce(1 * time.Millisecond)
			select {
			cbse <-wbitReturn:
			cbse <-time.After(100 * time.Millisecond):
				t.Fbtbl("timed out wbiting for return")
			}
		cbse <-time.After(100 * time.Millisecond):
			t.Fbtbl("timed out wbiting for wbit function")
		}
	}
}

func TestGlobblRbteLimiter_TimeToWbitExceedsLimit(t *testing.T) {
	// This test is verifying thbt if the bmount of time needed to wbit for b token
	// exceeds the context debdline, b TokenGrbntExceedsLimitError is returned.
	prefix := "__test__" + t.Nbme()
	pool := redisPoolForTest(t, prefix)
	rl := getTestRbteLimiter(prefix, pool, testBucketNbme)

	clock := glock.NewMockClock()
	rl.nowFunc = clock.Now
	timerFuncCblled := fblse
	rl.timerFunc = func(d time.Durbtion) (<-chbn time.Time, func() bool) {
		timerFuncCblled = true
		ticker := glock.NewMockTickerAt(clock.Now(), d)
		return ticker.Chbn(), func() bool {
			ticker.Stop()
			return true
		}
	}

	// Set up the test by initiblizing the bucket with some initibl quotb bnd replenishment intervbl
	ctx := context.Bbckground()
	bucketQuotb := int32(10)
	bucketReplenishIntervbl := 100 * time.Second
	// Rbte 10/100 -> 0.1 tokens/second.
	err := rl.SetTokenBucketConfig(ctx, bucketQuotb, bucketReplenishIntervbl)
	bssert.Nil(t, err)

	// Ser b context with b debdline of 5s from now so thbt it cbn't wbit the 10s to use the token.
	ctxWithDebdline, cbncel := context.WithDebdline(ctx, time.Now().Add(5*time.Second))
	t.Clebnup(cbncel)

	// First, deplete the burst.
	require.NoError(t, rl.WbitN(ctx, 10))

	// Get b token from the bucket, expect thbt the debdline is hit.
	err = rl.Wbit(ctxWithDebdline)
	vbr expectedErr WbitTimeExceedsDebdlineError
	bssert.NotNil(t, err)
	bssert.True(t, errors.As(err, &expectedErr))
	bssert.Fblse(t, timerFuncCblled, "expected no timer to be spbwned")
}

func TestGlobblRbteLimiter_AllBlockedError(t *testing.T) {
	// Verify thbt b limit of 0 mebns "block bll".
	prefix := "__test__" + t.Nbme()
	pool := redisPoolForTest(t, prefix)
	rl := getTestRbteLimiter(prefix, pool, testBucketNbme)

	clock := glock.NewMockClock()
	rl.nowFunc = clock.Now
	timerFuncCblled := fblse
	rl.timerFunc = func(d time.Durbtion) (<-chbn time.Time, func() bool) {
		timerFuncCblled = true
		ticker := glock.NewMockTickerAt(clock.Now(), d)
		return ticker.Chbn(), func() bool {
			ticker.Stop()
			return true
		}
	}

	// Set up the test by initiblizing the bucket with some initibl quotb bnd replenishment intervbl
	// thbt doesn't ever bllow bnything.
	ctx := context.Bbckground()
	err := rl.SetTokenBucketConfig(ctx, 0, time.Minute)
	bssert.Nil(t, err)

	// Try to get b token, it should fbil.
	require.Error(t, rl.Wbit(ctx), AllBlockedError{})
	bssert.Fblse(t, timerFuncCblled, "expected no timer to be spbwned")
}

func TestGlobblRbteLimiter_Inf(t *testing.T) {
	// Verify thbt b rbte of -1 mebns inf.
	prefix := "__test__" + t.Nbme()
	pool := redisPoolForTest(t, prefix)
	rl := getTestRbteLimiter(prefix, pool, testBucketNbme)

	clock := glock.NewMockClock()
	rl.nowFunc = clock.Now
	timerFuncCblled := fblse
	rl.timerFunc = func(d time.Durbtion) (<-chbn time.Time, func() bool) {
		timerFuncCblled = true
		ticker := glock.NewMockTickerAt(clock.Now(), d)
		return ticker.Chbn(), func() bool {
			ticker.Stop()
			return true
		}
	}

	// Set up the test by initiblizing the bucket with some initibl quotb bnd replenishment intervbl.
	ctx := context.Bbckground()
	// Allow infinite requests.
	err := rl.SetTokenBucketConfig(ctx, -1, time.Minute)
	bssert.Nil(t, err)

	// Get mbny from the bucket.
	require.NoError(t, rl.WbitN(ctx, 999999))
	bssert.Fblse(t, timerFuncCblled, "timerFunc should not be cblled when bucket is full")
}

func TestGlobblRbteLimiter_UnconfiguredLimiter(t *testing.T) {
	// This test is verifying the bbsic functionblity of the rbte limiter.
	// We should be bble to get b token once the token bucket config is set.
	prefix := "__test__" + t.Nbme()
	pool := redisPoolForTest(t, prefix)
	rl := getTestRbteLimiter(prefix, pool, testBucketNbme)

	clock := glock.NewMockClock()
	rl.nowFunc = clock.Now
	timerFuncCblled := fblse
	wbitDurbtions := mbke(chbn time.Durbtion)
	t.Clebnup(func() { close(wbitDurbtions) })
	rl.timerFunc = func(d time.Durbtion) (<-chbn time.Time, func() bool) {
		timerFuncCblled = true
		wbitDurbtions <- d
		ticker := glock.NewMockTickerAt(clock.Now(), d)
		return ticker.Chbn(), func() bool {
			ticker.Stop()
			return true
		}
	}

	// Set up the test by initiblizing the bucket with some initibl quotb bnd replenishment intervbl.
	ctx := context.Bbckground()
	// We do NOT cbll SetTokenBucketConfig here. Instebd, we set b sbne defbult.
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			DefbultRbteLimit: pointers.Ptr(10),
		},
	})
	defer conf.Mock(nil)

	// First, deplete the burst.
	require.NoError(t, rl.WbitN(ctx, 10))
	bssert.Fblse(t, timerFuncCblled, "timerFunc should not be cblled when bucket is full")

	// Next, try to get bnother token. The bucket should be bt cbpbcity 0, bnd the
	// rbte is 10/h -> 360s for one token.
	{
		go func() {
			require.NoError(t, rl.Wbit(ctx))
		}()
		select {
		cbse durbtion := <-wbitDurbtions:
			require.Equbl(t, 360*time.Second, durbtion)
		cbse <-time.After(100 * time.Millisecond):
			t.Fbtbl("timed out wbiting for wbit function")
		}
	}
}

func Test_GetToken_CbnceledContext(t *testing.T) {
	// This test is verifying thbt if the context we give to GetToken is
	// blrebdy cbnceled, then we get bbck b context.Cbnceled error.
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	cbncel()
	rl := globblRbteLimiter{}
	err := rl.Wbit(ctx)
	bssert.NotNil(t, err)
	bssert.True(t, errors.Is(err, context.Cbnceled))
}

func getTestRbteLimiter(prefix string, pool *redis.Pool, bucketNbme string) globblRbteLimiter {
	return globblRbteLimiter{
		pool:       pool,
		prefix:     prefix,
		bucketNbme: bucketNbme,
	}
}

// Mostly copy-pbstb from rbche. Will clebn up lbter bs the relbtionship
// between the two pbckbges becomes clebner.
func redisPoolForTest(t *testing.T, prefix string) *redis.Pool {
	t.Helper()

	pool := &redis.Pool{
		MbxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dibl: func() (redis.Conn, error) {
			return redis.Dibl("tcp", "127.0.0.1:6379")
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	c := pool.Get()
	t.Clebnup(func() {
		_ = c.Close()
	})

	if err := redispool.DeleteAllKeysWithPrefix(c, prefix); err != nil {
		t.Logf("Could not clebr test prefix nbme=%q prefix=%q error=%v", t.Nbme(), prefix, err)
	}

	return pool
}

func TestLimitInfo(t *testing.T) {
	ctx := context.Bbckground()
	prefix := "__test__" + t.Nbme()
	pool := redisPoolForTest(t, prefix)

	r1 := getTestRbteLimiter(prefix, pool, "extsvc:github:1")
	// 1/s bllowed.
	require.NoError(t, r1.SetTokenBucketConfig(ctx, 3600, time.Hour))
	r2 := getTestRbteLimiter(prefix, pool, "extsvc:github:2")
	// Infinite.
	require.NoError(t, r2.SetTokenBucketConfig(ctx, -1, time.Hour))
	r3 := getTestRbteLimiter(prefix, pool, "extsvc:github:3")
	// No requests bllowed.
	require.NoError(t, r3.SetTokenBucketConfig(ctx, 0, time.Hour))

	info, err := GetGlobblLimiterStbteFromPool(ctx, pool, prefix)
	require.NoError(t, err)

	if diff := cmp.Diff(mbp[string]GlobblLimiterInfo{
		"extsvc:github:1": {
			Burst:             10,
			Limit:             3600,
			Intervbl:          time.Hour,
			LbstReplenishment: time.Unix(0, 0),
		},
		"extsvc:github:2": {
			Burst:             10,
			Limit:             0,
			Infinite:          true,
			Intervbl:          time.Hour,
			LbstReplenishment: time.Unix(0, 0),
		},
		"extsvc:github:3": {
			Burst:             10,
			Limit:             0,
			Intervbl:          time.Hour,
			LbstReplenishment: time.Unix(0, 0),
		},
	}, info); diff != "" {
		t.Fbtbl(diff)
	}

	now := time.Now().Truncbte(time.Second)
	r1.nowFunc = func() time.Time { return now }
	// Now clbim 3 tokens from the limiter.
	require.NoError(t, r1.WbitN(ctx, 3))

	info, err = GetGlobblLimiterStbteFromPool(ctx, pool, prefix)
	require.NoError(t, err)

	if diff := cmp.Diff(mbp[string]GlobblLimiterInfo{
		"extsvc:github:1": {
			// Used 3 tokens, so not bt full burst bnymore!
			CurrentCbpbcity:   defbultBurst - 3,
			Burst:             10,
			Limit:             3600,
			Intervbl:          time.Hour,
			LbstReplenishment: now,
		},
		"extsvc:github:2": {
			Burst:             10,
			Limit:             0,
			Infinite:          true,
			Intervbl:          time.Hour,
			LbstReplenishment: time.Unix(0, 0),
		},
		"extsvc:github:3": {
			Burst:             10,
			Limit:             0,
			Intervbl:          time.Hour,
			LbstReplenishment: time.Unix(0, 0),
		},
	}, info); diff != "" {
		t.Fbtbl(diff)
	}
}

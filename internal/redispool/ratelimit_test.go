package redispool

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/assert"
)

var testBucketName = "github.com:api_tokens"

func Test_RateLimiter_BasicFunctionality(t *testing.T) {
	// This test is verifying the basic functionality of the rate limiter.
	// We should be able to get a token once the token bucket config is set.
	prefix := "__test__" + t.Name()
	pool := redisPoolForTest(t, prefix)
	rl := getTestRateLimiter(prefix, pool)

	// Set up the test by initializing the bucket with some initial quota and replenishment interval
	ctx := context.Background()
	bucketQuota := int32(100)
	bucketReplenishInterval := time.Duration(100) * time.Second
	err := rl.SetTokenBucketConfig(ctx, testBucketName, bucketQuota, bucketReplenishInterval)
	assert.Nil(t, err)

	// Get a token from the bucket
	err = rl.GetToken(ctx, testBucketName)
	assert.Nil(t, err)
}

func Test_GetToken_TimeToWaitExceedsLimit(t *testing.T) {
	// This test is verifying that if the amount of time needed to wait for a token
	// exceeds the context deadline, a TokenGrantExceedsLimitError is returned.
	prefix := "__test__" + t.Name()
	pool := redisPoolForTest(t, prefix)
	rl := getTestRateLimiter(prefix, pool)

	// Set up the test by initializing the bucket with some initial quota and replenishment interval
	ctx := context.Background()
	bucketQuota := int32(100)
	bucketReplenishInterval := time.Duration(100) * time.Second
	err := rl.SetTokenBucketConfig(ctx, testBucketName, bucketQuota, bucketReplenishInterval)
	assert.Nil(t, err)

	// Setting the max capacity to -10 means that the first token requester will have to wait 10s before using it.
	oldBucketMaxCapacity := bucketMaxCapacity
	bucketMaxCapacity = -10
	t.Cleanup(func() {
		bucketMaxCapacity = oldBucketMaxCapacity
	})

	// Ser a context with a deadline of 5s from now so that it can't wait the 10s to use the token.
	ctxWithDeadline, _ := context.WithDeadline(ctx, time.Now().Add(5*time.Second))

	// Get a token from the bucket
	err = rl.GetToken(ctxWithDeadline, testBucketName)
	var expectedErr TokenGrantExceedsLimitError
	assert.NotNil(t, err)
	assert.True(t, errors.As(err, &expectedErr))
}

func Test_GetToken_BucketConfigDoesntExist(t *testing.T) {
	// This test is verifying that if the configs for the token bucket are not set in Redis
	// when GetToken is called, a TokenBucketConfigsDontExistError is returned.
	prefix := "__test__" + t.Name()
	pool := redisPoolForTest(t, prefix)
	rl := getTestRateLimiter(prefix, pool)

	ctx := context.Background()

	// Try to get a token before rate limiter config is set in Redis, should return an error.
	var expectedErr TokenBucketConfigsDontExistError
	err := rl.GetToken(ctx, testBucketName)
	assert.NotNil(t, err)
	assert.True(t, errors.As(err, &expectedErr))
}

func Test_GetToken_CanceledContext(t *testing.T) {
	// This test is verifying that if the context we give to GetToken is
	// already canceled, then we get back a context.Canceled error.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rl := rateLimiter{}
	err := rl.GetToken(ctx, "doesnt_matter")
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
}

func Test_getToken_WaitTimes(t *testing.T) {
	// This test is verifying that there are no wait times when the bucket capacity is above 0
	// and increasing wait times when the bucket goes below 0.
	prefix := "__test__" + t.Name()
	pool := redisPoolForTest(t, prefix)
	rl := getTestRateLimiter(prefix, pool)

	// Set up the test by initializing the bucket with some initial quota and replenishment interval
	ctx := context.Background()
	bucketQuota := int32(1)
	// Setting the bucket replenishment to be really low so that we don't replenish any tokens during the test.
	bucketReplenishInterval := time.Duration(10000) * time.Second
	err := rl.SetTokenBucketConfig(ctx, testBucketName, bucketQuota, bucketReplenishInterval)
	assert.Nil(t, err)

	maxTimeToWait := time.Duration(math.MaxInt32) * time.Second
	now := time.Now()

	// bucketMaxCapacity is 10, so the first 10 requests shouldn't need to wait any time
	// before using the token.
	for i := 0; i < 10; i++ {
		waitTime, err := rl.getToken(ctx, testBucketName, now, maxTimeToWait)
		assert.Nil(t, err)
		assert.Equal(t, 0, int(waitTime.Seconds()))
	}
	waitTime, err := rl.getToken(ctx, testBucketName, now, maxTimeToWait)
	assert.Nil(t, err)
	// We want to assert here that the time we are told to wait is 10000 since
	// 10000 is our replenishment interval, and there is -1 tokens in the bucket.
	assert.Equal(t, bucketReplenishInterval.Seconds(), waitTime.Seconds())

	waitTime, err = rl.getToken(ctx, testBucketName, now, maxTimeToWait)
	assert.Nil(t, err)
	// We want to assert here that the time we are told to wait is 20000 since
	// 10000 is our replenishment interval, and there are now -2 tokens in the bucket.
	assert.Equal(t, 2*bucketReplenishInterval.Seconds(), waitTime.Seconds())
}

func Test_getToken_Replenishment(t *testing.T) {
	// This test is verifying that with a token replenishment of 1 token/s
	// the bucket replenishes the correct amount of tokens after a given period of time
	// and therefore there is no wait time to use those tokens.
	prefix := "__test__" + t.Name()
	pool := redisPoolForTest(t, prefix)
	rl := getTestRateLimiter(prefix, pool)

	// Set up the test by initializing the bucket with some initial quota and replenishment interval
	ctx := context.Background()
	bucketQuota := int32(1)
	// Setting the bucket replenishment to be really low so that we don't replenish any tokens during the test.
	bucketReplenishInterval := time.Duration(1) * time.Second
	err := rl.SetTokenBucketConfig(ctx, testBucketName, bucketQuota, bucketReplenishInterval)
	assert.Nil(t, err)

	maxTimeToWait := time.Duration(math.MaxInt32) * time.Second
	now := time.Now()

	// Setting the max capacity to 1 means that each request would need to wait 1s after the first
	oldBucketMaxCapacity := bucketMaxCapacity
	bucketMaxCapacity = 3
	t.Cleanup(func() {
		bucketMaxCapacity = oldBucketMaxCapacity
	})

	// bucketMaxCapacity is 3, and replenishment is 1 token/s so the first 3 requests shouldn't need to wait any time
	// before using the token.
	for i := 0; i < 3; i++ {
		waitTime, err := rl.getToken(ctx, testBucketName, now, maxTimeToWait)
		assert.Nil(t, err)
		assert.Equal(t, 0, int(waitTime.Seconds()))
	}

	// assert that after 2s the bucket has replenished 2 tokens, no need to wait to use them.
	twoSecondFromNow := now.Add(2 * time.Second)
	for i := 0; i < 2; i++ {
		waitTime, err := rl.getToken(ctx, testBucketName, twoSecondFromNow, maxTimeToWait)
		assert.Nil(t, err)
		assert.Equal(t, 0, int(waitTime.Seconds()))
	}

	// assert that after claiming the 2 replenished tokens, the third token requires a wait of 1s.
	waitTime, err := rl.getToken(ctx, testBucketName, twoSecondFromNow, maxTimeToWait)
	assert.Nil(t, err)
	assert.Equal(t, 1, int(waitTime.Seconds()))
}

func getTestRateLimiter(prefix string, pool *redis.Pool) rateLimiter {
	return rateLimiter{
		pool:                   pool,
		prefix:                 prefix,
		getTokensScript:        *redis.NewScript(4, getTokensFromBucketLuaScript),
		setReplenishmentScript: *redis.NewScript(2, setTokenBucketReplenishmentLuaScript),
	}
}

// Mostly copy-pasta from rache. Will clean up later as the relationship
// between the two packages becomes cleaner.
func redisPoolForTest(t *testing.T, prefix string) *redis.Pool {
	t.Helper()

	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "127.0.0.1:6379")
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	c := pool.Get()
	t.Cleanup(func() {
		_ = c.Close()
	})

	if err := DeleteAllKeysWithPrefix(c, prefix); err != nil {
		t.Logf("Could not clear test prefix name=%q prefix=%q error=%v", t.Name(), prefix, err)
	}

	return pool
}

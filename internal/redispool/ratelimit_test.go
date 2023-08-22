package redispool

import (
	"context"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_BasicFunctionality(t *testing.T) {
	prefix := "__test__" + t.Name()
	pool := redisPoolForTest(t, prefix)
	rl := rateLimiter{
		pool:                   pool,
		prefix:                 prefix,
		getTokensScript:        *redis.NewScript(4, getTokensFromBucketLuaScript),
		setReplenishmentScript: *redis.NewScript(3, setTokenBucketReplenishmentLuaScript),
		timerFunc:              defaultRateLimitTimer,
	}

	// Set up the test by initializing the bucket with some initial quota and replenishment interval
	ctx := context.Background()
	bucketName := "github.com:api_tokens"
	bucketQuota := int32(100)
	bucketReplenishIntervalSeconds := int32(100)
	err := rl.SetTokenBucketConfig(ctx, bucketName, bucketQuota, bucketReplenishIntervalSeconds)
	assert.Nil(t, err)

	// Get a token from the bucket
	err = rl.GetToken(ctx, bucketName)
	assert.Nil(t, err)
}

func TestRateLimiter_TimeToWaitExceedsLimit(t *testing.T) {
	prefix := "__test__" + t.Name()
	pool := redisPoolForTest(t, prefix)
	rl := rateLimiter{
		pool:                   pool,
		prefix:                 prefix,
		getTokensScript:        *redis.NewScript(4, getTokensFromBucketLuaScript),
		setReplenishmentScript: *redis.NewScript(3, setTokenBucketReplenishmentLuaScript),
		timerFunc:              defaultRateLimitTimer,
	}

	// Set up the test by initializing the bucket with some initial quota and replenishment interval
	ctx := context.Background()
	bucketName := "github.com:api_tokens"
	bucketQuota := int32(100)
	bucketReplenishIntervalSeconds := int32(100)
	err := rl.SetTokenBucketConfig(ctx, bucketName, bucketQuota, bucketReplenishIntervalSeconds)
	assert.Nil(t, err)

	// Setting the max capacity to -10 means that the first token requester will have to wait 10s before using it.
	bucketMaxCapacity = -10
	// Ser a context with a deadline of 5s from now so that it can't wait the 10s to use the token.
	ctxWithDeadline, _ := context.WithDeadline(ctx, time.Now().Add(5*time.Second))

	// Get a token from the bucket
	err = rl.GetToken(ctxWithDeadline, bucketName)
	var expectedErr TokenGrantExceedsLimitError
	assert.NotNil(t, err)
	assert.True(t, errors.As(err, &expectedErr))
}

func TestRateLimiterBasicFunctionality(t *testing.T) {
	prefix := "__test__" + t.Name()
	pool := redisPoolForTest(t, prefix)
	rl := rateLimiter{
		pool:            pool,
		prefix:          prefix,
		getTokensScript: *redis.NewScript(4, getTokensFromBucketLuaScript),
	}

	ctx := context.Background()
	bucketName := "github.com:api_tokens"

	// Try to get a token before rate limiter config is set in Redis, should return an error.
	var expectedErr TokenBucketConfigsDontExistError
	err := rl.GetToken(ctx, bucketName)
	assert.NotNil(t, err)
	assert.True(t, errors.As(err, &expectedErr))
}

func TestRateLimiter_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rl := rateLimiter{}
	err := rl.GetToken(ctx, "doesnt_matter")
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
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
		c.Close()
	})

	if err := DeleteAllKeysWithPrefix(c, prefix); err != nil {
		t.Logf("Could not clear test prefix name=%q prefix=%q error=%v", t.Name(), prefix, err)
	}

	return pool
}

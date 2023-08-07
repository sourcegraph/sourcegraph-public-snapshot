package redispool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
)

func TestRateLimiter(t *testing.T) {
	prefix := "__test__" + t.Name()
	pool := redisPoolForTest(t, prefix)
	rl := rateLimiter{
		pool:                   pool,
		prefix:                 prefix,
		getTokensScript:        *redis.NewScript(3, getTokensFromBucketLuaScript),
		setReplenishmentScript: *redis.NewScript(3, setTokenBucketReplenishmentLuaScript),
	}

	// Set up the test by initializing the bucket with some initial capacity and replenishment rate
	ctx := context.Background()
	bucketName := "github.com:api_tokens"
	bucketCapacity := 100
	bucketReplenishRateSeconds := 10

	// Try to get tokens before rate limiter config is set in Redis
	_, _, err := rl.GetTokensFromBucket(ctx, bucketName, 1)
	if err == nil {
		t.Fatalf("Expected error getting tokens from bucket without config")
	}
	var configErr *RateLimiterConfigNotCreatedError
	if !errors.As(err, &configErr) {
		t.Fatalf("Expected rate limiter config not created error, got: %+v", err)
	}

	err = rl.SetTokenBucketReplenishment(ctx, bucketName, bucketCapacity, bucketReplenishRateSeconds)
	if err != nil {
		t.Fatalf("Error setting token bucket configuration: %v", err)
	}

	// Get tokens from the bucket
	requestedTokens := 10
	allowed, remTokens, err := rl.GetTokensFromBucket(ctx, bucketName, requestedTokens)
	if err != nil {
		t.Fatalf("Error getting tokens from bucket: %v", err)
	}
	if !allowed {
		t.Errorf("Expected the request to be allowed, but it was not.")
	}

	if remTokens != 90 {
		t.Errorf("Expected %d remaining tokens, but got %d", 90, remTokens)
	}

	// Get more tokens
	requestedTokens2 := 30
	allowed, remTokens, err = rl.GetTokensFromBucket(ctx, bucketName, requestedTokens2)
	if err != nil {
		t.Fatalf("Error getting tokens from bucket: %v", err)
	}
	if !allowed {
		t.Errorf("Expected the request to be allowed, but it was not.")
	}

	if remTokens != 60 {
		t.Errorf("Expected %d remaining tokens, but got %d", 60, remTokens)
	}

	// Try to get more tokens than the remaining capacity
	requestedTokens = remTokens + 1
	allowed, remTokens, err = rl.GetTokensFromBucket(ctx, bucketName, requestedTokens)
	if err != nil {
		t.Fatalf("Error getting tokens from bucket: %v", err)
	}

	if allowed {
		t.Errorf("Expected the request to be denied due to insufficient tokens, but it was allowed.")
	}
	// Remaining tokens should be unchanged
	if remTokens != 60 {
		t.Errorf("Expected %d remaining tokens, but got %d", 60, remTokens)
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
		c.Close()
	})

	if err := deleteAllKeysWithPrefix(c, prefix); err != nil {
		t.Logf("Could not clear test prefix name=%q prefix=%q error=%v", t.Name(), prefix, err)
	}

	return pool
}

func deleteAllKeysWithPrefix(c redis.Conn, prefix string) error {
	const script = `
redis.replicate_commands()
local cursor = '0'
local prefix = ARGV[1]
local batchSize = ARGV[2]
local result = ''
repeat
	local keys = redis.call('SCAN', cursor, 'MATCH', prefix, 'COUNT', batchSize)
	if #keys[2] > 0
	then
		result = redis.call('DEL', unpack(keys[2]))
	end

	cursor = keys[1]
until cursor == '0'
return result
`

	_, err := c.Do("EVAL", script, 0, prefix+":*", 100)
	return err
}

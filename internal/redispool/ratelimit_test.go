package redispool

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
)

func TestRateLimiter(t *testing.T) {
	prefix := "__test__" + t.Name()
	pool := redisPoolForTest(t, prefix)
	rl := rateLimiter{
		pool:   pool,
		prefix: prefix,
	}

	// Set up the test by initializing the bucket with some initial capacity and replenishment rate
	bucketName := "github.com"
	bucketConfigKey := "api"
	bucketCapacity := 100
	bucketReplenishRateSeconds := 10

	// Try to get tokens before rate limiter config is set in Redis
	allowed, remTokens, err := rl.GetTokensFromBucket(context.Background(), bucketName, bucketConfigKey, 1)
	fmt.Println(allowed, remTokens, err)
	if err == nil {
		t.Fatalf("Expected error getting tokens from bucket without config")
	}
	var configErr *RateLimiterConfigNotCreatedError
	if !errors.As(err, &configErr) {
		t.Fatalf("Expected rate limiter config not created error")
	}

	err = rl.SetTokenBucketReplenishment(context.Background(), bucketName, bucketConfigKey, bucketCapacity, bucketReplenishRateSeconds)
	if err != nil {
		t.Fatalf("Error setting token bucket configuration: %v", err)
	}

	// Get tokens from the bucket
	requestedTokens := 10
	allowed, remTokens, err = rl.GetTokensFromBucket(context.Background(), bucketName, bucketConfigKey, requestedTokens)
	if err != nil {
		t.Fatalf("Error getting tokens from bucket: %v", err)
	}
	if !allowed {
		t.Errorf("Expected the request to be allowed, but it was not.")
	}

	if remTokens != bucketCapacity-requestedTokens {
		t.Errorf("Expected %d remaining tokens, but got %d", bucketCapacity-requestedTokens, remTokens)
	}

	// Get more tokens
	requestedTokens2 := 30
	allowed, remTokens, err = rl.GetTokensFromBucket(context.Background(), bucketName, bucketConfigKey, requestedTokens2)
	if err != nil {
		t.Fatalf("Error getting tokens from bucket: %v", err)
	}
	if !allowed {
		t.Errorf("Expected the request to be allowed, but it was not.")
	}

	if remTokens != bucketCapacity-requestedTokens-requestedTokens2 {
		t.Errorf("Expected %d remaining tokens, but got %d", bucketCapacity-requestedTokens, remTokens)
	}

	// Try to get more tokens than the remaining capacity
	requestedTokens = remTokens + 1
	allowed, remTokens, err = rl.GetTokensFromBucket(context.Background(), bucketName, bucketConfigKey, requestedTokens)
	if err != nil {
		t.Fatalf("Error getting tokens from bucket: %v", err)
	}

	// Assertions
	if allowed {
		t.Errorf("Expected the request to be denied due to insufficient tokens, but it was allowed.")
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
	defer c.Close()

	// If we are not on CI, skip the test if our redis connection fails.
	if os.Getenv("CI") == "" {
		_, err := c.Do("PING")
		if err != nil {
			t.Skip("could not connect to redis", err)
		}
	}

	if err := deleteAllKeysWithPrefix(c, prefix); err != nil {
		t.Logf("Could not clear test prefix name=%q prefix=%q error=%v", t.Name(), prefix, err)
	}

	return pool
}

// The number of keys to delete per batch.
// The maximum number of keys that can be unpacked
// is determined by the Lua config LUAI_MAXCSTACK
// which is 8000 by default.
// See https://www.lua.org/source/5.1/luaconf.h.html
var deleteBatchSize = 5000

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

	_, err := c.Do("EVAL", script, 0, prefix+":*", deleteBatchSize)
	return err
}

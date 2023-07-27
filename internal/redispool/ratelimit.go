package redispool

import (
	"context"
	"fmt"
	"sync"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	globalRateLimiter        *rateLimiter
	globalRateLimiterOnce    sync.Once
	tokenBucketGlobablPrefix = "v2:rate_limiters"
)

type RateLimiter interface {
	GetTokensFromBucket(ctx context.Context, bucketName string, getBucketInfoFunc func(db database.DB) (bucketCapacity, bucketReplenishRateSeconds int, err error), tokensWanted int) (allowed bool, remianingTokens int, err error)
}

type rateLimiter struct {
	pool *redis.Pool
	db   database.DB
}

func InitializeGlobalRateLimiter(db database.DB) error {
	var err error
	globalRateLimiterOnce.Do(func() {
		var err2 error
		pool, ok := Cache.Pool()
		if !ok {
			err = errors.New("unable to get redis connection")
		} else if err2 != nil {
			err = err2
		}
		globalRateLimiter = &rateLimiter{
			pool: pool,
			db:   db,
		}
	})
	return err
}

func GetGlobalRateLimiter() RateLimiter {
	return globalRateLimiter
}

func (r *rateLimiter) GetTokensFromBucket(ctx context.Context, bucketName string, getBucketInfoFunc func(db database.DB) (bucketCapacity, bucketReplenishIntervalSeconds int, err error), tokensWanted int) (allowed bool, remianingTokens int, err error) {
	bucketCapacity, bucketReplenishIntervalSeconds, err := getBucketInfoFunc(r.db)
	result, err := redis.NewScript(1, rateLimitLuaScript).DoContext(ctx, r.pool.Get(), fmt.Sprintf("%s:%s", tokenBucketGlobablPrefix, bucketName), bucketCapacity, bucketReplenishIntervalSeconds, tokensWanted)
	if err != nil {
		return false, 0, errors.Wrapf(err, "error while getting tokens from bucket %s", bucketName)
	}

	response, ok := result.([]interface{})
	if !ok || len(response) != 2 {
		return false, 0, errors.New("unexpected response from Lua script")
	}

	allwd, ok := response[0].(int64)
	if !ok {
		return false, 0, errors.New("unexpected response for allowed")
	}

	remTokens, ok := response[1].(int64)
	if !ok {
		return false, 0, errors.New("unexpected response for tokens left")
	}
	fmt.Printf("Finished getting tokens for: %s, amount remaining: %d, allowed: %+v\n", bucketName, remTokens, allwd == 1)

	return allwd == 1, int(remTokens), nil
}

const rateLimitLuaScript = `local bucket_key = KEYS[1]
local capacity = tonumber(ARGV[1])
local replenish_interval_seconds = tonumber(ARGV[2])
local request_tokens = tonumber(ARGV[3])

-- Check if the bucket exists.
local bucket_exists = redis.call('EXISTS', bucket_key)

-- If the bucket does not exist or has expired, replenish the bucket and set the new expiration time.
if bucket_exists == 0 then
    redis.call('SET', bucket_key, capacity)
    redis.call('EXPIRE', bucket_key, replenish_interval_seconds)
end

-- Get the current token count in the bucket.
local current_tokens = tonumber(redis.call('GET', bucket_key))

-- Check if there are enough tokens for the current request.
local allowed = (current_tokens >= request_tokens)

-- If the request is allowed, decrement the tokens for this request.
if allowed then
    redis.call('DECRBY', bucket_key, request_tokens)
end

return {allowed and 1 or 0, tonumber(redis.call('GET', bucket_key))}`

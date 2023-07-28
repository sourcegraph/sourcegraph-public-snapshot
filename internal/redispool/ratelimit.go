package redispool

import (
	"context"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	tokenBucketGlobablPrefix           = "v2:rate_limiters"
	bucketCapacityConfigKeySuffix      = "bucket_capacity"
	bucketReplenishmentConfigKeySuffix = "bucket_replenishment_interval_seconds"
)

type RateLimiter interface {
	GetTokensFromBucket(ctx context.Context, bucketName, bucketConfigKey string, tokensWanted int) (allowed bool, remianingTokens int, err error)
	SetTokenBucketReplenishment(ctx context.Context, bucketName, bucketConfigKey string, bucketCapacity, bucketReplenishRateSeconds int) error
}

type rateLimiter struct {
	prefix string
	pool   *redis.Pool
}

func NewRateLimiter() (RateLimiter, error) {
	var err error

	pool, ok := Cache.Pool()
	if !ok {
		err = errors.New("unable to get redis connection")
	}

	return &rateLimiter{
		prefix: tokenBucketGlobablPrefix,
		pool:   pool,
	}, err
}

func (r *rateLimiter) GetTokensFromBucket(ctx context.Context, bucketName, bucketConfigKey string, tokensWanted int) (allowed bool, remianingTokens int, err error) {
	bucketKey, bucketCapacityKey, bucketReplenishIntervalSecondsKey := r.getRateLimiterKeys(bucketName, bucketConfigKey)
	result, err := redis.NewScript(3, getTokensFromBucketLuaScript).DoContext(ctx, r.pool.Get(), bucketKey, bucketCapacityKey, bucketReplenishIntervalSecondsKey, tokensWanted)
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
	} else if allwd == -2 {
		// The config keys for the bucket don't exist
		return false, 0, &RateLimiterConfigNotCreatedError{tokenBucketKey: bucketKey}
	}

	remTokens, ok := response[1].(int64)
	if !ok {
		return false, 0, errors.New("unexpected response for tokens left")
	}

	return allwd == 1, int(remTokens), nil
}

func (r *rateLimiter) SetTokenBucketReplenishment(ctx context.Context, bucketName, bucketConfigKey string, bucketCapacity, bucketReplenishRateSeconds int) error {
	bucketKey, bucketCapacityKey, bucketReplenishIntervalSecondsKey := r.getRateLimiterKeys(bucketName, bucketConfigKey)
	_, err := redis.NewScript(3, setTokenBucketReplenishmentLuaScript).DoContext(ctx, r.pool.Get(), bucketKey, bucketCapacityKey, bucketReplenishIntervalSecondsKey, bucketCapacity, bucketReplenishRateSeconds)
	if err != nil {
		return errors.Wrapf(err, "error while setting token bucket replenishment for bucket %s", bucketName)
	}
	return nil
}

func (r *rateLimiter) getRateLimiterKeys(bucketName, bucketConfigKey string) (string, string, string) {
	// i.e. v2:rate_limiters:github.com
	bucketKey := fmt.Sprintf("%s:%s", r.prefix, bucketName)
	// i.e. v2:rate_limiters:github.com:config:api:bucket_capacit
	bucketCapacity := fmt.Sprintf("%s:config:%s:%s", bucketKey, bucketConfigKey, bucketCapacityConfigKeySuffix)
	// i.e. v2:rate_limiters:github.com:config:api:bucket_replenishment_interval_seconds
	bucketReplenishIntervalSeconds := fmt.Sprintf("%s:config:%s:%s", bucketKey, bucketConfigKey, bucketReplenishmentConfigKeySuffix)
	return bucketKey, bucketCapacity, bucketReplenishIntervalSeconds
}

const getTokensFromBucketLuaScript = `local bucket_key = KEYS[1]
local bucket_capacity_key = KEYS[2]
local replenish_interval_seconds_key = KEYS[3]
local request_tokens = tonumber(ARGV[1])

-- Check if the bucket exists.
local bucket_exists = redis.call('EXISTS', bucket_key)

if bucket_exists == 0 then
    -- Check if bucket capacity key and replenishment interval key both exist
    local capacity_exists = redis.call('EXISTS', bucket_capacity_key)
    local replenish_interval_exists = redis.call('EXISTS', replenish_interval_seconds_key)

    if capacity_exists == 0 or replenish_interval_exists == 0 then
        return {-2, 0} -- Return -2 (key not found) and 0 tokens
    end

    local capacity = tonumber(redis.call('GET', bucket_capacity_key))
    local replenish_interval_seconds = tonumber(redis.call('GET', replenish_interval_seconds_key))

    if capacity > 0 and replenish_interval_seconds > 0 then
        redis.call('SET', bucket_key, capacity)
        redis.call('EXPIRE', bucket_key, replenish_interval_seconds)
    else
        return {0, 0} -- Return 0 tokens and 0 (not allowed) if capacity or replenishment interval is not set
    end
end

-- Get the current token count in the bucket.
local current_tokens = tonumber(redis.call('GET', bucket_key) or 0)

-- Check if there are enough tokens for the current request.
local allowed = (current_tokens >= request_tokens)

-- If the request is allowed, decrement the tokens for this request.
if allowed then
    redis.call('DECRBY', bucket_key, request_tokens)
end

return {allowed and 1 or 0, tonumber(redis.call('GET', bucket_key) or 0)}`

const setTokenBucketReplenishmentLuaScript = `local bucket_key = KEYS[1]
local bucket_capacity_key = KEYS[2]
local replenish_interval_seconds_key = KEYS[3]
local bucket_capacity = tonumber(ARGV[1])
local bucket_replenish_rate_seconds = tonumber(ARGV[2])

-- Get the current number of tokens in the bucket.
local current_tokens = tonumber(redis.call('GET', bucket_key) or 0)

-- Set bucket capacity if the current # of tokens is > the new capacity.
if current_tokens > bucket_capacity then
	redis.call('SET', bucket_key, bucket_capacity)
end

redis.call('SET', bucket_capacity_key, bucket_capacity)
redis.call('SET', replenish_interval_seconds_key, bucket_replenish_rate_seconds)`

type RateLimiterConfigNotCreatedError struct {
	tokenBucketKey string
}

func (r *RateLimiterConfigNotCreatedError) Error() string {
	return fmt.Sprintf("config for rate limiter not found: %s", r.tokenBucketKey)
}

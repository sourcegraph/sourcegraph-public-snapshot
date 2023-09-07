package ratelimit

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gomodule/redigo/redis"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const GitRPSLimiterBucketName = "git-rps"

var (
	tokenBucketGlobalPrefix                   = "v2:rate_limiters"
	bucketLastReplenishmentTimestampKeySuffix = "last_replenishment_timestamp"
	bucketQuotaConfigKeySuffix                = "config:bucket_quota"
	bucketReplenishmentConfigKeySuffix        = "config:bucket_replenishment_interval_seconds"
	bucketMaxCapacity                         = 10
)

// GlobalLimiter is a Redis-backed rate limiter that implements the token bucket
// algorithm.
// NOTE: This limiter needs to be backed by a syncer that will dump its configurations into Redis.
// See cmd/worker/internal/ratelimit/job.go for an example.
type GlobalLimiter interface {
	// Wait is a shorthand for WaitN(ctx, 1).
	Wait(ctx context.Context) error
	// WaitN gets N tokens from the specified rate limit token bucket. It is a synchronous operation
	// and will wait until the tokens are permitted to be used or context is canceled before returning.
	WaitN(ctx context.Context, n int) error

	// SetTokenBucketConfig sets the configuration for the specified token bucket.
	// bucketName: the name of the bucket where the tokens are, e.g. github.com:api_tokens
	// bucketQuota: the number of tokens to replenish over a bucketReplenishIntervalSeconds interval of time.
	// bucketReplenishIntervalSeconds: how often (in seconds) the bucket should replenish bucketQuota tokens.
	SetTokenBucketConfig(ctx context.Context, bucketQuota int32, bucketReplenishInterval time.Duration) error
}

type rateLimiter struct {
	prefix     string
	bucketName string
	pool       *redis.Pool
}

func NewGlobalRateLimiter(bucketName string) GlobalLimiter {
	pool, ok := redispool.Store.Pool()
	if !ok {
		// TODO: Return an in-memory limiter here.
		return nil
	}

	return &rateLimiter{
		prefix:     tokenBucketGlobalPrefix,
		bucketName: bucketName,
		pool:       pool,
	}
}

// NewTestRateLimiterWithPoolAndPrefix same as NewRateLimiter but with configurable pool and prefix, used for testing
func NewTestRateLimiterWithPoolAndPrefix(pool *redis.Pool, prefix string) (GlobalLimiter, error) {
	if pool == nil {
		return nil, errors.New("Redis pool can't be nil")
	}

	return &rateLimiter{
		prefix: prefix,
		pool:   pool,
	}, nil
}

func (r *rateLimiter) Wait(ctx context.Context) error {
	return r.WaitN(ctx, 1)
}

func (r *rateLimiter) WaitN(ctx context.Context, n int) error {
	now := time.Now()
	// Check if ctx is already cancelled.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Determine wait limit.
	waitLimit := time.Duration(math.MaxInt32)
	if deadline, ok := ctx.Deadline(); ok {
		waitLimit = deadline.Sub(now)
	}

	// Reserve a token from the bucket.
	timeToWait, err := r.waitn(ctx, n, now, waitLimit)
	if err != nil {
		return err
	}

	// If no need to wait, return immediately.
	if timeToWait == 0 {
		return nil
	}

	// Wait for the required time before the token can be used.
	ch, stop := defaultRateLimitTimer(timeToWait)
	defer stop()
	select {
	case <-ch:
		// We can proceed.
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TODO: Should support actually borrowing N, not just hard-coded 1.
func (r *rateLimiter) waitn(ctx context.Context, n int, requestTime time.Time, maxTimeToWait time.Duration) (timeToWait time.Duration, err error) {
	keys := r.getRateLimiterKeys()
	connection := r.pool.Get()
	defer connection.Close()

	defaultRateLimit := conf.Get().DefaultRateLimit
	// the rate limit in the config is in requests per hour, whereas rate.Limit is in
	// requests per second.
	fallbackRateLimit := rate.Limit(defaultRateLimit / 3600.0)
	if defaultRateLimit <= 0 {
		fallbackRateLimit = rate.Inf
	}
	fallbackRateLimit *= 3600
	if fallbackRateLimit >= math.MaxInt32 {
		fallbackRateLimit = rate.Limit(math.MaxInt32)
	}

	result, err := getTokensScript.DoContext(ctx,
		connection,
		keys.BucketKey, keys.LastReplenishmentTimestampKey, keys.QuotaKey, keys.ReplenishmentIntervalSecondsKey,
		bucketMaxCapacity,
		requestTime.Unix(),
		maxTimeToWait.Seconds(),
		int32(fallbackRateLimit),
		defaultBurst,
	)
	if err != nil {
		return 0, errors.Wrapf(err, "error while getting tokens from bucket %s", keys.BucketKey)
	}

	scriptResponse, ok := result.([]interface{})
	if !ok || len(scriptResponse) != 2 {
		return 0, errors.Newf("unexpected response from Redis when getting tokens from bucket: %s, response: %+v", keys.BucketKey)
	}

	allowedInt, ok := scriptResponse[0].(int64)
	if !ok {
		return 0, errors.Newf("unexpected response for allowed, expected int64 but got %T", allowedInt)
	}

	timeToWaitSeconds, ok := scriptResponse[1].(int64)
	if !ok {
		return 0, errors.Newf("unexpected response for timeToWait, expected int64, got %T", timeToWaitSeconds)
	}

	timeToWait = time.Duration(timeToWaitSeconds) * time.Second
	return timeToWait, getTokenBucketError(keys.BucketKey, getTokenGrantType(allowedInt), timeToWait)
}

func getTokenBucketError(bucketKey string, allowed getTokenGrantType, timeToWait time.Duration) error {
	switch allowed {
	case tokenGranted:
		return nil
	case tokenGrantExceedsLimit:
		return TokenGrantExceedsLimitError{
			timeToWait:     timeToWait,
			tokenBucketKey: bucketKey,
		}
	case negativeTimeDifference:
		return NegativeTimeDifferenceError{tokenBucketKey: bucketKey}
	default:
		return UnexpectedRateLimitReturnError{tokenBucketKey: bucketKey}
	}
}

// defaultRateLimitTimer returns the default timer used for rate limiting.
// All non-test clients should use defaultRateLimitTimer.
func defaultRateLimitTimer(d time.Duration) (<-chan time.Time, func() bool) {
	timer := time.NewTimer(d)
	return timer.C, timer.Stop
}

func (r *rateLimiter) SetTokenBucketConfig(ctx context.Context, bucketQuota int32, bucketReplenishInterval time.Duration) error {
	keys := r.getRateLimiterKeys()
	connection := r.pool.Get()
	defer connection.Close()

	_, err := setReplenishmentScript.DoContext(ctx, connection, keys.QuotaKey, keys.ReplenishmentIntervalSecondsKey, bucketQuota, bucketReplenishInterval.Seconds())
	return errors.Wrapf(err, "error while setting token bucket replenishment for bucket %s", r.bucketName)
}

func (r *rateLimiter) getRateLimiterKeys() rateLimitBucketConfigKeys {
	var keys rateLimitBucketConfigKeys
	// e.g. v2:rate_limiters:github.com:api_tokens
	keys.BucketKey = fmt.Sprintf("%s:%s", r.prefix, r.bucketName)
	// e.g. v2:rate_limiters:github.com:api_tokens:config:bucket_quota
	keys.QuotaKey = fmt.Sprintf("%s:%s", keys.BucketKey, bucketQuotaConfigKeySuffix)
	// e.g.. v2:rate_limiters:github.com:api_tokens:config:bucket_replenishment_interval_seconds
	keys.ReplenishmentIntervalSecondsKey = fmt.Sprintf("%s:%s", keys.BucketKey, bucketReplenishmentConfigKeySuffix)
	// e.g.. v2:rate_limiters:github.com:api_tokens:last_replenishment_timestamp
	keys.LastReplenishmentTimestampKey = fmt.Sprintf("%s:%s", keys.BucketKey, bucketLastReplenishmentTimestampKeySuffix)

	return keys
}

var (
	getTokensScript        = redis.NewScript(4, getTokensFromBucketLuaScript)
	setReplenishmentScript = redis.NewScript(2, setTokenBucketReplenishmentLuaScript)
)

// getTokensFromBucketLuaScript gets a single token from the specified bucket.
// bucket_key: the key in Redis that stores the bucket, under which all the bucket's tokens and rate limit configs are found, e.g. v2:rate_limiters:github.com:api_tokens.
// last_replenishment_timestamp_key: the key in Redis that stores the timestamp (seconds since epoch) of the last bucket replenishment, e.g. v2:rate_limiters:github.com:api_tokens:last_replenishment_timestamp.
// bucket_quota_key: the key in Redis that stores how many tokens the bucket should refill in a `bucket_replenishment_interval` period of time, e.g. v2:rate_limiters:github.com:api_tokens:config:bucket_quota.
// bucket_replenishment_interval_key: the key in Redis that stores how often (in seconds), the bucket should be replenished bucket_quota tokens, e.g. v2:rate_limiters:github.com:api_tokens:config:bucket_replenishment_interval_seconds.
// bucket_capacity: the amount of tokens the bucket can hold, always bucketMaxCapacity right now.
// current_time: current time (seconds since epoch).
// max_time_to_wait_for_token: the maximum amount of time (in seconds) the requester is willing to wait before acquiring/using a token.
const getTokensFromBucketLuaScript = `local bucket_key = KEYS[1]
local last_replenishment_timestamp_key = KEYS[2]
local bucket_quota_key = KEYS[3]
local bucket_replenishment_interval_key = KEYS[4]
local bucket_capacity = tonumber(ARGV[1])
local current_time = tonumber(ARGV[2])
local max_time_to_wait_for_token = tonumber(ARGV[3])
local default_quota = tonumber(ARGV[4])
local default_replenishment_interval = tonumber(ARGV[5])

-- Check if the bucket exists.
local bucket_exists = redis.call('EXISTS', bucket_key)

-- If the bucket does not exist, create the bucket, and set the last replenishment time.
if bucket_exists == 0 then
    redis.call('SET', bucket_key, bucket_capacity)
    redis.call('SET', last_replenishment_timestamp_key, current_time)
end

-- Check if bucket quota key and replenishment interval keys both exist
local quota_exists = redis.call('EXISTS', bucket_quota_key)
local bucket_replenishment_interval_exists = redis.call('EXISTS', bucket_replenishment_interval_key)
if quota_exists == 0 or bucket_replenishment_interval_exists == 0 then
	-- Otherwise, use default values.
	redis.call('SET', bucket_quota_key, default_quota)
	redis.call('SET', bucket_replenishment_interval_key, default_replenishment_interval)
end

-- Calculate the time difference in seconds since last replenishment
local last_replenishment_timestamp = tonumber(redis.call('GET', last_replenishment_timestamp_key))
local time_difference = current_time - last_replenishment_timestamp
-- Shouldn't happen, but check just in case.
if time_difference < 0 then
	return {-2, 0} -- Return -2 (negative time difference)
end

-- Get the rate (tokens/second) that the bucket should replenish
local bucket_quota = tonumber(redis.call('GET', bucket_quota_key))
local bucket_replenishment_interval =  tonumber(redis.call('GET', bucket_replenishment_interval_key))
local replenishment_rate = bucket_quota / bucket_replenishment_interval

-- Calculate the amount of tokens to replenish, round down for the number of 'full' tokens.
local num_tokens_to_replenish = math.floor(replenishment_rate * time_difference)

-- Get the current token count in the bucket.
local current_tokens = tonumber(redis.call('GET', bucket_key))

-- Replenish the bucket if there are tokens to replenish
if num_tokens_to_replenish > 0 then
    local available_capacity = bucket_capacity - current_tokens
    if available_capacity > 0 then
		-- The number of tokens we add is either the number of tokens we have replenished over
		-- the last time_difference, or enough tokens to refill the bucket completely, whichever
		-- is lower.
		current_tokens = math.min(bucket_capacity, current_tokens + num_tokens_to_replenish)
    	redis.call('SET', bucket_key, math.min(bucket_capacity, current_tokens))
    	redis.call('SET', last_replenishment_timestamp_key, current_time)
    end
end

local time_to_wait_for_token = 0
-- This is for calculations with us removing a token.
local tokens_after_consumption = current_tokens - 1

-- If the bucket will be negative, calculate the needed to 'wait' before using the token.
-- i.e. if we are going to be at -15 tokens after this consumption, and the token replenishment
-- rate is 0.33/s, then we need to wait 45.45 (46 because we round up) seconds before making the request.
if tokens_after_consumption < 0 then
    time_to_wait_for_token = math.ceil((tokens_after_consumption * -1) / replenishment_rate)
end

if time_to_wait_for_token >= max_time_to_wait_for_token then
    return {0, time_to_wait_for_token} -- Return 0 (token grant wait time exceeds limit)
end

-- Decrement the token bucket by 1, we are granted a token
redis.call('DECRBY', bucket_key, 1)

return {1, time_to_wait_for_token}`

const setTokenBucketReplenishmentLuaScript = `local bucket_quota_key = KEYS[1]
local replenish_interval_seconds_key = KEYS[2]
local bucket_quota = tonumber(ARGV[1])
local bucket_replenish_interval = tonumber(ARGV[2])

redis.call('SET', bucket_quota_key, bucket_quota)
redis.call('SET', replenish_interval_seconds_key, bucket_replenish_interval)`

type getTokenGrantType int64

var (
	tokenGranted           getTokenGrantType = 1
	tokenGrantExceedsLimit getTokenGrantType = 0
	negativeTimeDifference getTokenGrantType = -2
)

type rateLimitBucketConfigKeys struct {
	BucketKey                       string
	QuotaKey                        string
	ReplenishmentIntervalSecondsKey string
	LastReplenishmentTimestampKey   string
}

type TokenGrantExceedsLimitError struct {
	tokenBucketKey string
	timeToWait     time.Duration
}

func (e TokenGrantExceedsLimitError) Error() string {
	return fmt.Sprintf("bucket:%s, acquiring token would require a wait of %s which exceeds the limit", e.tokenBucketKey, e.timeToWait.String())
}

type NegativeTimeDifferenceError struct {
	tokenBucketKey string
}

func (e NegativeTimeDifferenceError) Error() string {
	return fmt.Sprintf("bucket:%s, time difference between now and the last replenishment is negative", e.tokenBucketKey)
}

type UnexpectedRateLimitReturnError struct {
	tokenBucketKey string
}

func (e UnexpectedRateLimitReturnError) Error() string {
	return fmt.Sprintf("bucket:%s, unexpected return code from rate limit script", e.tokenBucketKey)
}

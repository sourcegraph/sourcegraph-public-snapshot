package ratelimit

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"time"

	"github.com/gomodule/redigo/redis"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const GitRPSLimiterBucketName = "git-rps"

var (
	tokenBucketGlobalPrefix                   = "v2:rate_limiters"
	bucketLastReplenishmentTimestampKeySuffix = "last_replenishment_timestamp"
	bucketAllowedBurstKeySuffix               = "allowed_burst"
	bucketRateConfigKeySuffix                 = "config:bucket_rate"
	bucketReplenishmentConfigKeySuffix        = "config:bucket_replenishment_interval_seconds"
	// bucketMaxCapacity                         = 10
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

func (r *rateLimiter) WaitN(ctx context.Context, n int) (err error) {
	logger := log.Scoped("globalRateLimiter", "")
	// TODO: Debugging, remove.
	defer func() {
		if err != nil {
			logger.Warn("WaitN failed", log.Error(err))
		}
	}()

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

	// TODO: Debugging, remove.
	logger.Info("Enforcing wait before token can be consumed", log.Int("timeToWait", int(timeToWait.Seconds())))

	// Wait for the required time before the token can be used.
	ch, stop := defaultRateLimitTimer(timeToWait)
	defer stop()
	select {
	case <-ch:
		// We can proceed.
		return nil
	case <-ctx.Done():
		// Note: rate.Limiter would return the tokens to the bucket
		// here, we don't do that for simplicity.
		return ctx.Err()
	}
}

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
		keys.BucketKey, keys.LastReplenishmentTimestampKey, keys.RateKey, keys.ReplenishmentIntervalSecondsKey, keys.BurstKey,
		requestTime.Unix(),
		maxTimeToWait.Seconds(),
		int32(fallbackRateLimit),
		1,
		defaultBurst,
		n,
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
	case waitTimeExceedsDeadline:
		return WaitTimeExceedsDeadlineError{
			timeToWait:     timeToWait,
			tokenBucketKey: bucketKey,
		}
	case negativeTimeDifference:
		return NegativeTimeDifferenceError{tokenBucketKey: bucketKey}
	case allBlocked:
		return AllBlockedError{tokenBucketKey: bucketKey}
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

	_, err := setReplenishmentScript.DoContext(ctx, connection, keys.RateKey, keys.ReplenishmentIntervalSecondsKey, bucketQuota, bucketReplenishInterval.Seconds())
	return errors.Wrapf(err, "error while setting token bucket replenishment for bucket %s", r.bucketName)
}

func (r *rateLimiter) getRateLimiterKeys() rateLimitBucketConfigKeys {
	var keys rateLimitBucketConfigKeys
	// e.g. v2:rate_limiters:github.com:api_tokens
	keys.BucketKey = fmt.Sprintf("%s:%s", r.prefix, r.bucketName)
	// e.g. v2:rate_limiters:github.com:api_tokens:config:bucket_rate
	keys.RateKey = fmt.Sprintf("%s:%s", keys.BucketKey, bucketRateConfigKeySuffix)
	// e.g.. v2:rate_limiters:github.com:api_tokens:config:bucket_replenishment_interval_seconds
	keys.ReplenishmentIntervalSecondsKey = fmt.Sprintf("%s:%s", keys.BucketKey, bucketReplenishmentConfigKeySuffix)
	// e.g.. v2:rate_limiters:github.com:api_tokens:last_replenishment_timestamp
	keys.LastReplenishmentTimestampKey = fmt.Sprintf("%s:%s", keys.BucketKey, bucketLastReplenishmentTimestampKeySuffix)
	// e.g.. v2:rate_limiters:github.com:api_tokens:allowed_burst
	keys.BurstKey = fmt.Sprintf("%s:%s", keys.BucketKey, bucketAllowedBurstKeySuffix)

	return keys
}

var (
	getTokensScript        = redis.NewScript(5, getTokensFromBucketLuaScript)
	setReplenishmentScript = redis.NewScript(2, setTokenBucketReplenishmentLuaScript)
)

// getTokensFromBucketLuaScript gets a single token from the specified bucket.
// bucket_key: the key in Redis that stores the bucket, under which all the bucket's tokens and rate limit configs are found, e.g. v2:rate_limiters:github.com:api_tokens.
// last_replenishment_timestamp_key: the key in Redis that stores the timestamp (seconds since epoch) of the last bucket replenishment, e.g. v2:rate_limiters:github.com:api_tokens:last_replenishment_timestamp.
// bucket_quota_key: the key in Redis that stores how many tokens the bucket should refill in a `bucket_replenishment_interval` period of time, e.g. v2:rate_limiters:github.com:api_tokens:config:bucket_quota.
// bucket_replenishment_interval_key: the key in Redis that stores how often (in seconds), the bucket should be replenished bucket_quota tokens, e.g. v2:rate_limiters:github.com:api_tokens:config:bucket_replenishment_interval_seconds.
// burst: the amount of tokens the bucket can hold, always bucketMaxCapacity right now.
// current_time: current time (seconds since epoch).
// max_time_to_wait_for_token: the maximum amount of time (in seconds) the requester is willing to wait before acquiring/using a token.
//
//go:embed globallimitergettokens.lua
var getTokensFromBucketLuaScript string

const setTokenBucketReplenishmentLuaScript = `local bucket_quota_key = KEYS[1]
local replenish_interval_seconds_key = KEYS[2]
local bucket_quota = tonumber(ARGV[1])
local bucket_replenish_interval = tonumber(ARGV[2])

redis.call('SET', bucket_quota_key, bucket_quota)
redis.call('SET', replenish_interval_seconds_key, bucket_replenish_interval)`

type getTokenGrantType int64

var (
	tokenGranted            getTokenGrantType = 1
	waitTimeExceedsDeadline getTokenGrantType = -1
	negativeTimeDifference  getTokenGrantType = -2
	allBlocked              getTokenGrantType = -3
)

type rateLimitBucketConfigKeys struct {
	BucketKey                       string
	RateKey                         string
	ReplenishmentIntervalSecondsKey string
	LastReplenishmentTimestampKey   string
	BurstKey                        string
}

type WaitTimeExceedsDeadlineError struct {
	tokenBucketKey string
	timeToWait     time.Duration
}

func (e WaitTimeExceedsDeadlineError) Error() string {
	return fmt.Sprintf("bucket:%s, acquiring token would require a wait of %s which exceeds the context deadline", e.tokenBucketKey, e.timeToWait.String())
}

type NegativeTimeDifferenceError struct {
	tokenBucketKey string
}

func (e NegativeTimeDifferenceError) Error() string {
	return fmt.Sprintf("bucket:%s, time difference between now and the last replenishment is negative", e.tokenBucketKey)
}

type AllBlockedError struct {
	tokenBucketKey string
}

func (e AllBlockedError) Error() string {
	return fmt.Sprintf("bucket:%s, rate is 0, no requests permitted", e.tokenBucketKey)
}

type UnexpectedRateLimitReturnError struct {
	tokenBucketKey string
}

func (e UnexpectedRateLimitReturnError) Error() string {
	return fmt.Sprintf("bucket:%s, unexpected return code from rate limit script", e.tokenBucketKey)
}

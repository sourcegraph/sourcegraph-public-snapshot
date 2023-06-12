package limiter

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Limiter interface {
	// TryAcquire checks if the rate limit has been exceeded and returns no error
	// if the request can proceed. The commit callback should be called after
	// a successful request to upstream to update the rate limit counter and
	// actually consume a request. This allows us to easily avoid deducting from
	// the rate limit if the request to upstream fails, at the cost of slight
	// over-allowance.
	//
	// The commit callback accepts a parameter that dictates how much rate
	// limit to consume for this request.
	TryAcquire(ctx context.Context) (commit func(int) error, err error)
}

type StaticLimiter struct {
	// Identifier is the key used to identify the rate limit counter.
	Identifier string

	Redis    RedisStore
	Limit    int
	Interval time.Duration

	// UpdateRateLimitTTL, if true, indicates that the TTL of the rate limit count should
	// be updated if there is a significant deviance from the desired interval.
	UpdateRateLimitTTL bool

	NowFunc          func() time.Time
	RateLimitAlerter func(usagePercentage float32, ttl time.Duration)
}

// RetryAfterWithTTL consults the current TTL using the given identifier and
// returns the time should be retried.
func RetryAfterWithTTL(redis RedisStore, nowFunc func() time.Time, identifier string) (time.Time, error) {
	ttl, err := redis.TTL(identifier)
	if err != nil {
		return time.Time{}, err
	}
	return nowFunc().Add(time.Duration(ttl) * time.Second), nil
}

func (l StaticLimiter) TryAcquire(ctx context.Context) (commit func(int) error, _ error) {
	// Zero values implies no access - this is a fallback check, callers should
	// be checking independently if access is granted.
	if l.Identifier == "" || l.Limit <= 0 || l.Interval <= 0 {
		return nil, NoAccessError{}
	}

	// Check the current usage. If no record exists, redis will return 0.
	currentUsage, err := l.Redis.GetInt(l.Identifier)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read rate limit counter")
	}

	// If the usage exceeds the maximum, we return an error. Consumers can check if
	// the error is of type RateLimitExceededError and extract additional information
	// like the limit and the time by when they should retry.
	if currentUsage >= l.Limit {
		retryAfter, err := RetryAfterWithTTL(l.Redis, l.NowFunc, l.Identifier)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get TTL for rate limit counter")
		}

		if l.RateLimitAlerter != nil {
			go l.RateLimitAlerter(1, retryAfter.Sub(l.NowFunc()))
		}

		return nil, RateLimitExceededError{
			Limit: l.Limit,
			// Return the minimum value of currentUsage and limit to not return
			// confusing values when the limit was exceeded. This method increases
			// on every check, even if the limit was reached.
			Used:       min(currentUsage, l.Limit),
			RetryAfter: retryAfter,
		}
	}

	// Now that we know that we want to let the user pass, let's return our callback to
	// increment the rate limit counter for the user if the request succeeds.
	// Note that the rate limiter _may_ allow slightly more requests than the configured
	// limit, incrementing the rate limit counter and reading the usage futher up are currently
	// not an atomic operation, because there is no good way to read the TTL in a transaction
	// without a lua script.
	// This approach could also slightly overcount the usage if redis requests after
	// the INCR fail, but it will always recover safely.
	// If Incr works but then everything else fails (eg ctx cancelled) the user spent
	// a token without getting anything for it. This seems pretty rare and a fine trade-off
	// since its just one token. The most likely reason this would happen is user cancelling
	// the request and at that point its more likely to happen while the LLM is running than
	// in this quick redis block.
	// On the first request in the current time block, if the requests past Incr fail we don't
	// yet have a deadline set. This means if the user comes back later we wouldn't of expired
	// just one token. This seems fine. Note: this isn't an issue on subsequent requests in the
	// same time block since the TTL would have been set.
	return func(usage int) error {
		if _, err := l.Redis.Incrby(l.Identifier, usage); err != nil {
			return errors.Wrap(err, "failed to increment rate limit counter")
		}

		// Set expiry on the key. If the key didn't exist prior to the previous INCR,
		// it will set the expiry of the key to one day.
		// If it did exist before, it should have an expiry set already, so the TTL >= 0
		// makes sure that we don't overwrite it and restart the 1h bucket.
		ttl, err := l.Redis.TTL(l.Identifier)
		if err != nil {
			return errors.Wrap(err, "failed to get TTL for rate limit counter")
		}
		intervalSeconds := int(l.Interval / time.Second)
		var alertTTL time.Duration
		if ttl < 0 || (l.UpdateRateLimitTTL && ttl > intervalSeconds) {
			if err := l.Redis.Expire(l.Identifier, intervalSeconds); err != nil {
				return errors.Wrap(err, "failed to set expiry for rate limit counter")
			}
			alertTTL = time.Duration(intervalSeconds) * time.Second
		} else {
			alertTTL = time.Duration(ttl) * time.Second
		}

		if l.RateLimitAlerter != nil {
			go l.RateLimitAlerter(float32(currentUsage+usage)/float32(l.Limit), alertTTL)
		}

		return nil
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

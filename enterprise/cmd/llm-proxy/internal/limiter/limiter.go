package limiter

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Limiter interface {
	TryAcquire(ctx context.Context) error
}

type StaticLimiter struct {
	// Identifier is the key used to identify the rate limit counter.
	Identifier string

	Redis    RedisStore
	Limit    int
	Interval time.Duration

	// UpdateRateLimitTTL, if true, indicates that the TTL of the rate limit count should
	// updated if there is a significant deviance from the desired interval.
	UpdateRateLimitTTL bool

	// Optional stub for current time. Used for testing.
	nowFunc func() time.Time
}

func (l StaticLimiter) TryAcquire(ctx context.Context) error {
	// Zero values implies no access - this is a fallback check, callers should
	// be checking independently if access is granted.
	if l.Identifier == "" || l.Limit == 0 || l.Interval == 0 {
		return NoAccessError{}
	}

	// Check the current usage and increment the counter for the current user. If
	// no record exists, redis will initialize it with 1.
	currentUsage, err := l.Redis.Incr(l.Identifier)
	if err != nil {
		return errors.Wrap(err, "failed to increase rate limit counter")
	}

	// Set expiry on the key. If existing TTL has not yet been set, or TTL is
	// longer than the desired interval (indicates that the rate limit interval
	// has changed and been shortened)
	ttl, err := l.Redis.TTL(l.Identifier)
	if err != nil {
		return errors.Wrap(err, "failed to get TTL for rate limit counter")
	}
	intervalSeconds := int(l.Interval / time.Second)
	if ttl < 0 || (l.UpdateRateLimitTTL && ttl > intervalSeconds) {
		if err := l.Redis.Expire(l.Identifier, intervalSeconds); err != nil {
			return errors.Wrap(err, "failed to set expiry for rate limit counter")
		}
	}

	// If the usage exceeds the maximum, we return an error. Consumers can check if
	// the error is of type RateLimitExceededError and extract additional information
	// like the limit and the time by when they should retry.
	if currentUsage > l.Limit {
		var now time.Time
		if l.nowFunc != nil {
			now = l.nowFunc()
		} else {
			now = time.Now()
		}
		return RateLimitExceededError{
			Limit: l.Limit,
			// Return the minimum value of currentUsage and limit to not return
			// confusing values when the limit was exceeded. This method increases
			// on every check, even if the limit was reached.
			Used:       min(currentUsage, l.Limit),
			RetryAfter: now.Add(time.Duration(ttl) * time.Second),
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

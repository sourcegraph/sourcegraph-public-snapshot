package limiter

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Limiter interface {
	TryAcquire(ctx context.Context, identifier string) error
}

type StaticLimiter struct {
	Redis    RedisStore
	Limit    int
	Interval time.Duration

	// Optional stub for current time. Used for testing.
	nowFunc func() time.Time
}

func (l StaticLimiter) TryAcquire(ctx context.Context, identifier string) error {
	// Zero values implies no access - this is a fallback check, callers should
	// be checking independently if access is granted.
	if l.Limit == 0 || l.Interval == 0 {
		return NoAccessError{}
	}

	// Check the current usage and increment the counter for the current user. If
	// no record exists, redis will initialize it with 1.
	currentUsage, err := l.Redis.Incr(identifier)
	if err != nil {
		return errors.Wrap(err, "failed to increase rate limit counter")
	}

	// Set expiry on the key. If existing TTL has not yet been set, or TTL is
	// longer than the desired interval (indicates that the rate limit interval
	// has changed and been shortened)
	ttl, err := l.Redis.TTL(identifier)
	if err != nil {
		return errors.Wrap(err, "failed to get TTL for rate limit counter")
	}
	if ttl < 0 || ttl > int(l.Interval/time.Second) {
		if err := l.Redis.Expire(identifier, int(l.Interval/time.Second)); err != nil {
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

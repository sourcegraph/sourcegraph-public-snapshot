package limiter

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var tracer = otel.Tracer("internal/limiter")

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
	TryAcquire(ctx context.Context) (commit func(context.Context, int) error, err error)
	// Usage returns the current usage in this limiter and the expiry time.
	Usage(ctx context.Context) (int, time.Time, error)
}

type StaticLimiter struct {
	// LimiterName optionally identifies the limiter for instrumentation. If not
	// provided, 'StaticLimiter' is used.
	LimiterName string

	// Identifier is the key used to identify the rate limit counter.
	Identifier string

	Redis    RedisStore
	Limit    int64
	Interval time.Duration

	// UpdateRateLimitTTL, if true, indicates that the TTL of the rate limit count should
	// be updated if there is a significant deviance from the desired interval.
	UpdateRateLimitTTL bool

	NowFunc func() time.Time

	// RateLimitAlerter is always called with usageRatio whenever rate limits are acquired.
	RateLimitAlerter func(ctx context.Context, usageRatio float32, ttl time.Duration)
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

func (l StaticLimiter) TryAcquire(ctx context.Context) (_ func(context.Context, int) error, err error) {
	if l.LimiterName == "" {
		l.LimiterName = "StaticLimiter"
	}
	intervalSeconds := l.Interval.Seconds()
	var currentUsage int
	var span trace.Span
	ctx, span = tracer.Start(ctx, l.LimiterName+".TryAcquire",
		trace.WithAttributes(
			attribute.Int64("limit", l.Limit),
			attribute.Float64("intervalSeconds", intervalSeconds)))
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.SetAttributes(attribute.Int("currentUsage", currentUsage))
		span.End()
	}()

	// Zero values implies no access - this is a fallback check, callers should
	// be checking independently if access is granted.
	if l.Identifier == "" || l.Limit <= 0 || l.Interval <= 0 {
		return nil, NoAccessError{}
	}

	// To work better with the abuse detection system, we consider the rate limit of 1 as no access.
	if l.Limit == 1 {
		return nil, NoAccessError{}
	}

	// Check the current usage. If no record exists, redis will return 0.
	currentUsage, err = l.Redis.GetInt(l.Identifier)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read rate limit counter")
	}

	// If the usage exceeds the maximum, we return an error. Consumers can check if
	// the error is of type RateLimitExceededError and extract additional information
	// like the limit and the time by when they should retry.
	if int64(currentUsage) >= l.Limit {
		retryAfter, err := RetryAfterWithTTL(l.Redis, l.NowFunc, l.Identifier)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get TTL for rate limit counter")
		}

		if l.RateLimitAlerter != nil {
			// Call with usage 1 for 100% (rate limit exceeded)
			go l.RateLimitAlerter(context.WithoutCancel(ctx), 1, retryAfter.Sub(l.NowFunc()))
		}

		return nil, RateLimitExceededError{
			Limit:      l.Limit,
			RetryAfter: retryAfter,
		}
	}

	// Now that we know that we want to let the user pass, let's return our callback to
	// increment the rate limit counter for the user if the request succeeds.
	// Note that the rate limiter _may_ allow slightly more requests than the configured
	// limit, incrementing the rate limit counter and reading the usage further up are currently
	// not an atomic operation, because there is no good way to read the TTL in a transaction
	// without a lua script.
	// This approach could also slightly over-count the usage if redis requests after
	// the INCR fail, but it will always recover safely.
	// If Incr works but then everything else fails (for example, ctx cancelled) the user spent
	// a token without getting anything for it. This seems pretty rare and a fine trade-off
	// since its just one token. The most likely reason this would happen is user cancelling
	// the request and at that point its more likely to happen while the LLM is running than
	// in this quick redis block.
	// On the first request in the current time block, if the requests past Incr fail we don't
	// yet have a deadline set. This means if the user comes back later we wouldn't of expired
	// just one token. This seems fine. Note: this isn't an issue on subsequent requests in the
	// same time block since the TTL would have been set.
	return func(ctx context.Context, usage int) (err error) {
		// NOTE: This is to make sure we still commit usage even if the context was canceled.
		ctx = context.WithoutCancel(ctx)

		var incrementedTo, ttlSeconds int
		// We need to start a new span because the previous one has ended
		// TODO: ctx is unused after this, but if a usage is added, we need
		// to update this assignment - removed for now because of ineffassign
		_, span = tracer.Start(ctx, l.LimiterName+".commit",
			trace.WithAttributes(attribute.Int("usage", usage)))
		defer func() {
			span.SetAttributes(
				attribute.Int("incrementedTo", incrementedTo),
				attribute.Int("ttlSeconds", ttlSeconds))
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to commit rate limit usage")
			}
			span.End()
		}()

		if incrementedTo, err = l.Redis.Incrby(l.Identifier, usage); err != nil {
			return errors.Wrap(err, "failed to increment rate limit counter")
		}

		// Set expiry on the key. If the key didn't exist prior to the previous INCR,
		// it will set the expiry of the key to `intervalSeconds`.
		//
		// If it did exist before, it should have an expiry set already, so the TTL >= 0
		// makes sure that we don't overwrite it and restart the 1h bucket.
		ttl, err := l.Redis.TTL(l.Identifier)
		if err != nil {
			return errors.Wrap(err, "failed to get TTL for rate limit counter")
		}
		var alertTTL time.Duration
		if ttl < 0 || (l.UpdateRateLimitTTL && ttl > int(intervalSeconds)) {
			if err := l.Redis.Expire(l.Identifier, int(intervalSeconds)); err != nil {
				return errors.Wrap(err, "failed to set expiry for rate limit counter")
			}
			alertTTL = time.Duration(intervalSeconds) * time.Second
			ttlSeconds = int(intervalSeconds)
		} else {
			alertTTL = time.Duration(ttl) * time.Second
			ttlSeconds = ttl
		}

		if l.RateLimitAlerter != nil {
			go l.RateLimitAlerter(ctx, float32(currentUsage+usage)/float32(l.Limit), alertTTL)
		}

		return nil
	}, nil
}

func (l StaticLimiter) Usage(ctx context.Context) (_ int, _ time.Time, err error) {
	if l.LimiterName == "" {
		l.LimiterName = "StaticLimiter"
	}

	// TODO: ctx is unused after this, but if a usage is added, we need
	// to update this assignment - removed for now because of ineffassign
	_, span := tracer.Start(ctx, l.LimiterName+".Usage",
		trace.WithAttributes(
			attribute.Int64("limit", l.Limit),
		))
	defer func() {
		span.RecordError(err)
		span.End()
	}()

	// Zero values implies no access.
	if l.Identifier == "" || l.Limit <= 0 || l.Interval <= 0 {
		return 0, time.Time{}, NoAccessError{}
	}

	// Check the current usage. If no record exists, redis will return 0.
	currentUsage, err := l.Redis.GetInt(l.Identifier)
	if err != nil {
		return 0, time.Time{}, errors.Wrap(err, "failed to read rate limit counter")
	}
	if currentUsage == 0 {
		return 0, time.Time{}, nil
	}

	// Get the current expiry.
	ttl, err := l.Redis.TTL(l.Identifier)
	if err != nil {
		return 0, time.Time{}, errors.Wrap(err, "failed to get TTL for rate limit counter")
	}

	return currentUsage, time.Now().Add(time.Duration(ttl) * time.Second).Truncate(time.Second), nil
}

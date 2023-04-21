package streaming

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RateLimiter interface {
	TryAcquire(ctx context.Context) error
	ReadQuota(ctx context.Context, userID int32) (count, limit int, _ error)
}

type RateLimitExceededError struct {
	Limit      int
	Used       int
	RetryAfter time.Time
}

func (e RateLimitExceededError) Error() string {
	return fmt.Sprintf("you exceeded the rate limit for completions, only %d requests are allowed per hour at the moment to ensure the service stays functional. Current usage: %d. Retry after %s", e.Limit, e.Used, e.RetryAfter.Truncate(time.Second))
}

func NewRateLimiter(db database.DB, rstore redispool.KeyValue) RateLimiter {
	return &rateLimiter{db: db, rstore: rstore}
}

type rateLimiter struct {
	rstore redispool.KeyValue
	db     database.DB
}

func (r *rateLimiter) TryAcquire(ctx context.Context) (err error) {
	limit, err := getConfiguredLimit(ctx, r.db)
	if err != nil {
		return errors.Wrap(err, "failed to read configured rate limit")
	}

	if limit == 0 {
		// Rate limiting disabled.
		return nil
	}

	// Check that the user is authenticated.
	a := actor.FromContext(ctx)
	if a.IsInternal() {
		return nil
	}
	key := userKey(a.UID)
	if !a.IsAuthenticated() {
		// Fall back to the IP address, if provided in context (ie. this is a request handler).
		req := requestclient.FromContext(ctx)
		var ip string
		if req != nil {
			ip = req.IP
			// Note: ForwardedFor header in general can be spoofed. For
			// Sourcegraph.com we use a trusted value for this so this is a
			// reliable value to rate limit with.
			if req.ForwardedFor != "" {
				ip = req.ForwardedFor
			}
		}
		if ip == "" {
			return errors.Wrap(auth.ErrNotAuthenticated, "cannot claim rate limit for unauthenticated user without request context")
		}
		key = anonymousKey(ip)
	}

	// Check the current usage.
	current, retryAfter, err := r.get(ctx, key)
	if err != nil {
		return errors.Wrap(err, "failed to read current rate limit usage")
	}
	// If the usage exceeds the maximum, we return an error. Consumers can check if
	// the error is of type RateLimitExceededError and extract additional information
	// like the limit and the time by when they should retry.
	if current >= limit {
		return RateLimitExceededError{
			Limit:      limit,
			Used:       current,
			RetryAfter: retryAfter,
		}
	}

	// Open a new connection to redis so we can run a MULTI command.
	pool, ok := r.rstore.Pool()
	if !ok {
		return errors.New("redis pool is not available but rate limit has been configured")
	}
	conn, err := pool.Dial()
	if err != nil {
		return errors.Wrap(err, "failed to dial redis")
	}
	defer func() {
		err = errors.Append(err, conn.Close())
	}()

	// Now that we validated that we want to let the user pass, let's increment
	// the rate limit counter for the user.
	// Note that the rate limiter _may_ allow slightly more requests than the configured
	// limit, reading the current value and incrementing the rate limit counter are
	// currently not an atomic operation.
	if _, err := conn.Do("MULTI"); err != nil {
		return errors.Wrap(err, "failed to start redis transaction")
	}
	// Increment the counter for the current user. If no record exists, redis will
	// initialize it with 1.
	if _, err := conn.Do("INCR", key); err != nil {
		return errors.Wrap(err, "failed to increase rate limit counter")
	}

	// Set expiry on the key. If the key didn't exist prior to the previous INCR,
	// it will set the expiry of the key to one hour.
	// If it did exist before, it should have an expiry set already, so the NX makes
	// sure that we don't overwrite it.
	if _, err := conn.Do("EXPIRE", key, int(time.Hour/time.Second), "NX"); err != nil {
		return errors.Wrap(err, "failed to set expiry for rate limit counter")
	}

	// Submit the calls.
	if _, err := conn.Do("EXEC"); err != nil {
		return errors.Wrap(err, "failed to flush multi operation")
	}

	return nil
}

func (r *rateLimiter) ReadQuota(ctx context.Context, userID int32) (count, limit int, _ error) {
	limit, _, err := r.get(ctx, userKey(userID))
	if err != nil {
		return 0, 0, err
	}
	configuredLimit, err := getConfiguredLimit(actor.WithActor(context.Background(), actor.FromUser(userID)), r.db)
	if err != nil {
		return 0, 0, err
	}
	return limit, configuredLimit, nil
}

func (r *rateLimiter) get(ctx context.Context, key string) (int, time.Time, error) {
	rstore := r.rstore.WithContext(ctx)

	rv := rstore.Get(key)
	if rv.IsNil() {
		return 0, time.Time{}, nil
	}
	count, err := rv.Int()
	if err != nil {
		return 0, time.Time{}, errors.Wrap(err, "failed to get request counter")
	}
	ttl, err := rstore.TTL(key)
	if err != nil {
		return 0, time.Time{}, errors.Wrap(err, "failed to get ttl of key")
	}
	return count, time.Now().Add(time.Duration(ttl) * time.Second), nil
}

func userKey(userID int32) string {
	return fmt.Sprintf("user:%d:completion_requests", userID)
}

func anonymousKey(ip string) string {
	return fmt.Sprintf("anon:%s:completion_requests", ip)
}

func getConfiguredLimit(ctx context.Context, db database.DB) (int, error) {
	a := actor.FromContext(ctx)
	if a.IsAuthenticated() && !a.IsInternal() {
		// If an authenticated user exists, check if an override exists.
		limit, err := db.Users().GetCompletionsQuota(ctx, a.UID)
		if err != nil {
			return 0, err
		}
		if limit != nil {
			return *limit, err
		}
	}

	// Otherwise, fall back to the global limit.
	cfg := conf.Get()
	if cfg.Completions != nil && cfg.Completions.PerUserHourlyLimit > 0 {
		return cfg.Completions.PerUserHourlyLimit, nil
	}

	return 0, nil
}

package actor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RateLimit struct {
	// AllowedModels is a set of models in Cody Gateway's model configuration
	// format, "$PROVIDER/$MODEL_NAME".
	AllowedModels []string `json:"allowedModels"`

	Limit    int64         `json:"limit"`
	Interval time.Duration `json:"interval"`

	// ConcurrentRequests, ConcurrentRequestsInterval are generally applied
	// with NewRateLimitWithPercentageConcurrency.
	ConcurrentRequests         int           `json:"concurrentRequests"`
	ConcurrentRequestsInterval time.Duration `json:"concurrentRequestsInterval"`
}

func NewRateLimitWithPercentageConcurrency(limit int64, interval time.Duration, allowedModels []string, concurrencyConfig codygateway.ActorConcurrencyLimitConfig) RateLimit {
	// The actual type of time.Duration is int64, so we can use it to compute the
	// ratio of the rate limit interval to a day (24 hours).
	ratioToDay := float32(interval) / float32(24*time.Hour)
	// Then use the ratio to compute the rate limit for a day.
	dailyLimit := float32(limit) / ratioToDay
	// Finally, compute the concurrency limit with the given percentage of the daily limit.
	concurrencyLimit := int(dailyLimit * concurrencyConfig.Percentage)
	// Just in case a poor choice of percentage results in a concurrency limit less than 1.
	if concurrencyLimit < 1 {
		concurrencyLimit = 1
	}

	return RateLimit{
		AllowedModels: allowedModels,
		Limit:         limit,
		Interval:      interval,

		ConcurrentRequests:         concurrencyLimit,
		ConcurrentRequestsInterval: concurrencyConfig.Interval,
	}
}

func (r *RateLimit) IsValid() bool {
	return r != nil && r.Interval > 0 && r.Limit > 0 && len(r.AllowedModels) > 0
}

type concurrencyLimiter struct {
	logger  log.Logger
	actor   *Actor
	feature codygateway.Feature

	// redis must be a prefixed store
	redis limiter.RedisStore

	concurrentRequests int
	concurrentInterval time.Duration

	nextLimiter limiter.Limiter

	nowFunc func() time.Time
}

func (l *concurrencyLimiter) TryAcquire(ctx context.Context) (func(context.Context, int) error, error) {
	commit, err := (limiter.StaticLimiter{
		LimiterName:        "actor.concurrencyLimiter",
		Identifier:         l.actor.ID,
		Redis:              l.redis,
		Limit:              int64(l.concurrentRequests),
		Interval:           l.concurrentInterval,
		UpdateRateLimitTTL: true, // always adjust
		NowFunc:            l.nowFunc,
	}).TryAcquire(ctx)
	if err != nil {
		if errors.As(err, &limiter.NoAccessError{}) || errors.As(err, &limiter.RateLimitExceededError{}) {
			retryAfter, err := limiter.RetryAfterWithTTL(l.redis, l.nowFunc, l.actor.ID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get TTL for rate limit counter")
			}
			return nil, ErrConcurrencyLimitExceeded{
				feature:    l.feature,
				limit:      l.concurrentRequests,
				retryAfter: retryAfter,
			}
		}
		return nil, errors.Wrap(err, "check concurrent limit")
	}
	if err = commit(ctx, 1); err != nil {
		trace.Logger(ctx, l.logger).Error("failed to commit concurrency limit consumption", log.Error(err))
	}

	return l.nextLimiter.TryAcquire(ctx)
}

func (l *concurrencyLimiter) Usage(ctx context.Context) (int, time.Time, error) {
	return l.nextLimiter.Usage(ctx)
}

type ErrConcurrencyLimitExceeded struct {
	feature    codygateway.Feature
	limit      int
	retryAfter time.Time
}

// Error generates a simple string that is fairly static for use in logging.
// This helps with categorizing errors. For more detailed output use Summary().
func (e ErrConcurrencyLimitExceeded) Error() string {
	return fmt.Sprintf("%q: concurrency limit exceeded", e.feature)
}

func (e ErrConcurrencyLimitExceeded) Summary() string {
	return fmt.Sprintf("you have exceeded the concurrency limit of %d requests for %q. Retry after %s",
		e.limit, e.feature, e.retryAfter.Truncate(time.Second))
}

func (e ErrConcurrencyLimitExceeded) WriteResponse(w http.ResponseWriter) {
	// Rate limit exceeded, write well known headers and return correct status code.
	w.Header().Set("x-ratelimit-limit", strconv.Itoa(e.limit))
	w.Header().Set("x-ratelimit-remaining", "0")
	w.Header().Set("retry-after", e.retryAfter.Format(time.RFC1123))
	// Use Summary instead of Error for more informative text
	http.Error(w, e.Summary(), http.StatusTooManyRequests)
}

// updateOnErrorLimiter calls Actor.Update if nextLimiter responds with certain
// access errors.
type updateOnErrorLimiter struct {
	logger log.Logger
	actor  *Actor

	nextLimiter limiter.Limiter
}

func (u updateOnErrorLimiter) TryAcquire(ctx context.Context) (func(context.Context, int) error, error) {
	commit, err := u.nextLimiter.TryAcquire(ctx)
	// If we have an access issue, try to update the actor in case they have
	// been granted updated access.
	if errors.As(err, &limiter.NoAccessError{}) || errors.As(err, &limiter.RateLimitExceededError{}) {
		oteltrace.SpanFromContext(ctx).
			SetAttributes(attribute.Bool("update-on-error", true))
		// Do update transiently, outside request hotpath
		go func() {
			if updateErr := u.actor.Update(context.WithoutCancel(ctx)); updateErr != nil &&
				!IsErrActorRecentlyUpdated(updateErr) {
				u.logger.Warn("unexpected error updating actor",
					log.Error(updateErr),
					log.NamedError("originalError", err))
			}
		}()
	}
	return commit, err
}

func (u updateOnErrorLimiter) Usage(ctx context.Context) (int, time.Time, error) {
	return u.nextLimiter.Usage(ctx)
}

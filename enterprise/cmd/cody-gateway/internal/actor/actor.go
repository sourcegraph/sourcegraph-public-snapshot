package actor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RateLimit struct {
	// AllowedModels is a set of models in Cody Gateway's model configuration
	// format, "$PROVIDER/$MODEL_NAME".
	AllowedModels []string `json:"allowedModels"`

	Limit    int           `json:"limit"`
	Interval time.Duration `json:"interval"`
}

func (r *RateLimit) IsValid() bool {
	return r != nil && r.Interval > 0 && r.Limit > 0 && len(r.AllowedModels) > 0
}

type Actor struct {
	// Key is the original key used to identify the actor. It may be a sensitive value
	// so use with care!
	//
	// For example, for product subscriptions this is the license-based access token.
	Key string `json:"key"`
	// ID is the identifier for this actor's rate-limiting pool. It is not a sensitive
	// value.
	//
	// For example, for product subscriptions this is the subscription UUID. For
	// Sourcegraph.com users, this is the string representation of the user ID.
	ID string `json:"id"`
	// AccessEnabled is an evaluated field that summarizes whether or not Cody Gateway access
	// is enabled.
	//
	// For example, for product subscriptions it is based on whether the subscription is
	// archived, if access is enabled, and if any rate limits are set.
	AccessEnabled bool `json:"accessEnabled"`
	// RateLimits holds the rate limits for Cody Gateway features for this actor.
	RateLimits map[codygateway.Feature]RateLimit `json:"rateLimits"`
	// LastUpdated indicates when this actor's state was last updated.
	LastUpdated *time.Time `json:"lastUpdated"`
	// Source is a reference to the source of this actor's state.
	Source Source `json:"-"`
}

type contextKey int

const actorKey contextKey = iota

// FromContext returns a new Actor instance from a given context. It always
// returns a non-nil actor.
func FromContext(ctx context.Context) *Actor {
	a, ok := ctx.Value(actorKey).(*Actor)
	if !ok || a == nil {
		return &Actor{}
	}
	return a
}

// Logger returns a logger that has metadata about the actor attached to it.
func (a *Actor) Logger(logger log.Logger) log.Logger {
	if a == nil {
		return logger
	}
	return logger.With(
		log.String("actor.ID", a.ID),
		log.String("actor.Source", a.Source.Name()),
		log.Bool("actor.AccessEnabled", a.AccessEnabled),
		log.Timep("actor.LastUpdated", a.LastUpdated),
	)
}

// Update updates the given actor's state using the actor's originating source
// if it implements SourceUpdater.
//
// The source may define additional conditions for updates, such that an update
// does not necessarily occur on every call.
//
// If the actor has no source, this is a no-op.
func (a *Actor) Update(ctx context.Context) {
	if su, ok := a.Source.(SourceUpdater); ok && su != nil {
		su.Update(ctx, a)
	}
}

// WithActor returns a new context with the given Actor instance.
func WithActor(ctx context.Context, a *Actor) context.Context {
	return context.WithValue(ctx, actorKey, a)
}

func (a *Actor) Limiter(
	logger log.Logger,
	redis limiter.RedisStore,
	feature codygateway.Feature,
	concurrencyLimitConfig codygateway.ActorConcurrencyLimitConfig,
) (limiter.Limiter, bool) {
	if a == nil {
		// Not logged in, no limit applicable.
		return nil, false
	}
	limit, ok := a.RateLimits[feature]
	if !ok {
		return nil, false
	}

	if !limit.IsValid() {
		// No valid limit, cannot provide limiter.
		return nil, false
	}

	var concurrencyLimit int
	if feature == codygateway.FeatureEmbeddings {
		// For embeddings, we use a custom, hardcoded limit. Concurrency on the client side
		// should be 1, so 15 is a very safe default here.
		concurrencyLimit = 15
	} else {
		// The actual type of time.Duration is int64, so we can use it to compute the
		// ratio of the rate limit interval to a day (24 hours).
		ratioToDay := float32(limit.Interval) / float32(24*time.Hour)
		// Then use the ratio to compute the rate limit for a day.
		dailyLimit := float32(limit.Limit) / ratioToDay
		// Finally, compute the concurrency limit with the given percentage of the daily limit.
		concurrencyLimit = int(dailyLimit * concurrencyLimitConfig.Percentage)
		// Just in case a poor choice of percentage results in a concurrency limit less than 1.
		if concurrencyLimit < 1 {
			concurrencyLimit = 1
		}
	}

	// The redis store has to use a prefix for the given feature because we need to
	// rate limit by feature.
	featurePrefix := fmt.Sprintf("%s:", feature)
	concurrencyLimiter := &concurrencyLimiter{
		logger:  logger,
		actor:   a,
		feature: feature,
		redis:   limiter.NewPrefixRedisStore(fmt.Sprintf("concurrent:%s", featurePrefix), redis),
		rateLimit: RateLimit{
			Limit:    concurrencyLimit,
			Interval: concurrencyLimitConfig.Interval,
		},
		featureLimiter: updateOnFailureLimiter{
			Redis:     limiter.NewPrefixRedisStore(featurePrefix, redis),
			RateLimit: limit,
			Actor:     a,
		},
		nowFunc: time.Now,
	}
	return concurrencyLimiter, true
}

type concurrencyLimiter struct {
	logger         log.Logger
	actor          *Actor
	feature        codygateway.Feature
	redis          limiter.RedisStore
	rateLimit      RateLimit
	featureLimiter limiter.Limiter
	nowFunc        func() time.Time
}

func (l *concurrencyLimiter) TryAcquire(ctx context.Context) (func(int) error, error) {
	commit, err := (limiter.StaticLimiter{
		Identifier: l.actor.ID,
		Redis:      l.redis,
		Limit:      l.rateLimit.Limit,
		Interval:   l.rateLimit.Interval,
		// Only update rate limit TTL if the actor has been updated recently.
		UpdateRateLimitTTL: l.actor.LastUpdated != nil && l.nowFunc().Sub(*l.actor.LastUpdated) < 5*time.Minute,
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
				limit:      l.rateLimit.Limit,
				retryAfter: retryAfter,
			}
		}
		return nil, errors.Wrap(err, "check concurrent limit")
	}
	if err = commit(1); err != nil {
		l.logger.Error("failed to commit concurrency limit consumption", log.Error(err))
	}

	featureCommit, err := l.featureLimiter.TryAcquire(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "check feature rate limit")
	}
	return featureCommit, nil
}

type ErrConcurrencyLimitExceeded struct {
	feature    codygateway.Feature
	limit      int
	retryAfter time.Time
}

func (e ErrConcurrencyLimitExceeded) Error() string {
	return fmt.Sprintf("you exceeded the concurrency limit of %d requests for %q. Retry after %s",
		e.limit, e.feature, e.retryAfter.Truncate(time.Second))
}

func (e ErrConcurrencyLimitExceeded) WriteResponse(w http.ResponseWriter) {
	// Rate limit exceeded, write well known headers and return correct status code.
	w.Header().Set("x-ratelimit-limit", strconv.Itoa(e.limit))
	w.Header().Set("x-ratelimit-remaining", "0")
	w.Header().Set("retry-after", e.retryAfter.Format(time.RFC1123))
	http.Error(w, e.Error(), http.StatusTooManyRequests)
}

type updateOnFailureLimiter struct {
	Redis     limiter.RedisStore
	RateLimit RateLimit
	*Actor
}

func (u updateOnFailureLimiter) TryAcquire(ctx context.Context) (func(int) error, error) {
	commit, err := (limiter.StaticLimiter{
		Identifier: u.ID,
		Redis:      u.Redis,
		Limit:      u.RateLimit.Limit,
		Interval:   u.RateLimit.Interval,
		// Only update rate limit TTL if the actor has been updated recently.
		UpdateRateLimitTTL: u.LastUpdated != nil && time.Since(*u.LastUpdated) < 5*time.Minute,
		NowFunc:            time.Now,
	}).TryAcquire(ctx)

	if errors.As(err, &limiter.NoAccessError{}) || errors.As(err, &limiter.RateLimitExceededError{}) {
		u.Actor.Update(ctx) // TODO: run this in goroutine+background context maybe?
	}

	return commit, err
}

// ErrAccessTokenDenied is returned when the access token is denied due to the
// reason.
type ErrAccessTokenDenied struct {
	Reason string
}

func (e ErrAccessTokenDenied) Error() string {
	return fmt.Sprintf("access token denied: %s", e.Reason)
}

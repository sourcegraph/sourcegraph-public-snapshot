package actor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/limiter"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/notify"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayactor"
)

type Actor struct {
	// Key is the original key used to identify the actor. It may be a sensitive value
	// so use with care!
	//
	// For example, for product subscriptions this is the license-based access token.
	Key string `json:"key"`
	// ID is the identifier for this actor's rate-limiting pool. It is not a sensitive
	// value. It must be set for all valid actors - if empty, the actor must be invalid
	// and must not have any feature access.
	//
	// For example, for product subscriptions this is the subscription UUID. For
	// Sourcegraph.com users, this is the string representation of the user ID.
	ID string `json:"id"`
	// Name is the human-readable name for this actor, e.g. username, account name.
	// Optional for implementations - if unset, ID will be returned from GetName().
	Name string `json:"name"`
	// AccessEnabled is an evaluated field that summarizes whether or not Cody Gateway access
	// is enabled.
	//
	// For example, for product subscriptions it is based on whether the subscription is
	// archived, if access is enabled, and if any rate limits are set.
	AccessEnabled bool `json:"accessEnabled"`
	// EndpointAccess is a map that grants access to a specific endpoint prefix
	// in case not all access is enabled.
	EndpointAccess map[string]bool `json:"endpointAccess"`
	// RateLimits holds the rate limits for Cody Gateway features for this actor.
	RateLimits map[codygateway.Feature]RateLimit `json:"rateLimits"`
	// LastUpdated indicates when this actor's state was last updated.
	LastUpdated *time.Time `json:"lastUpdated"`
	// Source is a reference to the source of this actor's state.
	Source Source `json:"-"`
}

func (a *Actor) GetID() string {
	return a.ID
}

func (a *Actor) GetName() string {
	if a.Name == "" {
		return a.ID
	}
	return a.Name
}

func (a *Actor) GetSource() codygatewayactor.ActorSource {
	if a == nil || a.Source == nil {
		return "unknown"
	}
	return codygatewayactor.ActorSource(a.Source.Name())
}

// IsEmpty returns true if the actor is empty, e.g. has no ID.
// An empty actor is saved in the cache on fetch req failure,
// so that we aren't constantly hitting the dotcom API. Check
// the implementation of `Source.fetchAndCache`.
func (a *Actor) IsEmpty() bool {
	return a == nil || a.ID == ""
}

func (a *Actor) IsDotComActor() bool {
	// Corresponds to sourcegraph.com subscription ID, or using a dotcom access token
	return a != nil && (a.GetSource() == codygatewayactor.ActorSourceEnterpriseSubscription && a.ID == "d3d2b638-d0a2-4539-a099-b36860b09819") || a.GetSource() == codygatewayactor.ActorSourceDotcomUser
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
	// If there's no ID and no source and no key, this is probably just no
	// actor available. Possible in actor-less endpoints like diagnostics.
	if a == nil || (a.ID == "" && a.Source == nil && a.Key == "") {
		return logger.With(log.String("actor.ID", "<nil>"))
	}

	// TODO: We shouldn't ever have a nil source, but check just in case, since
	// we don't want to panic on some instrumentation.
	var sourceName string
	if a.Source != nil {
		sourceName = a.Source.Name()
	} else {
		sourceName = "<nil>"
	}

	return logger.With(
		log.String("actor.ID", a.ID),
		log.String("actor.Source", sourceName),
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
func (a *Actor) Update(ctx context.Context) error {
	if su, ok := a.Source.(SourceUpdater); ok && su != nil {
		return su.Update(ctx, a)
	}
	return nil
}

func (a *Actor) TraceAttributes() []attribute.KeyValue {
	if a == nil {
		return []attribute.KeyValue{attribute.String("actor", "<nil>")}
	}

	attrs := []attribute.KeyValue{
		attribute.String("actor.id", a.ID),
		attribute.Bool("actor.accessEnabled", a.AccessEnabled),
	}
	if a.LastUpdated != nil {
		attrs = append(attrs, attribute.String("actor.lastUpdated", a.LastUpdated.String()))
	}
	for f, rl := range a.RateLimits {
		key := fmt.Sprintf("actor.rateLimits.%s", f)
		if rlJSON, err := json.Marshal(rl); err != nil {
			attrs = append(attrs, attribute.String(key, err.Error()))
		} else {
			attrs = append(attrs, attribute.String(key, string(rlJSON)))
		}
	}
	return attrs
}

// WithActor returns a new context with the given Actor instance.
func WithActor(ctx context.Context, a *Actor) context.Context {
	return context.WithValue(ctx, actorKey, a)
}

// baseLimiterAndLimit returns the base limiter for the given feature,
// and the rate limits for the actor.
func (a *Actor) baseLimiterAndLimit(
	redis limiter.RedisStore,
	feature codygateway.Feature,
	rateLimitNotifier notify.RateLimitNotifier,
) (*limiter.StaticLimiter, RateLimit, bool) {
	if a == nil {
		// Not logged in, no limit applicable.
		return nil, RateLimit{}, false
	}

	limit, ok := a.RateLimits[feature]
	if !ok {
		return nil, RateLimit{}, false
	}
	if !limit.IsValid() {
		// No valid limit, cannot provide limiter.
		return nil, RateLimit{}, false
	}

	// BaseLimiter is the core Limiter that naively applies the specified
	// rate limits. This will get wrapped in various other layers of limiter
	// behaviour.
	baseLimiter := limiter.StaticLimiter{
		LimiterName: "actor.Limiter",
		Identifier:  a.ID,
		Redis:       limiter.NewFeatureUsageStore(redis, feature),
		Limit:       limit.Limit,
		Interval:    limit.Interval,
		// Only update rate limit TTL if the actor has been updated recently.
		UpdateRateLimitTTL: a.LastUpdated != nil && time.Since(*a.LastUpdated) < 5*time.Minute,
		NowFunc:            time.Now,
		RateLimitAlerter: func(ctx context.Context, usageRatio float32, ttl time.Duration) {
			rateLimitNotifier(ctx, a, feature, usageRatio, ttl)
		},
	}

	return &baseLimiter, limit, true
}

// Limiter is the main limiter used to apply rate limits to requests. It
// applies the base rate limit, and wraps it in layers of concurrency
// and update-on-error rate limiters.
func (a *Actor) Limiter(
	logger log.Logger,
	redis limiter.RedisStore,
	feature codygateway.Feature,
	rateLimitNotifier notify.RateLimitNotifier,
) (limiter.Limiter, bool) {
	// Start with the base limiter. This limiter enforces rate limits provided
	// the actor's Source, and each Source implementation owns ensuring that
	// Cody Gateway has an up-to-date view of the appropriate rate limits for
	// a particular actor based on the application (e.g. users from enterprise product
	// subscriptions, Self-Serve-Cody's tiers, etc.
	baseLimiter, limit, ok := a.baseLimiterAndLimit(redis, feature, rateLimitNotifier)
	if !ok {
		return nil, false
	}

	// Wrap the base limiter with updateOnErrorLimiter, to call Actor.Update to
	// refresh the actor cache on any errors.
	updateOnErrLimiter := updateOnErrorLimiter{
		logger:      logger.Scoped("updateOnError"),
		actor:       a,
		nextLimiter: baseLimiter,
	}

	// Finally return a concurrency limiter, to ensure that a user cannot have too many
	// requests in-flight at a time. This is generally a percentage of the rate limit
	// assigned to an Actor by its Source - see RateLimit for more details.
	concurrentStorePrefix := fmt.Sprintf("concurrent:%s:", feature)
	return &concurrencyLimiter{
		logger:             logger.Scoped("concurrency"),
		actor:              a,
		feature:            feature,
		redis:              limiter.NewPrefixRedisStore(concurrentStorePrefix, redis),
		concurrentRequests: limit.ConcurrentRequests,
		concurrentInterval: limit.ConcurrentRequestsInterval,
		nextLimiter:        updateOnErrLimiter,
		nowFunc:            time.Now,
	}, true
}

// ErrAccessTokenDenied is returned when the access token is denied due to the
// reason.
type ErrAccessTokenDenied struct {
	Reason string
	Source string
}

func (e ErrAccessTokenDenied) Error() string {
	return fmt.Sprintf("access token denied: %s", e.Reason)
}

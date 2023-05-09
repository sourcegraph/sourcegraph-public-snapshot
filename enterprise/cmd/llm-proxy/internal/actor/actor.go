package actor

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/limiter"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RateLimit struct {
	Limit    int
	Interval time.Duration
}

func (r *RateLimit) IsValid() bool {
	return r != nil && r.Interval > 0 && r.Limit > 0
}

type Actor struct {
	// Key is the original key used to identify the actor.
	//
	// For example, for product subscriptions this is the license-based access token.
	Key string `json:"key"`
	// ID is the identifier for this actor's rate-limiting pool.
	//
	// For example, for product subscriptions this is the subscription ID.
	ID string `json:"id"`
	// AccessEnabled is an evaluated field that summarizes whether or not LLM-proxy access
	// is enabled.
	//
	// For example, for product subscriptions it is based on whether the subscription is
	// archived, if access is enabled, and if any rate limits are set.
	AccessEnabled bool `json:"accessEnabled"`
	// RateLimit is the rate limit for LLM-proxy access for this actor.
	RateLimit RateLimit `json:"rateLimit"`
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

func (a *Actor) Limiter(redis limiter.RedisStore) limiter.Limiter {
	if a == nil {
		return &limiter.StaticLimiter{}
	}
	return updateOnFailureLimiter{Redis: redis, Actor: a}
}

type updateOnFailureLimiter struct {
	Redis limiter.RedisStore
	*Actor
}

func (u updateOnFailureLimiter) TryAcquire(ctx context.Context) error {
	err := (limiter.StaticLimiter{
		Identifier: u.ID,
		Redis:      u.Redis,
		Limit:      u.RateLimit.Limit,
		Interval:   u.RateLimit.Interval,
		// Only update rate limit TTL if the actor has been updated recently.
		UpdateRateLimitTTL: u.LastUpdated != nil && time.Since(*u.LastUpdated) < 5*time.Minute,
	}).TryAcquire(ctx)

	if errors.Is(err, limiter.NoAccessError{}) || errors.Is(err, limiter.RateLimitExceededError{}) {
		u.Actor.Update(ctx) // TODO: run this in goroutine+background context maybe?
	}

	return err
}

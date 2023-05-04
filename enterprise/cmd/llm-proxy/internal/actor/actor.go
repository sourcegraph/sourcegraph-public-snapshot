package actor

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/dotcom"
)

type SubscriptionRateLimit struct {
	Limit    int
	Interval time.Duration
}

type Subscription struct {
	// ID is the subscription ID.
	ID string `json:"id"`
	// AccessEnabled is an evaluated field that summarizes whether or not LLM-proxy access
	// is enabled, based on whether the subscription is archived, if access is enabled, and
	// if any rate limits are set.
	AccessEnabled bool `json:"accessEnabled"`
	// RateLimit is the rate limit for LLM-proxy access.
	RateLimit *SubscriptionRateLimit `json:"rateLimit"`
}

func NewSubscriptionFromDotCom(s dotcom.ProductSubscriptionState) *Subscription {
	var rateLimit *SubscriptionRateLimit
	if s.LlmProxyAccess.RateLimit != nil {
		rateLimit = &SubscriptionRateLimit{
			Limit:    s.LlmProxyAccess.RateLimit.Limit,
			Interval: time.Duration(s.LlmProxyAccess.RateLimit.IntervalSeconds) * time.Second,
		}
	}
	return &Subscription{
		ID:            s.Id,
		AccessEnabled: !s.IsArchived && rateLimit != nil && s.LlmProxyAccess.Enabled,
		RateLimit:     rateLimit,
	}
}

type Actor struct {
	Subscription *Subscription
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

// WithActor returns a new context with the given Actor instance.
func WithActor(ctx context.Context, a *Actor) context.Context {
	return context.WithValue(ctx, actorKey, a)
}

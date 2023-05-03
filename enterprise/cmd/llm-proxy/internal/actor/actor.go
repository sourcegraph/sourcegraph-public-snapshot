package actor

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/llm-proxy/internal/dotcom"
)

type Subscription struct {
	ID        string                                                                                                                                             `json:"id"`
	RateLimit *dotcom.CheckAccessTokenDotcomDotcomQueryProductSubscriptionByAccessTokenProductSubscriptionLlmProxyAccessLLMProxyAccessRateLimitLLMProxyRateLimit `json:"rateLimit"`
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

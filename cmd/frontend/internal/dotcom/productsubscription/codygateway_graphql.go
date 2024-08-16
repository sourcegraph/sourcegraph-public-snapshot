package productsubscription

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayactor"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
)

type codyGatewayRateLimitResolver struct {
	actorID     string
	actorSource codygatewayactor.ActorSource
	feature     types.CompletionsFeature
	source      graphqlbackend.CodyGatewayRateLimitSource
	v           licensing.CodyGatewayRateLimit
}

func (r *codyGatewayRateLimitResolver) Source() graphqlbackend.CodyGatewayRateLimitSource {
	return r.source
}

func (r *codyGatewayRateLimitResolver) AllowedModels() []string { return r.v.AllowedModels }

func (r *codyGatewayRateLimitResolver) Limit() graphqlbackend.BigInt {
	return graphqlbackend.BigInt(r.v.Limit)
}

func (r *codyGatewayRateLimitResolver) IntervalSeconds() int32 { return r.v.IntervalSeconds }

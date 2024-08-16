package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type DotcomRootResolver interface {
	DotcomResolver
	Dotcom() DotcomResolver
	NodeResolvers() map[string]NodeByIDFunc
}

// DotcomResolver is the interface for the GraphQL types DotcomMutation and DotcomQuery.
type DotcomResolver interface {
	// DotcomQuery
	CodyGatewayDotcomUserByToken(context.Context, *CodyGatewayUsersByAccessTokenArgs) (CodyGatewayUser, error)
	CodyGatewayRateLimitStatusByUserName(context.Context, *CodyGatewayRateLimitStatusByUserNameArgs) (*[]RateLimitStatus, error)
}

type UpdateCodyGatewayAccessInput struct {
	Enabled                                 *bool
	ChatCompletionsRateLimit                *BigInt
	ChatCompletionsRateLimitIntervalSeconds *int32
	ChatCompletionsAllowedModels            *[]string
	CodeCompletionsRateLimit                *BigInt
	CodeCompletionsRateLimitIntervalSeconds *int32
	CodeCompletionsAllowedModels            *[]string
	EmbeddingsRateLimit                     *BigInt
	EmbeddingsRateLimitIntervalSeconds      *int32
	EmbeddingsAllowedModels                 *[]string
}

type CodyGatewayUsersByAccessTokenArgs struct {
	Token string
}

type CodyGatewayRateLimitStatusByUserNameArgs struct {
	Username string
}

type CodyGatewayUser interface {
	Username() string
	CodyGatewayAccess() CodyGatewayAccess
	ID() graphql.ID
}

type CodyGatewayAccess interface {
	Enabled() bool
	ChatCompletionsRateLimit(context.Context) (CodyGatewayRateLimit, error)
	CodeCompletionsRateLimit(context.Context) (CodyGatewayRateLimit, error)
	EmbeddingsRateLimit(context.Context) (CodyGatewayRateLimit, error)
}

type CodyGatewayRateLimitSource string

const (
	CodyGatewayRateLimitSourceOverride CodyGatewayRateLimitSource = "OVERRIDE"
	CodyGatewayRateLimitSourcePlan     CodyGatewayRateLimitSource = "PLAN"
)

type CodyGatewayRateLimit interface {
	Source() CodyGatewayRateLimitSource
	AllowedModels() []string
	Limit() BigInt
	IntervalSeconds() int32
}

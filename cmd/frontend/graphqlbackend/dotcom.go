package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type DotcomRootResolver interface {
	DotcomResolver
	Dotcom() DotcomResolver
	NodeResolvers() map[string]NodeByIDFunc
}

// DotcomResolver is the interface for the GraphQL types DotcomMutation and DotcomQuery.
type DotcomResolver interface {
	// DotcomMutation
	CreateProductSubscription(context.Context, *CreateProductSubscriptionArgs) (ProductSubscription, error)
	UpdateProductSubscription(context.Context, *UpdateProductSubscriptionArgs) (*EmptyResponse, error)
	GenerateProductLicenseForSubscription(context.Context, *GenerateProductLicenseForSubscriptionArgs) (ProductLicense, error)
	ArchiveProductSubscription(context.Context, *ArchiveProductSubscriptionArgs) (*EmptyResponse, error)
	RevokeLicense(context.Context, *RevokeLicenseArgs) (*EmptyResponse, error)

	// DotcomQuery
	ProductSubscription(context.Context, *ProductSubscriptionArgs) (ProductSubscription, error)
	ProductSubscriptions(context.Context, *ProductSubscriptionsArgs) (ProductSubscriptionConnection, error)
	ProductSubscriptionByAccessToken(context.Context, *ProductSubscriptionByAccessTokenArgs) (ProductSubscription, error)
	ProductLicenses(context.Context, *ProductLicensesArgs) (ProductLicenseConnection, error)
	ProductLicenseByID(ctx context.Context, id graphql.ID) (ProductLicense, error)
	ProductSubscriptionByID(ctx context.Context, id graphql.ID) (ProductSubscription, error)
	CodyGatewayDotcomUserByToken(context.Context, *CodyGatewayUsersByAccessTokenArgs) (CodyGatewayUser, error)
}

// ProductSubscription is the interface for the GraphQL type ProductSubscription.
type ProductSubscription interface {
	ID() graphql.ID
	UUID() string
	Name() string
	Account(context.Context) (*UserResolver, error)
	ActiveLicense(context.Context) (ProductLicense, error)
	ProductLicenses(context.Context, *graphqlutil.ConnectionArgs) (ProductLicenseConnection, error)
	CodyGatewayAccess() CodyGatewayAccess
	CreatedAt() gqlutil.DateTime
	IsArchived() bool
	URL(context.Context) (string, error)
	URLForSiteAdmin(context.Context) *string
	CurrentSourcegraphAccessToken(context.Context) (*string, error)
	SourcegraphAccessTokens(context.Context) ([]string, error)
}

type CreateProductSubscriptionArgs struct {
	AccountID graphql.ID
}

type GenerateProductLicenseForSubscriptionArgs struct {
	ProductSubscriptionID graphql.ID
	License               *ProductLicenseInput
}

type GenerateAccessTokenForSubscriptionArgs struct {
	ProductSubscriptionID graphql.ID
}

// ProductSubscriptionAccessToken is the interface for the GraphQL type ProductSubscriptionAccessToken.
type ProductSubscriptionAccessToken interface {
	AccessToken() string
}

type ArchiveProductSubscriptionArgs struct{ ID graphql.ID }

type ProductSubscriptionArgs struct {
	UUID string
}

type ProductSubscriptionsArgs struct {
	graphqlutil.ConnectionArgs
	Account *graphql.ID
	Query   *string
}

// ProductSubscriptionConnection is the interface for the GraphQL type
// ProductSubscriptionConnection.
type ProductSubscriptionConnection interface {
	Nodes(context.Context) ([]ProductSubscription, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}

// ProductLicense is the interface for the GraphQL type ProductLicense.
type ProductLicense interface {
	ID() graphql.ID
	Subscription(context.Context) (ProductSubscription, error)
	Info() (*ProductLicenseInfo, error)
	LicenseKey() string
	SiteID() *string
	CreatedAt() gqlutil.DateTime
	RevokedAt() *gqlutil.DateTime
	RevokeReason() *string
	Version() int32
}

// ProductLicenseInput implements the GraphQL type ProductLicenseInput.
type ProductLicenseInput struct {
	Tags                     []string
	UserCount                int32
	ExpiresAt                int32
	SalesforceSubscriptionID *string
	SalesforceOpportunityID  *string
}

type ProductLicensesArgs struct {
	graphqlutil.ConnectionArgs
	LicenseKeySubstring   *string
	ProductSubscriptionID *graphql.ID
}

// ProductLicenseConnection is the interface for the GraphQL type ProductLicenseConnection.
type ProductLicenseConnection interface {
	Nodes(context.Context) ([]ProductLicense, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}

type ProductSubscriptionByAccessTokenArgs struct {
	AccessToken string
}

type UpdateProductSubscriptionArgs struct {
	ID     graphql.ID
	Update UpdateProductSubscriptionInput
}

type RevokeLicenseArgs struct {
	ID     graphql.ID
	Reason string
}

type UpdateProductSubscriptionInput struct {
	CodyGatewayAccess *UpdateCodyGatewayAccessInput
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

type CodyGatewayUsageDatapoint interface {
	Date() gqlutil.DateTime
	Model() string
	Count() BigInt
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
	Usage(context.Context) ([]CodyGatewayUsageDatapoint, error)
}

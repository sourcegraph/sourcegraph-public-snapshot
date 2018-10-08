package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Dotcom is the implementation of the GraphQL type DotcomMutation. If it is not set at runtime, a
// "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Dotcom DotcomResolver

func (schemaResolver) Dotcom() (DotcomResolver, error) {
	if Dotcom == nil {
		return nil, errors.New("dotcom is not implemented")
	}
	return Dotcom, nil
}

// DotcomResolver is the interface for the GraphQL types DotcomMutation and DotcomQuery.
type DotcomResolver interface {
	// DotcomMutation
	SetUserBilling(context.Context, *SetUserBillingArgs) (*EmptyResponse, error)
	CreateProductSubscription(context.Context, *CreateProductSubscriptionArgs) (ProductSubscription, error)
	SetProductSubscriptionBilling(context.Context, *SetProductSubscriptionBillingArgs) (*EmptyResponse, error)
	GenerateProductLicenseForSubscription(context.Context, *GenerateProductLicenseForSubscriptionArgs) (ProductLicense, error)
	CreatePaidProductSubscription(context.Context, *CreatePaidProductSubscriptionArgs) (*CreatePaidProductSubscriptionResult, error)
	ArchiveProductSubscription(context.Context, *ArchiveProductSubscriptionArgs) (*EmptyResponse, error)

	// DotcomQuery
	ProductSubscription(context.Context, *ProductSubscriptionArgs) (ProductSubscription, error)
	ProductSubscriptions(context.Context, *ProductSubscriptionsArgs) (ProductSubscriptionConnection, error)
	PreviewProductSubscriptionInvoice(context.Context, *PreviewProductSubscriptionInvoiceArgs) (ProductSubscriptionPreviewInvoice, error)
	ProductLicenses(context.Context, *ProductLicensesArgs) (ProductLicenseConnection, error)
	ProductPlans(context.Context) ([]ProductPlan, error)
}

// ProductSubscriptionByID is called to look up a ProductSubscription given its GraphQL ID.
//
// This is contributed by enterprise.
var ProductSubscriptionByID func(context.Context, graphql.ID) (ProductSubscription, error)

// ProductSubscription is the interface for the GraphQL type ProductSubscription.
type ProductSubscription interface {
	ID() graphql.ID
	UUID() string
	Name() string
	Account(context.Context) (*UserResolver, error)
	Plan(context.Context) (ProductPlan, error)
	UserCount(context.Context) (*int32, error)
	ExpiresAt(context.Context) (*string, error)
	Events(context.Context) ([]ProductSubscriptionEvent, error)
	ActiveLicense(context.Context) (ProductLicense, error)
	ProductLicenses(context.Context, *graphqlutil.ConnectionArgs) (ProductLicenseConnection, error)
	CreatedAt() string
	IsArchived() bool
	URL(context.Context) (string, error)
	URLForSiteAdmin(context.Context) *string
	URLForSiteAdminBilling(context.Context) (*string, error)
}

type SetUserBillingArgs struct {
	User              graphql.ID
	BillingCustomerID *string
}

type CreateProductSubscriptionArgs struct {
	AccountID graphql.ID
}

type SetProductSubscriptionBillingArgs struct {
	ID                    graphql.ID
	BillingSubscriptionID *string
}

type GenerateProductLicenseForSubscriptionArgs struct {
	ProductSubscriptionID graphql.ID
	License               *ProductLicenseInput
}

type CreatePaidProductSubscriptionArgs struct {
	AccountID           graphql.ID
	ProductSubscription ProductSubscriptionInput
	PaymentToken        string
}

// ProductSubscriptionInput implements the GraphQL type ProductSubscriptionInput.
type ProductSubscriptionInput struct {
	BillingPlanID string
	UserCount     int32
}

// CreatePaidProductSubscriptionResult implements the GraphQL type CreatePaidProductSubscriptionResult.
type CreatePaidProductSubscriptionResult struct {
	ProductSubscriptionValue ProductSubscription
}

func (r *CreatePaidProductSubscriptionResult) ProductSubscription() ProductSubscription {
	return r.ProductSubscriptionValue
}

type ArchiveProductSubscriptionArgs struct{ ID graphql.ID }

type ProductSubscriptionArgs struct {
	UUID string
}

type ProductSubscriptionsArgs struct {
	graphqlutil.ConnectionArgs
	Account *graphql.ID
}

// ProductSubscriptionConnection is the interface for the GraphQL type
// ProductSubscriptionConnection.
type ProductSubscriptionConnection interface {
	Nodes(context.Context) ([]ProductSubscription, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}

type PreviewProductSubscriptionInvoiceArgs struct {
	Account              graphql.ID
	SubscriptionToUpdate *graphql.ID
	ProductSubscription  ProductSubscriptionInput
}

// ProductLicenseByID is called to look up a ProductLicense given its GraphQL ID.
//
// This is contributed by enterprise.
var ProductLicenseByID func(context.Context, graphql.ID) (ProductLicense, error)

// ProductLicense is the interface for the GraphQL type ProductLicense.
type ProductLicense interface {
	ID() graphql.ID
	Subscription(context.Context) (ProductSubscription, error)
	Info() (*ProductLicenseInfo, error)
	LicenseKey() string
	CreatedAt() string
}

// ProductLicenseInput implements the GraphQL type ProductLicenseInput.
type ProductLicenseInput struct {
	Tags      []string
	UserCount int32
	ExpiresAt int32
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

// ProductSubscriptionPreviewInvoice is the interface for the GraphQL type
// ProductSubscriptionPreviewInvoice.
type ProductSubscriptionPreviewInvoice interface {
	AmountDue() int32
	ProrationDate() int32
}

// ProductPlan is the interface for the GraphQL type ProductPlan.
type ProductPlan interface {
	BillingPlanID() string
	Name() string
	NameWithBrand() string
	PricePerUserPerYear() int32
}

// ProductSubscriptionEvent is the interface for the GraphQL type ProductSubscriptionEvent.
type ProductSubscriptionEvent interface {
	ID() string
	Date() string
	Title() string
	Description() *string
	URL() *string
}

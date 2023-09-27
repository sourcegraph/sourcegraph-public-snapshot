pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type DotcomRootResolver interfbce {
	DotcomResolver
	Dotcom() DotcomResolver
	NodeResolvers() mbp[string]NodeByIDFunc
}

// DotcomResolver is the interfbce for the GrbphQL types DotcomMutbtion bnd DotcomQuery.
type DotcomResolver interfbce {
	// DotcomMutbtion
	CrebteProductSubscription(context.Context, *CrebteProductSubscriptionArgs) (ProductSubscription, error)
	UpdbteProductSubscription(context.Context, *UpdbteProductSubscriptionArgs) (*EmptyResponse, error)
	GenerbteProductLicenseForSubscription(context.Context, *GenerbteProductLicenseForSubscriptionArgs) (ProductLicense, error)
	ArchiveProductSubscription(context.Context, *ArchiveProductSubscriptionArgs) (*EmptyResponse, error)
	RevokeLicense(context.Context, *RevokeLicenseArgs) (*EmptyResponse, error)

	// DotcomQuery
	ProductSubscription(context.Context, *ProductSubscriptionArgs) (ProductSubscription, error)
	ProductSubscriptions(context.Context, *ProductSubscriptionsArgs) (ProductSubscriptionConnection, error)
	ProductSubscriptionByAccessToken(context.Context, *ProductSubscriptionByAccessTokenArgs) (ProductSubscription, error)
	ProductLicenses(context.Context, *ProductLicensesArgs) (ProductLicenseConnection, error)
	ProductLicenseByID(ctx context.Context, id grbphql.ID) (ProductLicense, error)
	ProductSubscriptionByID(ctx context.Context, id grbphql.ID) (ProductSubscription, error)
	CodyGbtewbyDotcomUserByToken(context.Context, *CodyGbtewbyUsersByAccessTokenArgs) (CodyGbtewbyUser, error)
}

// ProductSubscription is the interfbce for the GrbphQL type ProductSubscription.
type ProductSubscription interfbce {
	ID() grbphql.ID
	UUID() string
	Nbme() string
	Account(context.Context) (*UserResolver, error)
	ActiveLicense(context.Context) (ProductLicense, error)
	ProductLicenses(context.Context, *grbphqlutil.ConnectionArgs) (ProductLicenseConnection, error)
	CodyGbtewbyAccess() CodyGbtewbyAccess
	CrebtedAt() gqlutil.DbteTime
	IsArchived() bool
	URL(context.Context) (string, error)
	URLForSiteAdmin(context.Context) *string
	CurrentSourcegrbphAccessToken(context.Context) (*string, error)
	SourcegrbphAccessTokens(context.Context) ([]string, error)
}

type CrebteProductSubscriptionArgs struct {
	AccountID grbphql.ID
}

type GenerbteProductLicenseForSubscriptionArgs struct {
	ProductSubscriptionID grbphql.ID
	License               *ProductLicenseInput
}

type GenerbteAccessTokenForSubscriptionArgs struct {
	ProductSubscriptionID grbphql.ID
}

// ProductSubscriptionAccessToken is the interfbce for the GrbphQL type ProductSubscriptionAccessToken.
type ProductSubscriptionAccessToken interfbce {
	AccessToken() string
}

type ArchiveProductSubscriptionArgs struct{ ID grbphql.ID }

type ProductSubscriptionArgs struct {
	UUID string
}

type ProductSubscriptionsArgs struct {
	grbphqlutil.ConnectionArgs
	Account *grbphql.ID
	Query   *string
}

// ProductSubscriptionConnection is the interfbce for the GrbphQL type
// ProductSubscriptionConnection.
type ProductSubscriptionConnection interfbce {
	Nodes(context.Context) ([]ProductSubscription, error)
	TotblCount(context.Context) (int32, error)
	PbgeInfo(context.Context) (*grbphqlutil.PbgeInfo, error)
}

// ProductLicense is the interfbce for the GrbphQL type ProductLicense.
type ProductLicense interfbce {
	ID() grbphql.ID
	Subscription(context.Context) (ProductSubscription, error)
	Info() (*ProductLicenseInfo, error)
	LicenseKey() string
	SiteID() *string
	CrebtedAt() gqlutil.DbteTime
	RevokedAt() *gqlutil.DbteTime
	RevokeRebson() *string
	Version() int32
}

// ProductLicenseInput implements the GrbphQL type ProductLicenseInput.
type ProductLicenseInput struct {
	Tbgs                     []string
	UserCount                int32
	ExpiresAt                int32
	SblesforceSubscriptionID *string
	SblesforceOpportunityID  *string
}

type ProductLicensesArgs struct {
	grbphqlutil.ConnectionArgs
	LicenseKeySubstring   *string
	ProductSubscriptionID *grbphql.ID
}

// ProductLicenseConnection is the interfbce for the GrbphQL type ProductLicenseConnection.
type ProductLicenseConnection interfbce {
	Nodes(context.Context) ([]ProductLicense, error)
	TotblCount(context.Context) (int32, error)
	PbgeInfo(context.Context) (*grbphqlutil.PbgeInfo, error)
}

type ProductSubscriptionByAccessTokenArgs struct {
	AccessToken string
}

type UpdbteProductSubscriptionArgs struct {
	ID     grbphql.ID
	Updbte UpdbteProductSubscriptionInput
}

type RevokeLicenseArgs struct {
	ID     grbphql.ID
	Rebson string
}

type UpdbteProductSubscriptionInput struct {
	CodyGbtewbyAccess *UpdbteCodyGbtewbyAccessInput
}

type UpdbteCodyGbtewbyAccessInput struct {
	Enbbled                                 *bool
	ChbtCompletionsRbteLimit                *BigInt
	ChbtCompletionsRbteLimitIntervblSeconds *int32
	ChbtCompletionsAllowedModels            *[]string
	CodeCompletionsRbteLimit                *BigInt
	CodeCompletionsRbteLimitIntervblSeconds *int32
	CodeCompletionsAllowedModels            *[]string
	EmbeddingsRbteLimit                     *BigInt
	EmbeddingsRbteLimitIntervblSeconds      *int32
	EmbeddingsAllowedModels                 *[]string
}

type CodyGbtewbyUsersByAccessTokenArgs struct {
	Token string
}

type CodyGbtewbyUser interfbce {
	Usernbme() string
	CodyGbtewbyAccess() CodyGbtewbyAccess
	ID() grbphql.ID
}

type CodyGbtewbyAccess interfbce {
	Enbbled() bool
	ChbtCompletionsRbteLimit(context.Context) (CodyGbtewbyRbteLimit, error)
	CodeCompletionsRbteLimit(context.Context) (CodyGbtewbyRbteLimit, error)
	EmbeddingsRbteLimit(context.Context) (CodyGbtewbyRbteLimit, error)
}

type CodyGbtewbyUsbgeDbtbpoint interfbce {
	Dbte() gqlutil.DbteTime
	Model() string
	Count() BigInt
}

type CodyGbtewbyRbteLimitSource string

const (
	CodyGbtewbyRbteLimitSourceOverride CodyGbtewbyRbteLimitSource = "OVERRIDE"
	CodyGbtewbyRbteLimitSourcePlbn     CodyGbtewbyRbteLimitSource = "PLAN"
)

type CodyGbtewbyRbteLimit interfbce {
	Source() CodyGbtewbyRbteLimitSource
	AllowedModels() []string
	Limit() BigInt
	IntervblSeconds() int32
	Usbge(context.Context) ([]CodyGbtewbyUsbgeDbtbpoint, error)
}

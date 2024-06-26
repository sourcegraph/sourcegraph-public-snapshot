package subscriptionsservice

import (
	"context"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/iam"
)

// StoreV1 is the data layer carrier for subscriptions service v1. This interface
// is meant to abstract away and limit the exposure of the underlying data layer
// to the handler through a thin-wrapper.
type StoreV1 interface {
	// UpsertEnterpriseSubscription upserts a enterprise subscription record based
	// on the given options.
	UpsertEnterpriseSubscription(ctx context.Context, subscriptionID string, opts subscriptions.UpsertSubscriptionOptions) (*subscriptions.Subscription, error)
	// ListEnterpriseSubscriptions returns a list of enterprise subscriptions based
	// on the given options.
	ListEnterpriseSubscriptions(ctx context.Context, opts subscriptions.ListEnterpriseSubscriptionsOptions) ([]*subscriptions.Subscription, error)

	// ListDotcomEnterpriseSubscriptionLicenses returns a list of enterprise
	// subscription license attributes with the given filters. It silently ignores
	// any non-matching filters. The caller should check the length of the returned
	// slice to ensure all requested licenses were found.
	ListDotcomEnterpriseSubscriptionLicenses(context.Context, []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter, int) ([]*dotcomdb.LicenseAttributes, error)
	// ListDotcomEnterpriseSubscriptions returns a list of enterprise subscription
	// attributes with the given IDs from dotcom DB. It silently ignores any
	// non-existent subscription IDs. The caller should check the length of the
	// returned slice to ensure all requested subscriptions were found.
	//
	// If no IDs are provided, it returns all subscriptions.
	ListDotcomEnterpriseSubscriptions(ctx context.Context, opts dotcomdb.ListEnterpriseSubscriptionsOptions) ([]*dotcomdb.SubscriptionAttributes, error)

	// IntrospectSAMSToken takes a SAMS access token and returns relevant metadata.
	//
	// 🚨SECURITY: SAMS will return a successful result if the token is valid, but
	// is no longer active. It is critical that the caller not honor tokens where
	// `.Active == false`.
	IntrospectSAMSToken(ctx context.Context, token string) (*sams.IntrospectTokenResponse, error)
	// GetSAMSUserByID returns the SAMS user with the given ID. It returns
	// sams.ErrNotFound if no such user exists.
	//
	// Required scope: profile
	GetSAMSUserByID(ctx context.Context, id string) (*clientsv1.User, error)

	// IAMListObjects returns a list of object IDs that satisfy the given options.
	IAMListObjects(ctx context.Context, opts iam.ListObjectsOptions) ([]string, error)
	// IAMWrite adds, updates, and/or deletes the IAM relation tuples.
	IAMWrite(ctx context.Context, opts iam.WriteOptions) error
	// IAMCheck checks whether a relationship exists (thus permission allowed) using
	// the given tuple key as the check condition.
	IAMCheck(ctx context.Context, opts iam.CheckOptions) (allowed bool, _ error)
}

type storeV1 struct {
	db         *database.DB
	dotcomDB   *dotcomdb.Reader
	SAMSClient *sams.ClientV1
	IAMClient  *iam.ClientV1
}

type NewStoreV1Options struct {
	DB         *database.DB
	DotcomDB   *dotcomdb.Reader
	SAMSClient *sams.ClientV1
	IAMClient  *iam.ClientV1
}

// NewStoreV1 returns a new StoreV1 using the given resource handles.
func NewStoreV1(opts NewStoreV1Options) StoreV1 {
	return &storeV1{
		db:         opts.DB,
		dotcomDB:   opts.DotcomDB,
		SAMSClient: opts.SAMSClient,
		IAMClient:  opts.IAMClient,
	}
}

func (s *storeV1) UpsertEnterpriseSubscription(ctx context.Context, subscriptionID string, opts subscriptions.UpsertSubscriptionOptions) (*subscriptions.Subscription, error) {
	return s.db.Subscriptions().Upsert(ctx, subscriptionID, opts)
}

func (s *storeV1) ListEnterpriseSubscriptions(ctx context.Context, opts subscriptions.ListEnterpriseSubscriptionsOptions) ([]*subscriptions.Subscription, error) {
	return s.db.Subscriptions().List(ctx, opts)
}

func (s *storeV1) ListDotcomEnterpriseSubscriptionLicenses(ctx context.Context, filters []*subscriptionsv1.ListEnterpriseSubscriptionLicensesFilter, limit int) ([]*dotcomdb.LicenseAttributes, error) {
	return s.dotcomDB.ListEnterpriseSubscriptionLicenses(ctx, filters, limit)
}

func (s *storeV1) ListDotcomEnterpriseSubscriptions(ctx context.Context, opts dotcomdb.ListEnterpriseSubscriptionsOptions) ([]*dotcomdb.SubscriptionAttributes, error) {
	return s.dotcomDB.ListEnterpriseSubscriptions(ctx, opts)
}

func (s *storeV1) IntrospectSAMSToken(ctx context.Context, token string) (*sams.IntrospectTokenResponse, error) {
	return s.SAMSClient.Tokens().IntrospectToken(ctx, token)
}

func (s *storeV1) GetSAMSUserByID(ctx context.Context, id string) (*clientsv1.User, error) {
	return s.SAMSClient.Users().GetUserByID(ctx, id)
}

func (s *storeV1) IAMListObjects(ctx context.Context, opts iam.ListObjectsOptions) ([]string, error) {
	return s.IAMClient.ListObjects(ctx, opts)
}

func (s *storeV1) IAMWrite(ctx context.Context, opts iam.WriteOptions) error {
	return s.IAMClient.Write(ctx, opts)
}

func (s *storeV1) IAMCheck(ctx context.Context, opts iam.CheckOptions) (allowed bool, _ error) {
	return s.IAMClient.Check(ctx, opts)
}

package productsubscription

import (
	"context"
	"encoding/hex"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// productSubscriptionAccessTokenPrefix is the prefix used for identifying tokens
// generated for product subscriptions.
const productSubscriptionAccessTokenPrefix = "sgs_"

// defaultAccessToken creates a prefixed, encoded token for users to use from raw token contents.
func defaultAccessToken(rawToken []byte) string {
	return productSubscriptionAccessTokenPrefix + hex.EncodeToString(rawToken)
}

type productSubscriptionAccessToken struct {
	accessToken string
}

func (t productSubscriptionAccessToken) AccessToken() string { return t.accessToken }

// GenerateAccessTokenForSubscription currently creates an access token from the hash of
// the current active license of a subscription.
func (r ProductSubscriptionLicensingResolver) GenerateAccessTokenForSubscription(ctx context.Context, args *graphqlbackend.GenerateAccessTokenForSubscriptionArgs) (graphqlbackend.ProductSubscriptionAccessToken, error) {
	// 🚨 SECURITY: Only site admins may generate product access tokens.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.DB); err != nil {
		return nil, err
	}

	sub, err := productSubscriptionByID(ctx, r.DB, args.ProductSubscriptionID)
	if err != nil {
		return nil, err
	}

	active, err := dbLicenses{db: r.DB}.Active(ctx, sub.v.ID)
	if err != nil {
		return nil, err
	} else if active == nil {
		return nil, errors.New("an active license is required")
	}

	// The token comprises of a prefix and the above token.
	accessToken := productSubscriptionAccessToken{
		accessToken: defaultAccessToken(defaultRawAccessToken([]byte(active.LicenseKey))),
	}

	// Token already enabled, just return the generated token
	if active.AccessTokenEnabled {
		return accessToken, nil
	}

	// Otherwise, enable before returning
	if err := newDBTokens(r.DB).EnableUseAsAccessToken(ctx, active.ID); err != nil {
		return nil, err
	}
	return accessToken, nil
}

// ProductSubscriptionByAccessToken retrieves the subscription corresponding to the
// given access token.
func (r ProductSubscriptionLicensingResolver) ProductSubscriptionByAccessToken(ctx context.Context, args *graphqlbackend.ProductSubscriptionByAccessTokenArgs) (graphqlbackend.ProductSubscription, error) {
	// 🚨 SECURITY: Only specific entities may use this functionality.
	if err := serviceAccountOrOwnerOrSiteAdmin(ctx, r.DB, nil); err != nil {
		return nil, err
	}

	subID, err := newDBTokens(r.DB).LookupAccessToken(ctx, args.AccessToken)
	if err != nil {
		return nil, err
	}
	v, err := dbSubscriptions{db: r.DB}.GetByID(ctx, subID)
	if err != nil {
		return nil, err
	}
	return &productSubscription{v: v, db: r.DB}, nil
}

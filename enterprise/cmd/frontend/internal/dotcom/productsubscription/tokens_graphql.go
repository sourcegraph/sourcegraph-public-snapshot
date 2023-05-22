package productsubscription

import (
	"context"
	"encoding/hex"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// productSubscriptionAccessTokenPrefix is the prefix used for identifying tokens
// generated for product subscriptions.
const productSubscriptionAccessTokenPrefix = "sgs_"

// defaultAccessToken creates a prefixed, encoded token for users to use from raw token contents.
func defaultAccessToken(rawToken []byte) string {
	return productSubscriptionAccessTokenPrefix + hex.EncodeToString(rawToken)
}

// ProductSubscriptionByAccessToken retrieves the subscription corresponding to the
// given access token.
func (r ProductSubscriptionLicensingResolver) ProductSubscriptionByAccessToken(ctx context.Context, args *graphqlbackend.ProductSubscriptionByAccessTokenArgs) (graphqlbackend.ProductSubscription, error) {
	// 🚨 SECURITY: Only specific entities may use this functionality.
	if err := serviceAccountOrSiteAdmin(ctx, r.DB, false); err != nil {
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

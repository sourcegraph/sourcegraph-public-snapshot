package productsubscription

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
)

// featureFlagProductSubscriptionsServiceAccount is a feature flag that should be
// set on service accounts that can read product subscriptions.
const featureFlagProductSubscriptionsServiceAccount = "product-subscriptions-service-account"

// 🚨 SECURITY: Use this to check if access to a subscription query or mutation
// is authorized for service accounts, the owning user, and site admins.
func serviceAccountOrOwnerOrSiteAdmin(ctx context.Context, db database.DB, ownerUserID *int32) error {
	if featureflag.FromContext(ctx).GetBoolOr(featureFlagProductSubscriptionsServiceAccount, false) {
		return nil
	}

	if ownerUserID != nil {
		return auth.CheckSiteAdminOrSameUser(ctx, db, *ownerUserID)
	}

	return auth.CheckCurrentUserIsSiteAdmin(ctx, db)
}

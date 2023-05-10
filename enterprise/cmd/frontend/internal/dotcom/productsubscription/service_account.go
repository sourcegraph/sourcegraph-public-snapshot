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
	// Check if the user is a service account.
	if featureflag.FromContext(ctx).GetBoolOr(featureFlagProductSubscriptionsServiceAccount, false) {
		return nil
	}

	// If ownerUserID is specified, the user must be the owner, or a site admin.
	if ownerUserID != nil {
		return auth.CheckSiteAdminOrSameUser(ctx, db, *ownerUserID)
	}

	// Otherwise, the user must be a site admin.
	return auth.CheckCurrentUserIsSiteAdmin(ctx, db)
}

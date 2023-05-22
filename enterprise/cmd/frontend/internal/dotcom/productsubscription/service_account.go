package productsubscription

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
)

const (
	// featureFlagProductSubscriptionsServiceAccount is a feature flag that should be
	// set on service accounts that can read AND write product subscriptions.
	featureFlagProductSubscriptionsServiceAccount = "product-subscriptions-service-account"

	// featureFlagProductSubscriptionsReaderServiceAccount is a feature flag that should be
	// set on service accounts that can only read product subscriptions.
	featureFlagProductSubscriptionsReaderServiceAccount = "product-subscriptions-reader-service-account"
)

// 🚨 SECURITY: Use this to check if access to a subscription query or mutation
// is authorized for service accounts and site admins.
func serviceAccountOrSiteAdmin(
	ctx context.Context,
	db database.DB,
	// requiresWriterServiceAccount, if true, requires "product-subscriptions-service-account",
	// otherwise just "product-subscriptions-reader-service-account" is sufficient.
	requiresWriterServiceAccount bool,
) error {
	return serviceAccountOrOwnerOrSiteAdmin(ctx, db, nil, requiresWriterServiceAccount)
}

// 🚨 SECURITY: Use this to check if access to a subscription query or mutation
// is authorized for service accounts, the owning user, and site admins.
func serviceAccountOrOwnerOrSiteAdmin(
	ctx context.Context,
	db database.DB,
	ownerUserID *int32,
	// requiresWriterServiceAccount, if true, requires "product-subscriptions-service-account",
	// otherwise just "product-subscriptions-reader-service-account" is sufficient.
	requiresWriterServiceAccount bool,
) error {
	// Check if the user is has the prerequisite service account.
	serviceAccountIsWriter := featureflag.FromContext(ctx).GetBoolOr(featureFlagProductSubscriptionsServiceAccount, false)
	if requiresWriterServiceAccount {
		// 🚨 SECURITY: Require the more strict featureFlagProductSubscriptionsServiceAccount
		// if requiresWriterServiceAccount=true
		if serviceAccountIsWriter {
			return nil
		}
		// Otherwise, fall through to check if actor is owner or site admin.
	} else {
		// If requiresWriterServiceAccount==false, then just
		// featureFlagProductSubscriptionsReaderServiceAccount is sufficient.
		if serviceAccountIsWriter || featureflag.FromContext(ctx).GetBoolOr(featureFlagProductSubscriptionsReaderServiceAccount, false) {
			return nil
		}
	}

	// If ownerUserID is specified, the user must be the owner, or a site admin.
	if ownerUserID != nil {
		return auth.CheckSiteAdminOrSameUser(ctx, db, *ownerUserID)
	}

	// Otherwise, the user must be a site admin.
	return auth.CheckCurrentUserIsSiteAdmin(ctx, db)
}

package productsubscription

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
) (string, error) {
	return serviceAccountOrOwnerOrSiteAdmin(ctx, db, nil, requiresWriterServiceAccount)
}

// 🚨 SECURITY: Use this to check if access to a subscription query or mutation
// is authorized for service accounts, the owning user, and site admins. Callers
// should record the returned grant reason in an audit log tracking the access.
func serviceAccountOrOwnerOrSiteAdmin(
	ctx context.Context,
	db database.DB,
	ownerUserID *int32,
	// requiresWriterServiceAccount, if true, requires "product-subscriptions-service-account",
	// otherwise just "product-subscriptions-reader-service-account" is sufficient.
	requiresWriterServiceAccount bool,
) (string, error) {
	// Check if the user is has the prerequisite service account.
	serviceAccountIsWriter := readFeatureFlagFromDB(ctx, db, featureFlagProductSubscriptionsServiceAccount, false)
	if requiresWriterServiceAccount {
		// 🚨 SECURITY: Require the more strict featureFlagProductSubscriptionsServiceAccount
		// if requiresWriterServiceAccount=true
		if serviceAccountIsWriter {
			return "writer_service_account", nil
		}
		// Otherwise, fall through to check if actor is owner or site admin.
	} else {
		// If requiresWriterServiceAccount==false, then just
		// featureFlagProductSubscriptionsReaderServiceAccount is sufficient.
		if serviceAccountIsWriter {
			return "writer_service_account", nil
		}
		if readFeatureFlagFromDB(ctx, db, featureFlagProductSubscriptionsReaderServiceAccount, false) {
			return "reader_service_account", nil
		}
	}

	// If ownerUserID is specified, the user must be the owner, or a site admin.
	if ownerUserID != nil {
		return "same_user_or_site_admin", auth.CheckSiteAdminOrSameUser(ctx, db, *ownerUserID)
	}

	// Otherwise, the user must be a site admin.
	return "site_admin", auth.CheckCurrentUserIsSiteAdmin(ctx, db)
}

// readFeatureFlagFromDB explicitly reads the feature flag values from the database,
// instead of from the feature flag store in the context.
//
// 🚨 SECURITY: This makes it so that we solely look at the database as authority here,
// and not allow HTTP headers to override the feature flags for service accounts.
func readFeatureFlagFromDB(ctx context.Context, db database.DB, flag string, defaultValue bool) bool {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return defaultValue
	}
	ffs, err := db.FeatureFlags().GetUserFlags(ctx, a.UID)
	if err != nil {
		return defaultValue
	}
	v, ok := ffs[flag]
	if !ok {
		return defaultValue
	}
	return v
}

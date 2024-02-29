package productsubscription

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/rbac"
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
	requiresWriter bool,
) (string, error) {
	// Check if the user is has the prerequisite service account.
	subscriptionsWriter := rbac.CheckCurrentUserHasPermission(ctx, db,
		rbac.ProductsubscriptionsWritePermission) == nil
	if requiresWriter {
		// 🚨 SECURITY: Require the more strict featureFlagProductSubscriptionsServiceAccount
		// if requiresWriterServiceAccount=true
		if subscriptionsWriter {
			return rbac.ProductsubscriptionsWritePermission, nil
		}
		// Otherwise, fall through to check if actor is owner or site admin.
	} else {
		// If requiresWriterServiceAccount==false, then just reader account is
		// sufficient.
		if subscriptionsWriter {
			return rbac.ProductsubscriptionsWritePermission, nil
		}
		if rbac.CheckCurrentUserHasPermission(ctx, db,
			rbac.ProductsubscriptionsReadPermission) == nil {
			return rbac.ProductsubscriptionsReadPermission, nil
		}
	}

	// If ownerUserID is specified, the user must be the owner, or a site admin.
	if ownerUserID != nil {
		return "same_user_or_site_admin", auth.CheckSiteAdminOrSameUser(ctx, db, *ownerUserID)
	}

	// Otherwise, the user must be a site admin.
	return "site_admin", auth.CheckCurrentUserIsSiteAdmin(ctx, db)
}

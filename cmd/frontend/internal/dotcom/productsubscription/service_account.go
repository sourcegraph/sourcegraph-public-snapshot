package productsubscription

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/rbac"
)

// 🚨 SECURITY: Use this to check if access to a subscription query or mutation
// is authorized for users with the correct RBAC permissions and site admins.
func hasRBACPermsOrSiteAdmin(
	ctx context.Context,
	db database.DB,
	requiresSubscriptionsWriter bool,
) (string, error) {
	return hasRBACPermsOrOwnerOrSiteAdmin(ctx, db, nil, requiresSubscriptionsWriter)
}

// 🚨 SECURITY: Use this to check if access to a subscription query or mutation
// is authorized for service accounts, the owning user, and site admins. Callers
// should record the returned grant reason in an audit log tracking the access.
func hasRBACPermsOrOwnerOrSiteAdmin(
	ctx context.Context,
	db database.DB,
	ownerUserID *int32,
	requiresSubscriptionsWriter bool,
) (string, error) {
	// Check if the user is has the prerequisite service account.
	subscriptionsWriter := rbac.CheckCurrentUserHasPermission(ctx, db,
		rbac.ProductSubscriptionsWritePermission) == nil
	if requiresSubscriptionsWriter {
		// 🚨 SECURITY: Require the more strict featureFlagProductSubscriptionsServiceAccount
		// if requiresWriterServiceAccount=true
		if subscriptionsWriter {
			return rbac.ProductSubscriptionsWritePermission, nil
		}
		// Otherwise, fall through to check if actor is owner or site admin.
	} else {
		// If requiresWriterServiceAccount==false, then just reader account is
		// sufficient.
		if subscriptionsWriter {
			return rbac.ProductSubscriptionsWritePermission, nil
		}
		if rbac.CheckCurrentUserHasPermission(ctx, db,
			rbac.ProductSubscriptionsReadPermission) == nil {
			return rbac.ProductSubscriptionsReadPermission, nil
		}
	}

	// If ownerUserID is specified, the user must be the owner, or a site admin.
	if ownerUserID != nil {
		return "same_user_or_site_admin", auth.CheckSiteAdminOrSameUser(ctx, db, *ownerUserID)
	}

	// Otherwise, the user must be a site admin.
	return "site_admin", auth.CheckCurrentUserIsSiteAdmin(ctx, db)
}

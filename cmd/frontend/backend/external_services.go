package backend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrNoAccessExternalService = errors.New("the authenticated user does not have access to this external service")

// CheckExternalServiceAccess checks whether the current user is allowed to
// access the supplied external service.
func CheckExternalServiceAccess(ctx context.Context, db database.DB, namespaceUserID, namespaceOrgID int32) error {
	// Fast path that doesn't need to hit DB as we can get id from context
	a := actor.FromContext(ctx)
	if namespaceUserID > 0 && a.IsAuthenticated() && namespaceUserID == a.UID {
		return nil
	}

	if namespaceOrgID > 0 && CheckOrgAccess(ctx, db, namespaceOrgID) == nil {
		return nil
	}

	// Special case when external service has no owner
	if namespaceUserID == 0 && namespaceOrgID == 0 && CheckCurrentUserIsSiteAdmin(ctx, db) == nil {
		return nil
	}

	return ErrNoAccessExternalService
}

// CheckOrgExternalServices checks if the feature organization can own external services
// is allowed or not
func CheckOrgExternalServices(ctx context.Context, db database.DB, orgID int32) error {
	enabled, err := db.FeatureFlags().GetOrgFeatureFlag(ctx, orgID, "org-code")
	if err != nil {
		return err
	} else if enabled {
		return nil
	}

	return errors.New("organization code host connections are not enabled")
}

// OrgExternalServicesQuotaReached checks if the maximum mumber of external services has been
// reached for a given org. This is currenlty used only for Cloud orgs.
func OrgExternalServicesQuotaReached(ctx context.Context, db database.DB, orgID int32) (bool, error) {
	externalServicesAllowed := []string{extsvc.KindGitHub, extsvc.KindGitLab}
	options := database.ExternalServicesListOptions{NamespaceOrgID: orgID, Kinds: externalServicesAllowed}

	totalUsed, err := db.ExternalServices().Count(ctx, options)
	if err != nil {
		return true, err
	}

	if totalUsed == extsvc.MaxCodeHostsForCloudOrgs {
		return true, errors.Errorf("maximum number of external servcies has been reached for organization %v ", orgID)
	}

	return false, nil
}

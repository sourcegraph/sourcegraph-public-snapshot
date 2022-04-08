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
// reached for a given org on Cloud. Max of two services, one for GitHub and one for GitLab, can be added per org
func OrgExternalServicesQuotaReached(ctx context.Context, db database.DB, orgID int32, kind string) (bool, error) {
	services, err := servicesCountPerType(ctx, db, orgID)
	if err != nil {
		return true, err
	}

	if services[extsvc.KindGitHub] == 0 || services[extsvc.KindGitLab] == 0 {
		return false, nil
	}

	if (services[extsvc.KindGitHub] == 1 && kind == extsvc.KindGitHub) || (services[extsvc.KindGitLab] == 1 && kind == extsvc.KindGitLab) {
		return true, errors.New("only one external service of  type %s can be added per org")
	}

	return true, nil
}

// servicesCountPerType returns a dictionary with the total count for each type of service
func servicesCountPerType(ctx context.Context, db database.DB, orgID int32) (map[string]int, error) {
	options := database.ExternalServicesListOptions{NamespaceOrgID: orgID}

	services, err := db.ExternalServices().List(ctx, options)
	if err != nil {
		return nil, err
	}

	svcCountMap := map[string]int{}
	for _, svc := range services {
		if _, ok := svcCountMap[svc.Kind]; ok {
			svcCountMap[svc.Kind] += 1
		}
		svcCountMap[svc.Kind] = 1
	}

	return svcCountMap, nil
}

// IsExternalServiceAllowed checks if a given external service can be added to an org on Cloud.
// Services currently allowed are GitHub and GitLab
func IsExternalServiceAllowed(kind string) (bool, error) {
	allowed := []string{extsvc.KindGitHub, extsvc.KindGitLab}

	for _, allowed := range allowed {
		if allowed == kind {
			return true, nil
		}
	}

	return false, nil
}

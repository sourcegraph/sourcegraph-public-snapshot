package backend

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
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

// CheckExternalServicesQuota returns an error message if the maximum mumber of external services has been
// reached for a given org or user on Cloud. Max of 2 services - one for GitHub, one for GitLab - can be added
func CheckExternalServicesQuota(ctx context.Context, db database.DB, kind string, orgID, userID int32) error {
	// TODO: CONFIRM if there's a case when orgID and userID are both null
	services, err := servicesMap(ctx, db, orgID, userID)
	if err != nil {
		return err
	}

	if (kind == extsvc.KindGitHub && services[extsvc.KindGitHub] >= 1) || (kind == extsvc.KindGitLab services[extsvc.KindGitLab] >= 1) {
		return errors.New(fmt.Sprintf("a max of one %s external service is allowed", kind))
	}

	if services[extsvc.KindGitHub] == 0 || services[extsvc.KindGitLab] == 0 {
		return nil
	}

	return errors.New("maximum number of external services has been reached")
}

// servicesMap returns a dictionary with the total count for each type of service
func servicesMap(ctx context.Context, db database.DB, orgID, userID int32) (map[string]int, error) {
	var services []*types.ExternalService
	var err error

	if orgID > 0 {
		services, err = db.ExternalServices().List(ctx, database.ExternalServicesListOptions{NamespaceOrgID: orgID})
		if err != nil {
			return nil, err
		}
	}

	if userID > 0 {
		services, err = db.ExternalServices().List(ctx, database.ExternalServicesListOptions{NamespaceUserID: userID})
		if err != nil {
			return nil, err
		}
	}

	svcMap := map[string]int{}
	for _, svc := range services {
		if _, ok := svcMap[svc.Kind]; ok {
			svcMap[svc.Kind] += 1
		}
		svcMap[svc.Kind] = 1
	}

	return svcMap, nil
}

// ExternalServiceSupported checks if a given external service is supported on Cloud mode.
// Services currently supported are GitHub and GitLab
func ExternalServiceSupported(kind string) error {
	if kind == extsvc.KindGitHub || kind == extsvc.KindGitLab {
		return nil
	}

	return errors.Errorf("external service of kind %v is currently not supported on Cloud mode", kind)
}

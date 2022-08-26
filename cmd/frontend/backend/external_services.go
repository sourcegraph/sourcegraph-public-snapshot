package backend

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrNoAccessExternalService = errors.New("the authenticated user does not have access to this external service")
var ErrExternalServiceLimitPerKindReached = errors.New("cannot add more than one external service of a given kind")
var ErrExternalServiceKindNotSupported = errors.New("external service kind not supported on Cloud mode")

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

// CheckExternalServicesQuota checks if the maximum mumber of external services has been
// reached. Max of 2 services - one for GitHub, one for GitLab - can be added per org or user
func CheckExternalServicesQuota(ctx context.Context, db database.DB, kind string, orgID, userID int32) error {
	const limitPerKind = 1

	services, err := servicesCountPerKind(ctx, db, orgID, userID)
	if err != nil {
		return err
	}

	if kind == extsvc.KindGitHub {
		if services[extsvc.KindGitHub] >= limitPerKind {
			return ErrExternalServiceLimitPerKindReached
		}
		return nil
	}

	if kind == extsvc.KindGitLab {
		if services[extsvc.KindGitLab] >= limitPerKind {
			return ErrExternalServiceLimitPerKindReached
		}
		return nil
	}

	return nil
}

// servicesCountPerKind returns a map with the total count for each type of service
func servicesCountPerKind(ctx context.Context, db database.DB, orgID, userID int32) (map[string]int, error) {
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

// ExternalServiceKindSupported checks if a given external service is supported on Cloud mode.
// Services currently supported are GitHub and GitLab
func ExternalServiceKindSupported(kind string) error {
	if kind == extsvc.KindGitHub || kind == extsvc.KindGitLab {
		return nil
	}

	return ErrExternalServiceKindNotSupported
}

// repoupdaterClient is an interface with only the methods required in SyncExternalService. As a
// result instead of using the entire repoupdater client implementation, we use a thinner API which
// only needs the SyncExternalService method to be defined on the object.
type repoupdaterClient interface {
	SyncExternalService(ctx context.Context, externalServiceID int64) (*protocol.ExternalServiceSyncResult, error)
}

// SyncExternalService will eagerly trigger a repo-updater sync. It accepts a
// timeout as an argument which is recommended to be 5 seconds unless the caller
// has special requirements for it to be larger or smaller.
func SyncExternalService(ctx context.Context, svc *types.ExternalService, timeout time.Duration, client repoupdaterClient) (err error) {
	// Set a timeout to validate external service sync. It usually fails in
	// under 5s if there is a problem.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	defer func() {
		// err is either nil or contains an actual error from the API call. And we return it
		// nonetheless.
		err = errors.Wrapf(err, "error in SyncExternalService for service %q with ID %d", svc.Kind, svc.ID)

		// If context error is anything but a deadline exceeded error, we do not want to propagate
		// it. But we definitely want to log the error as a warning.
		if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
			log15.Warn("SyncExternalService: context error discarded", "err", ctx.Err())
			err = nil
		}
	}()

	_, err = client.SyncExternalService(ctx, svc.ID)
	return err
}

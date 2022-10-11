package backend

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
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

	if namespaceOrgID > 0 && auth.CheckOrgAccess(ctx, db, namespaceOrgID) == nil {
		return nil
	}

	// Special case when external service has no owner
	if namespaceUserID == 0 && namespaceOrgID == 0 && auth.CheckCurrentUserIsSiteAdmin(ctx, db) == nil {
		return nil
	}

	return ErrNoAccessExternalService
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
func SyncExternalService(ctx context.Context, logger log.Logger, svc *types.ExternalService, timeout time.Duration, client repoupdaterClient) (err error) {
	logger = logger.Scoped("SyncExternalService", "handles triggering of repo-updater syncing for a particular external service")
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
			logger.Warn("context error discarded", log.Error(ctx.Err()))
			err = nil
		}
	}()

	_, err = client.SyncExternalService(ctx, svc.ID)
	return err
}

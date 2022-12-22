package backend

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

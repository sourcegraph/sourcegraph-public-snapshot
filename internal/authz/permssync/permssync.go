package permssync

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

const featureFlagName = "database-permission-sync-worker"

func PermissionSyncWorkerEnabled(ctx context.Context) bool {
	return featureflag.FromContext(ctx).GetBoolOr(featureFlagName, false)
}

// SchedulePermsSync enqueues a permission sync job for the given
// PermsSyncRequest. If the feature flag for the database-backed permission
// sync worker is enabled then that will be used. Otherwise a request to
// repo-updater will be sent to enqueue a sync job directly on the PermsSyncer.
//
// Errors are not fatal since our background permissions syncer will eventually
// sync the user anyway, so we just log any errors and don't return them here.
func SchedulePermsSync(ctx context.Context, logger log.Logger, db database.DB, req protocol.PermsSyncRequest) {
	// If the new permission sync worker is enabled, we create sync jobs, otherwise we send a request to repo-updater
	if PermissionSyncWorkerEnabled(ctx) {
		for _, userID := range req.UserIDs {
			opts := database.PermissionSyncJobOpts{
				HighPriority:     true,
				InvalidateCaches: req.Options.InvalidateCaches,
			}
			err := db.PermissionSyncJobs().CreateUserSyncJob(ctx, userID, opts)
			if err != nil {
				logger.Warn("Error enqueueing permissions sync job", log.Error(err), log.Int32("user_id", userID))
			}
		}

		for _, repoID := range req.RepoIDs {
			opts := database.PermissionSyncJobOpts{
				HighPriority:     true,
				InvalidateCaches: req.Options.InvalidateCaches,
			}
			err := db.PermissionSyncJobs().CreateUserSyncJob(ctx, int32(repoID), opts)
			if err != nil {
				logger.Warn("Error enqueueing permissions sync job", log.Error(err), log.Int32("repo_id", int32(repoID)))
			}
		}

		return
	}

	if err := repoupdater.DefaultClient.SchedulePermsSync(ctx, req); err != nil {
		repoIDs := make([]int32, 0, len(req.RepoIDs))
		for _, id := range req.RepoIDs {
			repoIDs = append(repoIDs, int32(id))
		}
		logger.Warn("Error scheduling permissions sync", log.Error(err), log.Int32s("user_ids", req.UserIDs), log.Int32s("repo_ids", repoIDs))
	}
}

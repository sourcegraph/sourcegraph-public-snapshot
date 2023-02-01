package permssync

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

const featureFlagName = "database-permission-sync-worker"

func PermissionSyncWorkerEnabled(ctx context.Context, db database.DB, logger log.Logger) bool {
	globalFeatureFlags, err := db.FeatureFlags().GetGlobalFeatureFlags(ctx)
	if err != nil {
		logger.
			Scoped("PermissionSyncWorkerEnabled", "checking feature flag for permission sync worker").
			Warn("failed to query global feature flags", log.Error(err))
		return false
	}
	return globalFeatureFlags[featureFlagName]
}

var MockSchedulePermsSync func(ctx context.Context, logger log.Logger, db database.DB, req protocol.PermsSyncRequest)

// SchedulePermsSync enqueues a permission sync job for the given
// PermsSyncRequest.
//
// If the feature flag for the database-backed permission sync worker is enabled
// then that will be used. Otherwise, a request to repo-updater will be sent to
// enqueue a sync job directly on the PermsSyncer.
//
// Errors are not fatal since our background permissions syncer will eventually
// sync the user anyway, so we just log any errors and don't return them here.
func SchedulePermsSync(ctx context.Context, logger log.Logger, db database.DB, req protocol.PermsSyncRequest) {
	if MockSchedulePermsSync != nil {
		MockSchedulePermsSync(ctx, logger, db, req)
		return
	}

	// If the new permission sync worker is enabled, we create sync jobs, otherwise we send a request to repo-updater
	if PermissionSyncWorkerEnabled(ctx, db, logger) {
		for _, userID := range req.UserIDs {
			opts := database.PermissionSyncJobOpts{
				Priority:          database.HighPriorityPermissionSync,
				InvalidateCaches:  req.Options.InvalidateCaches,
				Reason:            req.Reason,
				TriggeredByUserID: req.TriggeredByUserID,
				ProcessAfter:      req.ProcessAfter,
			}
			err := db.PermissionSyncJobs().CreateUserSyncJob(ctx, userID, opts)
			if err != nil {
				logger.Warn("Error enqueueing permissions sync job", log.Error(err), log.Int32("user_id", userID))
			}
		}

		for _, repoID := range req.RepoIDs {
			opts := database.PermissionSyncJobOpts{
				Priority:          database.HighPriorityPermissionSync,
				InvalidateCaches:  req.Options.InvalidateCaches,
				Reason:            req.Reason,
				TriggeredByUserID: req.TriggeredByUserID,
				ProcessAfter:      req.ProcessAfter,
			}
			err := db.PermissionSyncJobs().CreateRepoSyncJob(ctx, repoID, opts)
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

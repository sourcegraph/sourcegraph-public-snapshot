package permssync

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// ScheduleSyncOpts captures options to sync permissions. The provided options are
// used to sync all provided users and repos.
type ScheduleSyncOpts struct {
	UserIDs           []int32                           `json:"user_ids"`
	RepoIDs           []api.RepoID                      `json:"repo_ids"`
	Options           authz.FetchPermsOptions           `json:"options"`
	Reason            database.PermissionsSyncJobReason `json:"reason"`
	TriggeredByUserID int32                             `json:"triggered_by_user_id"`
	ProcessAfter      time.Time                         `json:"process_after"`
}

var MockSchedulePermsSync func(ctx context.Context, logger log.Logger, db database.DB, req ScheduleSyncOpts)

// SchedulePermsSync enqueues a permission sync job for the given
// ScheduleSyncOpts.
//
// If the feature flag for the database-backed permission sync worker is enabled
// then that will be used. Otherwise, a request to repo-updater will be sent to
// enqueue a sync job directly on the PermsSyncer.
//
// Errors are not fatal since our background permissions syncer will eventually
// sync the user anyway, so we just log any errors and don't return them here.
func SchedulePermsSync(ctx context.Context, logger log.Logger, db database.DB, req ScheduleSyncOpts) {
	if MockSchedulePermsSync != nil {
		MockSchedulePermsSync(ctx, logger, db, req)
		return
	}

	// If the new permission sync worker is enabled, we create sync jobs, otherwise we send a request to repo-updater
	for _, userID := range req.UserIDs {
		opts := database.PermissionSyncJobOpts{
			Priority:          database.HighPriorityPermissionsSync,
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
			Priority:          database.HighPriorityPermissionsSync,
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
}

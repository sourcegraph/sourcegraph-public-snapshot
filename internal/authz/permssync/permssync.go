package permssync

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

const (
	featureFlagName = "database-permission-sync-worker"

	// ReasonUserOutdatedPermissions and below are reasons of scheduled permission
	// syncs.
	ReasonUserOutdatedPermissions = "REASON_USER_OUTDATED_PERMS"
	ReasonUserNoPermissions       = "REASON_USER_NO_PERMS"
	ReasonUserEmailRemoved        = "REASON_USER_EMAIL_REMOVED"
	ReasonUserEmailVerified       = "REASON_USER_EMAIL_VERIFIED"
	ReasonUserAddedToOrg          = "REASON_USER_ADDED_TO_ORG"
	ReasonUserRemovedFromOrg      = "REASON_USER_REMOVED_FROM_ORG"
	ReasonUserAcceptedOrgInvite   = "REASON_USER_ACCEPTED_ORG_INVITE"
	ReasonRepoOutdatedPermissions = "REASON_REPO_OUTDATED_PERMS"
	ReasonRepoNoPermissions       = "REASON_REPO_NO_PERMS"
	ReasonRepoUpdatedFromCodeHost = "REASON_REPO_UPDATED_FROM_CODE_HOST"

	// ReasonGitHubUserEvent and below are reasons of permission syncs triggered by
	// webhook events.
	ReasonGitHubUserEvent                  = "REASON_GITHUB_USER_EVENT"
	ReasonGitHubUserAddedEvent             = "REASON_GITHUB_USER_ADDED_EVENT"
	ReasonGitHubUserRemovedEvent           = "REASON_GITHUB_USER_REMOVED_EVENT"
	ReasonGitHubUserMembershipAddedEvent   = "REASON_GITHUB_USER_MEMBERSHIP_ADDED_EVENT"
	ReasonGitHubUserMembershipRemovedEvent = "REASON_GITHUB_USER_MEMBERSHIP_REMOVED_EVENT"
	ReasonGitHubTeamAddedToRepoEvent       = "REASON_GITHUB_TEAM_ADDED_TO_REPO_EVENT"
	ReasonGitHubTeamRemovedFromRepoEvent   = "REASON_GITHUB_TEAM_REMOVED_FROM_REPO_EVENT"
	ReasonGitHubOrgMemberAddedEvent        = "REASON_GITHUB_ORG_MEMBER_ADDED_EVENT"
	ReasonGitHubOrgMemberRemovedEvent      = "REASON_GITHUB_ORG_MEMBER_REMOVED_EVENT"
	ReasonGitHubRepoEvent                  = "REASON_GITHUB_REPO_EVENT"
	ReasonGitHubRepoMadePrivateEvent       = "REASON_GITHUB_REPO_MADE_PRIVATE_EVENT"

	// ReasonManualRepoSync and below are reasons of permission syncs triggered
	// manually.
	ReasonManualRepoSync = "REASON_MANUAL_REPO_SYNC"
	ReasonManualUserSync = "REASON_MANUAL_USER_SYNC"
)

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
				HighPriority:      true,
				InvalidateCaches:  req.Options.InvalidateCaches,
				Reason:            req.Reason,
				TriggeredByUserID: req.TriggeredByUserID,
			}
			err := db.PermissionSyncJobs().CreateUserSyncJob(ctx, userID, opts)
			if err != nil {
				logger.Warn("Error enqueueing permissions sync job", log.Error(err), log.Int32("user_id", userID))
			}
		}

		for _, repoID := range req.RepoIDs {
			opts := database.PermissionSyncJobOpts{
				HighPriority:      true,
				InvalidateCaches:  req.Options.InvalidateCaches,
				Reason:            req.Reason,
				TriggeredByUserID: req.TriggeredByUserID,
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

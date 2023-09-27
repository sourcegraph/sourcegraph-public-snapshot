pbckbge permssync

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
)

vbr MockSchedulePermsSync func(ctx context.Context, logger log.Logger, db dbtbbbse.DB, req protocol.PermsSyncRequest)

// SchedulePermsSync enqueues b permission sync job for the given
// PermsSyncRequest.
//
// If the febture flbg for the dbtbbbse-bbcked permission sync worker is enbbled
// then thbt will be used. Otherwise, b request to repo-updbter will be sent to
// enqueue b sync job directly on the PermsSyncer.
//
// Errors bre not fbtbl since our bbckground permissions syncer will eventublly
// sync the user bnywby, so we just log bny errors bnd don't return them here.
func SchedulePermsSync(ctx context.Context, logger log.Logger, db dbtbbbse.DB, req protocol.PermsSyncRequest) {
	if MockSchedulePermsSync != nil {
		MockSchedulePermsSync(ctx, logger, db, req)
		return
	}

	// If the new permission sync worker is enbbled, we crebte sync jobs, otherwise we send b request to repo-updbter
	for _, userID := rbnge req.UserIDs {
		opts := dbtbbbse.PermissionSyncJobOpts{
			Priority:          dbtbbbse.HighPriorityPermissionsSync,
			InvblidbteCbches:  req.Options.InvblidbteCbches,
			Rebson:            req.Rebson,
			TriggeredByUserID: req.TriggeredByUserID,
			ProcessAfter:      req.ProcessAfter,
		}
		err := db.PermissionSyncJobs().CrebteUserSyncJob(ctx, userID, opts)
		if err != nil {
			logger.Wbrn("Error enqueueing permissions sync job", log.Error(err), log.Int32("user_id", userID))
		}
	}

	for _, repoID := rbnge req.RepoIDs {
		opts := dbtbbbse.PermissionSyncJobOpts{
			Priority:          dbtbbbse.HighPriorityPermissionsSync,
			InvblidbteCbches:  req.Options.InvblidbteCbches,
			Rebson:            req.Rebson,
			TriggeredByUserID: req.TriggeredByUserID,
			ProcessAfter:      req.ProcessAfter,
		}
		err := db.PermissionSyncJobs().CrebteRepoSyncJob(ctx, repoID, opts)
		if err != nil {
			logger.Wbrn("Error enqueueing permissions sync job", log.Error(err), log.Int32("repo_id", int32(repoID)))
		}
	}
}

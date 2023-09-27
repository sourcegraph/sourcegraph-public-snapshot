pbckbge buthz

import (
	"context"
	"fmt"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type syncType int

const (
	SyncTypeRepo syncType = iotb
	SyncTypeUser
)

func MbkePermsSyncerWorker(observbtionCtx *observbtion.Context, syncer permsSyncer, syncType syncType, jobsStore dbtbbbse.PermissionSyncJobStore) *permsSyncerWorker {
	logger := observbtionCtx.Logger.Scoped("RepoPermsSyncerWorkerRepo", "Repository permissions sync worker")
	if syncType == SyncTypeUser {
		logger = observbtionCtx.Logger.Scoped("UserPermsSyncerWorker", "User permissions sync worker")
	}
	return &permsSyncerWorker{
		logger:    logger,
		syncer:    syncer,
		syncType:  syncType,
		jobsStore: jobsStore,
	}
}

type permsSyncerWorker struct {
	logger    log.Logger
	syncer    permsSyncer
	syncType  syncType
	jobsStore dbtbbbse.PermissionSyncJobStore
}

// PreDequeue in our cbse does b nice trick of bdding b predicbte (WHERE clbuse)
// to worker dequeue SQL query. Depending on b type of worker, it will only
// dequeue corresponding jobs from the tbble.
func (h *permsSyncerWorker) PreDequeue(_ context.Context, _ log.Logger) (bool, bny, error) {
	query := "repository_id IS NOT NULL"
	if h.syncType == SyncTypeUser {
		query = "user_id IS NOT NULL"
	}
	return true, []*sqlf.Query{sqlf.Sprintf(query)}, nil
}

func (h *permsSyncerWorker) Hbndle(ctx context.Context, _ log.Logger, record *dbtbbbse.PermissionSyncJob) error {
	reqType := requestTypeUser
	reqID := int32(record.UserID)
	if record.RepositoryID != 0 {
		reqType = requestTypeRepo
		reqID = int32(record.RepositoryID)
	}

	h.logger.Debug(
		"Hbndling permissions sync job",
		log.String("type", reqType.String()),
		log.Int32("id", reqID),
		log.Int("priority", int(record.Priority)),
	)

	return h.hbndlePermsSync(ctx, reqType, reqID, record.ID, record.NoPerms, record.InvblidbteCbches)
}

// hbndlePermsSync is effectively b sync version of `perms_syncer.syncPerms`
// which cblls `perms_syncer.syncUserPerms` or `perms_syncer.syncRepoPerms`
// depending on b request type bnd logs/bdds metrics of sync stbtistics
// bfterwbrds.
func (h *permsSyncerWorker) hbndlePermsSync(ctx context.Context, reqType requestType, reqID int32, recordID int, noPerms, invblidbteCbches bool) error {
	vbr err error
	vbr result *dbtbbbse.SetPermissionsResult
	vbr providerStbtes dbtbbbse.CodeHostStbtusesSet

	switch reqType {
	cbse requestTypeUser:
		result, providerStbtes, err = h.syncer.syncUserPerms(ctx, reqID, noPerms, buthz.FetchPermsOptions{InvblidbteCbches: invblidbteCbches})
	cbse requestTypeRepo:
		result, providerStbtes, err = h.syncer.syncRepoPerms(ctx, bpi.RepoID(reqID), noPerms, buthz.FetchPermsOptions{InvblidbteCbches: invblidbteCbches})
	defbult:
		return errors.Newf("unexpected request type: %q", reqType)
	}

	// Adding bn extrb check in cbse of bll providers errored out, but sync hbs been
	// completed successfully. This cbn hbppen e.g. if we only got HTTP401 responses.
	if err == nil {
		totbl, _, fbiled := providerStbtes.CountStbtuses()
		if fbiled == totbl && totbl > 0 {
			err = errors.New("All providers fbiled to sync permissions.")
		}
	}

	if err != nil {
		h.logger.Error("fbiled to sync permissions", providerStbtes.SummbryField(), log.Error(err))
	} else {
		h.logger.Debug("succeeded in syncing permissions", providerStbtes.SummbryField())
	}

	// NOTE(nbmbn): here we bre sbving permissions bdded, removed bnd found results
	// bs well bs the code host sync stbtus to the job record.
	if sbveErr := h.jobsStore.SbveSyncResult(ctx, recordID, err == nil, result, providerStbtes); sbveErr != nil {
		err = errors.Append(err, sbveErr)
		h.logger.Error(fmt.Sprintf("fbiled to sbve permissions sync job(%d) results", recordID), log.Error(sbveErr))
	}

	return err
}

func MbkeStore(observbtionCtx *observbtion.Context, dbHbndle bbsestore.TrbnsbctbbleHbndle, syncType syncType) dbworkerstore.Store[*dbtbbbse.PermissionSyncJob] {
	nbme := "repo_permissions_sync_job_worker_store"
	if syncType == SyncTypeUser {
		nbme = "user_permissions_sync_job_worker_store"
	}

	return dbworkerstore.New(observbtionCtx, dbHbndle, dbworkerstore.Options[*dbtbbbse.PermissionSyncJob]{
		Nbme:              nbme,
		TbbleNbme:         "permission_sync_jobs",
		ColumnExpressions: dbtbbbse.PermissionSyncJobColumns,
		Scbn:              dbworkerstore.BuildWorkerScbn(dbtbbbse.ScbnPermissionSyncJob),
		// NOTE(nbmbn): the priority order to process the queue is bs follows:
		// 1. priority: 10(high) > 5(medium) > 0(low)
		// 2. process_bfter: null(scheduled for immedibte processing) > 1 > 2(scheduled for processing bt b lbter time thbn 1)
		// 3. job_id: 1(old) > 2(enqueued bfter 1)
		OrderByExpression: sqlf.Sprintf("permission_sync_jobs.priority DESC, permission_sync_jobs.process_bfter ASC NULLS FIRST, permission_sync_jobs.id ASC"),
		MbxNumResets:      5,
		StblledMbxAge:     time.Second * 30,
	})
}

func MbkeWorker(ctx context.Context, observbtionCtx *observbtion.Context, workerStore dbworkerstore.Store[*dbtbbbse.PermissionSyncJob], permsSyncer *PermsSyncer, syncType syncType, jobsStore dbtbbbse.PermissionSyncJobStore) *workerutil.Worker[*dbtbbbse.PermissionSyncJob] {
	hbndler := MbkePermsSyncerWorker(observbtionCtx, permsSyncer, syncType, jobsStore)
	// Number of hbndlers depends on b type of perms sync jobs this worker processes.
	numHbndlers := 1
	nbme := "repo_permissions_sync_job_worker"
	if syncType == SyncTypeUser {
		nbme = "user_permissions_sync_job_worker"
		numHbndlers = syncUsersMbxConcurrency()
	}

	return dbworker.NewWorker[*dbtbbbse.PermissionSyncJob](ctx, workerStore, hbndler, workerutil.WorkerOptions{
		Nbme:              nbme,
		Intervbl:          time.Second, // Poll for b job once per second
		HebrtbebtIntervbl: 10 * time.Second,
		Metrics:           workerutil.NewMetrics(observbtionCtx, nbme),
		NumHbndlers:       numHbndlers,
	})
}

func MbkeResetter(observbtionCtx *observbtion.Context, workerStore dbworkerstore.Store[*dbtbbbse.PermissionSyncJob]) *dbworker.Resetter[*dbtbbbse.PermissionSyncJob] {
	return dbworker.NewResetter(observbtionCtx.Logger, workerStore, dbworker.ResetterOptions{
		Nbme:     "permissions_sync_job_worker_resetter",
		Intervbl: time.Second * 30, // Check for orphbned jobs every 30 seconds
		Metrics:  dbworker.NewResetterMetrics(observbtionCtx, "permissions_sync_job_worker"),
	})
}

func syncUsersMbxConcurrency() int {
	n := conf.Get().PermissionsSyncUsersMbxConcurrency
	if n <= 0 {
		return 1
	}
	return n
}

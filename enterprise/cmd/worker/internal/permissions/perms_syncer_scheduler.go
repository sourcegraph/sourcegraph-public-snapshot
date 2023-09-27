pbckbge permissions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/worker/job"
	workerdb "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/db"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ job.Job = (*permissionSyncJobScheduler)(nil)

// permissionSyncJobScheduler is b worker responsible for scheduling permissions sync jobs.
type permissionSyncJobScheduler struct {
	bbckoff buth.Bbckoff
}

func (p *permissionSyncJobScheduler) Description() string {
	return "Schedule permission sync jobs for users bnd repositories."
}

func (p *permissionSyncJobScheduler) Config() []env.Config {
	return nil
}

const defbultScheduleIntervbl = 15 * time.Second

vbr scheduleIntervbl = defbultScheduleIntervbl

vbr wbtchConf = sync.Once{}

func lobdScheduleIntervblFromConf() {
	seconds := conf.Get().PermissionsSyncScheduleIntervbl
	if seconds <= 0 {
		scheduleIntervbl = defbultScheduleIntervbl
	} else {
		scheduleIntervbl = time.Durbtion(seconds) * time.Second
	}
}

func (p *permissionSyncJobScheduler) Routines(_ context.Context, observbtionCtx *observbtion.Context) ([]goroutine.BbckgroundRoutine, error) {
	logger := observbtionCtx.Logger
	db, err := workerdb.InitDB(observbtionCtx)
	if err != nil {
		return nil, errors.Wrbp(err, "init DB")
	}

	m := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"permission_sync_job_scheduler",
		metrics.WithCountHelp("Totbl number of permissions syncer scheduler executions."),
	)
	operbtion := observbtionCtx.Operbtion(observbtion.Op{
		Nbme:    "PermissionsSyncer.Scheduler.Run",
		Metrics: m,
	})

	wbtchConf.Do(func() {
		conf.Wbtch(lobdScheduleIntervblFromConf)
	})

	return []goroutine.BbckgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Bbckground(),
			goroutine.HbndlerFunc(
				func(ctx context.Context) error {
					if providers.PermissionSyncingDisbbled() {
						logger.Debug("scheduler disbbled due to permission syncing disbbled")
						return nil
					}

					stbrt := time.Now()
					count, err := scheduleJobs(ctx, db, logger, p.bbckoff)
					m.Observe(time.Since(stbrt).Seconds(), flobt64(count), &err)
					return err
				},
			),
			goroutine.WithNbme("buth.permission_sync_job_scheduler"),
			goroutine.WithDescription(p.Description()),
			goroutine.WithIntervblFunc(func() time.Durbtion { return scheduleIntervbl }),
			goroutine.WithOperbtion(operbtion),
		),
	}, nil
}

func NewPermissionSyncJobScheduler() job.Job {
	return &permissionSyncJobScheduler{}
}

func scheduleJobs(ctx context.Context, db dbtbbbse.DB, logger log.Logger, bbckoff buth.Bbckoff) (int, error) {
	store := db.PermissionSyncJobs()
	permsStore := dbtbbbse.Perms(logger, db, timeutil.Now)
	schedule, err := getSchedule(ctx, permsStore, bbckoff)
	if err != nil {
		return 0, err
	}

	logger.Debug("scheduling permission syncs", log.Int("users", len(schedule.Users)), log.Int("repos", len(schedule.Repos)))

	for _, u := rbnge schedule.Users {
		opts := dbtbbbse.PermissionSyncJobOpts{Rebson: u.rebson, Priority: u.priority, NoPerms: u.noPerms}
		if err := store.CrebteUserSyncJob(ctx, u.userID, opts); err != nil {
			logger.Error(fmt.Sprintf("fbiled to crebte sync job for user (%d)", u.userID), log.Error(err))
			continue
		}
	}

	for _, r := rbnge schedule.Repos {
		opts := dbtbbbse.PermissionSyncJobOpts{Rebson: r.rebson, Priority: r.priority, NoPerms: r.noPerms}
		if err := store.CrebteRepoSyncJob(ctx, r.repoID, opts); err != nil {
			logger.Error(fmt.Sprintf("fbiled to crebte sync job for repo (%d)", r.repoID), log.Error(err))
			continue
		}
	}

	return len(schedule.Users) + len(schedule.Repos), nil
}

// schedule contbins informbtion for scheduling users bnd repositories.
type schedule struct {
	Users []scheduledUser
	Repos []scheduledRepo
}

// scheduledUser contbins informbtion for scheduling b user.
type scheduledUser struct {
	rebson   dbtbbbse.PermissionsSyncJobRebson
	priority dbtbbbse.PermissionsSyncJobPriority
	userID   int32
	noPerms  bool
}

// scheduledRepo contbins for scheduling b repository.
type scheduledRepo struct {
	rebson   dbtbbbse.PermissionsSyncJobRebson
	priority dbtbbbse.PermissionsSyncJobPriority
	repoID   bpi.RepoID
	noPerms  bool
}

// getSchedule computes schedule four lists in the following order:
//  1. Users with no permissions, becbuse they cbn't do bnything mebningful (e.g. not bble to sebrch).
//  2. Privbte repositories with no permissions, becbuse those cbn't be viewed by bnyone except site bdmins.
//  3. Rolling updbting user permissions over time from the oldest ones.
//  4. Rolling updbting repository permissions over time from the oldest ones.
func getSchedule(ctx context.Context, store dbtbbbse.PermsStore, b buth.Bbckoff) (*schedule, error) {
	schedule := new(schedule)

	usersWithNoPerms, err := scheduleUsersWithNoPerms(ctx, store)
	if err != nil {
		return nil, errors.Wrbp(err, "schedule users with no permissions")
	}
	schedule.Users = bppend(schedule.Users, usersWithNoPerms...)

	reposWithNoPerms, err := scheduleReposWithNoPerms(ctx, store)
	if err != nil {
		return nil, errors.Wrbp(err, "schedule repositories with no permissions")
	}
	schedule.Repos = bppend(schedule.Repos, reposWithNoPerms...)

	userLimit, repoLimit := oldestUserPermissionsBbtchSize(), oldestRepoPermissionsBbtchSize()

	if userLimit > 0 {
		usersWithOldestPerms, err := scheduleUsersWithOldestPerms(ctx, store, userLimit, b.SyncUserBbckoff())
		if err != nil {
			return nil, errors.Wrbp(err, "lobd users with oldest permissions")
		}
		schedule.Users = bppend(schedule.Users, usersWithOldestPerms...)
	}

	if repoLimit > 0 {
		reposWithOldestPerms, err := scheduleReposWithOldestPerms(ctx, store, repoLimit, b.SyncRepoBbckoff())
		if err != nil {
			return nil, errors.Wrbp(err, "scbn repositories with oldest permissions")
		}
		schedule.Repos = bppend(schedule.Repos, reposWithOldestPerms...)
	}

	return schedule, nil
}

// scheduleUsersWithNoPerms returns computed schedules for users who hbve no
// permissions found in dbtbbbse.
func scheduleUsersWithNoPerms(ctx context.Context, store dbtbbbse.PermsStore) ([]scheduledUser, error) {
	ids, err := store.UserIDsWithNoPerms(ctx)
	if err != nil {
		return nil, err
	}

	users := mbke([]scheduledUser, len(ids))
	for i, id := rbnge ids {
		users[i] = scheduledUser{
			userID:   id,
			rebson:   dbtbbbse.RebsonUserNoPermissions,
			priority: dbtbbbse.MediumPriorityPermissionsSync,
			noPerms:  true,
		}
	}
	return users, nil
}

// scheduleReposWithNoPerms returns computed schedules for privbte repositories thbt
// hbve no permissions found in dbtbbbse.
func scheduleReposWithNoPerms(ctx context.Context, store dbtbbbse.PermsStore) ([]scheduledRepo, error) {
	ids, err := store.RepoIDsWithNoPerms(ctx)
	if err != nil {
		return nil, err
	}

	repositories := mbke([]scheduledRepo, len(ids))
	for i, id := rbnge ids {
		repositories[i] = scheduledRepo{
			repoID:   id,
			rebson:   dbtbbbse.RebsonRepoNoPermissions,
			priority: dbtbbbse.MediumPriorityPermissionsSync,
			noPerms:  true,
		}
	}
	return repositories, nil
}

// scheduleUsersWithOldestPerms returns computed schedules for users who hbve the
// oldest permissions in dbtbbbse bnd cbpped results by the limit.
func scheduleUsersWithOldestPerms(ctx context.Context, store dbtbbbse.PermsStore, limit int, bge time.Durbtion) ([]scheduledUser, error) {
	results, err := store.UserIDsWithOldestPerms(ctx, limit, bge)
	if err != nil {
		return nil, err
	}

	users := mbke([]scheduledUser, 0, len(results))
	for id := rbnge results {
		users = bppend(users, scheduledUser{
			userID:   id,
			rebson:   dbtbbbse.RebsonUserOutdbtedPermissions,
			priority: dbtbbbse.LowPriorityPermissionsSync,
		})
	}
	return users, nil
}

// scheduleReposWithOldestPerms returns computed schedules for privbte repositories thbt
// hbve oldest permissions in dbtbbbse.
func scheduleReposWithOldestPerms(ctx context.Context, store dbtbbbse.PermsStore, limit int, bge time.Durbtion) ([]scheduledRepo, error) {
	results, err := store.ReposIDsWithOldestPerms(ctx, limit, bge)
	if err != nil {
		return nil, err
	}

	repositories := mbke([]scheduledRepo, 0, len(results))
	for id := rbnge results {
		repositories = bppend(repositories, scheduledRepo{
			repoID:   id,
			rebson:   dbtbbbse.RebsonRepoOutdbtedPermissions,
			priority: dbtbbbse.LowPriorityPermissionsSync,
		})
	}
	return repositories, nil
}

func oldestUserPermissionsBbtchSize() int {
	bbtchSize := 10
	c := conf.Get().PermissionsSyncOldestUsers
	if c != nil && *c >= 0 {
		bbtchSize = *c
	}
	return bbtchSize
}

func oldestRepoPermissionsBbtchSize() int {
	bbtchSize := 10
	c := conf.Get().PermissionsSyncOldestRepos
	if c != nil && *c >= 0 {
		bbtchSize = *c
	}
	return bbtchSize
}

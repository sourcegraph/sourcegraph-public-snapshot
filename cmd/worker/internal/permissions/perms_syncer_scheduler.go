package permissions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz/providers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ job.Job = (*permissionSyncJobScheduler)(nil)

// permissionSyncJobScheduler is a worker responsible for scheduling permissions sync jobs.
type permissionSyncJobScheduler struct {
	backoff auth.Backoff
}

func (p *permissionSyncJobScheduler) Description() string {
	return "Schedule permission sync jobs for users and repositories."
}

func (p *permissionSyncJobScheduler) Config() []env.Config {
	return nil
}

const defaultScheduleInterval = 15 * time.Second

var scheduleInterval = defaultScheduleInterval

var watchConf = sync.Once{}

func loadScheduleIntervalFromConf() {
	seconds := conf.Get().PermissionsSyncScheduleInterval
	if seconds <= 0 {
		scheduleInterval = defaultScheduleInterval
	} else {
		scheduleInterval = time.Duration(seconds) * time.Second
	}
}

func (p *permissionSyncJobScheduler) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	logger := observationCtx.Logger
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, errors.Wrap(err, "init DB")
	}

	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"permission_sync_job_scheduler",
		metrics.WithCountHelp("Total number of permissions syncer scheduler executions."),
	)
	operation := observationCtx.Operation(observation.Op{
		Name:    "PermissionsSyncer.Scheduler.Run",
		Metrics: m,
	})

	watchConf.Do(func() {
		conf.Watch(loadScheduleIntervalFromConf)
	})

	return []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Background(),
			goroutine.HandlerFunc(
				func(ctx context.Context) error {
					if providers.PermissionSyncingDisabled() {
						logger.Debug("scheduler disabled due to permission syncing disabled")
						return nil
					}

					start := time.Now()
					count, err := scheduleJobs(ctx, db, logger, p.backoff)
					m.Observe(time.Since(start).Seconds(), float64(count), &err)
					return err
				},
			),
			goroutine.WithName("auth.permission_sync_job_scheduler"),
			goroutine.WithDescription(p.Description()),
			goroutine.WithIntervalFunc(func() time.Duration { return scheduleInterval }),
			goroutine.WithOperation(operation),
		),
	}, nil
}

func NewPermissionSyncJobScheduler() job.Job {
	return &permissionSyncJobScheduler{}
}

func scheduleJobs(ctx context.Context, db database.DB, logger log.Logger, backoff auth.Backoff) (int, error) {
	store := db.PermissionSyncJobs()
	permsStore := database.Perms(logger, db, timeutil.Now)
	schedule, err := getSchedule(ctx, permsStore, backoff)
	if err != nil {
		return 0, err
	}

	logger.Debug("scheduling permission syncs", log.Int("users", len(schedule.Users)), log.Int("repos", len(schedule.Repos)))

	for _, u := range schedule.Users {
		opts := database.PermissionSyncJobOpts{Reason: u.reason, Priority: u.priority, NoPerms: u.noPerms}
		if err := store.CreateUserSyncJob(ctx, u.userID, opts); err != nil {
			logger.Error(fmt.Sprintf("failed to create sync job for user (%d)", u.userID), log.Error(err))
			continue
		}
	}

	for _, r := range schedule.Repos {
		opts := database.PermissionSyncJobOpts{Reason: r.reason, Priority: r.priority, NoPerms: r.noPerms}
		if err := store.CreateRepoSyncJob(ctx, r.repoID, opts); err != nil {
			logger.Error(fmt.Sprintf("failed to create sync job for repo (%d)", r.repoID), log.Error(err))
			continue
		}
	}

	return len(schedule.Users) + len(schedule.Repos), nil
}

// schedule contains information for scheduling users and repositories.
type schedule struct {
	Users []scheduledUser
	Repos []scheduledRepo
}

// scheduledUser contains information for scheduling a user.
type scheduledUser struct {
	reason   database.PermissionsSyncJobReason
	priority database.PermissionsSyncJobPriority
	userID   int32
	noPerms  bool
}

// scheduledRepo contains for scheduling a repository.
type scheduledRepo struct {
	reason   database.PermissionsSyncJobReason
	priority database.PermissionsSyncJobPriority
	repoID   api.RepoID
	noPerms  bool
}

// getSchedule computes schedule four lists in the following order:
//  1. Users with no permissions, because they can't do anything meaningful (e.g. not able to search).
//  2. Private repositories with no permissions, because those can't be viewed by anyone except site admins.
//  3. Rolling updating user permissions over time from the oldest ones.
//  4. Rolling updating repository permissions over time from the oldest ones.
func getSchedule(ctx context.Context, store database.PermsStore, b auth.Backoff) (*schedule, error) {
	schedule := new(schedule)

	usersWithNoPerms, err := scheduleUsersWithNoPerms(ctx, store)
	if err != nil {
		return nil, errors.Wrap(err, "schedule users with no permissions")
	}
	schedule.Users = append(schedule.Users, usersWithNoPerms...)

	reposWithNoPerms, err := scheduleReposWithNoPerms(ctx, store)
	if err != nil {
		return nil, errors.Wrap(err, "schedule repositories with no permissions")
	}
	schedule.Repos = append(schedule.Repos, reposWithNoPerms...)

	userLimit, repoLimit := oldestUserPermissionsBatchSize(), oldestRepoPermissionsBatchSize()

	if userLimit > 0 {
		usersWithOldestPerms, err := scheduleUsersWithOldestPerms(ctx, store, userLimit, b.SyncUserBackoff())
		if err != nil {
			return nil, errors.Wrap(err, "load users with oldest permissions")
		}
		schedule.Users = append(schedule.Users, usersWithOldestPerms...)
	}

	if repoLimit > 0 {
		reposWithOldestPerms, err := scheduleReposWithOldestPerms(ctx, store, repoLimit, b.SyncRepoBackoff())
		if err != nil {
			return nil, errors.Wrap(err, "scan repositories with oldest permissions")
		}
		schedule.Repos = append(schedule.Repos, reposWithOldestPerms...)
	}

	return schedule, nil
}

// scheduleUsersWithNoPerms returns computed schedules for users who have no
// permissions found in database.
func scheduleUsersWithNoPerms(ctx context.Context, store database.PermsStore) ([]scheduledUser, error) {
	ids, err := store.UserIDsWithNoPerms(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]scheduledUser, len(ids))
	for i, id := range ids {
		users[i] = scheduledUser{
			userID:   id,
			reason:   database.ReasonUserNoPermissions,
			priority: database.MediumPriorityPermissionsSync,
			noPerms:  true,
		}
	}
	return users, nil
}

// scheduleReposWithNoPerms returns computed schedules for private repositories that
// have no permissions found in database.
func scheduleReposWithNoPerms(ctx context.Context, store database.PermsStore) ([]scheduledRepo, error) {
	ids, err := store.RepoIDsWithNoPerms(ctx)
	if err != nil {
		return nil, err
	}

	repositories := make([]scheduledRepo, len(ids))
	for i, id := range ids {
		repositories[i] = scheduledRepo{
			repoID:   id,
			reason:   database.ReasonRepoNoPermissions,
			priority: database.MediumPriorityPermissionsSync,
			noPerms:  true,
		}
	}
	return repositories, nil
}

// scheduleUsersWithOldestPerms returns computed schedules for users who have the
// oldest permissions in database and capped results by the limit.
func scheduleUsersWithOldestPerms(ctx context.Context, store database.PermsStore, limit int, age time.Duration) ([]scheduledUser, error) {
	results, err := store.UserIDsWithOldestPerms(ctx, limit, age)
	if err != nil {
		return nil, err
	}

	users := make([]scheduledUser, 0, len(results))
	for id := range results {
		users = append(users, scheduledUser{
			userID:   id,
			reason:   database.ReasonUserOutdatedPermissions,
			priority: database.LowPriorityPermissionsSync,
		})
	}
	return users, nil
}

// scheduleReposWithOldestPerms returns computed schedules for private repositories that
// have oldest permissions in database.
func scheduleReposWithOldestPerms(ctx context.Context, store database.PermsStore, limit int, age time.Duration) ([]scheduledRepo, error) {
	results, err := store.ReposIDsWithOldestPerms(ctx, limit, age)
	if err != nil {
		return nil, err
	}

	repositories := make([]scheduledRepo, 0, len(results))
	for id := range results {
		repositories = append(repositories, scheduledRepo{
			repoID:   id,
			reason:   database.ReasonRepoOutdatedPermissions,
			priority: database.LowPriorityPermissionsSync,
		})
	}
	return repositories, nil
}

func oldestUserPermissionsBatchSize() int {
	batchSize := 10
	c := conf.Get().PermissionsSyncOldestUsers
	if c != nil && *c >= 0 {
		batchSize = *c
	}
	return batchSize
}

func oldestRepoPermissionsBatchSize() int {
	batchSize := 10
	c := conf.Get().PermissionsSyncOldestRepos
	if c != nil && *c >= 0 {
		batchSize = *c
	}
	return batchSize
}

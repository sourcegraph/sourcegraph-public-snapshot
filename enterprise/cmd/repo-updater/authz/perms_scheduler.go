package authz

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"gopkg.in/inconshreveable/log15.v2"
)

// PermsScheduler is a permissions syncing scheduler that is in charge of
// keeping permissions up-to-date for users and repositories at best effort.
type PermsScheduler struct {
	// The time duration of how often to compute schedule for users and repositories.
	interval time.Duration
	// The database interface.
	db dbutil.DB
}

// NewPermsScheduler returns a new permissions syncing scheduler.
func NewPermsScheduler(interval time.Duration, db dbutil.DB) *PermsScheduler {
	return &PermsScheduler{
		interval: interval,
		db:       db,
	}
}

type scanResult struct {
	id   int32
	time time.Time
}

// TODO(jchen): Move this to authz.PermsStore.
func (s *PermsScheduler) scanIDsWithTime(ctx context.Context, q *sqlf.Query) ([]scanResult, error) {
	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []scanResult
	for rows.Next() {
		var id int32
		var t time.Time
		if err = rows.Scan(&id, &t); err != nil {
			return nil, err
		}

		results = append(results, scanResult{id: id, time: t})
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// scheduleUsersWithNoPerms returns computed schedules for users who have no permissions
// found in database.
func (s *PermsScheduler) scheduleUsersWithNoPerms(ctx context.Context) ([]ScheduledUser, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/repo-updater/authz/perms_scheduler.go:PermsScheduler.scheduleUsersWithNoPerms
SELECT users.id, '1970-01-01 00:00:00+00' FROM users
WHERE users.id NOT IN
	(SELECT perms.user_id FROM user_permissions AS perms)
`)
	results, err := s.scanIDsWithTime(ctx, q)
	if err != nil {
		return nil, err
	}

	users := make([]ScheduledUser, len(results))
	for i := range results {
		users[i] = ScheduledUser{
			Priority: PriorityHigh,
			UserID:   results[i].id,
		}
	}
	return users, nil
}

// scheduleReposWithNoPerms returns computed schedules for private repositories that
// have no permissions found in database.
func (s *PermsScheduler) scheduleReposWithNoPerms(ctx context.Context) ([]ScheduledRepo, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/repo-updater/authz/perms_scheduler.go:PermsScheduler.scheduleReposWithNoPerms
SELECT repo.id, '1970-01-01 00:00:00+00' FROM repo
WHERE repo.private = TRUE AND repo.id NOT IN
	(SELECT perms.repo_id FROM repo_permissions AS perms)
`)

	results, err := s.scanIDsWithTime(ctx, q)
	if err != nil {
		return nil, err
	}

	repos := make([]ScheduledRepo, len(results))
	for i := range results {
		repos[i] = ScheduledRepo{
			Priority: PriorityHigh,
			RepoID:   api.RepoID(results[i].id),
		}
	}
	return repos, nil
}

// scheduleUsersWithOldestPerms returns computed schedules for users who have oldest
// permissions in database and capped results by the limit.
func (s *PermsScheduler) scheduleUsersWithOldestPerms(ctx context.Context, limit int) ([]ScheduledUser, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/repo-updater/authz/perms_scheduler.go:PermsScheduler.scheduleUsersWithOldestPerms
SELECT user_id, updated_at FROM user_permissions
ORDER BY updated_at ASC
LIMIT %s
`, limit)

	results, err := s.scanIDsWithTime(ctx, q)
	if err != nil {
		return nil, err
	}

	users := make([]ScheduledUser, len(results))
	for i := range results {
		users[i] = ScheduledUser{
			Priority:      PriorityLow,
			UserID:        results[i].id,
			LastUpdatedAt: results[i].time,
		}
	}
	return users, nil
}

// scheduleReposWithOldestPerms returns computed schedules for private repositories that
// have oldest permissions in database.
func (s *PermsScheduler) scheduleReposWithOldestPerms(ctx context.Context, limit int) ([]ScheduledRepo, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/repo-updater/authz/perms_scheduler.go:PermsScheduler.scheduleReposWithOldestPerms
SELECT repo_id, updated_at FROM repo_permissions
ORDER BY updated_at ASC
LIMIT %s
`, limit)

	results, err := s.scanIDsWithTime(ctx, q)
	if err != nil {
		return nil, err
	}

	repos := make([]ScheduledRepo, len(results))
	for i := range results {
		repos[i] = ScheduledRepo{
			Priority:      PriorityLow,
			RepoID:        api.RepoID(results[i].id),
			LastUpdatedAt: results[i].time,
		}
	}
	return repos, nil
}

// Schedule contains information for scheduling users and repositories.
type Schedule struct {
	Users []ScheduledUser
	Repos []ScheduledRepo
}

// ScheduledRepo contains information for scheduling a user.
type ScheduledUser struct {
	Priority
	UserID        int32
	LastUpdatedAt time.Time
}

// ScheduledRepo contains for scheduling a repository.
type ScheduledRepo struct {
	Priority
	api.RepoID
	LastUpdatedAt time.Time
}

// schedule computes schedule four lists in the following order:
//   1. Users with no permissions, because they can't do anything meaningful (e.g. not able to search).
//   2. Private repositories with no permissions, because those can't be viewed by anyone except site admins.
//   3. Rolling updating user permissions over time from oldest ones.
//   4. Rolling updating repository permissions over time from oldest ones.
func (s *PermsScheduler) schedule(ctx context.Context) (*Schedule, error) {
	schedule := new(Schedule)

	users, err := s.scheduleUsersWithNoPerms(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "schedule users with no permissions")
	}
	schedule.Users = append(schedule.Users, users...)

	repos, err := s.scheduleReposWithNoPerms(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "schedule repositories with no permissions")
	}
	schedule.Repos = append(schedule.Repos, repos...)

	// TODO(jchen): Predict a threshold based on total repos and users that make sense to
	// finish syncing before next scans, so we don't waste database bandwidth.
	// Formula (in worse case scenario, at the pace of 1 req/s):
	//   initial threshold  = schdule interval in seconds
	//	 consumed by users = <initial threshold> / (<total repos> / <page size>)
	//   consumed by repos = (<initial threshold> - <consumed by users>) / (<total users> / <page size>)
	// Hard coded both to 100 for now.
	const threshold = 100

	users, err = s.scheduleUsersWithOldestPerms(ctx, threshold)
	if err != nil {
		return nil, errors.Wrap(err, "load users with oldest permissions")
	}
	schedule.Users = append(schedule.Users, users...)

	repos, err = s.scheduleReposWithOldestPerms(ctx, threshold)
	if err != nil {
		return nil, errors.Wrap(err, "scan repositories with oldest permissions")
	}
	schedule.Repos = append(schedule.Repos, repos...)

	return schedule, nil
}

// Sync kicks off the permissions syncing process, this method is blocking and
// should be called as a goroutine.
func Sync(ctx context.Context, scheduler *PermsScheduler, syncer *PermsSyncer) {
	go syncer.Run(ctx)

	log15.Debug("started perms scheduler")
	defer log15.Info("stopped perms scheduler")

	ticker := time.NewTicker(scheduler.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}

		schedule, err := scheduler.schedule(ctx)
		if err != nil {
			log15.Error("Failed to compute schedule", "err", err)
			continue
		}

		syncer.ScheduleUsers(ctx, schedule.Users...)
		syncer.ScheduleRepos(ctx, schedule.Repos...)
	}
}

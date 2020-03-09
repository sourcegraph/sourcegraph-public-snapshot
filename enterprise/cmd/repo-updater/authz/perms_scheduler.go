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
	// The time duration of how often a schedule happens.
	interval time.Duration
	// The database interface.
	store *store
}

// NewPermsScheduler returns a new permissions syncing scheduler.
func NewPermsScheduler(internal time.Duration, db dbutil.DB) *PermsScheduler {
	return &PermsScheduler{
		internal: internal,
		store:    &store{db: db},
	}
}

// scanUsersWithNoPerms returns a list of user IDs who have no permissions found in
// database.
func (s *PermsScheduler) scanUsersWithNoPerms(ctx context.Context) ([]int32, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/repo-updater/authz/perms_scheduler.go:PermsScheduler.scanUsersWithNoPerms
SELECT users.id FROM users
WHERE users.id NOT IN
	(SELECT perms.user_id FROM user_permissions AS perms)
`)
	return s.store.scanIDs(ctx, q)
}

// scanReposWithNoPerms returns a list of private repositories IDs that have no
// permissions found in database.
func (s *PermsScheduler) scanReposWithNoPerms(ctx context.Context) ([]api.RepoID, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/repo-updater/authz/perms_scheduler.go:PermsScheduler.scanReposWithNoPerms
SELECT repo.id FROM repo
WHERE repo.private = TRUE AND repo.id NOT IN
	(SELECT perms.repo_id FROM repo_permissions AS perms)
`)

	ids, err := s.store.scanIDs(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "scan IDs")
	}
	return toRepoIDs(ids), nil
}

// scanUsersWithOldestPerms returns a list of user IDs who have oldest permissions
// in database and capped results by the limit.
func (s *PermsScheduler) scanUsersWithOldestPerms(ctx context.Context, limit int) ([]int32, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/repo-updater/authz/perms_scheduler.go:PermsScheduler.scanUsersWithOldestPerms
SELECT user_id FROM user_permissions
ORDER BY updated_at ASC
LIMIT %s
`, limit)
	return s.store.scanIDs(ctx, q)
}

// scanReposWithOldestPerms returns a list of repository IDs that have oldest permissions
// in database and capped results by the limit.
func (s *PermsScheduler) scanReposWithOldestPerms(ctx context.Context, limit int) ([]api.RepoID, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/repo-updater/authz/perms_scheduler.go:PermsScheduler.scanReposWithOldestPerms
SELECT repo_id FROM repo_permissions
ORDER BY updated_at ASC
LIMIT %s
`, limit)

	ids, err := s.store.scanIDs(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "scan IDs")
	}
	return toRepoIDs(ids), nil
}

// toRepoIDs conerts a slice of int32 to []api.RepoID.
func toRepoIDs(ids []int32) []api.RepoID {
	repoIDs := make([]api.RepoID, len(ids))
	for i := range ids {
		repoIDs[i] = api.RepoID(ids[i])
	}
	return repoIDs
}

// schedule does scan and schedule four lists in the following order:
//   1. Users with no permissions, because they can't do anything meaningful (e.g. not able to search).
//   2. Private repositories with no permissions, because those can't be viewed by anyone except site admins.
//   3. Rolling updating user permissions over time from oldest ones.
//   4. Rolling updating repository permissions over time from oldest ones.
func (s *PermsScheduler) schedule(ctx context.Context, syncer *PermsSyncer) error {
	userIDs, err := s.scanUsersWithNoPerms(ctx)
	if err != nil {
		return errors.Wrap(err, "scan users with no permissions")
	}
	err = syncer.ScheduleUsers(ctx, PriorityHigh, userIDs...)
	if err != nil {
		return errors.Wrap(err, "schedule requests for users with no permissions")
	}

	repoIDs, err := s.scanReposWithNoPerms(ctx)
	if err != nil {
		return errors.Wrap(err, "scan repositories with no permissions")
	}
	err = syncer.ScheduleRepos(ctx, PriorityHigh, repoIDs...)
	if err != nil {
		return errors.Wrap(err, "schedule requests for repositories with no permissions")
	}

	// TODO(jchen): Predict a threshold based on total repos and users that make sense to
	// finish syncing before next scans, so we don't waste database bandwidth.
	// Formula (in worse case scenario, at the pace of 1 req/s):
	//   initial threshold  = schdule interval in seconds
	//	 consumed by users = <initial threshold> / (<total repos> / <page size>)
	//   consumed by repos = (<initial threshold> - <consumed by users>) / (<total users> / <page size>)
	// Hard coded both to 100 for now.
	const threshold = 100

	userIDs, err = s.scanUsersWithOldestPerms(ctx, threshold)
	if err != nil {
		return errors.Wrap(err, "scan users with oldest permissions")
	}
	err = syncer.ScheduleUsers(ctx, PriorityLow, userIDs...)
	if err != nil {
		return errors.Wrap(err, "schedule requests for users with oldest permissions")
	}

	repoIDs, err = s.scanReposWithOldestPerms(ctx, threshold)
	if err != nil {
		return errors.Wrap(err, "scan repositories with oldest permissions")
	}
	err = syncer.ScheduleRepos(ctx, PriorityLow, repoIDs...)
	if err != nil {
		return errors.Wrap(err, "schedule requests for repositories with oldest permissions")
	}

	return nil
}

// StartPermsSyncing kicks off the permissions syncing process, this method is blocking
// and should be called as a goroutine.
func Sync(ctx context.Context, scheduler *PermsScheduler, syncer *PermsSyncer) {
	go syncer.Run(ctx)

	log15.Debug("started perms scheduler")
	defer log15.Info("stopped perms scheduler")

	ticker := time.NewTicker(scheduler.internal)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}

		err := scheduler.schedule(ctx, syncer)
		if err != nil {
			log15.Error("Failed to schedule permissions syncing", "err", err)
			continue
		}
	}
}

package authz

import (
	"context"
	"fmt"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	edb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"gopkg.in/inconshreveable/log15.v2"
)

// PermsSyncer is a permissions syncing manager that is in charge of keeping
// permissions up-to-date for users and repositories.
//
// It is meant to be running in the background.
type PermsSyncer struct {
	// The priority queue to maintain the permissions syncing requests.
	queue *requestQueue
	// The database interface for any repos and external services operations.
	// TODO(jchen): Move all DB calls to authz.PermsStore and remove this field.
	reposStore repos.Store
	// The database interface for any permissions operations.
	permsStore *edb.PermsStore
	// TODO(jchen): Move all DB calls to authz.PermsStore and remove this field.
	db dbutil.DB
	// The mockable function to return the current time.
	clock func() time.Time
	// The time duration of how often to re-compute schedule for users and repositories.
	scheduleInterval time.Duration
}

// PermsFetcher is an authz.Provider that could also fetch permissions in both
// user-centric and repository-centric ways.
type PermsFetcher interface {
	authz.Provider
	// FetchUserPerms returns a list of repository IDs (on code host) that the given
	// account has read access on the code host. The repository ID should be the same
	// value as it would be used as api.ExternalRepoSpec.ID. The returned list should
	// only include private repositories.
	FetchUserPerms(ctx context.Context, account *extsvc.ExternalAccount) ([]string, error)
	// FetchRepoPerms returns a list of user IDs (on code host) who have read ccess to
	// the given repository on the code host. The user ID should be the same value as it
	// would be used as extsvc.ExternalAccount.AccountID. The returned list should include
	// both direct access and inherited from the group/organization/team membership.
	FetchRepoPerms(ctx context.Context, repo *api.ExternalRepoSpec) ([]string, error)
}

// NewPermsSyncer returns a new permissions syncing manager.
func NewPermsSyncer(
	reposStore repos.Store,
	permsStore *edb.PermsStore,
	db dbutil.DB,
	clock func() time.Time,
) *PermsSyncer {
	return &PermsSyncer{
		queue:            newRequestQueue(),
		reposStore:       reposStore,
		permsStore:       permsStore,
		db:               db,
		clock:            clock,
		scheduleInterval: 10 * time.Minute,
	}
}

// ScheduleUsers schedules new permissions syncing requests for given users
// in desired priority.
func (s *PermsSyncer) ScheduleUsers(ctx context.Context, users ...ScheduledUser) {
	for i := range users {
		updated := s.queue.enqueue(&requestMeta{
			priority:   users[i].Priority,
			typ:        requestTypeUser,
			id:         users[i].UserID,
			nextSyncAt: users[i].NextSyncAt,
		})
		log15.Debug("PermsSyncer.queue.enqueued", "userID", users[i].UserID, "updated", updated)
	}
}

// ScheduleRepos schedules new permissions syncing requests for given repositories
// in desired priority.
func (s *PermsSyncer) ScheduleRepos(ctx context.Context, repos ...ScheduledRepo) {
	for i := range repos {
		updated := s.queue.enqueue(&requestMeta{
			priority:   repos[i].Priority,
			typ:        requestTypeRepo,
			id:         int32(repos[i].RepoID),
			nextSyncAt: repos[i].NextSyncAt,
		})
		log15.Debug("PermsSyncer.queue.enqueued", "repoID", repos[i].RepoID, "updated", updated)
	}
}

// fetchers returns a list of authz.Provider that also implemented PermsFetcher.
// Keys are ServiceID (e.g. https://gitlab.com/). The current approach is to
// minimize the changes required by keeping the authz.Provider interface as-is,
// so each authz provider could be opt-in progressively until we fully complete
// the transition of moving permissions syncing process to the background for
// all authz providers.
func (s *PermsSyncer) fetchers() map[string]PermsFetcher {
	_, providers := authz.GetProviders()
	fetchers := make(map[string]PermsFetcher, len(providers))

	for i := range providers {
		if f, ok := providers[i].(PermsFetcher); ok {
			fetchers[f.ServiceID()] = f
		}
	}
	return fetchers
}

// syncUserPerms processes permissions syncing request in user-centric way.
func (s *PermsSyncer) syncUserPerms(ctx context.Context, userID int32) error {
	// TODO(jchen): Remove the use of dbconn.Global().
	accts, err := db.ExternalAccounts.List(ctx, db.ExternalAccountsListOptions{
		UserID: userID,
	})
	if err != nil {
		return errors.Wrap(err, "list external accounts")
	}

	var repoSpecs []api.ExternalRepoSpec
	for _, acct := range accts {
		fetcher := s.fetchers()[acct.ServiceID]
		if fetcher == nil {
			// We have no authz provider configured for this external account.
			continue
		}

		extIDs, err := fetcher.FetchUserPerms(ctx, acct)
		if err != nil {
			return errors.Wrap(err, "fetch user permissions")
		}

		for i := range extIDs {
			repoSpecs = append(repoSpecs, api.ExternalRepoSpec{
				ID:          extIDs[i],
				ServiceType: fetcher.ServiceType(),
				ServiceID:   fetcher.ServiceID(),
			})
		}
	}

	// Get corresponding internal database IDs
	rs, err := s.reposStore.ListRepos(ctx, repos.StoreListReposArgs{
		ExternalRepos: repoSpecs,
		PerPage:       int64(len(repoSpecs)), // We want to get all repositories in one shot
	})
	if err != nil {
		return errors.Wrap(err, "list external repositories")
	}

	// Save permissions to database
	p := &authz.UserPermissions{
		UserID: userID,
		Perm:   authz.Read, // Note: We currently only support read for repository permissions.
		Type:   authz.PermRepos,
		IDs:    roaring.NewBitmap(),
	}
	for i := range rs {
		p.IDs.Add(uint32(rs[i].ID))
	}

	err = s.permsStore.SetUserPermissions(ctx, p)
	if err != nil {
		return errors.Wrap(err, "set user permissions")
	}

	return nil
}

// syncRepoPerms processes permissions syncing request in repository-centric way.
// It discards requests that are made for non-private repositories based on the
// value of "repo.private" column.
func (s *PermsSyncer) syncRepoPerms(ctx context.Context, repoID api.RepoID) error {
	rs, err := s.reposStore.ListRepos(ctx, repos.StoreListReposArgs{
		IDs: []api.RepoID{repoID},
	})
	if err != nil {
		return errors.Wrap(err, "list repositories")
	} else if len(rs) == 0 {
		return nil
	}

	repo := rs[0]
	if !repo.Private {
		return nil
	}

	fetcher := s.fetchers()[repo.ExternalRepo.ServiceID]
	if fetcher == nil {
		// We have no authz provider configured for this repository.
		return nil
	}

	// NOTE: The following logic is based on the assumption that we have accurate
	// one-to-one username mapping between the internal database and the code host.
	// See last paragraph of https://docs.sourcegraph.com/admin/auth#username-normalization
	// for details.
	// TODO(jchen): Ship the initial design to unblock working on authz providers,
	// but should revisit the feasibility of using ExternalAccount before final delivery.

	usernames, err := fetcher.FetchRepoPerms(ctx, &repo.ExternalRepo)
	if err != nil {
		return errors.Wrap(err, "fetch repository permissions")
	}

	// Get corresponding internal database IDs
	// TODO(jchen): Remove the use of dbconn.Global().
	users, err := db.Users.GetByUsernames(ctx, usernames...)
	if err != nil {
		return errors.Wrap(err, "get users by usernames")
	}

	// Set up set of all usernames that need to be bound to permissions
	bindUsernamesSet := make(map[string]struct{}, len(usernames))
	for i := range usernames {
		bindUsernamesSet[usernames[i]] = struct{}{}
	}

	// Save permissions to database
	p := &authz.RepoPermissions{
		RepoID:  int32(repoID),
		Perm:    authz.Read, // Note: We currently only support read for repository permissions.
		UserIDs: roaring.NewBitmap(),
	}

	for i := range users {
		// Add existing user to permissions
		p.UserIDs.Add(uint32(users[i].ID))

		// Remove existing user from set of pending users
		delete(bindUsernamesSet, users[i].Username)
	}

	pendingBindUsernames := make([]string, 0, len(bindUsernamesSet))
	for id := range bindUsernamesSet {
		pendingBindUsernames = append(pendingBindUsernames, id)
	}

	txs, err := s.permsStore.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer txs.Done(&err)

	if err = txs.SetRepoPermissions(ctx, p); err != nil {
		return errors.Wrap(err, "set repository permissions")
	} else if err = txs.SetRepoPendingPermissions(ctx, pendingBindUsernames, p); err != nil {
		return errors.Wrap(err, "set repository pending permissions")
	}

	return nil
}

// syncPerms processes the permissions syncing request and remove the request from
// the quque once it is done (independent of success or failure).
func (s *PermsSyncer) syncPerms(ctx context.Context, request *syncRequest) error {
	defer s.queue.remove(request.typ, request.id, true)

	var err error
	switch request.typ {
	case requestTypeUser:
		err = s.syncUserPerms(ctx, request.id)
	case requestTypeRepo:
		err = s.syncRepoPerms(ctx, api.RepoID(request.id))
	default:
		err = fmt.Errorf("unexpected request type: %v", request.typ)
	}

	return err
}

func (s *PermsSyncer) runSync(ctx context.Context) {
	log15.Debug("PermsSyncer.runSync.started")
	defer log15.Info("PermsSyncer.runSync.stopped")

	// To unblock the "select" on the next loop iteration if no enqueue happened in between.
	notifyDequeued := make(chan struct{}, 1)
	for {
		select {
		case <-notifyDequeued:
		case <-s.queue.notifyEnqueue:
		case <-ctx.Done():
			return
		}

		request := s.queue.acquireNext()
		if request == nil {
			// No waiting request is in the queue
			continue
		}

		// Check if it's the time to sync the request
		if wait := request.nextSyncAt.Sub(s.clock()); wait > 0 {
			time.AfterFunc(wait, func() {
				notify(s.queue.notifyEnqueue)
			})

			log15.Debug("PermsSyncer.Run.waitForNextSync", "duration", wait)
			continue
		}

		notify(notifyDequeued)

		err := s.syncPerms(ctx, request)
		if err != nil {
			log15.Warn("Failed to sync permissions", "type", request.typ, "id", request.id, "err", err)
			continue
		}
	}
}

type scanResult struct {
	id   int32
	time time.Time
}

// TODO(jchen): Move this to authz.PermsStore.
func (s *PermsSyncer) loadIDsWithTime(ctx context.Context, q *sqlf.Query) ([]scanResult, error) {
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
// TODO(jchen): Move this to authz.PermsStore.
func (s *PermsSyncer) scheduleUsersWithNoPerms(ctx context.Context) ([]ScheduledUser, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/repo-updater/authz/perms_scheduler.go:PermsScheduler.scheduleUsersWithNoPerms
SELECT users.id, '1970-01-01 00:00:00+00' FROM users
WHERE users.id NOT IN
	(SELECT perms.user_id FROM user_permissions AS perms)
`)
	results, err := s.loadIDsWithTime(ctx, q)
	if err != nil {
		return nil, err
	}

	users := make([]ScheduledUser, len(results))
	for i := range results {
		users[i] = ScheduledUser{
			Priority: PriorityLow,
			UserID:   results[i].id,
			// NOTE: Have NextSyncAt with zero value (i.e. not set) gives it higher priority.
		}
	}
	return users, nil
}

// scheduleReposWithNoPerms returns computed schedules for private repositories that
// have no permissions found in database.
// TODO(jchen): Move this to authz.PermsStore.
func (s *PermsSyncer) scheduleReposWithNoPerms(ctx context.Context) ([]ScheduledRepo, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/repo-updater/authz/perms_scheduler.go:PermsScheduler.scheduleReposWithNoPerms
SELECT repo.id, '1970-01-01 00:00:00+00' FROM repo
WHERE repo.private = TRUE AND repo.id NOT IN
	(SELECT perms.repo_id FROM repo_permissions AS perms)
`)

	results, err := s.loadIDsWithTime(ctx, q)
	if err != nil {
		return nil, err
	}

	repos := make([]ScheduledRepo, len(results))
	for i := range results {
		repos[i] = ScheduledRepo{
			Priority: PriorityLow,
			RepoID:   api.RepoID(results[i].id),
			// NOTE: Have NextSyncAt with zero value (i.e. not set) gives it higher priority.
		}
	}
	return repos, nil
}

// scheduleUsersWithOldestPerms returns computed schedules for users who have oldest
// permissions in database and capped results by the limit.
// TODO(jchen): Move this to authz.PermsStore.
func (s *PermsSyncer) scheduleUsersWithOldestPerms(ctx context.Context, limit int) ([]ScheduledUser, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/repo-updater/authz/perms_scheduler.go:PermsScheduler.scheduleUsersWithOldestPerms
SELECT user_id, updated_at FROM user_permissions
ORDER BY updated_at ASC
LIMIT %s
`, limit)

	results, err := s.loadIDsWithTime(ctx, q)
	if err != nil {
		return nil, err
	}

	users := make([]ScheduledUser, len(results))
	for i := range results {
		users[i] = ScheduledUser{
			Priority:   PriorityLow,
			UserID:     results[i].id,
			NextSyncAt: results[i].time,
		}
	}
	return users, nil
}

// scheduleReposWithOldestPerms returns computed schedules for private repositories that
// have oldest permissions in database.
// TODO(jchen): Move this to authz.PermsStore.
func (s *PermsSyncer) scheduleReposWithOldestPerms(ctx context.Context, limit int) ([]ScheduledRepo, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/repo-updater/authz/perms_scheduler.go:PermsScheduler.scheduleReposWithOldestPerms
SELECT repo_id, updated_at FROM repo_permissions
ORDER BY updated_at ASC
LIMIT %s
`, limit)

	results, err := s.loadIDsWithTime(ctx, q)
	if err != nil {
		return nil, err
	}

	repos := make([]ScheduledRepo, len(results))
	for i := range results {
		repos[i] = ScheduledRepo{
			Priority:   PriorityLow,
			RepoID:     api.RepoID(results[i].id),
			NextSyncAt: results[i].time,
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
	UserID     int32
	NextSyncAt time.Time
}

// ScheduledRepo contains for scheduling a repository.
type ScheduledRepo struct {
	Priority
	api.RepoID
	NextSyncAt time.Time
}

// schedule computes schedule four lists in the following order:
//   1. Users with no permissions, because they can't do anything meaningful (e.g. not able to search).
//   2. Private repositories with no permissions, because those can't be viewed by anyone except site admins.
//   3. Rolling updating user permissions over time from oldest ones.
//   4. Rolling updating repository permissions over time from oldest ones.
func (s *PermsSyncer) schedule(ctx context.Context) (*Schedule, error) {
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

	// TODO(jchen): Predict a limit taking account into:
	//   1. Based on total repos and users that make sense to finish syncing before
	//      next schedule call, so we don't waste database bandwidth.
	//   2. How we're doing in terms of rate limiting.
	// Formula (in worse case scenario, at the pace of 1 req/s):
	//   initial limit  = <predicted from the previous step>
	//	 consumed by users = <initial limit> / (<total repos> / <page size>)
	//   consumed by repos = (<initial limit> - <consumed by users>) / (<total users> / <page size>)
	// Hard coded both to 100 for now.
	const limit = 100

	// TODO(jchen): Use better heuristics for setting NexySyncAt, the initial version
	// just uses the value of LastUpdatedAt get from the perms tables.

	users, err = s.scheduleUsersWithOldestPerms(ctx, limit)
	if err != nil {
		return nil, errors.Wrap(err, "load users with oldest permissions")
	}
	schedule.Users = append(schedule.Users, users...)

	repos, err = s.scheduleReposWithOldestPerms(ctx, limit)
	if err != nil {
		return nil, errors.Wrap(err, "scan repositories with oldest permissions")
	}
	schedule.Repos = append(schedule.Repos, repos...)

	return schedule, nil
}

func (s *PermsSyncer) runSchedule(ctx context.Context) {
	log15.Debug("PermsSyncer.runSchedule.started")
	defer log15.Info("PermsSyncer.runSchedule.stopped")

	ticker := time.NewTicker(s.scheduleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}

		schedule, err := s.schedule(ctx)
		if err != nil {
			log15.Error("Failed to compute schedule", "err", err)
			continue
		}

		s.ScheduleUsers(ctx, schedule.Users...)
		s.ScheduleRepos(ctx, schedule.Repos...)
	}
}

// Run kicks off the permissions syncing process, this method is blocking and
// should be called as a goroutine.
func (s *PermsSyncer) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go s.runSync(ctx)
	go s.runSchedule(ctx)

	<-ctx.Done()
}

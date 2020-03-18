package authz

import (
	"container/heap"
	"context"
	"fmt"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	edb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/internal/api"
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
	// The mockable function to return the current time.
	clock func() time.Time
	// The time duration of how often to re-compute schedule for users and repositories.
	scheduleInterval time.Duration
}

// PermsFetcher is an authz.Provider that could also fetch permissions in both
// user-centric and repository-centric ways.
type PermsFetcher interface {
	authz.Provider
	// FetchUserPerms returns a list of repository/project IDs (on code host) that the
	// given account has read access on the code host. The repository ID should be the
	// same value as it would be used as api.ExternalRepoSpec.ID. The returned list
	// should only include private repositories/project IDs.
	//
	// Because permissions fetching APIs are often expensive, the implementation should
	// try to return partial but valid results in case of error, and it is up to callers
	// to decide whether to discard.
	FetchUserPerms(ctx context.Context, account *extsvc.ExternalAccount) ([]extsvc.ExternalRepoID, error)
	// FetchRepoPerms returns a list of user IDs (on code host) who have read ccess to
	// the given repository/project on the code host. The user ID should be the same value
	// as it would be used as extsvc.ExternalAccount.AccountID. The returned list should
	// include both direct access and inherited from the group/organization/team membership.
	//
	// Because permissions fetching APIs are often expensive, the implementation should
	// try to return partial but valid results in case of error, and it is up to callers
	// to decide whether to discard.
	FetchRepoPerms(ctx context.Context, repo *api.ExternalRepoSpec) ([]extsvc.ExternalAccountID, error)
}

// NewPermsSyncer returns a new permissions syncing manager.
func NewPermsSyncer(
	reposStore repos.Store,
	permsStore *edb.PermsStore,
	clock func() time.Time,
) *PermsSyncer {
	return &PermsSyncer{
		queue:            newRequestQueue(),
		reposStore:       reposStore,
		permsStore:       permsStore,
		clock:            clock,
		scheduleInterval: time.Minute,
	}
}

// ScheduleUsers schedules new permissions syncing requests for given users
// in desired priority.
//
// This method implements the authz.PermsSyncer in the OSS namespace.
func (s *PermsSyncer) ScheduleUsers(ctx context.Context, priority Priority, userIDs ...int32) {
	users := make([]scheduledUser, len(userIDs))
	for i := range userIDs {
		users[i] = scheduledUser{
			Priority: priority,
			UserID:   userIDs[i],
			// NOTE: Have NextSyncAt with zero value (i.e. not set) gives it higher priority,
			// as the request is most likely triggered by a user action from OSS namespace.
		}
	}

	s.scheduleUsers(ctx, users...)
}

func (s *PermsSyncer) scheduleUsers(ctx context.Context, users ...scheduledUser) {
	for i := range users {
		select {
		case <-ctx.Done():
			log15.Debug("PermsSyncer.scheduleUsers.canceled")
			return
		default:
		}

		updated := s.queue.enqueue(&requestMeta{
			Priority:   users[i].Priority,
			Type:       requestTypeUser,
			ID:         users[i].UserID,
			NextSyncAt: users[i].NextSyncAt,
		})
		log15.Debug("PermsSyncer.queue.enqueued", "userID", users[i].UserID, "updated", updated)
	}
}

// ScheduleRepos schedules new permissions syncing requests for given repositories
// in desired priority.
//
// This method implements the authz.PermsSyncer in the OSS namespace.
func (s *PermsSyncer) ScheduleRepos(ctx context.Context, priority Priority, repoIDs ...api.RepoID) {
	repos := make([]scheduledRepo, len(repoIDs))
	for i := range repoIDs {
		repos[i] = scheduledRepo{
			Priority: priority,
			RepoID:   repoIDs[i],
			// NOTE: Have NextSyncAt with zero value (i.e. not set) gives it higher priority,
			// as the request is most likely triggered by a user action from OSS namespace.
		}
	}

	s.scheduleRepos(ctx, repos...)
}

func (s *PermsSyncer) scheduleRepos(ctx context.Context, repos ...scheduledRepo) {
	for i := range repos {
		select {
		case <-ctx.Done():
			log15.Debug("PermsSyncer.scheduleRepos.canceled")
			return
		default:
		}

		updated := s.queue.enqueue(&requestMeta{
			Priority:   repos[i].Priority,
			Type:       requestTypeRepo,
			ID:         int32(repos[i].RepoID),
			NextSyncAt: repos[i].NextSyncAt,
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
	accts, err := s.permsStore.ListExternalAccounts(ctx, userID)
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
				ID:          string(extIDs[i]),
				ServiceType: fetcher.ServiceType(),
				ServiceID:   fetcher.ServiceID(),
			})
		}
	}

	var rs []*repos.Repo
	if len(repoSpecs) > 0 {
		// Get corresponding internal database IDs
		rs, err = s.reposStore.ListRepos(ctx, repos.StoreListReposArgs{
			ExternalRepos: repoSpecs,
			PerPage:       int64(len(repoSpecs)), // We want to get all repositories in one shot
		})
		if err != nil {
			return errors.Wrap(err, "list external repositories")
		}
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

	log15.Info("PermsSyncer.syncUserPerms.synced", "userID", userID)
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

	extAccountIDs, err := fetcher.FetchRepoPerms(ctx, &repo.ExternalRepo)
	if err != nil {
		return errors.Wrap(err, "fetch repository permissions")
	}

	pendingAccountIDsSet := make(map[string]struct{})
	var userIDs map[string]int32 // Account ID -> User ID
	if len(extAccountIDs) > 0 {
		accountIDs := make([]string, len(extAccountIDs))
		for i := range extAccountIDs {
			accountIDs[i] = string(extAccountIDs[i])
		}

		// Get corresponding internal database IDs
		userIDs, err = s.permsStore.GetUserIDsByExternalAccounts(ctx, &extsvc.ExternalAccounts{
			ServiceType: fetcher.ServiceType(),
			ServiceID:   fetcher.ServiceID(),
			AccountIDs:  accountIDs,
		})
		if err != nil {
			return errors.Wrap(err, "get user IDs by external accounts")
		}

		// Set up the set of all account IDs that need to be bound to permissions
		pendingAccountIDsSet = make(map[string]struct{}, len(accountIDs))
		for i := range accountIDs {
			pendingAccountIDsSet[accountIDs[i]] = struct{}{}
		}
	}

	// Save permissions to database
	p := &authz.RepoPermissions{
		RepoID:  int32(repoID),
		Perm:    authz.Read, // Note: We currently only support read for repository permissions.
		UserIDs: roaring.NewBitmap(),
	}

	for aid, uid := range userIDs {
		// Add existing user to permissions
		p.UserIDs.Add(uint32(uid))

		// Remove existing user from the set of pending users
		delete(pendingAccountIDsSet, aid)
	}

	pendingAccountIDs := make([]string, 0, len(pendingAccountIDsSet))
	for aid := range pendingAccountIDsSet {
		pendingAccountIDs = append(pendingAccountIDs, aid)
	}

	txs, err := s.permsStore.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer txs.Done(&err)

	accounts := &extsvc.ExternalAccounts{
		ServiceType: fetcher.ServiceType(),
		ServiceID:   fetcher.ServiceID(),
		AccountIDs:  pendingAccountIDs,
	}

	if err = txs.SetRepoPermissions(ctx, p); err != nil {
		return errors.Wrap(err, "set repository permissions")
	} else if err = txs.SetRepoPendingPermissions(ctx, accounts, p); err != nil {
		return errors.Wrap(err, "set repository pending permissions")
	}

	log15.Info("PermsSyncer.syncRepoPerms.synced", "repoID", repo.ID, "name", repo.Name)
	return nil
}

// syncPerms processes the permissions syncing request and remove the request from
// the queue once it is done (independent of success or failure).
func (s *PermsSyncer) syncPerms(ctx context.Context, request *syncRequest) error {
	defer s.queue.remove(request.Type, request.ID, true)

	var err error
	switch request.Type {
	case requestTypeUser:
		err = s.syncUserPerms(ctx, request.ID)
	case requestTypeRepo:
		err = s.syncRepoPerms(ctx, api.RepoID(request.ID))
	default:
		err = fmt.Errorf("unexpected request type: %v", request.Type)
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
		if wait := request.NextSyncAt.Sub(s.clock()); wait > 0 {
			s.queue.release(request.Type, request.ID)
			time.AfterFunc(wait, func() {
				notify(s.queue.notifyEnqueue)
			})

			log15.Debug("PermsSyncer.Run.waitForNextSync", "duration", wait)
			continue
		}

		notify(notifyDequeued)

		err := s.syncPerms(ctx, request)
		if err != nil {
			log15.Warn("Failed to sync permissions", "type", request.Type, "id", request.ID, "err", err)
			continue
		}
	}
}

// scheduleUsersWithNoPerms returns computed schedules for users who have no permissions
// found in database.
func (s *PermsSyncer) scheduleUsersWithNoPerms(ctx context.Context) ([]scheduledUser, error) {
	ids, err := s.permsStore.UserIDsWithNoPerms(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]scheduledUser, len(ids))
	for i, id := range ids {
		users[i] = scheduledUser{
			Priority: PriorityLow,
			UserID:   id,
			// NOTE: Have NextSyncAt with zero value (i.e. not set) gives it higher priority.
		}
	}
	return users, nil
}

// scheduleReposWithNoPerms returns computed schedules for private repositories that
// have no permissions found in database.
func (s *PermsSyncer) scheduleReposWithNoPerms(ctx context.Context) ([]scheduledRepo, error) {
	ids, err := s.permsStore.RepoIDsWithNoPerms(ctx)
	if err != nil {
		return nil, err
	}

	repos := make([]scheduledRepo, len(ids))
	for i, id := range ids {
		repos[i] = scheduledRepo{
			Priority: PriorityLow,
			RepoID:   id,
			// NOTE: Have NextSyncAt with zero value (i.e. not set) gives it higher priority.
		}
	}
	return repos, nil
}

// scheduleUsersWithOldestPerms returns computed schedules for users who have oldest
// permissions in database and capped results by the limit.
func (s *PermsSyncer) scheduleUsersWithOldestPerms(ctx context.Context, limit int) ([]scheduledUser, error) {
	results, err := s.permsStore.UserIDsWithOldestPerms(ctx, limit)
	if err != nil {
		return nil, err
	}

	users := make([]scheduledUser, 0, len(results))
	for id, t := range results {
		users = append(users, scheduledUser{
			Priority:   PriorityLow,
			UserID:     id,
			NextSyncAt: t,
		})
	}
	return users, nil
}

// scheduleReposWithOldestPerms returns computed schedules for private repositories that
// have oldest permissions in database.
func (s *PermsSyncer) scheduleReposWithOldestPerms(ctx context.Context, limit int) ([]scheduledRepo, error) {
	results, err := s.permsStore.ReposIDsWithOldestPerms(ctx, limit)
	if err != nil {
		return nil, err
	}

	repos := make([]scheduledRepo, 0, len(results))
	for id, t := range results {
		repos = append(repos, scheduledRepo{
			Priority:   PriorityLow,
			RepoID:     id,
			NextSyncAt: t,
		})
	}
	return repos, nil
}

// schedule contains information for scheduling users and repositories.
type schedule struct {
	Users []scheduledUser
	Repos []scheduledRepo
}

// scheduledUser contains information for scheduling a user.
type scheduledUser struct {
	Priority
	UserID     int32
	NextSyncAt time.Time
}

// scheduledRepo contains for scheduling a repository.
type scheduledRepo struct {
	Priority
	api.RepoID
	NextSyncAt time.Time
}

// schedule computes schedule four lists in the following order:
//   1. Users with no permissions, because they can't do anything meaningful (e.g. not able to search).
//   2. Private repositories with no permissions, because those can't be viewed by anyone except site admins.
//   3. Rolling updating user permissions over time from oldest ones.
//   4. Rolling updating repository permissions over time from oldest ones.
func (s *PermsSyncer) schedule(ctx context.Context) (*schedule, error) {
	schedule := new(schedule)

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
	// Hard coded both to 10 for now.
	const limit = 10

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

		if !globals.PermissionsBackgroundSync().Enabled {
			continue
		}

		schedule, err := s.schedule(ctx)
		if err != nil {
			log15.Error("Failed to compute schedule", "err", err)
			continue
		}

		s.scheduleUsers(ctx, schedule.Users...)
		s.scheduleRepos(ctx, schedule.Repos...)
	}
}

// DebugDump returns the state of the permissions syncer for debugging.
func (s *PermsSyncer) DebugDump() interface{} {
	type requestInfo struct {
		Meta     *requestMeta
		Acquired bool
	}
	data := struct {
		Name  string
		Size  int
		Queue []*requestInfo
	}{
		Name: "permissions",
	}

	queue := requestQueue{
		heap: make([]*syncRequest, len(s.queue.heap)),
	}

	s.queue.mu.RLock()
	defer s.queue.mu.RUnlock()

	for i, request := range s.queue.heap {
		// Copy the syncRequest as a value so that poping off the heap here won't
		// update the index value of the real heap, and we don't do a racy read on
		// the repo pointer which may change concurrently in the real heap.
		requestCopy := *request
		queue.heap[i] = &requestCopy
	}

	for len(queue.heap) > 0 {
		// Copy values of the syncRequest so that the requestMeta pointer
		// won't change concurrently after we release the lock.
		request := heap.Pop(&queue).(*syncRequest)
		data.Queue = append(data.Queue, &requestInfo{
			Meta: &requestMeta{
				Priority:   request.Priority,
				Type:       request.Type,
				ID:         request.ID,
				NextSyncAt: request.NextSyncAt,
			},
			Acquired: request.acquired,
		})
	}
	data.Size = len(data.Queue)

	return &data
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

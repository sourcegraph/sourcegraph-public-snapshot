package authz

import (
	"context"
	"fmt"
	"time"

	"github.com/RoaringBitmap/roaring"
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

// PermsSyncer is a permissions syncing request manager that is in charge of
// both accepting and processing those requests. It is meant to be running
// in the background.
type PermsSyncer struct {
	// The priority queue to maintain the permissions syncing requests.
	queue *requestQueue
	// fetchers is a list of authz.Provider implementations that also
	// implemented PermsFetcher. Keys are ServiceID (e.g. https://gitlab.com/).
	// TODO(jchen): Use conf.Watch to get up-to-date authz providers.
	// The current approach is to minimize the changes required by keeping
	// the authz.Provider interface as-is, so each authz provider could be
	// opt-in progressively until we fully complete the transition of moving
	// permissions syncing process to the background for all authz providers.
	fetchers map[string]PermsFetcher
	// The database interface for any repos and external services operations.
	reposStore repos.Store
	// The database interface for any permissions operations.
	permsStore *edb.PermsStore
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

// NewPermsSyncer returns a new permissions syncing request manager.
func NewPermsSyncer(fetchers map[string]PermsFetcher, reposStore repos.Store, db dbutil.DB, clock func() time.Time) *PermsSyncer {
	return &PermsSyncer{
		queue:      newRequestQueue(),
		fetchers:   fetchers,
		reposStore: reposStore,
		permsStore: edb.NewPermsStore(db, clock),
	}
}

// ScheduleUser schedules a new permissions syncing request for given user
// in desired priority.
func (s *PermsSyncer) ScheduleUser(ctx context.Context, priority Priority, userID int32) error {
	p := &authz.UserPermissions{
		UserID: userID,
		Perm:   authz.Read,
		Type:   authz.PermRepos,
	}
	err := s.permsStore.LoadUserPermissions(ctx, p)
	if err != nil && err != authz.ErrPermsNotFound {
		return errors.Wrap(err, "load user permissions")
	}

	// NOTE: It is OK to have p.UpdatedAt with zero value that gets higher priority in the queue.
	updated := s.queue.enqueue(&requestMeta{
		priority:    priority,
		typ:         requestTypeUser,
		id:          userID,
		lastUpdated: p.UpdatedAt,
	})
	log15.Debug("PermsSyncer.queue.enqueued", "userID", userID, "updated", updated)
	return nil
}

// ScheduleRepo schedules a new permissions syncing request for given repository
// in desired priority.
func (s *PermsSyncer) ScheduleRepo(ctx context.Context, priority Priority, repoID api.RepoID) error {
	p := &authz.RepoPermissions{
		RepoID: int32(repoID),
		Perm:   authz.Read,
	}
	err := s.permsStore.LoadRepoPermissions(ctx, p)
	if err != nil && err != authz.ErrPermsNotFound {
		return errors.Wrap(err, "load repo permissions")
	}

	// NOTE: It is OK to have p.UpdatedAt with zero value that gets higher priority in the queue.
	updated := s.queue.enqueue(&requestMeta{
		priority:    priority,
		typ:         requestTypeRepo,
		id:          int32(repoID),
		lastUpdated: p.UpdatedAt,
	})
	log15.Debug("PermsSyncer.queue.enqueued", "repoID", repoID, "updated", updated)
	return nil
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
		fetcher := s.fetchers[acct.ServiceID]
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

	fetcher := s.fetchers[repo.ExternalRepo.ServiceID]
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
func (s PermsSyncer) syncPerms(ctx context.Context, request *syncRequest) {
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

	if err != nil {
		log15.Warn("Error syncing permissions", "type", request.typ, "id", request.id, "err", err)
		return
	}
}

// RunPermsSyncer starts running the given syncer in the background.
func RunPermsSyncer(ctx context.Context, syncer *PermsSyncer) {
	log15.Debug("started perms syncer")
	defer log15.Info("stopped perms syncer")

	// To unblock the "select" on the next loop iteration if no enqueue happened in between.
	notifyDequeued := make(chan struct{}, 1)
	for {
		select {
		case <-notifyDequeued:
		case <-syncer.queue.notifyEnqueue:
		case <-ctx.Done():
			return
		}

		request := syncer.queue.acquireNext()
		if request == nil {
			// No waiting request is in the queue
			continue
		}

		syncer.syncPerms(ctx, request)
		notify(notifyDequeued)
	}
}

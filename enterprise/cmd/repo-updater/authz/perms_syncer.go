package authz

import (
	"context"
	"fmt"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	edb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/internal/actor"
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
	//implemented PermsFetcher. Keys are ServiceID (e.g. https://gitlab.com/).
	// TODO(jchen): Use conf.Watch to get up-to-date authz providers.
	// The current approach is to minimize the changes required by keeping
	// the authz.Provider interface as-is, so each authz provider could be
	// opt-in progressively until we fully complete the transition of moving
	// permissions syncing process to the background for all authz providers.
	fetchers map[string]PermsFetcher
	// The database interface.
	db dbutil.DB
	// The clock to mock time.
	clock func() time.Time
}

// PermsFetcher is an authz.Provider that could also fetch permissions in both
// user-centric and repository-centric ways.
type PermsFetcher interface {
	authz.Provider
	// FetchUserPerms returns a list of repository IDs (on code host) that the given
	// account has read access on the code host. The returned list should only include
	// private repositories.
	FetchUserPerms(ctx context.Context, account *extsvc.ExternalAccount) ([]string, error)
	// FetchRepoPerms returns a list of (code host) usernames who have read ccess to
	// the given repository on the code host. Including direct access and inherited
	// from the group/organization/team membership.
	FetchRepoPerms(ctx context.Context, repo *api.ExternalRepoSpec) ([]string, error)
}

// NewPermsSyncer returns a new permissions syncing request manager.
func NewPermsSyncer(fetchers map[string]PermsFetcher, db dbutil.DB, clock func() time.Time) *PermsSyncer {
	return &PermsSyncer{
		queue:    newRequestQueue(),
		fetchers: fetchers,
		db:       db,
		clock:    clock,
	}
}

// ScheduleUser schedules a new permissions syncing request for given user
// in desired priority.
func (s *PermsSyncer) ScheduleUser(userID int32, priority Priority) error {
	updated := s.queue.enqueue(requestTypeUser, userID, priority)
	log15.Debug("PermsSyncer.queue.enqueued", "userID", userID, "updated", updated)
	return nil
}

// ScheduleRepo schedules a new permissions syncing request for given repository
// in desired priority.
func (s *PermsSyncer) ScheduleRepo(repoID api.RepoID, priority Priority) error {
	updated := s.queue.enqueue(requestTypeRepo, int32(repoID), priority)
	log15.Debug("PermsSyncer.queue.enqueued", "repoID", repoID, "updated", updated)
	return nil
}

// syncUserPerms processes permissions syncing request in user-centric way.
func (s *PermsSyncer) syncUserPerms(ctx context.Context, userID int32) error {
	accts, err := db.ExternalAccounts.List(ctx, db.ExternalAccountsListOptions{
		UserID: userID,
	})
	if err != nil {
		return errors.Wrap(err, "list external accounts")
	}

	// Set internal actor to bypass checking permissions.
	ctx = actor.WithActor(ctx, &actor.Actor{Internal: true})
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

		// Get corresponding internal database IDs
		repos, err := db.Repos.GetByExternalIDs(ctx, fetcher.ServiceType(), fetcher.ServiceID(), extIDs...)
		if err != nil {
			return errors.Wrap(err, "get repositories by external IDs")
		}

		// Save permissions to database
		p := &authz.UserPermissions{
			UserID: userID,
			Perm:   authz.Read, // Note: We currently only support read for repository permissions.
			Type:   authz.PermRepos,
			IDs:    roaring.NewBitmap(),
		}
		for i := range repos {
			p.IDs.Add(uint32(repos[i].ID))
		}

		store := edb.NewPermsStore(s.db, s.clock)
		err = store.SetUserPermissions(ctx, p)
		if err != nil {
			return errors.Wrap(err, "set user permissions")
		}
	}

	return nil
}

// syncRepoPerms processes permissions syncing request in repository-centric way.
// It discards requests that are made for non-private repositories based on the
// value of "repo.private" column.
func (s *PermsSyncer) syncRepoPerms(ctx context.Context, repoID api.RepoID) error {
	repo, err := db.Repos.Get(ctx, repoID)
	if err != nil {
		return errors.Wrap(err, "get repositroy")
	}

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

	usernames, err := fetcher.FetchRepoPerms(ctx, &repo.ExternalRepo)
	if err != nil {
		return errors.Wrap(err, "fetch repository permissions")
	}

	// Get corresponding internal database IDs
	users, err := db.Users.GetByUsernames(ctx, usernames...)
	if err != nil {
		return errors.Wrap(err, "get users by usernames")
	}

	// Compute bind IDs for late-binding users
	bindIDSet := make(map[string]struct{}, len(usernames))
	for i := range usernames {
		bindIDSet[usernames[i]] = struct{}{}
	}
	for i := range users {
		delete(bindIDSet, users[i].Username)
	}
	pendingBindIDs := make([]string, 0, len(bindIDSet))
	for id := range bindIDSet {
		pendingBindIDs = append(pendingBindIDs, id)
	}

	// Save permissions to database
	p := &authz.RepoPermissions{
		RepoID:  int32(repoID),
		Perm:    authz.Read, // Note: We currently only support read for repository permissions.
		UserIDs: roaring.NewBitmap(),
	}
	for i := range users {
		p.UserIDs.Add(uint32(users[i].ID))
	}

	store := edb.NewPermsStore(s.db, s.clock)
	txs, err := store.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer txs.Done(&err)

	if err = txs.SetRepoPermissions(ctx, p); err != nil {
		return errors.Wrap(err, "set repository permissions")
	} else if err = txs.SetRepoPendingPermissions(ctx, pendingBindIDs, p); err != nil {
		return errors.Wrap(err, "set repository pending permissions")
	}

	return nil
}

// RunPermsSyncer starts running the given syncer in the background.
func RunPermsSyncer(ctx context.Context, syncer *PermsSyncer) {
	log15.Debug("started perms syncer")
	defer log15.Info("stopped perms syncer")

	for {
		select {
		case <-syncer.queue.notifyEnqueue:
		case <-ctx.Done():
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			request := syncer.queue.acquireNext()
			if request == nil {
				// No waiting request is in the queue
				break
			}

			syncPerms := func(ctx context.Context, request *syncRequest) {
				defer syncer.queue.remove(request.typ, request.id, true)

				var err error
				switch request.typ {
				case requestTypeUser:
					err = syncer.syncUserPerms(ctx, request.id)
				case requestTypeRepo:
					err = syncer.syncRepoPerms(ctx, api.RepoID(request.id))
				default:
					err = fmt.Errorf("unexpected request type: %v", request.typ)
				}

				if err != nil {
					log15.Warn("Error syncing permissions", "type", request.typ, "id", request.id, "err", err)
					return
				}
			}
			syncPerms(ctx, request)
		}
	}
}

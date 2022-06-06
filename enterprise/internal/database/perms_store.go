package database

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	ErrPermsUpdatedAtNotSet = errors.New("permissions UpdatedAt timestamp must be set")
	ErrPermsSyncedAtNotSet  = errors.New("permissions SyncedAt timestamp must be set")
)

// PermsStore is the unified interface for managing permissions in the database.
type PermsStore interface {
	basestore.ShareableStore
	With(other basestore.ShareableStore) PermsStore
	// Transact begins a new transaction and make a new PermsStore over it.
	Transact(ctx context.Context) (PermsStore, error)
	Done(err error) error

	// LoadUserPermissions loads stored user permissions into p. An ErrPermsNotFound
	// is returned when there are no valid permissions available.
	LoadUserPermissions(ctx context.Context, p *authz.UserPermissions) error
	// LoadRepoPermissions loads stored repository permissions into p. An
	// ErrPermsNotFound is returned when there are no valid permissions available.
	LoadRepoPermissions(ctx context.Context, p *authz.RepoPermissions) error
	// SetUserPermissions performs a full update for p, new object IDs found in p
	// will be upserted and object IDs no longer in p will be removed. This method
	// updates both `user_permissions` and `repo_permissions` tables.
	//
	// Example input:
	// &UserPermissions{
	//     UserID: 1,
	//     Perm: authz.Read,
	//     Type: authz.PermRepos,
	//     IDs: {1, 2},
	// }
	//
	// Table states for input:
	// 	"user_permissions":
	//   user_id | permission | object_type | object_ids_ints | updated_at | synced_at
	//  ---------+------------+-------------+-----------------+------------+-----------
	//         1 |       read |       repos |          {1, 2} |      NOW() |     NOW()
	//
	//  "repo_permissions":
	//   repo_id | permission | user_ids_ints | updated_at |  synced_at
	//  ---------+------------+---------------+------------+-------------
	//         1 |       read |           {1} |      NOW() | <Unchanged>
	//         2 |       read |           {1} |      NOW() | <Unchanged>
	SetUserPermissions(ctx context.Context, p *authz.UserPermissions) error
	// SetRepoPermissions performs a full update for p, new user IDs found in p will
	// be upserted and user IDs no longer in p will be removed. This method updates
	// both `user_permissions` and `repo_permissions` tables.
	//
	// This method starts its own transaction for update consistency if the caller hasn't started one already.
	//
	// Example input:
	//  &RepoPermissions{
	//      RepoID: 1,
	//      Perm: authz.Read,
	//      UserIDs: {1, 2},
	//  }
	//
	// Table states for input:
	// 	"user_permissions":
	//   user_id | permission | object_type | object_ids_ints | updated_at |  synced_at
	//  ---------+------------+-------------+-----------------+------------+-------------
	//         1 |       read |       repos |             {1} |      NOW() | <Unchanged>
	//         2 |       read |       repos |             {1} |      NOW() | <Unchanged>
	//
	//  "repo_permissions":
	//   repo_id | permission | user_ids_ints | updated_at | synced_at
	//  ---------+------------+---------------+------------+-----------
	//         1 |       read |        {1, 2} |      NOW() |     NOW()
	SetRepoPermissions(ctx context.Context, p *authz.RepoPermissions) error
	// SetRepoPermissionsUnrestricted sets the unrestricted on the repo_permissions
	// table for all the provided repos. Either all or non are updated. Passing a
	// non-existent id is a noop.
	SetRepoPermissionsUnrestricted(ctx context.Context, ids []int32, unrestricted bool) error
	// TouchRepoPermissions only updates the value of both `updated_at` and
	// `synced_at` columns of the `repo_permissions` table without modifying the
	// permissions bits. It inserts a new row when the row does not yet exist. The
	// use case is to trick the scheduler to skip the repository for syncing
	// permissions when we can't sync permissions for the repository (e.g. due to
	// insufficient permissions of the access token).
	TouchRepoPermissions(ctx context.Context, repoID int32) error
	// LoadUserPendingPermissions returns pending permissions found by given
	// parameters. An ErrPermsNotFound is returned when there are no pending
	// permissions available.
	LoadUserPendingPermissions(ctx context.Context, p *authz.UserPendingPermissions) error
	// SetRepoPendingPermissions performs a full update for p with given accounts,
	// new account IDs found will be upserted and account IDs no longer in AccountIDs
	// will be removed.
	//
	// This method updates both `user_pending_permissions` and
	// `repo_pending_permissions` tables.
	//
	// This method starts its own transaction for update consistency if the caller
	// hasn't started one already.
	//
	// Example input:
	//  &extsvc.Accounts{
	//      ServiceType: "sourcegraph",
	//      ServiceID:   "https://sourcegraph.com/",
	//      AccountIDs:  []string{"alice", "bob"},
	//  }
	//  &authz.RepoPermissions{
	//      RepoID: 1,
	//      Perm: authz.Read,
	//  }
	//
	// Table states for input:
	// 	"user_pending_permissions":
	//   id | service_type |        service_id        | bind_id | permission | object_type | object_ids_ints | updated_at
	//  ----+--------------+--------------------------+---------+------------+-------------+-----------------+-----------
	//    1 | sourcegraph  | https://sourcegraph.com/ |   alice |       read |       repos |             {1} | <DateTime>
	//    2 | sourcegraph  | https://sourcegraph.com/ |     bob |       read |       repos |             {1} | <DateTime>
	//
	//  "repo_pending_permissions":
	//   repo_id | permission | user_ids_ints | updated_at
	//  ---------+------------+---------------+------------
	//         1 |       read |        {1, 2} | <DateTime>
	SetRepoPendingPermissions(ctx context.Context, accounts *extsvc.Accounts, p *authz.RepoPermissions) error
	GrantPendingPermissions(ctx context.Context, userID int32, p *authz.UserPendingPermissions) error
	// ListPendingUsers returns a list of bind IDs who have pending permissions by
	// given service type and ID.
	ListPendingUsers(ctx context.Context, serviceType, serviceID string) (bindIDs []string, _ error)
	// DeleteAllUserPermissions deletes all rows with given user ID from the
	// "user_permissions" table, which effectively removes access to all repositories
	// for the user.
	DeleteAllUserPermissions(ctx context.Context, userID int32) error
	// DeleteAllUserPendingPermissions deletes all rows with given bind IDs from the
	// "user_pending_permissions" table. It accepts list of bind IDs because a user
	// has multiple bind IDs, e.g. username and email addresses.
	DeleteAllUserPendingPermissions(ctx context.Context, accounts *extsvc.Accounts) error
	// GetUserIDsByExternalAccounts returns all user IDs matched by given external
	// account specs. The returned set has mapping relation as "account ID -> user
	// ID". The number of results could be less than the candidate list due to some
	// users are not associated with any external account.
	GetUserIDsByExternalAccounts(ctx context.Context, accounts *extsvc.Accounts) (map[string]int32, error)
	// UserIDsWithNoPerms returns a list of user IDs with no permissions found in the
	// database.
	UserIDsWithNoPerms(ctx context.Context) ([]int32, error)
	// UserIDsWithOutdatedPerms returns a list of user IDs who have had repository
	// syncing from either user or organization code host connection (that the user
	// is a member of) after last permissions sync.
	UserIDsWithOutdatedPerms(ctx context.Context) (map[int32]time.Time, error)
	// RepoIDsWithNoPerms returns a list of private repository IDs with no
	// permissions found in the database.
	RepoIDsWithNoPerms(ctx context.Context) ([]api.RepoID, error)
	// UserIDsWithOldestPerms returns a list of user ID and last updated pairs for
	// users who have the least recent synced permissions in the database and capped
	// results by the limit. If a user's permissions have been recently synced, based
	// on "age" they are ignored.
	UserIDsWithOldestPerms(ctx context.Context, limit int, age time.Duration) (map[int32]time.Time, error)
	// ReposIDsWithOldestPerms returns a list of repository ID and last updated pairs
	// for repositories that have the least recent synced permissions in the database
	// and caps results by the limit. If a repo's permissions have been recently
	// synced, based on "age" they are ignored.
	ReposIDsWithOldestPerms(ctx context.Context, limit int, age time.Duration) (map[api.RepoID]time.Time, error)
	// UserIsMemberOfOrgHasCodeHostConnection returns true if the user is a member of
	// any organization that has added code host connection.
	UserIsMemberOfOrgHasCodeHostConnection(ctx context.Context, userID int32) (has bool, err error)
	// Metrics returns calculated metrics values by querying the database. The
	// "staleDur" argument indicates how long ago was the last update to be
	// considered as stale.
	Metrics(ctx context.Context, staleDur time.Duration) (*PermsMetrics, error)
}

// It is concurrency-safe and maintains data consistency over the 'user_permissions',
// 'repo_permissions', 'user_pending_permissions', and 'repo_pending_permissions' tables.
type permsStore struct {
	*basestore.Store

	clock func() time.Time
}

var _ PermsStore = (*permsStore)(nil)

// Perms returns a new PermsStore with given parameters.
func Perms(db dbutil.DB, clock func() time.Time) PermsStore {
	return perms(db, clock)
}

func perms(db dbutil.DB, clock func() time.Time) *permsStore {
	return &permsStore{Store: basestore.NewWithDB(db, sql.TxOptions{}), clock: clock}
}

func PermsWith(other basestore.ShareableStore, clock func() time.Time) PermsStore {
	return &permsStore{Store: basestore.NewWithHandle(other.Handle()), clock: clock}
}

func (s *permsStore) With(other basestore.ShareableStore) PermsStore {
	return &permsStore{Store: s.Store.With(other), clock: s.clock}
}

func (s *permsStore) Transact(ctx context.Context) (PermsStore, error) {
	return s.transact(ctx)
}

func (s *permsStore) transact(ctx context.Context) (*permsStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &permsStore{Store: txBase, clock: s.clock}, err
}

func (s *permsStore) Done(err error) error {
	return s.Store.Done(err)
}

func (s *permsStore) LoadUserPermissions(ctx context.Context, p *authz.UserPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadUserPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	ids, updatedAt, syncedAt, err := s.loadUserPermissions(ctx, p, "")
	if err != nil {
		return err
	}

	// Since this is the Permissions table and not pending permissions we still use bitmaps here
	p.IDs = make(map[int32]struct{}, len(ids))
	for _, id := range ids {
		p.IDs[id] = struct{}{}
	}

	p.UpdatedAt = updatedAt
	p.SyncedAt = syncedAt
	return nil
}

func (s *permsStore) LoadRepoPermissions(ctx context.Context, p *authz.RepoPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	ids, updatedAt, syncedAt, unrestricted, err := s.loadRepoPermissions(ctx, p, "")
	if err != nil {
		return err
	}
	// Since this is the Permissions table and not pending permissions we still use bitmaps here
	p.UserIDs = make(map[int32]struct{}, len(ids))
	for _, id := range ids {
		p.UserIDs[id] = struct{}{}
	}
	p.UpdatedAt = updatedAt
	p.SyncedAt = syncedAt
	p.Unrestricted = unrestricted
	return nil
}

func (s *permsStore) SetUserPermissions(ctx context.Context, p *authz.UserPermissions) (err error) {
	ctx, save := s.observe(ctx, "SetUserPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	// Open a transaction for update consistency.
	txs, err := s.transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = txs.Done(err) }()

	// Retrieve currently stored object IDs of this user.
	oldIDs := map[int32]struct{}{}
	ids, _, _, err := txs.loadUserPermissions(ctx, p, "FOR UPDATE")
	if err != nil {
		if err != authz.ErrPermsNotFound {
			return errors.Wrap(err, "load user permissions")
		}
	} else {
		oldIDs = sliceToSet(ids)
	}

	if p.IDs == nil {
		p.IDs = map[int32]struct{}{}
	}

	added, removed := computeDiff(oldIDs, p.IDs)

	// Iterating over maps doesn't guarantee order so we sort the slices to avoid doing unnecessary DB updates.
	sort.Slice(added, func(i, j int) bool { return added[i] < added[j] })
	sort.Slice(removed, func(i, j int) bool { return removed[i] < removed[j] })

	updatedAt := txs.clock()
	if len(added) != 0 || len(removed) != 0 {
		var (
			allAdded    = added
			allRemoved  = removed
			addQueue    = allAdded
			removeQueue = allRemoved
			hasNextPage = true
		)

		for hasNextPage {
			var page *upsertRepoPermissionsPage
			page, addQueue, removeQueue, hasNextPage = newUpsertRepoPermissionsPage(addQueue, removeQueue)

			if q, err := upsertRepoPermissionsBatchQuery(page, allAdded, []int32{p.UserID}, p.Perm, updatedAt); err != nil {
				return err
			} else if err = txs.execute(ctx, q); err != nil {
				return errors.Wrap(err, "execute upsert repo permissions batch query")
			}
		}
	}

	// NOTE: The permissions background syncing heuristics relies on SyncedAt column
	// to do rolling update, if we don't always update the value of the column regardless,
	// we will end up checking the same set of oldest but up-to-date rows in the table.
	p.UpdatedAt = updatedAt
	p.SyncedAt = updatedAt
	if q, err := upsertUserPermissionsQuery(p); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute upsert user permissions query")
	}

	return nil
}

// upsertUserPermissionsQuery upserts single row of user permissions, it does the
// same thing as upsertUserPermissionsBatchQuery but also updates "synced_at"
// column to the value of p.SyncedAt field.
func upsertUserPermissionsQuery(p *authz.UserPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/internal/database/perms_store.go:upsertUserPermissionsQuery
INSERT INTO user_permissions
  (user_id, permission, object_type, object_ids_ints, updated_at, synced_at)
VALUES
  (%s, %s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  user_permissions_perm_object_unique
DO UPDATE SET
  object_ids_ints = excluded.object_ids_ints,
  updated_at = excluded.updated_at,
  synced_at = excluded.synced_at
`

	if p.UpdatedAt.IsZero() {
		return nil, ErrPermsUpdatedAtNotSet
	} else if p.SyncedAt.IsZero() {
		return nil, ErrPermsSyncedAtNotSet
	}

	idsArray := make([]int32, 0, len(p.IDs))

	for id := range p.IDs {
		idsArray = append(idsArray, id)
	}
	return sqlf.Sprintf(
		format,
		p.UserID,
		p.Perm.String(),
		p.Type,
		pq.Array(idsArray),
		p.UpdatedAt.UTC(),
		p.SyncedAt.UTC(),
	), nil
}

func (s *permsStore) SetRepoPermissions(ctx context.Context, p *authz.RepoPermissions) (err error) {
	ctx, save := s.observe(ctx, "SetRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	var txs *permsStore
	if s.InTransaction() {
		txs = s
	} else {
		txs, err = s.transact(ctx)
		if err != nil {
			return err
		}
		defer func() { err = txs.Done(err) }()
	}

	// Retrieve currently stored user IDs of this repository.
	oldIDs := map[int32]struct{}{}
	ids, _, _, _, err := txs.loadRepoPermissions(ctx, p, "FOR UPDATE")
	if err != nil {
		if err != authz.ErrPermsNotFound {
			return errors.Wrap(err, "load repo permissions")
		}
	} else {
		oldIDs = sliceToSet(ids)
	}

	if p.UserIDs == nil {
		p.UserIDs = map[int32]struct{}{}
	}

	added, removed := computeDiff(oldIDs, p.UserIDs)

	// Iterating over maps doesn't guarantee order, so we sort the slices to avoid doing unnecessary DB updates.
	sort.Slice(added, func(i, j int) bool { return added[i] < added[j] })
	sort.Slice(removed, func(i, j int) bool { return removed[i] < removed[j] })

	updatedAt := txs.clock()
	if len(added) != 0 || len(removed) != 0 {
		if q, err := upsertUserPermissionsBatchQuery(added, removed, []int32{p.RepoID}, p.Perm, authz.PermRepos, updatedAt); err != nil {
			return err
		} else if err = txs.execute(ctx, q); err != nil {
			return errors.Wrap(err, "execute upsert user permissions batch query")
		}
	}

	// NOTE: The permissions background syncing heuristics relies on SyncedAt column
	// to do rolling update, if we don't always update the value of the column regardless,
	// we will end up checking the same set of oldest but up-to-date rows in the table.
	p.UpdatedAt = updatedAt
	p.SyncedAt = updatedAt
	if q, err := upsertRepoPermissionsQuery(p); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute upsert repo permissions query")
	}

	return nil
}

// upsertUserPermissionsBatchQuery composes a SQL query that does both addition (for `addedUserIDs`) and deletion (
// for `removedUserIDs`) of `objectIDs` using upsert.
func upsertUserPermissionsBatchQuery(addedUserIDs, removedUserIDs, objectIDs []int32, perm authz.Perms, permType authz.PermType, updatedAt time.Time) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/internal/database/perms_store.go:upsertUserPermissionsBatchQuery
INSERT INTO user_permissions
	(user_id, permission, object_type, object_ids_ints, updated_at)
VALUES
	%s
ON CONFLICT ON CONSTRAINT
	user_permissions_perm_object_unique
DO UPDATE SET
	object_ids_ints = CASE
		-- When the user is part of "addedUserIDs"
		WHEN user_permissions.user_id = ANY (%s) THEN
			user_permissions.object_ids_ints | excluded.object_ids_ints
		ELSE
			user_permissions.object_ids_ints - %s::INT[]
		END,
	updated_at = excluded.updated_at
`
	if updatedAt.IsZero() {
		return nil, ErrPermsUpdatedAtNotSet
	}

	items := make([]*sqlf.Query, 0, len(addedUserIDs)+len(removedUserIDs))
	for _, userID := range addedUserIDs {
		items = append(items, sqlf.Sprintf("(%s, %s, %s, %s, %s)",
			userID,
			perm.String(),
			permType,
			pq.Array(objectIDs),
			updatedAt.UTC(),
		))
	}
	for _, userID := range removedUserIDs {
		items = append(items, sqlf.Sprintf("(%s, %s, %s, %s, %s)",
			userID,
			perm.String(),
			permType,

			// NOTE: Rows from `removedUserIDs` are expected to exist, but in case they do not,
			// we need to set it with empty object IDs to be safe (it's a removal anyway).
			pq.Array([]int32{}),

			updatedAt.UTC(),
		))
	}

	return sqlf.Sprintf(
		format,
		sqlf.Join(items, ","),
		pq.Array(addedUserIDs),

		// NOTE: Because we use empty object IDs for `removedUserIDs`, we can't reuse "excluded.object_ids_ints"
		// and have to explicitly set what object IDs to be removed.
		pq.Array(objectIDs),
	), nil
}

func (s *permsStore) SetRepoPermissionsUnrestricted(ctx context.Context, ids []int32, unrestricted bool) error {
	if len(ids) == 0 {
		return nil
	}

	const format = `
UPDATE repo_permissions
SET unrestricted = %s
WHERE repo_id = ANY (%s::int[])
`
	q := sqlf.Sprintf(format, unrestricted, pq.Array(ids))

	return errors.Wrap(s.Exec(ctx, q), "setting unrestricted flag")
}

// upsertRepoPermissionsQuery upserts single row of repository permissions.
func upsertRepoPermissionsQuery(p *authz.RepoPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/internal/database/perms_store.go:upsertRepoPermissionsQuery
INSERT INTO repo_permissions
  (repo_id, permission, user_ids_ints, updated_at, synced_at, unrestricted)
VALUES
  (%s, %s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  repo_permissions_perm_unique
DO UPDATE SET
  user_ids_ints = excluded.user_ids_ints,
  updated_at = excluded.updated_at,
  synced_at = excluded.synced_at,
  unrestricted = excluded.unrestricted
`

	if p.UpdatedAt.IsZero() {
		return nil, ErrPermsUpdatedAtNotSet
	} else if p.SyncedAt.IsZero() {
		return nil, ErrPermsSyncedAtNotSet
	}

	userIDs := make([]int32, 0, len(p.UserIDs))
	for id := range p.UserIDs {
		userIDs = append(userIDs, id)
	}

	return sqlf.Sprintf(
		format,
		p.RepoID,
		p.Perm.String(),
		pq.Array(userIDs),
		p.UpdatedAt.UTC(),
		p.SyncedAt.UTC(),
		p.Unrestricted,
	), nil
}

// upsertRepoPendingPermissionsQuery
func upsertRepoPendingPermissionsQuery(p *authz.RepoPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/internal/database/perms_store.go:upsertRepoPendingPermissionsQuery
INSERT INTO repo_pending_permissions
  (repo_id, permission, user_ids_ints, updated_at)
VALUES
  (%s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  repo_pending_permissions_perm_unique
DO UPDATE SET
  user_ids_ints = excluded.user_ids_ints,
  updated_at = excluded.updated_at
`

	if p.UpdatedAt.IsZero() {
		return nil, ErrPermsUpdatedAtNotSet
	}

	userIDs := make([]int64, 0, len(p.PendingUserIDs))
	for key := range p.PendingUserIDs {
		userIDs = append(userIDs, key)
	}
	return sqlf.Sprintf(
		format,
		p.RepoID,
		p.Perm.String(),
		pq.Array(userIDs),
		p.UpdatedAt.UTC(),
	), nil
}

func (s *permsStore) TouchRepoPermissions(ctx context.Context, repoID int32) (err error) {
	ctx, save := s.observe(ctx, "TouchRepoPermissions", "")
	defer func() { save(&err, otlog.Int32("repoID", repoID)) }()

	touchedAt := s.clock().UTC()
	perm := authz.Read.String() // Note: We currently only support read for repository permissions.
	q := sqlf.Sprintf(`
-- source: enterprise/internal/database/perms_store.go:TouchRepoPermissions
INSERT INTO repo_permissions
	(repo_id, permission, updated_at, synced_at)
VALUES
  (%s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  repo_permissions_perm_unique
DO UPDATE SET
  updated_at = excluded.updated_at,
  synced_at = excluded.synced_at
`, repoID, perm, touchedAt, touchedAt)
	if err = s.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute upsert repo permissions query")
	}
	return nil
}

func (s *permsStore) LoadUserPendingPermissions(ctx context.Context, p *authz.UserPendingPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadUserPendingPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	id, ids, updatedAt, err := s.loadUserPendingPermissions(ctx, p, "")
	if err != nil {
		return err
	}
	p.ID = id
	p.IDs = make(map[int32]struct{}, len(ids))
	for _, id := range ids {
		p.IDs[id] = struct{}{}
	}

	p.UpdatedAt = updatedAt
	return nil
}

func (s *permsStore) SetRepoPendingPermissions(ctx context.Context, accounts *extsvc.Accounts, p *authz.RepoPermissions) (err error) {
	ctx, save := s.observe(ctx, "SetRepoPendingPermissions", "")
	defer func() { save(&err, append(p.TracingFields(), accounts.TracingFields()...)...) }()

	var txs *permsStore
	if s.InTransaction() {
		txs = s
	} else {
		txs, err = s.transact(ctx)
		if err != nil {
			return err
		}
		defer func() { err = txs.Done(err) }()
	}

	var q *sqlf.Query

	p.PendingUserIDs = map[int64]struct{}{}
	p.UserIDs = map[int32]struct{}{}

	// Insert rows for AccountIDs without one in the "user_pending_permissions"
	// table. The insert does not store any permission data but uses auto-increment
	// key to generate unique ID. This guarantees that rows of all AccountIDs exist
	// when getting user IDs in the next load query.
	updatedAt := txs.clock()
	p.UpdatedAt = updatedAt
	if len(accounts.AccountIDs) > 0 {
		// NOTE: The primary key of "user_pending_permissions" table is auto-incremented,
		// and it is monotonically growing even with upsert in Postgres (i.e. the primary
		// key is increased internally by one even if the row exists). This means with
		// large number of AccountIDs, the primary key will grow very quickly every time
		// we do an upsert, and not far from reaching the largest number an int8 (64-bit
		// integer) can hold (9,223,372,036,854,775,807). Therefore, load existing rows
		// would help us only upsert rows that are newly discovered. See NOTE in below
		// for why we do upsert not insert.
		q = loadExistingUserPendingPermissionsBatchQuery(accounts, p)
		bindIDsToIDs, err := txs.loadExistingUserPendingPermissionsBatch(ctx, q)
		if err != nil {
			return errors.Wrap(err, "loading existing user pending permissions")
		}

		missingAccounts := &extsvc.Accounts{
			ServiceType: accounts.ServiceType,
			ServiceID:   accounts.ServiceID,
			AccountIDs:  make([]string, 0, len(accounts.AccountIDs)-len(bindIDsToIDs)),
		}

		for _, bindID := range accounts.AccountIDs {
			id, ok := bindIDsToIDs[bindID]
			if ok {
				p.PendingUserIDs[id] = struct{}{}
			} else {
				missingAccounts.AccountIDs = append(missingAccounts.AccountIDs, bindID)
			}
		}

		// Only do upsert when there are missing accounts
		if len(missingAccounts.AccountIDs) > 0 {
			// NOTE: Row-level locking is not needed here because we're creating stub rows
			//  and not modifying permissions, which is also why it is best to use upsert not
			//  insert to avoid unique violation in case other transactions are trying to
			//  create overlapping stub rows.
			q, err = upsertUserPendingPermissionsBatchQuery(missingAccounts, p)
			if err != nil {
				return err
			}

			ids, err := txs.loadUserPendingPermissionsIDs(ctx, q)
			if err != nil {
				return errors.Wrap(err, "load user pending permissions IDs from upsert pending permissions")
			}

			// Make up p.PendingUserIDs from the result set.
			for _, id := range ids {
				p.PendingUserIDs[id] = struct{}{}
			}
		}

	}

	// Retrieve currently stored user IDs of this repository.
	_, ids, _, _, err := txs.loadRepoPendingPermissions(ctx, p, "FOR UPDATE")
	if err != nil && err != authz.ErrPermsNotFound {
		return errors.Wrap(err, "load repo pending permissions")
	}

	oldIDs := sliceToSet(ids)
	added, removed := computeDiff(oldIDs, p.PendingUserIDs)

	// In case there is nothing added or removed.
	if len(added) == 0 && len(removed) == 0 {
		return nil
	}

	if q, err = updateUserPendingPermissionsBatchQuery(added, removed, []int32{p.RepoID}, p.Perm, authz.PermRepos, updatedAt); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute append user pending permissions batch query")
	}

	if q, err = upsertRepoPendingPermissionsQuery(p); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute upsert repo pending permissions query")
	}

	return nil
}

func (s *permsStore) loadUserPendingPermissionsIDs(ctx context.Context, q *sqlf.Query) (ids []int64, err error) {
	ctx, save := s.observe(ctx, "loadUserPendingPermissionsIDs", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()

	rows, err := s.Query(ctx, q)
	return basestore.ScanInt64s(rows, err)
}

func (s *permsStore) loadExistingUserPendingPermissionsBatch(ctx context.Context, q *sqlf.Query) (bindIDsToIDs map[string]int64, err error) {
	ctx, save := s.observe(ctx, "loadExistingUserPendingPermissionsBatch", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	bindIDsToIDs = make(map[string]int64)
	for rows.Next() {
		var bindID string
		var id int64
		if err = rows.Scan(&bindID, &id); err != nil {
			return nil, err
		}
		bindIDsToIDs[bindID] = id
	}
	return bindIDsToIDs, nil
}

// upsertUserPendingPermissionsBatchQuery generates a query for upserting the provided
// external service accounts into `user_pending_permissions`.
func upsertUserPendingPermissionsBatchQuery(
	accounts *extsvc.Accounts,
	p *authz.RepoPermissions,
) (*sqlf.Query, error) {
	// Above ~10,000 accounts (10,000 * 6 fields each = 60,000 parameters), we can run
	// into the Postgres parameter limit inserting with VALUES. Instead, we pass in fields
	// as arrays, where each array only counts for a single parameter.
	//
	// If changing the parameters used in this query, make sure to run relevant tests
	// named `postgresParameterLimitTest` using "go test -slow-tests".
	const format = `
-- source: enterprise/internal/database/perms_store.go:upsertUserPendingPermissionsBatchQuery
INSERT INTO user_pending_permissions
	(service_type, service_id, bind_id, permission, object_type, updated_at)
	(
		SELECT %s::TEXT, %s::TEXT, UNNEST(%s::TEXT[]), %s::TEXT, %s::TEXT, %s::TIMESTAMPTZ
	)
ON CONFLICT ON CONSTRAINT
	user_pending_permissions_service_perm_object_unique
DO UPDATE SET
	updated_at = excluded.updated_at
RETURNING id
`
	if p.UpdatedAt.IsZero() {
		return nil, ErrPermsUpdatedAtNotSet
	}

	return sqlf.Sprintf(
		format,

		accounts.ServiceType,
		accounts.ServiceID,
		pq.Array(accounts.AccountIDs),

		p.Perm.String(),
		string(authz.PermRepos),
		p.UpdatedAt.UTC(),
	), nil
}

func loadExistingUserPendingPermissionsBatchQuery(accounts *extsvc.Accounts, p *authz.RepoPermissions) *sqlf.Query {
	const format = `
-- source: enterprise/internal/database/perms_store.go:loadExistingUserPendingPermissionsBatchQuery
SELECT bind_id, id FROM user_pending_permissions
WHERE
	service_type = %s
AND service_id = %s
AND permission = %s
AND object_type = %s
AND bind_id IN (%s)
`

	bindIDs := make([]*sqlf.Query, len(accounts.AccountIDs))
	for i := range accounts.AccountIDs {
		bindIDs[i] = sqlf.Sprintf("%s", accounts.AccountIDs[i])
	}

	return sqlf.Sprintf(
		format,
		accounts.ServiceType,
		accounts.ServiceID,
		p.Perm.String(),
		authz.PermRepos,
		sqlf.Join(bindIDs, ","),
	)
}

// updateUserPendingPermissionsBatchQuery composes a SQL query that does both addition (for `addedUserIDs`) and deletion (
// for `removedUserIDs`) of `objectIDs` using update.
func updateUserPendingPermissionsBatchQuery(addedUserIDs, removedUserIDs []int64, objectIDs []int32, perm authz.Perms, permType authz.PermType, updatedAt time.Time) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/internal/database/perms_store.go:updateUserPendingPermissionsBatchQuery
UPDATE user_pending_permissions
SET
	object_ids_ints = CASE
		-- When the user is part of "addedUserIDs"
		WHEN user_pending_permissions.id = ANY (%s) THEN
			user_pending_permissions.object_ids_ints | %s
		ELSE
			user_pending_permissions.object_ids_ints - %s
		END,
	updated_at = %s
WHERE
	id = ANY (%s)
AND permission = %s
AND object_type = %s
`
	if updatedAt.IsZero() {
		return nil, ErrPermsUpdatedAtNotSet
	}

	return sqlf.Sprintf(
		format,
		pq.Array(addedUserIDs),
		pq.Array(objectIDs),
		pq.Array(objectIDs),
		updatedAt.UTC(),
		pq.Array(append(addedUserIDs, removedUserIDs...)),
		perm.String(),
		permType,
	), nil
}

// GrantPendingPermissions is used to grant pending permissions when the associated "ServiceType",
// "ServiceID" and "BindID" found in p becomes effective for a given user, e.g. username as bind ID when
// a user is created, email as bind ID when the email address is verified.
//
// Because there could be multiple external services and bind IDs that are associated with a single user
// (e.g. same user on different code hosts, multiple email addresses), it merges data from "repo_pending_permissions"
// and "user_pending_permissions" tables to "repo_permissions" and "user_permissions" tables for the user.
//
// Therefore, permissions are unioned not replaced, which is one of the main differences from SetRepoPermissions
// and SetRepoPendingPermissions methods. Another main difference is that multiple calls to this method
// are not idempotent as it conceptually does nothing when there is no data in the pending permissions
// tables for the user.
//
// This method starts its own transaction for update consistency if the caller hasn't started one already.
//
// 🚨 SECURITY: This method takes arbitrary string as a valid bind ID and does not interpret the meaning
// of the value it represents. Therefore, it is caller's responsibility to ensure the legitimate relation
// between the given user ID and the bind ID found in p.
func (s *permsStore) GrantPendingPermissions(ctx context.Context, userID int32, p *authz.UserPendingPermissions) (err error) {
	ctx, save := s.observe(ctx, "GrantPendingPermissions", "")
	defer func() { save(&err, append(p.TracingFields(), otlog.Int32("userID", userID))...) }()

	var txs *permsStore
	if s.InTransaction() {
		txs = s
	} else {
		txs, err = s.transact(ctx)
		if err != nil {
			return err
		}
		defer func() { err = txs.Done(err) }()
	}

	id, ids, _, err := txs.loadUserPendingPermissions(ctx, p, "FOR UPDATE")
	if err != nil {
		// Skip the whole grant process if the user has no pending permissions.
		if err == authz.ErrPermsNotFound {
			return nil
		}
		return errors.Wrap(err, "load user pending permissions")
	}
	p.ID = id
	p.IDs = make(map[int32]struct{}, len(ids))
	allRepoIDs := make([]int32, 0, len(ids))
	for _, id := range ids {
		p.IDs[id] = struct{}{}
		allRepoIDs = append(allRepoIDs, id)
	}
	// NOTE: We currently only have "repos" type, so avoid unnecessary type checking for now.
	if len(allRepoIDs) == 0 {
		return nil
	}

	var (
		updatedAt  = txs.clock()
		allUserIDs = []int32{userID}

		addQueue    = allRepoIDs
		hasNextPage = true
	)
	for hasNextPage {
		var page *upsertRepoPermissionsPage
		page, addQueue, _, hasNextPage = newUpsertRepoPermissionsPage(addQueue, nil)

		if q, err := upsertRepoPermissionsBatchQuery(page, allRepoIDs, allUserIDs, p.Perm, updatedAt); err != nil {
			return err
		} else if err = txs.execute(ctx, q); err != nil {
			return errors.Wrap(err, "execute upsert repo permissions batch query")
		}
	}

	if q, err := upsertUserPermissionsBatchQuery(allUserIDs, nil, allRepoIDs, p.Perm, authz.PermRepos, updatedAt); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute upsert user permissions batch query")
	}

	// NOTE: Practically, we don't need to clean up "repo_pending_permissions" table because the value of "id" column
	// that is associated with this user will be invalidated automatically by deleting this row. Thus, we are able to
	// avoid database deadlocks with other methods (e.g. SetRepoPermissions, SetRepoPendingPermissions).
	if err = txs.execute(ctx, deleteUserPendingPermissionsQuery(p)); err != nil {
		return errors.Wrap(err, "execute delete user pending permissions query")
	}

	return nil
}

// upsertRepoPermissionsPage tracks entries to upsert in a upsertRepoPermissionsBatchQuery.
type upsertRepoPermissionsPage struct {
	addedRepoIDs   []int32
	removedRepoIDs []int32
}

// upsertRepoPermissionsPageSize restricts page size for newUpsertRepoPermissionsPage to
// stay within the Postgres parameter limit (see `defaultUpsertRepoPermissionsPageSize`).
//
// May be modified for testing.
var upsertRepoPermissionsPageSize = defaultUpsertRepoPermissionsPageSize

// defaultUpsertRepoPermissionsPageSize sets a default for upsertRepoPermissionsPageSize.
//
// Value set to avoid parameter limit of ~65k, because each page element counts for 4
// parameters (65k / 4 ~= 16k rows at a time)
const defaultUpsertRepoPermissionsPageSize = 15000

// newUpsertRepoPermissionsPage instantiates a page from the given add/remove queues.
// Callers should reassign their queues to the ones returned by this constructor.
func newUpsertRepoPermissionsPage(addQueue, removeQueue []int32) (
	page *upsertRepoPermissionsPage,
	newAddQueue, newRemoveQueue []int32,
	hasNextPage bool,
) {
	quota := upsertRepoPermissionsPageSize
	page = &upsertRepoPermissionsPage{}

	if len(addQueue) > 0 {
		if len(addQueue) < quota {
			page.addedRepoIDs = addQueue
			addQueue = nil
		} else {
			page.addedRepoIDs = addQueue[:quota]
			addQueue = addQueue[quota:]
		}
		quota -= len(page.addedRepoIDs)
	}

	if len(removeQueue) > 0 {
		if len(removeQueue) < quota {
			page.removedRepoIDs = removeQueue
			removeQueue = nil
		} else {
			page.removedRepoIDs = removeQueue[:quota]
			removeQueue = removeQueue[quota:]
		}
	}

	return page,
		addQueue,
		removeQueue,
		len(addQueue) > 0 || len(removeQueue) > 0
}

// upsertRepoPermissionsBatchQuery composes a SQL query that does both addition (for `addedRepoIDs`)
// and deletion (for `removedRepoIDs`) of `userIDs` using upsert.
//
// Pages should be set up using the helper function `newUpsertRepoPermissionsPage`
func upsertRepoPermissionsBatchQuery(page *upsertRepoPermissionsPage, allAddedRepoIDs, userIDs []int32, perm authz.Perms, updatedAt time.Time) (*sqlf.Query, error) {
	// If changing the parameters used in this query, make sure to run relevant tests
	// named `postgresParameterLimitTest` using "go test -slow-tests".
	const format = `
-- source: enterprise/internal/database/perms_store.go:upsertRepoPermissionsBatchQuery
INSERT INTO repo_permissions
	(repo_id, permission, user_ids_ints, updated_at)
VALUES
	%s
ON CONFLICT ON CONSTRAINT
	repo_permissions_perm_unique
DO UPDATE SET
	user_ids_ints = CASE
		-- When the repository is part of "addedRepoIDs"
		WHEN repo_permissions.repo_id = ANY (%s) THEN
			repo_permissions.user_ids_ints | excluded.user_ids_ints
		ELSE
			repo_permissions.user_ids_ints - %s::INT[]
		END,
	updated_at = excluded.updated_at
`
	if updatedAt.IsZero() {
		return nil, ErrPermsUpdatedAtNotSet
	}

	items := make([]*sqlf.Query, 0, len(page.addedRepoIDs)+len(page.removedRepoIDs))
	for _, repoID := range page.addedRepoIDs {
		items = append(items, sqlf.Sprintf("(%s, %s, %s, %s)",
			repoID,
			perm.String(),
			pq.Array(userIDs),
			updatedAt.UTC(),
		))
	}
	for _, repoID := range page.removedRepoIDs {
		items = append(items, sqlf.Sprintf("(%s, %s, %s, %s)",
			repoID,
			perm.String(),

			// NOTE: Rows from `removedRepoIDs` are expected to exist, but in case they do not,
			// we need to set it with empty user IDs to be safe (it's a removal anyway).
			pq.Array([]int32{}),

			updatedAt.UTC(),
		))
	}

	return sqlf.Sprintf(
		format,
		sqlf.Join(items, ","),
		pq.Array(allAddedRepoIDs),

		// NOTE: Because we use empty user IDs for `removedRepoIDs`, we can't reuse "excluded.user_ids_ints"
		// and have to explicitly set what user IDs to be removed.
		pq.Array(userIDs),
	), nil
}

func deleteUserPendingPermissionsQuery(p *authz.UserPendingPermissions) *sqlf.Query {
	const format = `
-- source: enterprise/internal/database/perms_store.go:deleteUserPendingPermissionsQuery
DELETE FROM user_pending_permissions
WHERE service_type = %s
AND service_id = %s
AND permission = %s
AND object_type = %s
AND bind_id = %s
`

	return sqlf.Sprintf(
		format,
		p.ServiceType,
		p.ServiceID,
		p.Perm.String(),
		p.Type,
		p.BindID,
	)
}

func (s *permsStore) ListPendingUsers(ctx context.Context, serviceType, serviceID string) (bindIDs []string, err error) {
	ctx, save := s.observe(ctx, "ListPendingUsers", "")
	defer save(&err)

	q := sqlf.Sprintf(`
SELECT bind_id, object_ids_ints
FROM user_pending_permissions
WHERE service_type = %s
AND service_id = %s
`, serviceType, serviceID)

	var rows *sql.Rows
	rows, err = s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var bindID string
		var ids []int64
		if err = rows.Scan(&bindID, pq.Array(&ids)); err != nil {
			return nil, err
		}

		// This user has no pending permissions, only has an empty record
		if len(ids) == 0 {
			continue
		}
		bindIDs = append(bindIDs, bindID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bindIDs, nil
}

func (s *permsStore) DeleteAllUserPermissions(ctx context.Context, userID int32) (err error) {
	ctx, save := s.observe(ctx, "DeleteAllUserPermissions", "")
	defer func() { save(&err, otlog.Int32("userID", userID)) }()

	// NOTE: Practically, we don't need to clean up "repo_permissions" table because the value of "id" column
	// that is associated with this user will be invalidated automatically by deleting this row.
	if err = s.execute(ctx, sqlf.Sprintf(`DELETE FROM user_permissions WHERE user_id = %s`, userID)); err != nil {
		return errors.Wrap(err, "execute delete user permissions query")
	}

	return nil
}

func (s *permsStore) DeleteAllUserPendingPermissions(ctx context.Context, accounts *extsvc.Accounts) (err error) {
	ctx, save := s.observe(ctx, "DeleteAllUserPendingPermissions", "")
	defer func() { save(&err, accounts.TracingFields()...) }()

	// NOTE: Practically, we don't need to clean up "repo_pending_permissions" table because the value of "id" column
	// that is associated with this user will be invalidated automatically by deleting this row.
	items := make([]*sqlf.Query, len(accounts.AccountIDs))
	for i := range accounts.AccountIDs {
		items[i] = sqlf.Sprintf("%s", accounts.AccountIDs[i])
	}
	q := sqlf.Sprintf(`
-- source: enterprise/internal/database/perms_store.go:PermsStore.DeleteAllUserPendingPermissions
DELETE FROM user_pending_permissions
WHERE service_type = %s
AND service_id = %s
AND bind_id IN (%s)`,
		accounts.ServiceType, accounts.ServiceID, sqlf.Join(items, ","))
	if err = s.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute delete user pending permissions query")
	}

	return nil
}

func (s *permsStore) execute(ctx context.Context, q *sqlf.Query, vs ...any) (err error) {
	ctx, save := s.observe(ctx, "execute", "")
	defer func() { save(&err, otlog.Object("q", q)) }()

	var rows *sql.Rows
	rows, err = s.Query(ctx, q)
	if err != nil {
		return err
	}

	if len(vs) > 0 {
		if !rows.Next() {
			// One row is expected, return ErrPermsNotFound if no other errors occurred.
			err = rows.Err()
			if err == nil {
				err = authz.ErrPermsNotFound
			}
			return err
		}

		if err = rows.Scan(vs...); err != nil {
			return err
		}
	}

	return rows.Close()
}

// loadUserPermissions is a method that scans three values from one user_permissions table row:
// []int32 (ids), time.Time (updatedAt) and nullable time.Time (syncedAt).
func (s *permsStore) loadUserPermissions(ctx context.Context, p *authz.UserPermissions, lock string) (ids []int32, updatedAt, syncedAt time.Time, err error) {
	const format = `
-- source: enterprise/internal/database/perms_store.go:loadUserPermissions
SELECT object_ids_ints, updated_at, synced_at
FROM user_permissions
WHERE user_id = %s
AND permission = %s
AND object_type = %s
`

	q := sqlf.Sprintf(
		format+lock,
		p.UserID,
		p.Perm.String(),
		p.Type,
	)
	ctx, save := s.observe(ctx, "load", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()
	var rows *sql.Rows
	rows, err = s.Query(ctx, q)
	if err != nil {
		return nil, time.Time{}, time.Time{}, err
	}

	if !rows.Next() {
		// One row is expected, return ErrPermsNotFound if no other errors occurred.
		err = rows.Err()
		if err == nil {
			err = authz.ErrPermsNotFound
		}
		return nil, time.Time{}, time.Time{}, err
	}

	if err = rows.Scan(pq.Array(&ids), &updatedAt, &dbutil.NullTime{Time: &syncedAt}); err != nil {
		return nil, time.Time{}, time.Time{}, err
	}

	if err = rows.Close(); err != nil {
		return nil, time.Time{}, time.Time{}, err
	}

	return ids, updatedAt, syncedAt, nil
}

// loadRepoPermissions is a method that scans three values from one repo_permissions table row:
// []int32 (ids), time.Time (updatedAt) and nullable time.Time (syncedAt).
func (s *permsStore) loadRepoPermissions(ctx context.Context, p *authz.RepoPermissions, lock string) (ids []int32, updatedAt, syncedAt time.Time, unrestricted bool, err error) {
	const format = `
-- source: enterprise/internal/database/perms_store.go:loadRepoPermissions
SELECT user_ids_ints, updated_at, synced_at, unrestricted
FROM repo_permissions
WHERE repo_id = %s
AND permission = %s
`

	q := sqlf.Sprintf(
		format+lock,
		p.RepoID,
		p.Perm.String(),
	)

	ctx, save := s.observe(ctx, "load", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()
	var rows *sql.Rows
	rows, err = s.Query(ctx, q)
	if err != nil {
		return nil, time.Time{}, time.Time{}, false, err
	}

	if !rows.Next() {
		// One row is expected, return ErrPermsNotFound if no other errors occurred.
		err = rows.Err()
		if err == nil {
			err = authz.ErrPermsNotFound
		}
		return nil, time.Time{}, time.Time{}, false, err
	}

	if err = rows.Scan(pq.Array(&ids), &updatedAt, &dbutil.NullTime{Time: &syncedAt}, &unrestricted); err != nil {
		return nil, time.Time{}, time.Time{}, false, err
	}

	if err = rows.Close(); err != nil {
		return nil, time.Time{}, time.Time{}, false, err
	}
	return ids, updatedAt, syncedAt, unrestricted, nil
}

// loadUserPendingPermissions is a method that scans three values from one user_pending_permissions table row:
// int64 (id), []int32 (ids), time.Time (updatedAt).
func (s *permsStore) loadUserPendingPermissions(ctx context.Context, p *authz.UserPendingPermissions, lock string) (id int64, ids []int32, updatedAt time.Time, err error) {
	const format = `
-- source: enterprise/internal/database/perms_store.go:loadUserPendingPermissions
SELECT id, object_ids_ints, updated_at
FROM user_pending_permissions
WHERE service_type = %s
AND service_id = %s
AND permission = %s
AND object_type = %s
AND bind_id = %s
`
	q := sqlf.Sprintf(
		format+lock,
		p.ServiceType,
		p.ServiceID,
		p.Perm.String(),
		p.Type,
		p.BindID,
	)
	ctx, save := s.observe(ctx, "load", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()
	var rows *sql.Rows
	rows, err = s.Query(ctx, q)
	if err != nil {
		return -1, nil, time.Time{}, err
	}

	if !rows.Next() {
		// One row is expected, return ErrPermsNotFound if no other errors occurred.
		err = rows.Err()
		if err == nil {
			err = authz.ErrPermsNotFound
		}
		return -1, nil, time.Time{}, err
	}

	if err = rows.Scan(&id, pq.Array(&ids), &updatedAt); err != nil {
		return -1, nil, time.Time{}, err
	}
	if err = rows.Close(); err != nil {
		return -1, nil, time.Time{}, err
	}

	return id, ids, updatedAt, nil
}

// loadRepoPendingPermissions is a method that scans three values from one repo_pending_permissions table row:
// int32 (id), []int64 (ids), time.Time (updatedAt) and nullable time.Time (syncedAt).
func (s *permsStore) loadRepoPendingPermissions(ctx context.Context, p *authz.RepoPermissions, lock string) (id int32, ids []int64, updatedAt, syncedAt time.Time, err error) {
	const format = `
-- source: enterprise/internal/database/perms_store.go:loadRepoPendingPermissionsQuery
SELECT repo_id, user_ids_ints, updated_at, NULL
FROM repo_pending_permissions
WHERE repo_id = %s
AND permission = %s
`
	q := sqlf.Sprintf(
		format+lock,
		p.RepoID,
		p.Perm.String(),
	)
	ctx, save := s.observe(ctx, "load", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()
	var rows *sql.Rows
	rows, err = s.Query(ctx, q)
	if err != nil {
		return -1, nil, time.Time{}, time.Time{}, err
	}

	if !rows.Next() {
		// One row is expected, return ErrPermsNotFound if no other errors occurred.
		err = rows.Err()
		if err == nil {
			err = authz.ErrPermsNotFound
		}
		return -1, nil, time.Time{}, time.Time{}, err
	}

	if err = rows.Scan(&id, pq.Array(&ids), &updatedAt, &dbutil.NullTime{Time: &syncedAt}); err != nil {
		return -1, nil, time.Time{}, time.Time{}, err
	}
	if err = rows.Close(); err != nil {
		return -1, nil, time.Time{}, time.Time{}, err
	}

	return id, ids, updatedAt, syncedAt, nil
}

func (s *permsStore) GetUserIDsByExternalAccounts(ctx context.Context, accounts *extsvc.Accounts) (_ map[string]int32, err error) {
	ctx, save := s.observe(ctx, "ListUsersByExternalAccounts", "")
	defer func() { save(&err, accounts.TracingFields()...) }()

	items := make([]*sqlf.Query, len(accounts.AccountIDs))
	for i := range accounts.AccountIDs {
		items[i] = sqlf.Sprintf("%s", accounts.AccountIDs[i])
	}

	q := sqlf.Sprintf(`
-- source: enterprise/internal/database/perms_store.go:PermsStore.GetUserIDsByExternalAccounts
SELECT user_id, account_id
FROM user_external_accounts
WHERE service_type = %s
AND service_id = %s
AND account_id IN (%s)
AND deleted_at IS NULL
`, accounts.ServiceType, accounts.ServiceID, sqlf.Join(items, ","))
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	userIDs := make(map[string]int32)
	for rows.Next() {
		var userID int32
		var accountID string
		if err := rows.Scan(&userID, &accountID); err != nil {
			return nil, err
		}
		userIDs[accountID] = userID
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (s *permsStore) UserIDsWithNoPerms(ctx context.Context) ([]int32, error) {
	// By default, site admins can access any repo
	filterSiteAdmins := sqlf.Sprintf("users.site_admin = FALSE")
	// Unless we enforce it in config
	if conf.Get().AuthzEnforceForSiteAdmins {
		filterSiteAdmins = sqlf.Sprintf("TRUE")
	}

	q := sqlf.Sprintf(`
-- source: enterprise/internal/database/perms_store.go:PermsStore.UserIDsWithNoPerms
SELECT users.id, NULL
FROM users
WHERE
	users.deleted_at IS NULL
AND %s
AND NOT EXISTS (
		SELECT
		FROM user_permissions
		WHERE user_id = users.id
	)
`, filterSiteAdmins)
	results, err := s.loadIDsWithTime(ctx, q)
	if err != nil {
		return nil, err
	}

	ids := make([]int32, 0, len(results))
	for id := range results {
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *permsStore) UserIDsWithOutdatedPerms(ctx context.Context) (map[int32]time.Time, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/internal/database/perms_store.go:PermsStore.UserIDsWithOutdatedPerms
SELECT
	user_permissions.user_id,
	user_permissions.synced_at
FROM external_services
JOIN user_permissions ON user_permissions.user_id = external_services.namespace_user_id
WHERE
	external_services.deleted_at IS NULL
AND (
		user_permissions.synced_at IS NULL
	OR  external_services.last_sync_at >= user_permissions.synced_at
)

UNION

SELECT
	user_permissions.user_id,
	user_permissions.synced_at
FROM external_services
JOIN org_members ON org_members.org_id = external_services.namespace_org_id
JOIN user_permissions ON user_permissions.user_id = org_members.user_id
WHERE
	external_services.deleted_at IS NULL
AND (
		user_permissions.synced_at IS NULL
	OR  external_services.last_sync_at >= user_permissions.synced_at
)
`)
	return s.loadIDsWithTime(ctx, q)
}

func (s *permsStore) RepoIDsWithNoPerms(ctx context.Context) ([]api.RepoID, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/internal/database/perms_store.go:PermsStore.RepoIDsWithNoPerms
SELECT repo.id, NULL FROM repo
WHERE repo.deleted_at IS NULL
AND repo.private = TRUE
AND repo.id NOT IN
	(SELECT perms.repo_id FROM repo_permissions AS perms
	 UNION
	 SELECT pending.repo_id FROM repo_pending_permissions AS pending)
`)

	results, err := s.loadIDsWithTime(ctx, q)
	if err != nil {
		return nil, err
	}

	ids := make([]api.RepoID, 0, len(results))
	for id := range results {
		ids = append(ids, api.RepoID(id))
	}
	return ids, nil
}

// UserIDsWithOldestPerms lists the users with the oldest synced perms, limited
// to limit. If age is non-zero, users that have synced within "age" since now
// will be filtered out.
func (s *permsStore) UserIDsWithOldestPerms(ctx context.Context, limit int, age time.Duration) (map[int32]time.Time, error) {
	cutoffClause := sqlf.Sprintf("TRUE")
	if age > 0 {
		cutoff := s.clock().Add(-1 * age)
		cutoffClause = sqlf.Sprintf("(perms.synced_at IS NULL OR perms.synced_at < %s)", cutoff)
	}
	q := sqlf.Sprintf(`
-- source: enterprise/internal/database/perms_store.go:PermsStore.UserIDsWithOldestPerms
SELECT perms.user_id, perms.synced_at FROM user_permissions AS perms
WHERE perms.user_id IN
	(SELECT users.id FROM users
	 WHERE users.deleted_at IS NULL)
AND %s
ORDER BY perms.synced_at ASC NULLS FIRST
LIMIT %s
`, cutoffClause, limit)
	return s.loadIDsWithTime(ctx, q)
}

func (s *permsStore) ReposIDsWithOldestPerms(ctx context.Context, limit int, age time.Duration) (map[api.RepoID]time.Time, error) {
	cutoffClause := sqlf.Sprintf("TRUE")
	if age > 0 {
		cutoff := s.clock().Add(-1 * age)
		cutoffClause = sqlf.Sprintf("(perms.synced_at IS NULL OR perms.synced_at < %s)", cutoff)
	}
	q := sqlf.Sprintf(`
-- source: enterprise/internal/database/perms_store.go:PermsStore.ReposIDsWithOldestPerms
SELECT perms.repo_id, perms.synced_at FROM repo_permissions AS perms
WHERE perms.repo_id IN
	(SELECT repo.id FROM repo
	 WHERE repo.deleted_at IS NULL)
AND %s
ORDER BY perms.synced_at ASC NULLS FIRST
LIMIT %s
`, cutoffClause, limit)

	pairs, err := s.loadIDsWithTime(ctx, q)
	if err != nil {
		return nil, err
	}

	results := make(map[api.RepoID]time.Time, len(pairs))
	for id, t := range pairs {
		results[api.RepoID(id)] = t
	}
	return results, nil
}

// loadIDsWithTime runs the query and returns a list of ID and nullable time pairs.
func (s *permsStore) loadIDsWithTime(ctx context.Context, q *sqlf.Query) (map[int32]time.Time, error) {
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	results := make(map[int32]time.Time)
	for rows.Next() {
		var id int32
		var t time.Time
		if err = rows.Scan(&id, &dbutil.NullTime{Time: &t}); err != nil {
			return nil, err
		}

		results[id] = t
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *permsStore) UserIsMemberOfOrgHasCodeHostConnection(ctx context.Context, userID int32) (has bool, err error) {
	ctx, save := s.observe(ctx, "UserIsMemberOfOrgHasCodeHostConnection", "")
	defer func() { save(&err, otlog.Int32("userID", userID)) }()

	q := sqlf.Sprintf(`
-- source: enterprise/internal/database/perms_store.go:PermsStore.UserIsMemberOfOrgHasCodeHostConnection
SELECT EXISTS (
	SELECT
	FROM org_members
	JOIN external_services ON external_services.namespace_org_id = org_id
	WHERE user_id = %s
)
`, userID)
	err = s.QueryRow(ctx, q).Scan(&has)
	if err != nil {
		return false, err
	}
	return has, nil
}

// PermsMetrics contains metrics values calculated by querying the database.
type PermsMetrics struct {
	// The number of users with stale permissions.
	UsersWithStalePerms int64
	// The seconds between users with oldest and the most up-to-date permissions.
	UsersPermsGapSeconds float64
	// The number of repositories with stale permissions.
	ReposWithStalePerms int64
	// The seconds between repositories with oldest and the most up-to-date permissions.
	ReposPermsGapSeconds float64
	// The number of repositories with stale sub-repo permissions.
	SubReposWithStalePerms int64
	// The seconds between repositories with oldest and the most up-to-date sub-repo
	// permissions.
	SubReposPermsGapSeconds float64
}

func (s *permsStore) Metrics(ctx context.Context, staleDur time.Duration) (*PermsMetrics, error) {
	m := &PermsMetrics{}

	stale := s.clock().Add(-1 * staleDur)
	q := sqlf.Sprintf(`
SELECT COUNT(*) FROM user_permissions AS perms
WHERE
	perms.user_id IN
		(
			SELECT users.id FROM users
			WHERE users.deleted_at IS NULL
		)
AND perms.updated_at <= %s
`, stale)
	if err := s.execute(ctx, q, &m.UsersWithStalePerms); err != nil {
		return nil, errors.Wrap(err, "users with stale perms")
	}

	var seconds sql.NullFloat64
	q = sqlf.Sprintf(`
SELECT EXTRACT(EPOCH FROM (MAX(updated_at) - MIN(updated_at)))
FROM user_permissions AS perms
WHERE perms.user_id IN
	(
		SELECT users.id FROM users
		WHERE users.deleted_at IS NULL
	)
`)
	if err := s.execute(ctx, q, &seconds); err != nil {
		return nil, errors.Wrap(err, "users perms gap seconds")
	}
	m.UsersPermsGapSeconds = seconds.Float64

	q = sqlf.Sprintf(`
SELECT COUNT(*) FROM repo_permissions AS perms
WHERE perms.repo_id IN
	(
		SELECT repo.id FROM repo
		WHERE
			repo.deleted_at IS NULL
		AND repo.private = TRUE
	)
AND perms.updated_at <= %s
`, stale)
	if err := s.execute(ctx, q, &m.ReposWithStalePerms); err != nil {
		return nil, errors.Wrap(err, "repos with stale perms")
	}

	q = sqlf.Sprintf(`
SELECT EXTRACT(EPOCH FROM (MAX(perms.updated_at) - MIN(perms.updated_at)))
FROM repo_permissions AS perms
WHERE perms.repo_id IN
	(
		SELECT repo.id FROM repo
		WHERE
			repo.deleted_at IS NULL
		AND repo.private = TRUE
	)
`)
	if err := s.execute(ctx, q, &seconds); err != nil {
		return nil, errors.Wrap(err, "repos perms gap seconds")
	}
	m.ReposPermsGapSeconds = seconds.Float64

	q = sqlf.Sprintf(`
SELECT COUNT(*) FROM sub_repo_permissions AS perms
WHERE perms.repo_id IN
	(
		SELECT repo.id FROM repo
		WHERE
			repo.deleted_at IS NULL
		AND repo.private = TRUE
	)
AND perms.updated_at <= %s
`, stale)
	if err := s.execute(ctx, q, &m.SubReposWithStalePerms); err != nil {
		return nil, errors.Wrap(err, "repos with stale sub-repo perms")
	}

	q = sqlf.Sprintf(`
SELECT EXTRACT(EPOCH FROM (MAX(perms.updated_at) - MIN(perms.updated_at)))
FROM sub_repo_permissions AS perms
WHERE perms.repo_id IN
	(
		SELECT repo.id FROM repo
		WHERE
			repo.deleted_at IS NULL
		AND repo.private = TRUE
	)
`)
	if err := s.execute(ctx, q, &seconds); err != nil {
		return nil, errors.Wrap(err, "sub-repo perms gap seconds")
	}
	m.SubReposPermsGapSeconds = seconds.Float64

	return m, nil
}

//nolint:unparam // unparam complains that `title` always has same value across call-sites, but that's OK
func (s *permsStore) observe(ctx context.Context, family, title string) (context.Context, func(*error, ...otlog.Field)) {
	began := s.clock()
	tr, ctx := trace.New(ctx, "database.PermsStore."+family, title)

	return ctx, func(err *error, fs ...otlog.Field) {
		now := s.clock()
		took := now.Sub(began)

		fs = append(fs, otlog.String("Duration", took.String()))

		tr.LogFields(fs...)

		success := err == nil || *err == nil
		if !success {
			tr.SetError(*err)
		}

		tr.Finish()
	}
}

// computeDiff determines which ids were added or removed when comparing the old
// list of ids, oldIDs, with the new set.
func computeDiff[T comparable](oldIDs map[T]struct{}, set map[T]struct{}) (added []T, removed []T) {
	for key := range set {
		if _, ok := oldIDs[key]; !ok {
			added = append(added, key)
		}
	}
	for key := range oldIDs {
		if _, ok := set[key]; !ok {
			removed = append(removed, key)
		}
	}
	return added, removed
}

func sliceToSet[T comparable](s []T) map[T]struct{} {
	m := make(map[T]struct{}, len(s))
	for _, n := range s {
		m[n] = struct{}{}
	}
	return m
}

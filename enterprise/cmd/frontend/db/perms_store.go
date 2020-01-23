package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	iauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	ErrPermsNotFound        = errors.New("permissions not found")
	ErrPermsUpdatedAtNotSet = errors.New("permissions UpdatedAt timestamp must be set")
)

// PermsStore is the unified interface for managing permissions explicitly in the database.
// It is concurrency-safe and maintains data consistency over the 'user_permissions',
// 'repo_permissions', 'user_pending_permissions', and 'repo_pending_permissions' tables.
type PermsStore struct {
	db    dbutil.DB
	clock func() time.Time
}

// NewPermsStore returns a new PermsStore with given parameters.
func NewPermsStore(db dbutil.DB, clock func() time.Time) *PermsStore {
	return &PermsStore{
		db:    db,
		clock: clock,
	}
}

// LoadUserPermissions loads stored user permissions into p. An ErrPermsNotFound is returned
// when there are no valid permissions available.
func (s *PermsStore) LoadUserPermissions(ctx context.Context, p *iauthz.UserPermissions) (err error) {
	if Mocks.Perms.LoadUserPermissions != nil {
		return Mocks.Perms.LoadUserPermissions(ctx, p)
	}

	ctx, save := s.observe(ctx, "LoadUserPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	vals, err := s.load(ctx, loadUserPermissionsQuery(p, ""))
	if err != nil {
		return err
	}
	p.IDs = vals.ids
	p.UpdatedAt = vals.updatedAt
	return nil
}

func loadUserPermissionsQuery(p *iauthz.UserPermissions, lock string) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:loadUserPermissionsQuery
SELECT user_id, object_ids, updated_at
FROM user_permissions
WHERE user_id = %s
AND permission = %s
AND object_type = %s
AND provider = %s
`

	return sqlf.Sprintf(
		format+lock,
		p.UserID,
		p.Perm.String(),
		p.Type,
		p.Provider,
	)
}

// LoadRepoPermissions loads stored repository permissions into p. An ErrPermsNotFound is
// returned when there are no valid permissions available.
func (s *PermsStore) LoadRepoPermissions(ctx context.Context, p *iauthz.RepoPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	vals, err := s.load(ctx, loadRepoPermissionsQuery(p, ""))
	if err != nil {
		return err
	}
	p.UserIDs = vals.ids
	p.UpdatedAt = vals.updatedAt
	return nil
}

func loadRepoPermissionsQuery(p *iauthz.RepoPermissions, lock string) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:loadRepoPermissionsQuery
SELECT repo_id, user_ids, updated_at
FROM repo_permissions
WHERE repo_id = %s
AND permission = %s
AND provider = %s
`

	return sqlf.Sprintf(
		format+lock,
		p.RepoID,
		p.Perm.String(),
		p.Provider,
	)
}

// SetRepoPermissions performs a full update for p, new user IDs found in p will be upserted
// and user IDs no longer in p will be removed. This method updates both `user_permissions`
// and `repo_permissions` tables.
//
// Example input:
// &RepoPermissions{
//     RepoID: 1,
//     Perm: authz.Read,
//     UserIDs: bitmap{1, 2},
//     Provider: ProviderSourcegraph,
// }
//
// Table states for input:
// 	"user_permissions":
//   user_id | permission | object_type | object_ids | updated_at |  provider
//  ---------+------------+-------------+------------+------------+------------
//         1 |       read |       repos |  bitmap{1} | <DateTime> | sourcegraph
//         2 |       read |       repos |  bitmap{1} | <DateTime> | sourcegraph
//
//  "repo_permissions":
//   repo_id | permission |   user_ids   |   provider  | updated_at
//  ---------+------------+--------------+-------------+------------
//         1 |       read | bitmap{1, 2} | sourcegraph | <DateTime>
func (s *PermsStore) SetRepoPermissions(ctx context.Context, p *iauthz.RepoPermissions) (err error) {
	ctx, save := s.observe(ctx, "SetRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	// Open a transaction for update consistency.
	txs, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer txs.Done(&err)

	// Retrieve currently stored user IDs of this repository.
	var oldIDs *roaring.Bitmap
	vals, err := txs.load(ctx, loadRepoPermissionsQuery(p, "FOR UPDATE"))
	if err != nil {
		if err == ErrPermsNotFound {
			oldIDs = roaring.NewBitmap()
		} else {
			return errors.Wrap(err, "load repo permissions")
		}
	} else {
		oldIDs = vals.ids
	}

	if p.UserIDs == nil {
		p.UserIDs = roaring.NewBitmap()
	}

	// Compute differences between the old and new sets.
	added := roaring.AndNot(p.UserIDs, oldIDs)
	removed := roaring.AndNot(oldIDs, p.UserIDs)

	// Load stored user IDs of both added and removed.
	changedIDs := roaring.Or(added, removed).ToArray()

	// In case there is nothing to add or remove.
	if len(changedIDs) == 0 {
		return nil
	}

	q := loadUserPermissionsBatchQuery(changedIDs, p.Perm, authz.PermRepos, p.Provider, "FOR UPDATE")
	loadedIDs, err := txs.batchLoadIDs(ctx, q)
	if err != nil {
		return errors.Wrap(err, "batch load user permissions")
	}

	// We have two sets of IDs that one needs to add, and the other needs to remove.
	updatedAt := txs.clock()
	updatedPerms := make([]*iauthz.UserPermissions, 0, len(changedIDs))
	for _, id := range changedIDs {
		userID := int32(id)
		repoIDs := loadedIDs[userID]
		if repoIDs == nil {
			repoIDs = roaring.NewBitmap()
		}

		switch {
		case added.Contains(id):
			repoIDs.Add(uint32(p.RepoID))
		case removed.Contains(id):
			repoIDs.Remove(uint32(p.RepoID))
		}

		updatedPerms = append(updatedPerms, &iauthz.UserPermissions{
			UserID:    userID,
			Perm:      p.Perm,
			Type:      authz.PermRepos,
			IDs:       repoIDs,
			Provider:  p.Provider,
			UpdatedAt: updatedAt,
		})
	}

	if q, err = upsertUserPermissionsBatchQuery(updatedPerms...); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute upsert user permissions batch query")
	}

	p.UpdatedAt = updatedAt
	if q, err = upsertRepoPermissionsBatchQuery(p); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute upsert repo permissions batch query")
	}

	return nil
}

func loadUserPermissionsBatchQuery(
	userIDs []uint32,
	perm authz.Perms,
	typ authz.PermType,
	provider authz.ProviderType,
	lock string,
) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:loadUserPermissionsBatchQuery
SELECT user_id, object_ids
FROM user_permissions
WHERE user_id IN (%s)
AND permission = %s
AND object_type = %s
AND provider = %s
`

	items := make([]*sqlf.Query, len(userIDs))
	for i := range userIDs {
		items[i] = sqlf.Sprintf("%d", userIDs[i])
	}
	return sqlf.Sprintf(
		format+lock,
		sqlf.Join(items, ","),
		perm.String(),
		typ,
		provider,
	)
}

func upsertUserPermissionsBatchQuery(ps ...*iauthz.UserPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:upsertUserPermissionsBatchQuery
INSERT INTO user_permissions
  (user_id, permission, object_type, object_ids, provider, updated_at)
VALUES
  %s
ON CONFLICT ON CONSTRAINT
  user_permissions_perm_object_provider_unique
DO UPDATE SET
  object_ids = excluded.object_ids,
  updated_at = excluded.updated_at
`

	items := make([]*sqlf.Query, len(ps))
	for i := range ps {
		ps[i].IDs.RunOptimize()
		ids, err := ps[i].IDs.ToBytes()
		if err != nil {
			return nil, err
		}

		if ps[i].UpdatedAt.IsZero() {
			return nil, ErrPermsUpdatedAtNotSet
		}

		items[i] = sqlf.Sprintf("(%s, %s, %s, %s, %s, %s)",
			ps[i].UserID,
			ps[i].Perm.String(),
			ps[i].Type,
			ids,
			ps[i].Provider,
			ps[i].UpdatedAt.UTC(),
		)
	}

	return sqlf.Sprintf(
		format,
		sqlf.Join(items, ","),
	), nil
}

// LoadUserPendingPermissions returns pending permissions found by given parameters.
// An ErrPermsNotFound is returned when there are no pending permissions available.
func (s *PermsStore) LoadUserPendingPermissions(ctx context.Context, p *iauthz.UserPendingPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadUserPendingPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	vals, err := s.load(ctx, loadUserPendingPermissionsQuery(p, ""))
	if err != nil {
		return err
	}
	p.ID = vals.id
	p.IDs = vals.ids
	p.UpdatedAt = vals.updatedAt
	return nil
}

func loadUserPendingPermissionsQuery(p *iauthz.UserPendingPermissions, lock string) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:loadUserPendingPermissionsQuery
SELECT id, object_ids, updated_at
FROM user_pending_permissions
WHERE bind_id = %s
AND permission = %s
AND object_type = %s
`
	return sqlf.Sprintf(
		format+lock,
		p.BindID,
		p.Perm.String(),
		p.Type,
	)
}

// SetRepoPendingPermissions performs a full update for p with given bindIDs, new bind IDs
// found will be upserted and bind IDs no longer in bindIDs will be removed. This method
// updates both `user_pending_permissions` and `repo_pending_permissions` tables.
//
// Example input:
// string{"alice", "bob"}
// &RepoPermissions{
//     RepoID: 1,
//     Perm: authz.Read,
// }
//
// Table states for input:
// 	"user_pending_permissions":
//   id | bind_id | permission | object_type | object_ids | updated_at
//  ----+---------+------------+-------------+------------+------------
//    1 |   alice |       read |       repos |  bitmap{1} | <DateTime>
//    2 |     bob |       read |       repos |  bitmap{1} | <DateTime>
//
//  "repo_pending_permissions":
//   repo_id | permission |   user_ids   | updated_at
//  ---------+------------+--------------+------------
//         1 |       read | bitmap{1, 2} | <DateTime>
func (s *PermsStore) SetRepoPendingPermissions(ctx context.Context, bindIDs []string, p *iauthz.RepoPermissions) (err error) {
	ctx, save := s.observe(ctx, "SetRepoPendingPermissions", "")
	defer func() { save(&err, append(p.TracingFields(), otlog.String("bindIDs", strings.Join(bindIDs, ",")))...) }()

	// Open a transaction for update consistency.
	txs, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer txs.Done(&err)

	var q *sqlf.Query

	p.UserIDs = roaring.NewBitmap()

	// Insert rows for bindIDs without one in the "user_pending_permissions" table.
	// The insert does not store any permissions data but uses auto-increment key to generate unique ID.
	// This help guarantees rows of all bindIDs exist when getting user IDs in next load query.
	updatedAt := txs.clock()
	p.UpdatedAt = updatedAt
	if len(bindIDs) > 0 {
		// NOTE: Row-level locking is not needed here because we're creating stub rows and not modifying permissions.
		q, err = insertUserPendingPermissionsBatchQuery(bindIDs, p)
		if err != nil {
			return err
		}

		ids, err := txs.loadUserPendingPermissionsIDs(ctx, q)
		if err != nil {
			return errors.Wrap(err, "load user pending permissions IDs")
		}

		// Make up p.UserIDs from the result set.
		p.UserIDs.AddMany(ids)
	}

	// Retrieve currently stored user IDs of this repository.
	vals, err := txs.load(ctx, loadRepoPendingPermissionsQuery(p, "FOR UPDATE"))
	if err != nil && err != ErrPermsNotFound {
		return errors.Wrap(err, "load repo pending permissions")
	}
	oldIDs := roaring.NewBitmap()
	if vals != nil && vals.ids != nil {
		oldIDs = vals.ids
	}

	// Compute differences between the old and new sets.
	added := roaring.AndNot(p.UserIDs, oldIDs)
	removed := roaring.AndNot(oldIDs, p.UserIDs)

	// Load stored user IDs of both added and removed.
	changedIDs := roaring.Or(added, removed).ToArray()

	// In case there is nothing to add or remove.
	if len(changedIDs) == 0 {
		return nil
	}

	q = loadUserPendingPermissionsByIDBatchQuery(changedIDs, p.Perm, authz.PermRepos, "FOR UPDATE")
	bindIDSet, loadedIDs, err := txs.batchLoadUserPendingPermissions(ctx, q)
	if err != nil {
		return errors.Wrap(err, "batch load user pending permissions")
	}

	updatedPerms := make([]*iauthz.UserPendingPermissions, 0, len(bindIDSet))
	for _, id := range changedIDs {
		userID := int32(id)
		repoIDs := loadedIDs[userID]
		if repoIDs == nil {
			repoIDs = roaring.NewBitmap()
		}

		switch {
		case added.Contains(id):
			repoIDs.Add(uint32(p.RepoID))
		case removed.Contains(id):
			repoIDs.Remove(uint32(p.RepoID))
		}

		updatedPerms = append(updatedPerms, &iauthz.UserPendingPermissions{
			BindID:    bindIDSet[userID],
			Perm:      p.Perm,
			Type:      authz.PermRepos,
			IDs:       repoIDs,
			UpdatedAt: updatedAt,
		})
	}

	if q, err = upsertUserPendingPermissionsBatchQuery(updatedPerms...); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute upsert user pending permissions batch query")
	}

	if q, err = upsertRepoPendingPermissionsBatchQuery(p); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute upsert repo pending permissions batch query")
	}

	return nil
}

func (s *PermsStore) loadUserPendingPermissionsIDs(ctx context.Context, q *sqlf.Query) (ids []uint32, err error) {
	ctx, save := s.observe(ctx, "loadUserPendingPermissionsIDs", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint32
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}

func (s *PermsStore) batchLoadUserPendingPermissions(ctx context.Context, q *sqlf.Query) (
	bindIDSet map[int32]string,
	loaded map[int32]*roaring.Bitmap,
	err error,
) {
	ctx, save := s.observe(ctx, "batchLoadUserPendingPermissions", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	bindIDSet = make(map[int32]string)
	loaded = make(map[int32]*roaring.Bitmap)
	for rows.Next() {
		var id int32
		var bindID string
		var ids []byte
		if err = rows.Scan(&id, &bindID, &ids); err != nil {
			return nil, nil, err
		}

		bindIDSet[id] = bindID

		if len(ids) == 0 {
			continue
		}

		bm := roaring.NewBitmap()
		if err = bm.UnmarshalBinary(ids); err != nil {
			return nil, nil, err
		}
		loaded[id] = bm
	}
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}

	return bindIDSet, loaded, nil
}

func insertUserPendingPermissionsBatchQuery(
	bindIDs []string,
	p *iauthz.RepoPermissions,
) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:insertUserPendingPermissionsBatchQuery
INSERT INTO user_pending_permissions
  (bind_id, permission, object_type, object_ids, updated_at)
VALUES
  %s
ON CONFLICT ON CONSTRAINT
  user_pending_permissions_perm_object_unique
DO UPDATE SET
  updated_at = excluded.updated_at
RETURNING id
`

	if p.UpdatedAt.IsZero() {
		return nil, ErrPermsUpdatedAtNotSet
	}

	items := make([]*sqlf.Query, len(bindIDs))
	for i := range bindIDs {
		items[i] = sqlf.Sprintf("(%s, %s, %s, %s, %s)",
			bindIDs[i],
			p.Perm.String(),
			authz.PermRepos,
			[]byte{},
			p.UpdatedAt.UTC(),
		)
	}

	return sqlf.Sprintf(
		format,
		sqlf.Join(items, ","),
	), nil
}

func loadRepoPendingPermissionsQuery(p *iauthz.RepoPermissions, lock string) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:loadRepoPendingPermissionsQuery
SELECT repo_id, user_ids, updated_at
FROM repo_pending_permissions
WHERE repo_id = %s
AND permission = %s
`
	return sqlf.Sprintf(
		format+lock,
		p.RepoID,
		p.Perm.String(),
	)
}

func loadUserPendingPermissionsByIDBatchQuery(ids []uint32, perm authz.Perms, typ authz.PermType, lock string) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:loadUserPendingPermissionsByIDBatchQuery
SELECT id, bind_id, object_ids
FROM user_pending_permissions
WHERE id IN (%s)
AND permission = %s
AND object_type = %s
`

	items := make([]*sqlf.Query, len(ids))
	for i := range ids {
		items[i] = sqlf.Sprintf("%d", ids[i])
	}
	return sqlf.Sprintf(
		format+lock,
		sqlf.Join(items, ","),
		perm.String(),
		typ,
	)
}

func upsertUserPendingPermissionsBatchQuery(ps ...*iauthz.UserPendingPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:upsertUserPendingPermissionsBatchQuery
INSERT INTO user_pending_permissions
  (bind_id, permission, object_type, object_ids, updated_at)
VALUES
  %s
ON CONFLICT ON CONSTRAINT
  user_pending_permissions_perm_object_unique
DO UPDATE SET
  object_ids = excluded.object_ids,
  updated_at = excluded.updated_at
`

	items := make([]*sqlf.Query, len(ps))
	for i := range ps {
		ps[i].IDs.RunOptimize()
		ids, err := ps[i].IDs.ToBytes()
		if err != nil {
			return nil, err
		}

		if ps[i].UpdatedAt.IsZero() {
			return nil, ErrPermsUpdatedAtNotSet
		}

		items[i] = sqlf.Sprintf("(%s, %s, %s, %s, %s)",
			ps[i].BindID,
			ps[i].Perm.String(),
			ps[i].Type,
			ids,
			ps[i].UpdatedAt.UTC(),
		)
	}

	return sqlf.Sprintf(
		format,
		sqlf.Join(items, ","),
	), nil
}

func upsertRepoPendingPermissionsBatchQuery(ps ...*iauthz.RepoPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:upsertRepoPendingPermissionsBatchQuery
INSERT INTO repo_pending_permissions
  (repo_id, permission, user_ids, updated_at)
VALUES
  %s
ON CONFLICT ON CONSTRAINT
  repo_pending_permissions_perm_unique
DO UPDATE SET
  user_ids = excluded.user_ids,
  updated_at = excluded.updated_at
`

	items := make([]*sqlf.Query, len(ps))
	for i := range ps {
		ps[i].UserIDs.RunOptimize()
		ids, err := ps[i].UserIDs.ToBytes()
		if err != nil {
			return nil, err
		}

		if ps[i].UpdatedAt.IsZero() {
			return nil, ErrPermsUpdatedAtNotSet
		}

		items[i] = sqlf.Sprintf("(%s, %s, %s, %s)",
			ps[i].RepoID,
			ps[i].Perm.String(),
			ids,
			ps[i].UpdatedAt.UTC(),
		)
	}

	return sqlf.Sprintf(
		format,
		sqlf.Join(items, ","),
	), nil
}

// GrantPendingPermissions grants the user has given ID with pending permissions found in p.
// It "merges" rows in pending permissions tables to effective permissions tables, i.e. permissions
// are unioned not replaced.
// This method starts its own transaction if the caller hasn't started one already.
func (s *PermsStore) GrantPendingPermissions(ctx context.Context, userID int32, p *iauthz.UserPendingPermissions) (err error) {
	if Mocks.Perms.GrantPendingPermissions != nil {
		return Mocks.Perms.GrantPendingPermissions(ctx, userID, p)
	}

	ctx, save := s.observe(ctx, "GrantPendingPermissions", "")
	defer func() { save(&err, append(p.TracingFields(), otlog.Object("userID", userID))...) }()

	var txs *PermsStore
	if s.inTx() {
		txs = s
	} else {
		// Open a transaction for update consistency.
		txs, err = s.Transact(ctx)
		if err != nil {
			return err
		}
		defer txs.Done(&err)
	}

	vals, err := txs.load(ctx, loadUserPendingPermissionsQuery(p, "FOR UPDATE"))
	if err != nil {
		// Skip the whole grant process if the user has no pending permissions.
		if err == ErrPermsNotFound {
			return nil
		}
		return errors.Wrap(err, "load user pending permissions")
	}
	p.ID = vals.id
	p.IDs = vals.ids

	// NOTE: We currently only have "repos" type, so avoid unnecessary type checking for now.
	ids := p.IDs.ToArray()
	if len(ids) == 0 {
		return nil
	}

	// Batch query all repository permissions object IDs in one go.
	// NOTE: It is critical to always acquire row-level locks in the same order as SetRepoPermissions
	// (i.e. repo -> user) to prevent deadlocks.
	q := loadRepoPermissionsBatchQuery(ids, p.Perm, authz.ProviderSourcegraph, "FOR UPDATE")
	loadedIDs, err := txs.batchLoadIDs(ctx, q)
	if err != nil {
		return errors.Wrap(err, "batch load repo permissions")
	}

	updatedAt := txs.clock()
	updatedPerms := make([]*iauthz.RepoPermissions, 0, len(ids))
	for i := range ids {
		repoID := int32(ids[i])
		oldIDs := loadedIDs[repoID]
		if oldIDs == nil {
			oldIDs = roaring.NewBitmap()
		}

		oldIDs.Add(uint32(userID))
		updatedPerms = append(updatedPerms, &iauthz.RepoPermissions{
			RepoID:    repoID,
			Perm:      p.Perm,
			UserIDs:   oldIDs,
			Provider:  authz.ProviderSourcegraph,
			UpdatedAt: updatedAt,
		})
	}

	if q, err = upsertRepoPermissionsBatchQuery(updatedPerms...); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute upsert repo permissions batch query")
	}

	// Load existing user permissions to be merged if any. Since we're doing union of permissions,
	// whatever we have already in the "repo_permissions" table is all valid thus we don't
	// need to do any clean up.
	up := &iauthz.UserPermissions{
		UserID:   userID,
		Perm:     p.Perm,
		Type:     p.Type,
		Provider: authz.ProviderSourcegraph,
	}
	var oldIDs *roaring.Bitmap
	vals, err = txs.load(ctx, loadUserPermissionsQuery(up, "FOR UPDATE"))
	if err != nil {
		if err != ErrPermsNotFound {
			return errors.Wrap(err, "load user permissions")
		}
		oldIDs = roaring.NewBitmap()
	} else {
		oldIDs = vals.ids
	}
	up.IDs = roaring.Or(oldIDs, p.IDs)

	up.UpdatedAt = txs.clock()
	if q, err = upsertUserPermissionsQuery(up); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return errors.Wrap(err, "execute upsert user permissions query")
	}

	// NOTE: Practically, we don't need to clean up "repo_pending_permissions" table because the value of "id" column
	// that is associated with this user will be invalidated automatically by deleting this row. Thus, we are able to
	// avoid database deadlocks with other methods (e.g. SetRepoPermissions, SetRepoPendingPermissions).
	if err = txs.execute(ctx, deleteUserPendingPermissionsQuery(p)); err != nil {
		return errors.Wrap(err, "execute delete user pending permissions query")
	}

	return nil
}

func upsertUserPermissionsQuery(p *iauthz.UserPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:upsertUserPermissionsQuery
INSERT INTO user_permissions
  (user_id, permission, object_type, object_ids, provider, updated_at)
VALUES
  (%s, %s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  user_permissions_perm_object_provider_unique
DO UPDATE SET
  object_ids = excluded.object_ids,
  updated_at = excluded.updated_at
`

	p.IDs.RunOptimize()
	ids, err := p.IDs.ToBytes()
	if err != nil {
		return nil, err
	}

	if p.UpdatedAt.IsZero() {
		return nil, ErrPermsUpdatedAtNotSet
	}

	return sqlf.Sprintf(
		format,
		p.UserID,
		p.Perm.String(),
		p.Type,
		ids,
		p.Provider,
		p.UpdatedAt.UTC(),
	), nil
}

func loadRepoPermissionsBatchQuery(repoIDs []uint32, perm authz.Perms, provider authz.ProviderType, lock string) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:loadRepoPermissionsBatchQuery
SELECT repo_id, user_ids
FROM repo_permissions
WHERE repo_id IN (%s)
AND permission = %s
AND provider = %s
`

	items := make([]*sqlf.Query, len(repoIDs))
	for i := range repoIDs {
		items[i] = sqlf.Sprintf("%d", repoIDs[i])
	}
	return sqlf.Sprintf(
		format+lock,
		sqlf.Join(items, ","),
		perm.String(),
		provider,
	)
}

func upsertRepoPermissionsBatchQuery(ps ...*iauthz.RepoPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:upsertRepoPermissionsBatchQuery
INSERT INTO repo_permissions
  (repo_id, permission, user_ids, provider, updated_at)
VALUES
  %s
ON CONFLICT ON CONSTRAINT
  repo_permissions_perm_provider_unique
DO UPDATE SET
  user_ids = excluded.user_ids,
  updated_at = excluded.updated_at
`

	items := make([]*sqlf.Query, len(ps))
	for i := range ps {
		ps[i].UserIDs.RunOptimize()
		ids, err := ps[i].UserIDs.ToBytes()
		if err != nil {
			return nil, err
		}

		if ps[i].UpdatedAt.IsZero() {
			return nil, ErrPermsUpdatedAtNotSet
		}

		items[i] = sqlf.Sprintf("(%s, %s, %s, %s, %s)",
			ps[i].RepoID,
			ps[i].Perm.String(),
			ids,
			ps[i].Provider,
			ps[i].UpdatedAt.UTC(),
		)
	}

	return sqlf.Sprintf(
		format,
		sqlf.Join(items, ","),
	), nil
}

func deleteUserPendingPermissionsQuery(p *iauthz.UserPendingPermissions) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/db/perms_store.go:deleteUserPendingPermissionsQuery
DELETE FROM user_pending_permissions
WHERE bind_id = %s
AND permission = %s
AND object_type = %s
`

	return sqlf.Sprintf(
		format,
		p.BindID,
		p.Perm.String(),
		p.Type,
	)
}

// ListPendingUsers returns a list of bind IDs who have pending permissions.
func (s *PermsStore) ListPendingUsers(ctx context.Context) (bindIDs []string, err error) {
	ctx, save := s.observe(ctx, "ListPendingUsers", "")
	defer save(&err)

	q := sqlf.Sprintf(`SELECT bind_id, object_ids FROM user_pending_permissions`)

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bindID string
		var ids []byte
		if err = rows.Scan(&bindID, &ids); err != nil {
			return nil, err
		}

		if len(ids) == 0 {
			continue
		}

		bm := roaring.NewBitmap()
		if err = bm.UnmarshalBinary(ids); err != nil {
			return nil, err
		} else if bm.GetCardinality() == 0 {
			continue
		}

		bindIDs = append(bindIDs, bindID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bindIDs, nil
}

func (s *PermsStore) execute(ctx context.Context, q *sqlf.Query) (err error) {
	ctx, save := s.observe(ctx, "execute", "")
	defer func() { save(&err, otlog.Object("q", q)) }()

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

// permsLoadValues contains return values of (*PermsStore).load method.
type permsLoadValues struct {
	id        int32           // An integer ID
	ids       *roaring.Bitmap // Bitmap of unmarshalled IDs
	updatedAt time.Time       // Last updated time of the row
}

// load is a generic method that scans three values from one database table row, these values must have
// types and be scanned in the order of int32, []byte and time.Time. In addition, it unmarshalles the
// []byte into a *roaring.Bitmap.
func (s *PermsStore) load(ctx context.Context, q *sqlf.Query) (*permsLoadValues, error) {
	var err error
	ctx, save := s.observe(ctx, "load", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		// One row is expected, return ErrPermsNotFound if no other errors occurred.
		err = rows.Err()
		if err == nil {
			err = ErrPermsNotFound
		}
		return nil, err
	}

	var id int32
	var ids []byte
	var updatedAt time.Time
	if err = rows.Scan(&id, &ids, &updatedAt); err != nil {
		return nil, err
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	vals := &permsLoadValues{
		id:        id,
		ids:       roaring.NewBitmap(),
		updatedAt: updatedAt,
	}
	if len(ids) == 0 {
		return vals, nil
	} else if err = vals.ids.UnmarshalBinary(ids); err != nil {
		return nil, err
	}

	return vals, nil
}

// batchLoadIDs runs the query and returns unmarshalled IDs with their corresponding object ID value.
func (s *PermsStore) batchLoadIDs(ctx context.Context, q *sqlf.Query) (map[int32]*roaring.Bitmap, error) {
	var err error
	ctx, save := s.observe(ctx, "batchLoadIDs", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	loaded := make(map[int32]*roaring.Bitmap)
	for rows.Next() {
		var objID int32
		var ids []byte
		if err = rows.Scan(&objID, &ids); err != nil {
			return nil, err
		}

		if len(ids) == 0 {
			continue
		}

		bm := roaring.NewBitmap()
		if err = bm.UnmarshalBinary(ids); err != nil {
			return nil, err
		}
		loaded[objID] = bm
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return loaded, nil
}

// tx begins a new transaction.
func (s *PermsStore) tx(ctx context.Context) (*sql.Tx, error) {
	switch t := s.db.(type) {
	case *sql.Tx:
		return t, nil
	case *sql.DB:
		tx, err := t.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		return tx, nil
	default:
		panic(fmt.Sprintf("can't open transaction with unknown implementation of dbutil.DB: %T", t))
	}
}

// Transact begins a new transaction and make a new PermsStore over it.
func (s *PermsStore) Transact(ctx context.Context) (*PermsStore, error) {
	if Mocks.Perms.Transact != nil {
		return Mocks.Perms.Transact(ctx)
	}

	tx, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	return NewPermsStore(tx, s.clock), nil
}

// inTx returns true if the current PermsStore wraps an underlying transaction.
func (s *PermsStore) inTx() bool {
	_, ok := s.db.(*sql.Tx)
	return ok
}

// Done commits the transaction if error is nil. Otherwise, rolls back the transaction.
func (s *PermsStore) Done(err *error) {
	if !s.inTx() {
		return
	}

	tx := s.db.(*sql.Tx)
	if err == nil || *err == nil {
		_ = tx.Commit()
	} else {
		_ = tx.Rollback()
	}
}

func (s *PermsStore) observe(ctx context.Context, family, title string) (context.Context, func(*error, ...otlog.Field)) {
	began := s.clock()
	tr, ctx := trace.New(ctx, "db.PermsStore."+family, title)

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

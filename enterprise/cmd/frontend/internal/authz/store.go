package authz

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
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Store is the unified interface for managing permissions explicitly in the database.
// It is concurrency-safe and maintains data consistency over the 'user_permissions',
// 'repo_permissions', 'user_pending_permissions', and 'repo_pending_permissions' tables.
type Store struct {
	db    dbutil.DB
	clock func() time.Time
}

// NewStore returns a new Store with given parameters.
func NewStore(db dbutil.DB, clock func() time.Time) *Store {
	return &Store{
		db:    db,
		clock: clock,
	}
}

// LoadUserPermissions loads stored user permissions into p. An error is returned
// when there are no valid permissions available.
func (s *Store) LoadUserPermissions(ctx context.Context, p *UserPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadUserPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	p.IDs, p.UpdatedAt, err = s.load(ctx, loadUserPermissionsQuery(p))
	return err
}

func loadUserPermissionsQuery(p *UserPermissions) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/store.go:loadUserPermissionsQuery
SELECT object_ids, updated_at
FROM user_permissions
WHERE user_id = %s
AND permission = %s
AND object_type = %s
AND provider = %s
`

	return sqlf.Sprintf(
		format,
		p.UserID,
		p.Perm.String(),
		p.Type,
		p.Provider,
	)
}

// LoadRepoPermissions loads stored repository permissions into p. An error is
// returned when there are no valid permissions available.
func (s *Store) LoadRepoPermissions(ctx context.Context, p *RepoPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	p.UserIDs, p.UpdatedAt, err = s.load(ctx, loadRepoPermissionsQuery(p))
	return err
}

func loadRepoPermissionsQuery(p *RepoPermissions) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:loadRepoPermissionsQuery
SELECT user_ids, updated_at
FROM repo_permissions
WHERE repo_id = %s
AND permission = %s
AND provider = %s
`

	return sqlf.Sprintf(
		format,
		p.RepoID,
		p.Perm.String(),
		p.Provider,
	)
}

// SetRepoPermissions performs a full update for p, new user IDs found in p will be upserted
// and user IDs no longer in p will be removed. This method updates both the user and
// repository permissions tables.
func (s *Store) SetRepoPermissions(ctx context.Context, p *RepoPermissions) (err error) {
	ctx, save := s.observe(ctx, "SetRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	// Open a transaction for update consistency.
	var tx *sqlTx
	if tx, err = s.tx(ctx); err != nil {
		return err
	}
	defer tx.commitOrRollback(&err)

	// Make another Store with this underlying transaction.
	txs := NewStore(tx, s.clock)

	// Retrieve currently stored user IDs of this repository.
	oldIDs, _, err := txs.load(ctx, loadRepoPermissionsQuery(p))
	if err != nil {
		return err
	}
	if oldIDs == nil {
		oldIDs = roaring.NewBitmap()
	}

	// Fisrt get the intersection (And), then use the intersection to compute diffs (AndNot)
	// with both the old and new sets to get user IDs to remove and to add respectively.
	common := roaring.And(oldIDs, p.UserIDs)
	removed := roaring.Xor(common, oldIDs)
	added := roaring.Xor(common, p.UserIDs)

	// Load stored user IDs of both added and removed.
	changedIDs := roaring.Or(added, removed).ToArray()

	// In case there is nothing to add or remove.
	if len(changedIDs) == 0 {
		return nil
	}

	q := loadUserPermissionsBatchQuery(changedIDs, p.Perm, PermRepos, p.Provider)
	loadedIDs, err := txs.batchLoadIDs(ctx, q)
	if err != nil {
		return err
	}

	// We have two sets of IDs that one needs to add, and the other needs to remove.
	updatedAt := txs.clock()
	updatedPerms := make([]*UserPermissions, 0, len(changedIDs))
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

		updatedPerms = append(updatedPerms, &UserPermissions{
			UserID:    userID,
			Perm:      p.Perm,
			Type:      PermRepos,
			IDs:       repoIDs,
			Provider:  p.Provider,
			UpdatedAt: updatedAt,
		})
	}

	if q, err = upsertUserPermissionsBatchQuery(updatedPerms...); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return err
	}

	p.UpdatedAt = updatedAt
	if q, err = upsertRepoPermissionsBatchQuery(p); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return err
	}

	return nil
}

func loadUserPermissionsBatchQuery(
	userIDs []uint32,
	perm authz.Perms,
	typ PermType,
	provider ProviderType,
) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/store.go:loadUserPermissionsBatchQuery
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
		format,
		sqlf.Join(items, ","),
		perm.String(),
		typ,
		provider,
	)
}

func upsertUserPermissionsBatchQuery(ps ...*UserPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:upsertUserPermissionsBatchQuery
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
			return nil, errors.New("UpdatedAt timestamp must be set")
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
// An error is returned when there are no pending permissions available.
func (s *Store) LoadUserPendingPermissions(ctx context.Context, p *UserPendingPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadUserPendingPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	q := loadUserPendingPermissionsQuery(p)
	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	if !rows.Next() {
		return rows.Err()
	}

	var ids []byte
	if err = rows.Scan(&p.ID, &ids, &p.UpdatedAt); err != nil {
		return err
	}

	if err = rows.Close(); err != nil {
		return err
	}

	p.IDs = roaring.NewBitmap()
	if len(ids) == 0 {
		return nil
	} else if err = p.IDs.UnmarshalBinary(ids); err != nil {
		return err
	}

	return nil
}

func loadUserPendingPermissionsQuery(p *UserPendingPermissions) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:loadUserPendingPermissionsQuery
SELECT id, object_ids, updated_at
FROM user_pending_permissions
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

// SetRepoPendingPermissions performs a full update for p with given bindIDs, new bind IDs
// found will be upserted and bind IDs no longer in bindIDs will be removed. This method
// updates both the user and repository pending permissions tables.
func (s *Store) SetRepoPendingPermissions(
	ctx context.Context,
	bindIDs []string,
	p *RepoPermissions,
) (err error) {
	ctx, save := s.observe(ctx, "SetRepoPendingPermissions", "")
	defer func() { save(&err, append(p.TracingFields(), otlog.String("bindIDs", strings.Join(bindIDs, ",")))...) }()

	// Open a transaction for update consistency.
	var tx *sqlTx
	if tx, err = s.tx(ctx); err != nil {
		return err
	}
	defer tx.commitOrRollback(&err)

	// Make another Store with this underlying transaction.
	txs := NewStore(tx, s.clock)

	updatedAt := txs.clock()
	p.UpdatedAt = updatedAt

	// Insert rows for bindIDs without one in the user_pending_permissions table.
	// This help guarantees rows of all bindIDs exist in next load step.
	q, err := insertUserPendingPermissionsBatchQuery(bindIDs, p)
	if err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return err
	}

	// Load all user pending permissions by bindIDs from the table.
	q = loadUserPendingPermissionsBatchQuery(bindIDs, p.Perm, PermRepos)
	bindIDSet, loadedIDs, err := txs.loadUserPendingPermissions(ctx, q)
	if err != nil {
		return err
	}

	// Make up p.UserIDs from the result set.
	for id := range bindIDSet {
		p.UserIDs.Add(uint32(id))
	}

	// Retrieve currently stored user IDs of this repository.
	oldIDs, _, err := txs.load(ctx, loadRepoPendingPermissionsQuery(p))
	if err != nil {
		return err
	}

	// Fisrt get the intersection (And), then use the intersection to compute diffs (AndNot)
	// with both the old and new sets to get user IDs to remove and to add respectively.
	common := roaring.And(oldIDs, p.UserIDs)
	removed := roaring.Xor(common, oldIDs)
	added := roaring.Xor(common, p.UserIDs)

	q = loadUserPendingPermissionsByIDBatchQuery(removed.ToArray(), p.Perm, PermRepos)
	idSet, loaded, err := txs.loadUserPendingPermissions(ctx, q)
	if err != nil {
		return err
	}

	// Merge toRemove data into the full sets.
	for id, bindID := range idSet {
		bindIDSet[id] = bindID
	}
	for id, set := range loaded {
		loadedIDs[id] = set
	}

	// We have two sets of IDs that one needs to add, and the other needs to remove.
	changedIDs := roaring.Or(added, removed).ToArray()

	// In case there is nothing to add or remove.
	if len(changedIDs) == 0 {
		return nil
	}

	updatedPerms := make([]*UserPendingPermissions, 0, len(bindIDSet))
	for _, id := range changedIDs {
		userID := int32(id)
		repoIDs := loadedIDs[userID]
		if repoIDs == nil {
			repoIDs = roaring.NewBitmap()
		}

		// It is guaranteed only one of OR conditions could be true at a time because
		// an id could only be contained by either toAdd or toRemove. This check is
		// needed to filter out the new rows of user_pending_permissions table that
		// were inserted previously, which already contain the p.RepoID upon insertion.
		updated := added.Contains(id) && repoIDs.CheckedAdd(uint32(p.RepoID)) ||
			removed.Contains(id) && repoIDs.CheckedRemove(uint32(p.RepoID))
		if !updated {
			continue
		}

		updatedPerms = append(updatedPerms, &UserPendingPermissions{
			BindID:    bindIDSet[userID],
			Perm:      p.Perm,
			Type:      PermRepos,
			IDs:       repoIDs,
			UpdatedAt: updatedAt,
		})
	}

	if len(updatedPerms) > 0 {
		if q, err = upsertUserPendingPermissionsBatchQuery(updatedPerms...); err != nil {
			return err
		} else if err = txs.execute(ctx, q); err != nil {
			return err
		}
	}

	if q, err = upsertRepoPendingPermissionsBatchQuery(p); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return err
	}

	return nil
}

func (s *Store) loadUserPendingPermissions(ctx context.Context, q *sqlf.Query) (
	bindIDSet map[int32]string,
	loaded map[int32]*roaring.Bitmap,
	err error,
) {
	ctx, save := s.observe(ctx, "loadUserPendingPermissions", "")
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

	return bindIDSet, loaded, rows.Close()
}

func insertUserPendingPermissionsBatchQuery(
	bindIDs []string,
	p *RepoPermissions,
) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:insertUserPendingPermissionsBatchQuery
INSERT INTO user_pending_permissions
  (bind_id, permission, object_type, object_ids, updated_at)
VALUES
  %s
ON CONFLICT ON CONSTRAINT
  user_pending_permissions_perm_object_unique
DO NOTHING
`

	ids := roaring.NewBitmap()
	ids.Add(uint32(p.RepoID))
	bs, err := ids.ToBytes()
	if err != nil {
		return nil, err
	}

	if p.UpdatedAt.IsZero() {
		return nil, errors.New("UpdatedAt timestamp must be set")
	}

	items := make([]*sqlf.Query, len(bindIDs))
	for i := range bindIDs {
		items[i] = sqlf.Sprintf("(%s, %s, %s, %s, %s)",
			bindIDs[i],
			p.Perm,
			PermRepos,
			bs,
			p.UpdatedAt.UTC(),
		)
	}

	return sqlf.Sprintf(
		format,
		sqlf.Join(items, ","),
	), nil
}

func loadUserPendingPermissionsBatchQuery(
	bindIDs []string,
	perm authz.Perms,
	typ PermType,
) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/store.go:loadUserPendingPermissionsBatchQuery
SELECT id, bind_id, object_ids, updated_at
FROM user_pending_permissions
WHERE bind_id IN (%s)
AND permission = %s
AND object_type = %s
`

	items := make([]*sqlf.Query, len(bindIDs))
	for i := range bindIDs {
		items[i] = sqlf.Sprintf("%s", bindIDs[i])
	}
	return sqlf.Sprintf(
		format,
		sqlf.Join(items, ","),
		perm.String(),
		typ,
	)
}

func loadRepoPendingPermissionsQuery(p *RepoPermissions) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:loadRepoPendingPermissionsQuery
SELECT user_ids, updated_at
FROM repo_pending_permissions
WHERE repo_id = %s
AND permission = %s
`

	return sqlf.Sprintf(
		format,
		p.RepoID,
		p.Perm.String(),
	)
}

func loadUserPendingPermissionsByIDBatchQuery(
	ids []uint32,
	perm authz.Perms,
	typ PermType,
) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/store.go:loadUserPendingPermissionsByIDBatchQuery
SELECT id, bind_id, object_ids, updated_at
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
		format,
		sqlf.Join(items, ","),
		perm.String(),
		typ,
	)
}

func upsertUserPendingPermissionsBatchQuery(ps ...*UserPendingPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:upsertUserPendingPermissionsBatchQuery
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
			return nil, errors.New("UpdatedAt timestamp must be set")
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

func upsertRepoPendingPermissionsBatchQuery(ps ...*RepoPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:upsertRepoPendingPermissionsBatchQuery
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
			return nil, errors.New("UpdatedAt timestamp must be set")
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
func (s *Store) GrantPendingPermissions(ctx context.Context, userID int32, p *UserPendingPermissions) (err error) {
	ctx, save := s.observe(ctx, "GrantPendingPermissions", "")
	defer func() { save(&err, append(p.TracingFields(), otlog.Object("userID", userID))...) }()

	// Open a transaction for update consistency.
	var tx *sqlTx
	if tx, err = s.tx(ctx); err != nil {
		return err
	}
	defer tx.commitOrRollback(&err)

	// Make another Store with this underlying transaction.
	txs := NewStore(tx, s.clock)

	if err = txs.LoadUserPendingPermissions(ctx, p); err != nil {
		return err
	}

	if p.IDs == nil || p.IDs.GetCardinality() == 0 {
		return nil
	}

	up := &UserPermissions{
		UserID:    userID,
		Perm:      p.Perm,
		Type:      p.Type,
		IDs:       p.IDs,
		Provider:  ProviderSourcegraph,
		UpdatedAt: txs.clock(),
	}
	var q *sqlf.Query
	if q, err = upsertUserPermissionsBatchQuery(up); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return err
	}

	// NOTE: We currently only have "repos" type, so avoid unnecessary type checking for now.
	ids := p.IDs.ToArray()

	// Batch query all repository permissions object IDs in one go.
	q = loadRepoPermissionsBatchQuery(ids, p.Perm, ProviderSourcegraph)
	loadedIDs, err := s.batchLoadIDs(ctx, q)
	if err != nil {
		return err
	}

	updatedAt := txs.clock()
	updatedPerms := make([]*RepoPermissions, 0, len(ids))
	for i := range ids {
		repoID := int32(ids[i])
		oldIDs := loadedIDs[repoID]
		if oldIDs == nil {
			oldIDs = roaring.NewBitmap()
		}

		if !oldIDs.CheckedAdd(uint32(userID)) {
			continue
		}

		updatedPerms = append(updatedPerms, &RepoPermissions{
			RepoID:    repoID,
			Perm:      p.Perm,
			UserIDs:   oldIDs,
			Provider:  ProviderSourcegraph,
			UpdatedAt: updatedAt,
		})
	}

	if q, err = upsertRepoPermissionsBatchQuery(updatedPerms...); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return err
	}

	if err = txs.execute(ctx, deleteUserPendingPermissionsQuery(p)); err != nil {
		return err
	}

	// Clean up repo pending permissions table.
	q = loadRepoPendingPermissionsBatchQuery(ids, p.Perm)
	loadedIDs, err = s.batchLoadIDs(ctx, q)
	if err != nil {
		return err
	}

	updatedAt = txs.clock()
	updatedPerms = make([]*RepoPermissions, 0, len(ids))
	for i := range ids {
		repoID := int32(ids[i])
		oldIDs := loadedIDs[repoID]
		if oldIDs == nil {
			oldIDs = roaring.NewBitmap()
		}

		if !oldIDs.CheckedRemove(uint32(userID)) {
			continue
		}

		updatedPerms = append(updatedPerms, &RepoPermissions{
			RepoID:    repoID,
			Perm:      p.Perm,
			UserIDs:   oldIDs,
			UpdatedAt: updatedAt,
		})
	}

	if q, err = upsertRepoPendingPermissionsBatchQuery(updatedPerms...); err != nil {
		return err
	} else if err = txs.execute(ctx, q); err != nil {
		return err
	}

	return nil
}

func loadRepoPermissionsBatchQuery(repoIDs []uint32, perm authz.Perms, provider ProviderType) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/store.go:loadRepoPermissionsBatchQuery
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
		format,
		sqlf.Join(items, ","),
		perm,
		provider,
	)
}

func upsertRepoPermissionsBatchQuery(ps ...*RepoPermissions) (*sqlf.Query, error) {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permissions.go:upsertRepoPermissionsBatchQuery
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
			return nil, errors.New("UpdatedAt timestamp must be set")
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

func deleteUserPendingPermissionsQuery(p *UserPendingPermissions) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/store.go:deleteUserPendingPermissionsQuery
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

func loadRepoPendingPermissionsBatchQuery(repoIDs []uint32, perm authz.Perms) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/store.go:loadRepoPendingPermissionsBatchQuery
SELECT repo_id, user_ids
FROM repo_pending_permissions
WHERE repo_id IN (%s)
AND permission = %s
`

	items := make([]*sqlf.Query, len(repoIDs))
	for i := range repoIDs {
		items[i] = sqlf.Sprintf("%d", repoIDs[i])
	}
	return sqlf.Sprintf(
		format,
		sqlf.Join(items, ","),
		perm,
	)
}

func (s *Store) execute(ctx context.Context, q *sqlf.Query) (err error) {
	ctx, save := s.observe(ctx, "execute", "")
	defer func() { save(&err, otlog.Object("q", q)) }()

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

// load runs the query and returns unmarshalled IDs and last updated time.
func (s *Store) load(ctx context.Context, q *sqlf.Query) (*roaring.Bitmap, time.Time, error) {
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
		return nil, time.Time{}, err
	}

	if !rows.Next() {
		return nil, time.Time{}, rows.Err()
	}

	var ids []byte
	var updatedAt time.Time
	if err = rows.Scan(&ids, &updatedAt); err != nil {
		return nil, time.Time{}, err
	}

	if err = rows.Close(); err != nil {
		return nil, time.Time{}, err
	}

	bm := roaring.NewBitmap()
	if len(ids) == 0 {
		return bm, time.Time{}, nil
	} else if err = bm.UnmarshalBinary(ids); err != nil {
		return nil, time.Time{}, err
	}

	return bm, updatedAt, nil
}

// batchLoadIDs runs the query and returns unmarshalled IDs with their corresponding object ID value.
func (s *Store) batchLoadIDs(ctx context.Context, q *sqlf.Query) (map[int32]*roaring.Bitmap, error) {
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

	if err = rows.Close(); err != nil {
		return nil, err
	}

	return loaded, nil
}

func (s *Store) tx(ctx context.Context) (*sqlTx, error) {
	switch t := s.db.(type) {
	case *sql.Tx:
		return &sqlTx{t}, nil
	case *sql.DB:
		tx, err := t.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		return &sqlTx{tx}, nil
	default:
		panic(fmt.Sprintf("can't open transaction with unknown implementation of dbutil.DB: %T", t))
	}
}

func (s *Store) observe(ctx context.Context, family, title string) (context.Context, func(*error, ...otlog.Field)) {
	began := s.clock()
	tr, ctx := trace.New(ctx, "authz.Store."+family, title)

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

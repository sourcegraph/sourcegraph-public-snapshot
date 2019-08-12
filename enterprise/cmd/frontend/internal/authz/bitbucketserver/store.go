package bitbucketserver

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbutil"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
)

// A store of Permissions with in-memory caching, safe for concurrent use.
//
// It implements and optimistic locking strategy similar to compare and swap
// so that many concurrent requests during an expiration don't overload the Bitbucket Server API.
//
// The second in-memory read-through layer avoids the round-trip and serialization
// costs associated with talking to Postgres, further speeding up the steady state
// operations.
type store struct {
	db      dbutil.DB
	hardTTL time.Duration
	softTTL time.Duration
	cache   *cache
	clock   func() time.Time
}

func newStore(db dbutil.DB, ttl time.Duration, clock func() time.Time, cache *cache) *store {
	return &store{
		db:      db,
		softTTL: ttl,
		hardTTL: 3 * 24 * time.Hour, // 3 days
		cache:   cache,
		clock:   clock,
	}
}

// PermissionsID uniquely identifies a set of Permissions.
type PermissionsID struct {
	UserID int32
	Perm   authz.Perms
	Type   string
}

// Permissions of a user to perform an action on the
// given set of object IDs of the defined type.
type Permissions struct {
	PermissionsID
	IDs       *roaring.Bitmap
	UpdatedAt time.Time
	Version   uint64
}

// Authorized returns the intersection of the given ids with
// the authorized ids.
func (p *Permissions) Authorized(repos []*types.Repo) []authz.RepoPerms {
	perms := make([]authz.RepoPerms, 0, len(repos))
	for _, r := range repos {
		if r.ID != 0 && p.IDs.Contains(uint32(r.ID)) {
			perms = append(perms, authz.RepoPerms{Repo: r, Perms: p.Perm})
		}
	}
	return perms
}

// LoadPermissions loads stored permissions into p, calling the given update closure
// to fetch updated permissions when they expire.
func (s *store) LoadPermissions(
	ctx context.Context,
	ps *Permissions,
	update updateFunc,
) (err error) {
	ctx, save := s.observe(ctx, "LoadPermissions", "")
	defer func() { save(ps, &err) }()

	if s == nil {
		return nil
	}

	if !s.cache.load(ps) {
		if err = s.load(ctx, ps); err != nil {
			return err
		}
	}

	now := s.clock()
	expired := !now.Before(ps.UpdatedAt.Add(s.softTTL))
	stale := !now.Before(ps.UpdatedAt.Add(s.hardTTL))
	updating := ps.Version%2 == 1

	if !updating && (expired || stale) {
		if _, err = s.update(ctx, ps, update); err != nil {
			return err
		}
	}

	if stale {
		return &StalePermissionsError{Permissions: ps}
	}

	return nil
}

// StalePermissionsError is returned by LoadPermissions when the stored
// permissions are stale (i.e. surpassed the hard-TTL before being updated).
// Callers should pass this error up to the user.
type StalePermissionsError struct {
	*Permissions
}

func (e StalePermissionsError) Error() string {
	return fmt.Sprintf("%s:%s permissions for user=%d are stale", e.Perm, e.Type, e.UserID)
}

func (s *store) load(ctx context.Context, ps *Permissions) (err error) {
	ctx, save := s.observe(ctx, "load", "")
	defer func() { save(ps, &err) }()

	q := s.loadQuery(ps)

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	if !rows.Next() {
		return rows.Err()
	}

	var ids []byte
	if err = rows.Scan(&ids, &ps.UpdatedAt, &ps.Version); err != nil {
		return err
	}

	if err = rows.Close(); err != nil {
		return err
	}

	ps.IDs = roaring.New()
	return ps.IDs.UnmarshalBinary(ids)
}

func (s *store) loadQuery(p *Permissions) *sqlf.Query {
	return sqlf.Sprintf(
		loadQueryFmtStr,
		p.UserID,
		p.Perm.String(),
		p.Type,
	)
}

const loadQueryFmtStr = `
-- source: enterprise/cmd/frontend/internal/authz/bitbucketserver/store.go:store.load
SELECT object_ids, updated_at, version
FROM user_permissions
WHERE user_id = %s AND permission = %s AND object_type = %s`

type updateFunc func(context.Context) (objectIDs []uint32, err error)

func (s *store) update(ctx context.Context, p *Permissions, update updateFunc) (ok bool, err error) {
	ctx, save := s.observe(ctx, "update", "")
	defer save(p, &err)

	p.Version++
	if ok, err = s.upsert(ctx, p); !ok || err != nil {
		return ok, err
	}

	go func(ps *Permissions) {
		for {
			if ok, err := s.refresh(ps, update); !ok || err != nil {
				log15.Error("bitbucketserver.authz.store.update", "error", err, "updated", ok)
				time.Sleep(10 * time.Second)
			}
			return
		}
	}(&Permissions{
		PermissionsID: p.PermissionsID,
		IDs:           p.IDs,
		UpdatedAt:     p.UpdatedAt,
		Version:       p.Version + 1,
	})

	return true, nil
}

func (s *store) refresh(ps *Permissions, update updateFunc) (ok bool, err error) {
	// Slow cache update operation, talks to the code host.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	ids, err := update(ctx)
	cancel()

	if err != nil {
		return false, err
	}

	ps.UpdatedAt = s.clock()
	ps.IDs = roaring.BitmapOf(ids...)

	if ok, err = s.upsert(ctx, ps); !ok || err != nil {
		return ok, err
	}

	s.cache.update(ps)

	return true, nil
}

func (s *store) upsert(ctx context.Context, p *Permissions) (ok bool, err error) {
	ctx, save := s.observe(ctx, "upsert", "")
	defer save(p, &err)

	var q *sqlf.Query
	if q, err = s.upsertQuery(p); err != nil {
		return false, err
	}

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return false, err
	}

	var userID uint32
	for rows.Next() {
		if err = rows.Scan(&userID); err != nil {
			return false, err
		}
	}

	if err = rows.Err(); err != nil {
		return false, err
	}

	return userID != 0, rows.Close()
}

func (s *store) upsertQuery(p *Permissions) (_ *sqlf.Query, err error) {
	var ids []byte
	if p.IDs != nil {
		ids, err = p.IDs.ToBytes()
		if err != nil {
			return nil, err
		}
	}

	return sqlf.Sprintf(
		upsertQueryFmtStr,
		p.UserID,
		p.Perm.String(),
		p.Type,
		ids,
		p.UpdatedAt.UTC(),
		p.Version,
	), nil
}

const upsertQueryFmtStr = `
-- source: enterprise/cmd/frontend/internal/authz/bitbucketserver/store.go:store.upsert
INSERT INTO user_permissions
  (user_id, permission, object_type, object_ids, updated_at, version)
VALUES
  (%s, %s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  user_permissions_perm_object_unique
DO UPDATE SET
  object_ids = excluded.object_ids,
  updated_at = excluded.updated_at,
  version    = excluded.version
WHERE
  user_permissions.version = excluded.version - 1
RETURNING user_id
`

func (s *store) observe(ctx context.Context, family, title string) (context.Context, func(*Permissions, *error)) {
	began := s.clock()
	tr, ctx := trace.New(ctx, "bitbucket.authz.store."+family, title)

	return ctx, func(ps *Permissions, err *error) {
		now := s.clock()
		took := now.Sub(began)

		fields := []otlog.Field{
			otlog.String("Duration", took.String()),
			otlog.Int32("Permissions.UserID", ps.UserID),
			otlog.String("Permissions.Perm", string(ps.Perm)),
			otlog.String("Permissions.Type", ps.Type),
		}

		if ps.IDs != nil {
			fields = append(fields,
				otlog.Uint64("Permissions.IDs.Count", ps.IDs.GetCardinality()),
				otlog.String("Permissions.UpdatedAt", ps.UpdatedAt.String()),
			)
		}

		tr.LogFields(fields...)

		success := err == nil || *err == nil
		if !success {
			tr.SetError(*err)
		}

		tr.Finish()
	}
}

// A store's in-memory read-through cache used in LoadPermissions.
type cache struct {
	mu    sync.RWMutex
	cache map[PermissionsID]*Permissions
}

func newCache(ttl time.Duration, clock func() time.Time) *cache {
	return &cache{cache: map[PermissionsID]*Permissions{}}
}

func (c *cache) load(ps *Permissions) (ok bool) {
	if c == nil {
		return false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.cache[ps.PermissionsID]
	if !ok {
		return false
	}

	*ps = *cached
	return true
}

func (c *cache) update(p *Permissions) {
	if c == nil {
		return
	}

	c.mu.Lock()
	c.cache[p.PermissionsID] = p
	c.mu.Unlock()
}

func (c *cache) delete(p PermissionsID) {
	c.mu.Lock()
	delete(c.cache, p)
	c.mu.Unlock()
}

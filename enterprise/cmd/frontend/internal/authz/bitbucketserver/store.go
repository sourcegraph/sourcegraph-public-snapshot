package bitbucketserver

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbutil"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
)

// A store of Permissions with in-memory caching, safe for concurrent use.
//
// It leverages Postgres row locking for concurrency control of cache fill events,
// so that many concurrent requests during an expiration don't overload the Bitbucket Server API.
//
// The second in-memory read-through layer avoids the round-trip and serialization
// costs associated with talking to Postgres, further speeding up the steady state
// operations.
type store struct {
	db    dbutil.DB
	ttl   time.Duration
	cache *cache
	clock func() time.Time
}

func newStore(db dbutil.DB, ttl time.Duration, clock func() time.Time, cache *cache) *store {
	return &store{
		db:    db,
		ttl:   ttl,
		cache: cache,
		clock: clock,
	}
}

// Permissions of a user to perform an action on the
// given set of object IDs of the defined type.
type Permissions struct {
	UserID    int32
	Perm      authz.Perm
	Type      string
	IDs       *roaring.Bitmap
	UpdatedAt time.Time
}

// Authorized returns the intersection of the given ids with
// the authorized ids.
func (p *Permissions) Authorized(repos map[authz.Repo]struct{}) map[api.RepoName]map[authz.Perm]bool {
	perms := make(map[api.RepoName]map[authz.Perm]bool, len(repos))
	for r := range repos {
		if r.ID != 0 && p.IDs.Contains(uint32(r.ID)) {
			perms[r.RepoName] = map[authz.Perm]bool{p.Perm: true}
		}
	}
	return perms
}

// LoadPermissions loads stored permissions into p, calling the given update closure
// to fetch updated permissions when they expire.
func (s *store) LoadPermissions(
	ctx context.Context,
	p **Permissions,
	update func() (objectIDs []uint32, _ error),
) (err error) {
	ctx, save := s.observe(ctx, "LoadPermissions", "")
	defer func() { save(*p, &err) }()

	if s == nil {
		return nil
	}

	// Do we have unexpired permissions cached in-memory?
	if s.cache.load(p) {
		return nil // Yes, in-memory cache hit!
	}

	// Nope, in-memory cache miss. Sad. Do we have unexpired
	// permissions stored in the Postgres user_permissions table?
	// Let's find out...

	ps := *p
	stored := &Permissions{
		UserID: ps.UserID,
		Perm:   ps.Perm,
		Type:   ps.Type,
	}

	// Load the permissions with a read lock.
	if err = s.load(ctx, stored, "SHARE"); err != nil {
		return err
	}

	// Did these permissions expire? If so, it's time for a slow
	// cache filling event.
	now := s.clock()
	if expired := !now.Before(stored.UpdatedAt.Add(s.ttl)); expired {
		if err = s.update(ctx, stored, update); err != nil {
			return err
		}
	}

	// At this point we have fresh permissions in `stored`, either because
	// we loaded them or updated them. So we store them in our read-through
	// cache and update the passed Permissions pointer!
	s.cache.update(stored)
	*p = stored

	return nil
}

func (s *store) load(ctx context.Context, p *Permissions, lock string) (err error) {
	ctx, save := s.observe(ctx, "load", lock)
	defer save(p, &err)

	q := s.loadQuery(p, lock)

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	if !rows.Next() {
		return rows.Err()
	}

	var ids []byte
	if err = rows.Scan(&ids, &p.UpdatedAt); err != nil {
		return err
	}

	if err = rows.Close(); err != nil {
		return err
	}

	if p.IDs == nil {
		p.IDs = roaring.NewBitmap()
	} else {
		p.IDs.Clear()
	}

	return p.IDs.UnmarshalBinary(ids)
}

func (s *store) loadQuery(p *Permissions, lock string) *sqlf.Query {
	return sqlf.Sprintf(
		loadQueryFmtStr+lock,
		p.UserID,
		p.Perm,
		p.Type,
	)
}

const loadQueryFmtStr = `
-- source: enterprise/cmd/frontend/internal/authz/bitbucketserver/store.go:store.load
SELECT object_ids, updated_at
FROM user_permissions
WHERE user_id = %s AND permission = %s AND object_type = %s
FOR `

func (s *store) update(ctx context.Context, p *Permissions, update func() ([]uint32, error)) (err error) {
	ctx, save := s.observe(ctx, "update", "")
	defer save(p, &err)

	// Open a transaction for concurrency control.
	var tx *sql.Tx
	if tx, err = s.tx(ctx); err != nil {
		return err
	}

	// Either rollback or commit this transaction, when we're done.
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Make another store with this underlying transaction.
	txs := store{db: tx, clock: s.clock}

	// We're here because we need to update our permissions. In order
	// to prevent multiple concurrent (and distributed) cache fills,
	// which would hammer the upstream code host, we acquire an exclusive
	// write lock over the Postgres row for these permissions.
	//
	// This means other processes will block until the cache fill succeeds,
	// which is the desired behaviour.
	if err = txs.load(ctx, p, "UPDATE"); err != nil {
		return err
	}

	// We have acquired an exclusive write lock over these permissions,
	// but maybe another process already finished the cache fill event.
	// If so, we don't have to update it again until it expires.
	now := s.clock()
	if now.Before(p.UpdatedAt.Add(s.ttl)) { // Valid!
		return nil // Updated by another process!
	}

	// Slow cache update operation, talks to the code host.
	var ids []uint32
	if ids, err = update(); err != nil {
		return err
	}

	// Create a set of the given ids and update the timestamp.
	p.IDs = roaring.BitmapOf(ids...)
	p.UpdatedAt = now

	// Write back the updated permissions to the database.
	return txs.upsert(ctx, p)
}

func (s *store) tx(ctx context.Context) (*sql.Tx, error) {
	switch t := s.db.(type) {
	case *sql.Tx:
		return t, nil
	case *sql.DB:
		return t.BeginTx(ctx, nil)
	default:
		panic("can't open transaction with unknown implementation of dbutil.DB")
	}
}

func (s *store) upsert(ctx context.Context, p *Permissions) (err error) {
	ctx, save := s.observe(ctx, "upsert", "")
	defer save(p, &err)

	var q *sqlf.Query
	if q, err = s.upsertQuery(p); err != nil {
		return err
	}

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	return rows.Close()
}

func (s *store) upsertQuery(p *Permissions) (*sqlf.Query, error) {
	ids, err := p.IDs.ToBytes()
	if err != nil {
		return nil, err
	}

	if p.UpdatedAt.IsZero() {
		return nil, errors.New("UpdatedAt timestamp must be set")
	}

	return sqlf.Sprintf(
		upsertQueryFmtStr,
		p.UserID,
		p.Perm,
		p.Type,
		ids,
		p.UpdatedAt.UTC(),
	), nil
}

const upsertQueryFmtStr = `
-- source: enterprise/cmd/frontend/internal/authz/bitbucketserver/store.go:store.upsert
INSERT INTO user_permissions
  (user_id, permission, object_type, object_ids, updated_at)
VALUES
  (%s, %s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT
  user_permissions_perm_object_unique
DO UPDATE SET
  object_ids = excluded.object_ids,
  updated_at = excluded.updated_at
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
	cache map[cacheKey]*Permissions
	ttl   time.Duration
	clock func() time.Time
}

func newCache(ttl time.Duration, clock func() time.Time) *cache {
	return &cache{
		cache: map[cacheKey]*Permissions{},
		ttl:   ttl,
		clock: clock,
	}
}

type cacheKey struct {
	UserID int32
	Perm   authz.Perm
	Type   string
}

// load sets the given Permissions pointer with a matching cached
// Permissions. If no cached Permissions are found or if they are
// now expired,
func (c *cache) load(p **Permissions) (hit bool) {
	if c == nil {
		return false
	}

	k := newCacheKey(*p)

	c.mu.RLock()
	e, ok := c.cache[k]
	c.mu.RUnlock()

	now := c.clock()
	if hit = ok && now.Before(e.UpdatedAt.Add(c.ttl)); hit {
		*p = e
	}

	return
}

func (c *cache) update(p *Permissions) {
	if c == nil {
		return
	}

	k := newCacheKey(p)
	c.mu.Lock()
	c.cache[k] = p
	c.mu.Unlock()
}

func newCacheKey(p *Permissions) cacheKey {
	if p.Perm == "" {
		panic("empty Perm")
	}

	if p.Type == "" {
		panic("empty Type")
	}

	return cacheKey{
		UserID: p.UserID,
		Perm:   p.Perm,
		Type:   p.Type,
	}
}

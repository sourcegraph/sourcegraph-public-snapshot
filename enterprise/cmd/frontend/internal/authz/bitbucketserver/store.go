package bitbucketserver

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbutil"
)

// A store of Permissions with in-memory caching, safe for concurrent use.
//
// We use Postgres as a persistent cache with ACID guarantees and row-level locking
// for concurrency control of cache filling events on expiration of entries in order
// to avoid a thundering herd against the upstream code host when a single Sourcegraph
// user (bot or not) is performing many concurrent requests.
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
	if s == nil {
		return nil
	}

	if s.cache.load(p) {
		return nil
	}

	ps := *p
	stored := &Permissions{
		UserID: ps.UserID,
		Perm:   ps.Perm,
		Type:   ps.Type,
	}

	tx, err := s.tx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	txs := store{db: tx, clock: s.clock}
	now := s.clock()

	if err = txs.load(ctx, stored, false); err != nil {
		return err
	}

	if !now.Before(stored.UpdatedAt.Add(s.ttl)) {
		if err = txs.update(ctx, stored, update); err != nil {
			return err
		}
	}

	s.cache.update(stored)
	*p = stored

	return nil
}

func (s *store) load(ctx context.Context, p *Permissions, update bool) error {
	q := s.loadQuery(p, update)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
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
	}

	return p.IDs.UnmarshalBinary(ids)
}

func (s *store) loadQuery(p *Permissions, update bool) *sqlf.Query {
	var lock string
	if update {
		lock = "FOR UPDATE" // Exclusive write lock (one writer)
	} else {
		lock = "FOR SHARE" // Shared read lock (many readers)
	}

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
`

func (s *store) update(ctx context.Context, p *Permissions, update func() ([]uint32, error)) (err error) {
	if err = s.load(ctx, p, true); err != nil {
		return err
	}

	now := s.clock()
	if now.Before(p.UpdatedAt.Add(s.ttl)) {
		return nil // Updated by another tx
	}

	// Slow cache update operation, synchronized via a serializable transaction.
	var ids []uint32
	if ids, err = update(); err != nil {
		return err
	}

	p.IDs = roaring.BitmapOf(ids...)
	p.UpdatedAt = s.clock()

	if err = s.upsert(ctx, p); err != nil {
		return err
	}

	return nil
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

func (s *store) upsert(ctx context.Context, p *Permissions) error {
	q, err := s.upsertQuery(p)
	if err != nil {
		return err
	}

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
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

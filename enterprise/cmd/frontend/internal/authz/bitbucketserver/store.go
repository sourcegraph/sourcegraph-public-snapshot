package bitbucketserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/segmentio/fasthash/fnv1"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbutil"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	"gopkg.in/inconshreveable/log15.v2"
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
	db dbutil.DB
	// Duration after which a given user's cached permissions SHOULD be updated.
	// Previously cached permissions can still be used.
	ttl time.Duration
	// Duration after which a given user's cached permissions MUST be updated.
	// Previously cached permissions can no longer be used.
	hardTTL time.Duration
	cache   *cache
	clock   func() time.Time
	block   bool // Perform blocking updates if true.
	updates chan *Permissions
}

func newStore(db dbutil.DB, ttl, hardTTL time.Duration, clock func() time.Time, cache *cache) *store {
	if hardTTL < ttl {
		hardTTL = ttl
	}

	return &store{
		db:      db,
		ttl:     ttl,
		hardTTL: hardTTL,
		cache:   cache,
		clock:   clock,
	}
}

// Permissions of a user to perform an action on the
// given set of object IDs of the defined type.
type Permissions struct {
	UserID    int32
	Perm      authz.Perms
	Type      string
	IDs       *roaring.Bitmap
	UpdatedAt time.Time
}

// Expired returns true if these Permissions have elapsed the given ttl.
func (p *Permissions) Expired(ttl time.Duration, now time.Time) bool {
	return !now.Before(p.UpdatedAt.Add(ttl))
}

// Authorized returns the intersection of the given ids with
// the authorized ids.
func (p *Permissions) Authorized(repos []*types.Repo) []authz.RepoPerms {
	perms := make([]authz.RepoPerms, 0, len(repos))
	for _, r := range repos {
		if r.ID != 0 && p.IDs != nil && p.IDs.Contains(uint32(r.ID)) {
			perms = append(perms, authz.RepoPerms{Repo: r, Perms: p.Perm})
		}
	}
	return perms
}

func (p *Permissions) tracingFields() []otlog.Field {
	fs := []otlog.Field{
		otlog.Int32("Permissions.UserID", p.UserID),
		otlog.String("Permissions.Perm", string(p.Perm)),
		otlog.String("Permissions.Type", p.Type),
	}

	if p.IDs != nil {
		fs = append(fs,
			otlog.Uint64("Permissions.IDs.Count", p.IDs.GetCardinality()),
			otlog.String("Permissions.UpdatedAt", p.UpdatedAt.String()),
		)
	}

	return fs
}

// DefaultHardTTL is the default hard TTL used in the permissions store, after which
// cached permissions for a given user MUST be updated, and previously cached permissions
// can no longer be used, resulting in a call to LoadPermissions returning a StalePermissionsError.
const DefaultHardTTL = 3 * 24 * time.Hour

// PermissionsUpdateFunc fetches updated permissions from a source of truth,
// returning the correspondent external ids of those objects for which the
// authenticated user has permissions for, as well as the code host associated
// with those objects.
type PermissionsUpdateFunc func(context.Context) (
	ids []uint32,
	codeHost *extsvc.CodeHost,
	err error,
)

// LoadPermissions loads stored permissions into p, calling the given update closure
// to asynchronously fetch updated permissions when they expire. When there are no
// valid permissions available (i.e. the first time a user needs them), an error is
// returned.
//
// Callers must NOT mutate the resulting Permissions pointer. It's shared across go-routines
// and it's meant to be read-only. Any write to it is not thread-safe.
func (s *store) LoadPermissions(
	ctx context.Context,
	p **Permissions,
	update PermissionsUpdateFunc,
) (err error) {
	if s == nil || p == nil || *p == nil {
		return nil
	}

	ctx, save := s.observe(ctx, "LoadPermissions", "")
	defer func() { save(&err, (*p).tracingFields()...) }()

	now := s.clock()

	// Do we have valid permissions cached in-memory?
	if !s.cache.load(p) || (*p).Expired(s.ttl, now) {
		// Try to load updated permissions from the database.
		if err = s.load(ctx, *p); err != nil {
			return err
		} else if !(*p).Expired(s.ttl, now) { // Cache them if they're not expired.
			s.cache.update(*p)
		}
	}

	if !(*p).Expired(s.ttl, now) { // Are these permissions still valid?
		return nil
	}

	return s.UpdatePermissions(ctx, *p, update)
}

// UpdatePermissions updates the given Permissions, calling the update function
// to fetch fresh data from the source of truth.
func (s *store) UpdatePermissions(
	ctx context.Context,
	p *Permissions,
	update PermissionsUpdateFunc,
) (err error) {
	ctx, save := s.observe(ctx, "UpdatePermissions", "")
	defer func() { save(&err, p.tracingFields()...) }()

	now := s.clock()
	expired := *p
	expired.IDs = nil

	if !s.block { // Non blocking code path
		go func(expired *Permissions) {
			err := s.update(ctx, expired, update)
			if err != nil && err != errLockNotAvailable {
				log15.Error("bitbucketserver.authz.store.UpdatePermissions", "error", err)
			}
		}(&expired)

		// No valid permissions available yet or hard TTL expired.
		if p.UpdatedAt.IsZero() || p.Expired(s.hardTTL, now) {
			return &StalePermissionsError{Permissions: p}
		}

		return nil
	}

	// Blocking code path
	switch err = s.update(ctx, &expired, update); {
	case err == nil:
	case err == errLockNotAvailable:
		if p.Expired(s.hardTTL, now) {
			return &StalePermissionsError{Permissions: p}
		}
	default:
		return err
	}

	*p = expired
	return nil
}

// StalePermissionsError is returned by LoadPermissions when the stored
// permissions are stale (e.g. the first time a user needs them and they haven't
// been fetched yet). Callers should pass this error up to the user and show it
// in the UI.
type StalePermissionsError struct {
	*Permissions
}

// Error implements the error interface.
func (e StalePermissionsError) Error() string {
	return fmt.Sprintf("%s:%s permissions for user=%d are stale and being updated", e.Perm, e.Type, e.UserID)
}

var errLockNotAvailable = errors.New("lock not available")

// lock uses Postgres advisory locks to acquire an exclusive lock over the
// given Permissions. Concurrent processes that call this method while a lock is
// already held by another process will have errLockNotAvailable returned.
func (s *store) lock(ctx context.Context, p *Permissions) (err error) {
	ctx, save := s.observe(ctx, "lock", "")
	defer func() { save(&err, p.tracingFields()...) }()

	if _, ok := s.db.(*sql.Tx); !ok {
		return errors.Errorf("store.lock must be called inside a transaction")
	}

	q := lockQuery(p)

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	if !rows.Next() {
		return rows.Err()
	}

	locked := false
	if err = rows.Scan(&locked); err != nil {
		return err
	}

	if err = rows.Close(); err != nil {
		return err
	}

	if !locked {
		return errLockNotAvailable
	}

	return nil
}

var lockNamespace = int32(fnv1.HashString32("perms"))

func lockQuery(p *Permissions) *sqlf.Query {
	// Postgres advisory lock ids are a global namespace within one database.
	// It's very unlikely that another part of our application uses a lock
	// namespace identicaly to this one. It's equally unlikely that there are
	// lock id conflicts for different permissions, but if it'd happen, no safety
	// guarantees would be violated, since those two different users would simply
	// have to wait on the other's update to finish, using stale permissions until
	// it would.
	lockID := int32(fnv1.HashString32(fmt.Sprintf("%d:%s:%s", p.UserID, p.Perm, p.Type)))
	return sqlf.Sprintf(
		lockQueryFmtStr,
		lockNamespace,
		lockID,
	)
}

const lockQueryFmtStr = `
-- source: enterprise/cmd/frontend/internal/authz/bitbucketserver/store.go:store.lock
SELECT pg_try_advisory_xact_lock(%s, %s)
`

func (s *store) load(ctx context.Context, p *Permissions) (err error) {
	ctx, save := s.observe(ctx, "load", "")
	defer func() { save(&err, p.tracingFields()...) }()

	q := loadQuery(p)

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

	if p.IDs = roaring.NewBitmap(); len(ids) == 0 {
		return nil
	}

	return p.IDs.UnmarshalBinary(ids)
}

func loadRepoIDsQuery(c *extsvc.CodeHost, externalIDs []uint32) (*sqlf.Query, error) {
	ids, err := json.Marshal(externalIDs)
	if err != nil {
		return nil, err
	}

	return sqlf.Sprintf(
		loadRepoIDsQueryFmtStr,
		c.ServiceType,
		c.ServiceID,
		ids,
	), nil
}

const loadRepoIDsQueryFmtStr = `
-- source: enterprise/cmd/frontend/internal/authz/bitbucketserver/store.go:store.loadRepoIDs
SELECT id FROM repo
WHERE external_service_type = %s AND external_service_id = %s
AND external_id IN (SELECT jsonb_array_elements_text(%s))
ORDER BY id ASC
`

func (s *store) loadRepoIDs(ctx context.Context, c *extsvc.CodeHost, externalIDs []uint32) (ids *roaring.Bitmap, err error) {
	ctx, save := s.observe(ctx, "loadRepoIDs", "")
	defer func() {
		fs := []otlog.Field{otlog.Int("externalIDs.count", len(externalIDs))}
		if ids != nil {
			fs = append(fs, otlog.Uint64("repoIDs.count", ids.GetCardinality()))
		}
		save(&err, fs...)
	}()

	var q *sqlf.Query
	if q, err = loadRepoIDsQuery(c, externalIDs); err != nil {
		return nil, err
	}

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}

	ns := make([]uint32, 0, len(externalIDs))
	for rows.Next() {
		var id uint32
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ns = append(ns, id)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if err = rows.Close(); err != nil {
		return nil, err
	}

	ids = roaring.BitmapOf(ns...)
	ids.RunOptimize()
	return ids, nil
}

func loadQuery(p *Permissions) *sqlf.Query {
	return sqlf.Sprintf(
		loadQueryFmtStr,
		p.UserID,
		p.Perm.String(),
		p.Type,
	)
}

const loadQueryFmtStr = `
-- source: enterprise/cmd/frontend/internal/authz/bitbucketserver/store.go:store.load
SELECT object_ids, updated_at
FROM user_permissions
WHERE user_id = %s AND permission = %s AND object_type = %s
`

func (s *store) update(ctx context.Context, p *Permissions, update PermissionsUpdateFunc) (err error) {
	ctx, save := s.observe(ctx, "update", "")
	defer func() { save(&err, p.tracingFields()...) }()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	// Open a transaction for concurrency control.
	var tx *sql.Tx
	if tx, err = s.tx(ctx); err != nil {
		return err
	}

	expired := false

	// Either rollback or commit this transaction, when we're done.
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else if err = tx.Commit(); err == nil &&
			s.updates != nil && expired {
			s.updates <- p
		}
	}()

	// Make another store with this underlying transaction.
	txs := store{db: tx, clock: s.clock}

	// We're here because we need to update our permissions. In order
	// to prevent multiple concurrent (and distributed) cache fills,
	// which would hammer the upstream code host, we acquire an exclusive
	// lock over the Postgres row for these permissions. This lock gets
	// automatically released when the transaction finishes.
	//
	// If another processes is updating these permissions, we abort and return
	// stale data.
	if err = txs.lock(ctx, p); err != nil {
		return err
	}

	// We have acquired an exclusive lock over these permissions,
	// but maybe another process already finished the cache fill event.
	// If so, we don't have to update it again until it expires.
	if err = txs.load(ctx, p); err != nil {
		return err
	}

	now := s.clock()
	if expired = p.Expired(s.ttl, now); !expired { // Valid!
		return nil // Permissions were updated by another process.
	}

	// Slow cache update operation, talks to the code host.
	var (
		externalIDs []uint32
		c           *extsvc.CodeHost
	)

	if externalIDs, c, err = update(ctx); err != nil {
		return err
	}

	p.IDs, err = s.loadRepoIDs(ctx, c, externalIDs)
	if err != nil {
		return err
	}

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
	defer func() { save(&err, p.tracingFields()...) }()

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
		p.Perm.String(),
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

func (s *store) observe(ctx context.Context, family, title string) (context.Context, func(*error, ...otlog.Field)) {
	began := s.clock()
	tr, ctx := trace.New(ctx, "bitbucket.authz.store."+family, title)

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

// A store's in-memory read-through cache used in LoadPermissions.
type cache struct {
	mu    sync.RWMutex
	cache map[cacheKey]*Permissions
}

func newCache() *cache {
	return &cache{cache: map[cacheKey]*Permissions{}}
}

type cacheKey struct {
	UserID int32
	Perm   authz.Perms
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
	e, hit := c.cache[k]
	c.mu.RUnlock()

	if hit {
		*p = e
	}

	return hit
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
	if p.Type == "" {
		panic("empty Type")
	}

	return cacheKey{
		UserID: p.UserID,
		Perm:   p.Perm,
		Type:   p.Type,
	}
}

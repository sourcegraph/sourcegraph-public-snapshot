package permsstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var clock = func() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }

// Store is the unified interface for managing permissions explicitly in the database.
// It is concurrent-safe and maintains the data consistency over 'user_permissions',
// 'repo_permissions' and 'user_pending_permissions' tables.
var Store = NewStore(dbconn.Global, clock)

type store struct {
	db    dbutil.DB
	clock func() time.Time
}

// NewStore initializes a new store with given parameters.
func NewStore(db dbutil.DB, clock func() time.Time) *store {
	return &store{
		db:    db,
		clock: clock,
	}
}

// LoadUserPermissions loads stored user permissions into p. An error is returned
// when there are no valid permissions available.
func (s *store) LoadUserPermissions(ctx context.Context, p *UserPermissions) (err error) {
	if p == nil {
		return nil
	}

	ctx, save := s.observe(ctx, "LoadUserPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	p.IDs, p.UpdatedAt, err = s.loadIDs(ctx, p.LoadQuery())
	return err
}

// LoadRepoPermissions loads stored repository permissions into p. An error is
// returned when there are no valid permissions available.
func (s *store) LoadRepoPermissions(ctx context.Context, p *RepoPermissions) (err error) {
	if p == nil {
		return nil
	}

	ctx, save := s.observe(ctx, "LoadRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	p.IDs, p.UpdatedAt, err = s.loadIDs(ctx, p.loadQuery())
	return err
}

// LoadPendingPermissions loads stored pending user permissions into p. An
// error is returned when there are no pending permissions available.
func (s *store) LoadPendingPermissions(ctx context.Context, p *PendingPermissions) (err error) {
	if p == nil {
		return nil
	}

	ctx, save := s.observe(ctx, "LoadPendingPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	p.IDs, p.UpdatedAt, err = s.loadIDs(ctx, p.loadQuery())
	return err
}

// UpsertRepoPermissions stores new user IDs found in p to the permissions tables.
func (s *store) UpsertRepoPermissions(ctx context.Context, p *RepoPermissions) (err error) {
	_, save := s.observe(ctx, "UpsertRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	// Retrieve currently stored user IDs of this repository.
	var userIDs *roaring.Bitmap
	userIDs, _, err = s.loadIDs(ctx, p.loadQuery())
	if err != nil {
		return err
	}

	// Open a transaction for data consistency.
	var tx *Tx
	if tx, err = s.Tx(ctx); err != nil {
		return err
	}
	defer func() {
		// We need to use closure for a reference to the err.
		tx.CommitOrRollback(err)
	}()

	// Make another store with this underlying transaction.
	txs := &store{db: tx, clock: s.clock}

	p.UpdatedAt = txs.clock()

	var q *sqlf.Query
	if q, err = p.upsertQuery(); err != nil {
		return err
	}

	if err = txs.upsert(ctx, q); err != nil {
		return err
	}

	// Iterate over the new set of user IDs and upsert ones that do not exist in database.
	iter := p.IDs.Iterator()
	for iter.HasNext() {
		id := iter.Next()
		if userIDs.Contains(id) {
			continue
		}

		up := &UserPermissions{
			UserID:   int32(id),
			Perm:     p.Perm,
			Type:     PermRepos,
			Provider: p.Provider,
		}
		up.IDs, _, err = txs.loadIDs(ctx, up.LoadQuery())
		if err != nil {
			return err
		}

		if up.IDs.Contains(uint32(p.RepoID)) {
			continue
		}

		up.IDs.Add(uint32(p.RepoID))
		up.IDs.RunOptimize()
		up.UpdatedAt = txs.clock()

		if q, err = up.UpsertQuery(); err != nil {
			return err
		} else if err = txs.upsert(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

// UpsertPendingPermissions stores the repoID to the pending permissions table
// with given bindIDs.
func (s *store) UpsertPendingPermissions(
	ctx context.Context,
	bindIDs []string,
	perm authz.Perms,
	typ PermType,
	repoID int32,
) (err error) {
	_, save := s.observe(ctx, "UpsertPendingPermissions", "")
	defer func() {
		save(&err,
			otlog.Object("bindIDs", bindIDs),
			otlog.String("perm", string(perm)),
			otlog.String("type", string(typ)),
			otlog.Int32("repoID", repoID),
		)
	}()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	// Open a transaction for data consistency.
	var tx *Tx
	if tx, err = s.Tx(ctx); err != nil {
		return err
	}
	defer func() {
		// We need to use closure for a reference to the err.
		tx.CommitOrRollback(err)
	}()

	// Make another store with this underlying transaction.
	txs := &store{db: tx, clock: s.clock}

	var pp *PendingPermissions
	var q *sqlf.Query
	for _, bindID := range bindIDs {
		pp = &PendingPermissions{
			BindID: bindID,
			Perm:   perm,
			Type:   typ,
		}
		pp.IDs, _, err = txs.loadIDs(ctx, pp.loadQuery())
		if err != nil {
			return err
		}

		if pp.IDs.Contains(uint32(repoID)) {
			continue
		}

		pp.IDs.Add(uint32(repoID))
		pp.IDs.RunOptimize()
		pp.UpdatedAt = txs.clock()
		if q, err = pp.upsertQuery(); err != nil {
			return err
		} else if err = txs.upsert(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

// RemoveRepoPermissions removes user IDs found in p from the permissions tables.
func (s *store) RemoveRepoPermissions(ctx context.Context, p *RepoPermissions) (err error) {
	_, save := s.observe(ctx, "RemoveRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	// Retrieve currently stored user IDs of this repository.
	var userIDs *roaring.Bitmap
	userIDs, _, err = s.loadIDs(ctx, p.loadQuery())
	if err != nil {
		return err
	}

	// Iterate over the delete set of user IDs and remove them from the existing set.
	iter := p.IDs.Iterator()
	for iter.HasNext() {
		userIDs.Remove(iter.Next())
	}
	p.IDs = userIDs
	p.IDs.RunOptimize()
	p.UpdatedAt = s.clock()

	var q *sqlf.Query
	if q, err = p.upsertQuery(); err != nil {
		return err
	}

	if err = s.upsert(ctx, q); err != nil {
		return err
	}

	return nil
}

// RemovePendingPermissions removes the repoID from the pending permissions table
// with given bindIDs.
func (s *store) RemovePendingPermissions(
	ctx context.Context,
	bindIDs []string,
	perm authz.Perms,
	typ PermType,
	repoID int32,
) (err error) {
	_, save := s.observe(ctx, "RemovePendingPermissions", "")
	defer func() {
		save(&err,
			otlog.Object("bindIDs", bindIDs),
			otlog.String("perm", string(perm)),
			otlog.String("type", string(typ)),
			otlog.Int32("repoID", repoID),
		)
	}()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	// Open a transaction for data consistency.
	var tx *Tx
	if tx, err = s.Tx(ctx); err != nil {
		return err
	}
	defer func() {
		// We need to use closure for a reference to the err.
		tx.CommitOrRollback(err)
	}()

	// Make another store with this underlying transaction.
	txs := &store{db: tx, clock: s.clock}

	var pp *PendingPermissions
	var q *sqlf.Query
	for _, bindID := range bindIDs {
		pp = &PendingPermissions{
			BindID: bindID,
			Perm:   perm,
			Type:   typ,
		}
		pp.IDs, _, err = txs.loadIDs(ctx, pp.loadQuery())
		if err != nil {
			return err
		}

		if !pp.IDs.Contains(uint32(repoID)) {
			continue
		}
		pp.IDs.Remove(uint32(repoID))
		pp.IDs.RunOptimize()
		pp.UpdatedAt = txs.clock()

		if q, err = pp.upsertQuery(); err != nil {
			return err
		} else if err = txs.upsert(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

// SetRepoPermissions performs a full update for p to the permissions tables.
// New user IDs found in p will be upserted and user IDs no longer in p will
// be removed.
func (s *store) SetRepoPermissions(ctx context.Context, p *RepoPermissions) (err error) {
	_, save := s.observe(ctx, "SetRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	// Retrieve currently stored user IDs of this repository.
	var userIDs *roaring.Bitmap
	userIDs, _, err = s.loadIDs(ctx, p.loadQuery())
	if err != nil {
		return err
	}

	toAdd, toRemove := p.diffIDs(userIDs)
	for _, id := range toAdd {
		p.IDs.Add(id)
	}
	p.IDs.RunOptimize()

	// Open a transaction for data consistency.
	var tx *Tx
	if tx, err = s.Tx(ctx); err != nil {
		return err
	}
	defer func() {
		// We need to use closure for a reference to the err.
		tx.CommitOrRollback(err)
	}()

	// Make another store with this underlying transaction.
	txs := &store{db: tx, clock: s.clock}

	p.UpdatedAt = txs.clock()

	var q *sqlf.Query
	if q, err = p.upsertQuery(); err != nil {
		return err
	}

	if err = txs.upsert(ctx, q); err != nil {
		return err
	}

	for _, id := range toAdd {
		up := &UserPermissions{
			UserID:   int32(id),
			Perm:     p.Perm,
			Type:     PermRepos,
			Provider: p.Provider,
		}
		up.IDs, _, err = txs.loadIDs(ctx, up.LoadQuery())
		if err != nil {
			return err
		}

		if up.IDs.Contains(uint32(p.RepoID)) {
			continue
		}

		up.IDs.Add(uint32(p.RepoID))
		up.IDs.RunOptimize()
		up.UpdatedAt = txs.clock()

		if q, err = up.UpsertQuery(); err != nil {
			return err
		} else if err = txs.upsert(ctx, q); err != nil {
			return err
		}
	}

	for _, id := range toRemove {
		up := &UserPermissions{
			UserID:   int32(id),
			Perm:     p.Perm,
			Type:     PermRepos,
			Provider: p.Provider,
		}
		up.IDs, _, err = txs.loadIDs(ctx, up.LoadQuery())
		if err != nil {
			return err
		}

		if up.IDs.Contains(uint32(p.RepoID)) {
			continue
		}

		up.IDs.Remove(uint32(p.RepoID))
		up.IDs.RunOptimize()
		up.UpdatedAt = txs.clock()

		if q, err = up.UpsertQuery(); err != nil {
			return err
		} else if err = txs.upsert(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

// SetPendingPermissions performs a full update for p to the pending permissions
// table. New bindIDs found in p will be upserted and bindIDs no longer in p will
// be removed.
func (s *store) SetPendingPermissions(
	ctx context.Context,
	bindIDs []string,
	perm authz.Perms,
	typ PermType,
	repoID int32,
) (err error) {
	_, save := s.observe(ctx, "SetPendingPermissions", "")
	defer func() {
		save(&err,
			otlog.Object("bindIDs", bindIDs),
			otlog.String("perm", string(perm)),
			otlog.String("type", string(typ)),
			otlog.Int32("repoID", repoID),
		)
	}()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	inBindIDs := make(map[string]bool)
	for i := range bindIDs {
		inBindIDs[bindIDs[i]] = true
	}

	// Open a transaction for data consistency.
	var tx *Tx
	if tx, err = s.Tx(ctx); err != nil {
		return err
	}
	defer func() {
		// We need to use closure for a reference to the err.
		tx.CommitOrRollback(err)
	}()

	// Make another store with this underlying transaction.
	txs := &store{db: tx, clock: s.clock}

	pp := &PendingPermissions{
		Perm: perm,
		Type: typ,
	}
	q := pp.loadWithBindIDQuery()

	var rows *sql.Rows
	rows, err = txs.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	p := &PendingPermissions{
		Perm: pp.Perm,
		Type: pp.Type,
	}
	for rows.Next() {
		var ids []byte
		if err = rows.Scan(&p.BindID, &ids, &p.UpdatedAt); err != nil {
			return err
		}

		if len(ids) == 0 {
			continue
		}

		p.IDs = roaring.NewBitmap()
		if err = p.IDs.UnmarshalBinary(ids); err != nil {
			return err
		}

		needsUpdate := false
		if p.IDs.Contains(uint32(repoID)) && !inBindIDs[p.BindID] {
			needsUpdate = true
			p.IDs.Remove(uint32(repoID))
		} else if inBindIDs[p.BindID] && !p.IDs.Contains(uint32(repoID)) {
			needsUpdate = true
			p.IDs.Add(uint32(repoID))
		}

		if needsUpdate {
			p.IDs.RunOptimize()
			if q, err = p.upsertQuery(); err != nil {
				return err
			}

			if err = txs.upsert(ctx, q); err != nil {
				return err
			}
		}
	}

	return rows.Close()
}

// GrantPendingPermissions grants the user has given ID with pending permissions found in p.
func (s *store) GrantPendingPermissions(ctx context.Context, userID int32, p *PendingPermissions) (err error) {
	_, save := s.observe(ctx, "GrantPendingPermissions", "")
	defer func() {
		save(&err,
			append(p.TracingFields(), otlog.Object("userID", userID))...,
		)
	}()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	p.IDs, p.UpdatedAt, err = s.loadIDs(ctx, p.loadQuery())
	if err != nil {
		return err
	}

	// Open a transaction for data consistency.
	var tx *Tx
	if tx, err = s.Tx(ctx); err != nil {
		return err
	}
	defer func() {
		// We need to use closure for a reference to the err.
		tx.CommitOrRollback(err)
	}()

	// Make another store with this underlying transaction.
	txs := &store{db: tx, clock: s.clock}

	up := &UserPermissions{
		UserID:    userID,
		Perm:      p.Perm,
		Type:      p.Type,
		IDs:       p.IDs,
		Provider:  ProviderSourcegraph,
		UpdatedAt: txs.clock(),
	}
	var q *sqlf.Query
	if q, err = up.UpsertQuery(); err != nil {
		return err
	}
	if err = txs.upsert(ctx, q); err != nil {
		return err
	}

	// NOTE: We currently only have "repos" type, so avoid unnecessary type checking for now.
	// Upsert repository permissions
	rp := &RepoPermissions{
		Perm:     p.Perm,
		Provider: ProviderSourcegraph,
	}
	iter := p.IDs.Iterator()
	for iter.HasNext() {
		id := iter.Next()
		rp.RepoID = int32(id)
		rp.IDs, _, err = txs.loadIDs(ctx, rp.loadQuery())
		if err != nil {
			return err
		}

		if rp.IDs.Contains(uint32(userID)) {
			continue
		}
		rp.IDs.Add(uint32(userID))
		rp.IDs.RunOptimize()

		rp.UpdatedAt = txs.clock()
		if q, err = rp.upsertQuery(); err != nil {
			return err
		}
		if err = txs.upsert(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

func (s *store) upsert(ctx context.Context, q *sqlf.Query) (err error) {
	ctx, save := s.observe(ctx, "upsert", "")
	defer func() {
		save(&err,
			otlog.String("Query.Query", q.Query(sqlf.PostgresBindVar)),
			otlog.Object("Query.Args", q.Args()),
		)
	}()

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

// LoadIDsTx runs the query on given transaction object and returns unmarshalled
// IDs and last updated time.
func (s *store) LoadIDsTx(ctx context.Context, tx *Tx, q *sqlf.Query) (*roaring.Bitmap, time.Time, error) {
	var err error
	ctx, save := s.observe(ctx, "LoadIDsTx", "")
	defer func() {
		save(&err,
			otlog.Object("tx", tx),
			otlog.Object("q", q),
		)
	}()

	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
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

// loadIDs runs the query and returns unmarshalled IDs and last updated time.
func (s *store) loadIDs(ctx context.Context, q *sqlf.Query) (*roaring.Bitmap, time.Time, error) {
	var err error
	ctx, save := s.observe(ctx, "loadIDs", "")
	defer func() { save(&err, otlog.Object("q", q)) }()

	var tx *Tx
	tx, err = s.Tx(ctx)
	if err != nil {
		return nil, time.Time{}, err
	}
	return s.LoadIDsTx(ctx, tx, q)
}

func (s *store) Tx(ctx context.Context) (*Tx, error) {
	switch t := s.db.(type) {
	case *sql.Tx:
		return &Tx{t}, nil
	case *sql.DB:
		tx, err := t.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		return &Tx{tx}, nil
	default:
		panic(fmt.Sprintf("can't open transaction with unknown implementation of dbutil.DB: %T", t))
	}
}

var ErrLockNotAvailable = errors.New("lock not available")

// Lock uses Postgres advisory locks to acquire an exclusive lock over the
// given namespace and lock ID on given transaction object. Concurrent
// processes that call this method while a lock is  already held by another
// process will have ErrLockNotAvailable returned.
func (s *store) Lock(ctx context.Context, tx *Tx, namespace, id int32) (err error) {
	ctx, save := s.observe(ctx, "lock", "")
	defer func() {
		save(&err,
			otlog.Object("tx", tx),
			otlog.Int32("namespace", namespace),
			otlog.Int32("id", id),
		)
	}()

	q := lockQuery(namespace, id)

	var rows *sql.Rows
	rows, err = tx.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
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
		return ErrLockNotAvailable
	}

	return nil
}

func lockQuery(namespace, id int32) *sqlf.Query {
	const format = `
-- source: enterprise/cmd/frontend/internal/authz/permsstore/store.go:lockQuery
SELECT pg_try_advisory_xact_lock(%s, %s)
`

	return sqlf.Sprintf(
		format,
		namespace,
		id,
	)
}

func (s *store) observe(ctx context.Context, family, title string) (context.Context, func(*error, ...otlog.Field)) {
	began := s.clock()
	tr, ctx := trace.New(ctx, "authz.permsstore.store."+family, title)

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

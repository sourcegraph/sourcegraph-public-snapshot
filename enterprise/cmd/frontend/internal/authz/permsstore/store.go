package permsstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/keegancsmith/sqlf"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Store is the unified interface for managing permissions explicitly in the database.
// It is concurrent-safe and maintains the data consistency over 'user_permissions',
// 'repo_permissions' and 'user_pending_permissions' tables.
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

// Now returns the current time in the context.
func (s *Store) Now() time.Time {
	return s.clock()
}

// LoadUserPermissions loads stored user permissions into p. An error is returned
// when there are no valid permissions available.
func (s *Store) LoadUserPermissions(ctx context.Context, p *UserPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadUserPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	p.IDs, p.UpdatedAt, err = s.Load(ctx, p.LoadQuery())
	return err
}

// LoadRepoPermissions loads stored repository permissions into p. An error is
// returned when there are no valid permissions available.
func (s *Store) LoadRepoPermissions(ctx context.Context, p *RepoPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	p.IDs, p.UpdatedAt, err = s.Load(ctx, p.loadQuery())
	return err
}

// LoadPendingPermissions loads stored pending user permissions into p. An
// error is returned when there are no pending permissions available.
func (s *Store) LoadPendingPermissions(ctx context.Context, p *PendingPermissions) (err error) {
	ctx, save := s.observe(ctx, "LoadPendingPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	p.IDs, p.UpdatedAt, err = s.Load(ctx, p.loadQuery())
	return err
}

// UpsertRepoPermissions stores new user IDs found in p, this method updates both
// the user and repository permissions tables.
func (s *Store) UpsertRepoPermissions(ctx context.Context, p *RepoPermissions) (err error) {
	_, save := s.observe(ctx, "UpsertRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	// Open a transaction for update consistency.
	var tx *Tx
	if tx, err = s.Tx(ctx); err != nil {
		return err
	}
	defer func() { tx.commitOrRollback(err) }()

	// Make another Store with this underlying transaction.
	txs := &Store{db: tx, clock: s.clock}

	// Iterate over the new set of user IDs and upsert ones that do not exist in database.
	err = txs.iterateAndUpsertUserPermissions(ctx, p.IDs.Iterator(), p.Perm, p.Provider,
		func(set *roaring.Bitmap) bool {
			return set.CheckedAdd(uint32(p.RepoID))
		})
	if err != nil {
		return err
	}

	// Retrieve currently stored IDs of this repository.
	oldIDs, _, err := s.Load(ctx, p.loadQuery())
	if err != nil {
		return err
	}

	// Upsert is just a union of two sets.
	p.IDs.Or(oldIDs)

	p.UpdatedAt = txs.clock()
	q, err := p.upsertQuery()
	if err != nil {
		return err
	} else if err = txs.upsert(ctx, q); err != nil {
		return err
	}

	return nil
}

// RemoveRepoPermissions removes user IDs found in p, this method updates both
// the user and repository permissions tables.
func (s *Store) RemoveRepoPermissions(ctx context.Context, p *RepoPermissions) (err error) {
	_, save := s.observe(ctx, "RemoveRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	// Open a transaction for update consistency.
	var tx *Tx
	if tx, err = s.Tx(ctx); err != nil {
		return err
	}
	defer func() { tx.commitOrRollback(err) }()

	// Make another Store with this underlying transaction.
	txs := &Store{db: tx, clock: s.clock}

	// Iterate over the delete set of user IDs and remove ones that exist in database.
	err = txs.iterateAndUpsertUserPermissions(ctx, p.IDs.Iterator(), p.Perm, p.Provider,
		func(set *roaring.Bitmap) bool {
			return set.CheckedRemove(uint32(p.RepoID))
		})
	if err != nil {
		return err
	}

	// Retrieve currently stored IDs of this repository.
	oldIDs, _, err := s.Load(ctx, p.loadQuery())
	if err != nil {
		return err
	}

	// Intersection (And) is the set that should be removed, doing it just in case p.IDs
	// contains elements do not exist in oldIDs. Once we're sure p.IDs is a subset of oldIDs,
	// we can compute diff (AndNot) to get the set we want to keep.
	oldIDs.And(p.IDs)
	p.IDs.AndNot(oldIDs)

	p.UpdatedAt = txs.clock()
	q, err := p.upsertQuery()
	if err != nil {
		return err
	} else if err = txs.upsert(ctx, q); err != nil {
		return err
	}

	return nil
}

// SetRepoPermissions performs a full update for p, new IDs found in p will be upserted
// and IDs no longer in p will be removed. This method updates both the user and
// repository permissions tables.
func (s *Store) SetRepoPermissions(ctx context.Context, p *RepoPermissions) (err error) {
	_, save := s.observe(ctx, "SetRepoPermissions", "")
	defer func() { save(&err, p.TracingFields()...) }()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	// Retrieve currently stored IDs of this repository.
	oldIDs, _, err := s.Load(ctx, p.loadQuery())
	if err != nil {
		return err
	}

	// Fisrt get the intersection (And), then use the intersection to compute diffs (AndNot)
	// with both the old and new sets to get IDs to remove and to add.
	isec := p.IDs.Clone()
	isec.And(oldIDs)
	toRemove := isec.Clone()
	toRemove.AndNot(oldIDs)
	toAdd := isec.Clone()
	toAdd.AndNot(p.IDs)

	// Open a transaction for update consistency.
	var tx *Tx
	if tx, err = s.Tx(ctx); err != nil {
		return err
	}
	defer func() { tx.commitOrRollback(err) }()

	// Make another Store with this underlying transaction.
	txs := &Store{db: tx, clock: s.clock}

	var q *sqlf.Query

	err = txs.iterateAndUpsertUserPermissions(ctx, toAdd.Iterator(), p.Perm, p.Provider,
		func(set *roaring.Bitmap) bool {
			return set.CheckedAdd(uint32(p.RepoID))
		})
	if err != nil {
		return err
	}

	err = txs.iterateAndUpsertUserPermissions(ctx, toRemove.Iterator(), p.Perm, p.Provider,
		func(set *roaring.Bitmap) bool {
			return set.CheckedRemove(uint32(p.RepoID))
		})
	if err != nil {
		return err
	}

	isec.Or(toAdd)
	p.IDs = isec

	p.UpdatedAt = txs.clock()
	if q, err = p.upsertQuery(); err != nil {
		return err
	} else if err = txs.upsert(ctx, q); err != nil {
		return err
	}

	return nil
}

// iterateAndUpsertUserPermissions uses the iterator to check if the user permissions loaded
// for the user ID needs an upsert. The iterator is meant to return user IDs.
//
// The checker is the check and modify function that determines whether an update to the user
// permissions is needed. It should return true if the set provided is being modified during
// the check.
func (s *Store) iterateAndUpsertUserPermissions(
	ctx context.Context,
	iter roaring.IntPeekable,
	perm authz.Perms,
	provider ProviderType,
	checker func(set *roaring.Bitmap) bool,
) (err error) {
	_, save := s.observe(ctx, "iterateAndUpsertUserPermissions", "")
	defer func() {
		save(&err,
			otlog.String("perm", string(perm)),
			otlog.String("provider", string(provider)),
		)
	}()

	var q *sqlf.Query

	for iter.HasNext() {
		up := &UserPermissions{
			UserID:   int32(iter.Next()),
			Perm:     perm,
			Type:     PermRepos,
			Provider: provider,
		}
		up.IDs, _, err = s.Load(ctx, up.LoadQuery())
		if err != nil {
			return err
		}

		if !checker(up.IDs) {
			continue
		}

		up.UpdatedAt = s.clock()
		if q, err = up.UpsertQuery(); err != nil {
			return err
		} else if err = s.upsert(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

// PendingPermssionsOpts contains required options for functions of pending permissions.
type PendingPermssionsOpts struct {
	BindIDs []string
	Perm    authz.Perms
	Type    PermType
	RepoID  int32
}

func (opts PendingPermssionsOpts) tracingFields() []otlog.Field {
	return []otlog.Field{
		otlog.Object("PendingPermssionsOpts.BindIDs", opts.BindIDs),
		otlog.String("PendingPermssionsOpts.Perm", string(opts.Perm)),
		otlog.String("PendingPermssionsOpts.Type", string(opts.Type)),
		otlog.Int32("PendingPermssionsOpts.RepoID", opts.RepoID),
	}
}

// UpsertPendingPermissions stores the repoID to the pending permissions table
// with given bindIDs.
func (s *Store) UpsertPendingPermissions(ctx context.Context, opts PendingPermssionsOpts) (err error) {
	_, save := s.observe(ctx, "UpsertPendingPermissions", "")
	defer func() { save(&err, opts.tracingFields()...) }()

	return s.iterateAndUpsertPendingPermissions(context.Background(), opts,
		func(set *roaring.Bitmap) bool {
			return set.CheckedAdd(uint32(opts.RepoID))
		})
}

// RemovePendingPermissions removes the repoID from the pending permissions table
// with given bindIDs.
func (s *Store) RemovePendingPermissions(ctx context.Context, opts PendingPermssionsOpts) (err error) {
	_, save := s.observe(ctx, "RemovePendingPermissions", "")
	defer func() { save(&err, opts.tracingFields()...) }()

	return s.iterateAndUpsertPendingPermissions(context.Background(), opts,
		func(set *roaring.Bitmap) bool {
			return set.CheckedRemove(uint32(opts.RepoID))
		})
}

// iterateAndUpsertPendingPermissions loops over the bindIDs in opts and check if the
// pending permissions loaded for the bind ID needs an upsert.
//
// The checker is the check and modify function that determines whether an update to the user
// permissions is needed. It should return true if the set provided is being modified during
// the check.
func (s *Store) iterateAndUpsertPendingPermissions(
	ctx context.Context,
	opts PendingPermssionsOpts,
	checker func(set *roaring.Bitmap) bool,
) (err error) {
	_, save := s.observe(ctx, "iterateAndUpsertPendingPermissions", "")
	defer func() { save(&err, opts.tracingFields()...) }()

	var q *sqlf.Query
	for _, bindID := range opts.BindIDs {
		pp := &PendingPermissions{
			BindID: bindID,
			Perm:   opts.Perm,
			Type:   opts.Type,
		}
		pp.IDs, _, err = s.Load(ctx, pp.loadQuery())
		if err != nil {
			return err
		}

		if !checker(pp.IDs) {
			continue
		}

		pp.UpdatedAt = s.clock()
		if q, err = pp.upsertQuery(); err != nil {
			return err
		} else if err = s.upsert(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

// SetPendingPermissions performs a full update for p to the pending permissions
// table. New bindIDs found in p will be upserted and bindIDs no longer in p will
// be removed.
func (s *Store) SetPendingPermissions(ctx context.Context, opts PendingPermssionsOpts) (err error) {
	_, save := s.observe(ctx, "SetPendingPermissions", "")
	defer func() { save(&err, opts.tracingFields()...) }()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	inBindIDs := make(map[string]bool)
	for i := range opts.BindIDs {
		inBindIDs[opts.BindIDs[i]] = true
	}

	pp := &PendingPermissions{
		Perm: opts.Perm,
		Type: opts.Type,
	}
	q := pp.loadWithBindIDQuery()

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	for rows.Next() {
		var ids []byte
		if err = rows.Scan(&pp.BindID, &ids, &pp.UpdatedAt); err != nil {
			return err
		}

		if len(ids) == 0 {
			continue
		}

		pp.IDs = roaring.NewBitmap()
		if err = pp.IDs.UnmarshalBinary(ids); err != nil {
			return err
		}

		// It is guaranteed only one of OR conditions could be true at a time because of inBindIDs and !inBindIDs.
		needsUpdate := (inBindIDs[pp.BindID] && pp.IDs.CheckedAdd(uint32(opts.RepoID))) ||
			(!inBindIDs[pp.BindID] && pp.IDs.CheckedRemove(uint32(opts.RepoID)))
		if needsUpdate {
			pp.UpdatedAt = s.clock()
			if q, err = pp.upsertQuery(); err != nil {
				return err
			} else if err = s.upsert(ctx, q); err != nil {
				return err
			}
		}
	}

	return rows.Close()
}

// GrantPendingPermissions grants the user has given ID with pending permissions found in p.
func (s *Store) GrantPendingPermissions(ctx context.Context, userID int32, p *PendingPermissions) (err error) {
	_, save := s.observe(ctx, "GrantPendingPermissions", "")
	defer func() {
		save(&err,
			append(p.TracingFields(), otlog.Object("userID", userID))...,
		)
	}()

	// Set context to background without a request bound timeout,
	// but let the above instrumentation use the original request's context.
	ctx = context.Background()

	p.IDs, p.UpdatedAt, err = s.Load(ctx, p.loadQuery())
	if err != nil {
		return err
	}

	// Open a transaction for update consistency.
	var tx *Tx
	if tx, err = s.Tx(ctx); err != nil {
		return err
	}
	defer func() { tx.commitOrRollback(err) }()

	// Make another Store with this underlying transaction.
	txs := &Store{db: tx, clock: s.clock}

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
	} else if err = txs.upsert(ctx, q); err != nil {
		return err
	}

	// NOTE: We currently only have "repos" type, so avoid unnecessary type checking for now.
	iter := p.IDs.Iterator()
	for iter.HasNext() {
		rp := &RepoPermissions{
			RepoID:   int32(iter.Next()),
			Perm:     p.Perm,
			Provider: ProviderSourcegraph,
		}
		rp.IDs, _, err = txs.Load(ctx, rp.loadQuery())
		if err != nil {
			return err
		}

		if !rp.IDs.CheckedAdd(uint32(userID)) {
			continue
		}

		rp.UpdatedAt = txs.clock()
		if q, err = rp.upsertQuery(); err != nil {
			return err
		} else if err = txs.upsert(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) upsert(ctx context.Context, q *sqlf.Query) (err error) {
	ctx, save := s.observe(ctx, "upsert", "")
	defer func() { save(&err, otlog.Object("q", q)) }()

	var rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

// Load runs the query and returns unmarshalled IDs and last updated time.
func (s *Store) Load(ctx context.Context, q *sqlf.Query) (*roaring.Bitmap, time.Time, error) {
	var err error
	ctx, save := s.observe(ctx, "loadIDs", "")
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

// QueryContext is the wrapper method that operates on the underlying db object.
func (s *Store) QueryContext(ctx context.Context, q string, args ...interface{}) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, q, args...)
}

// Tx returns a new transaction on the underlying db object. If the caller is already in
// a transaction, it returns the existing transaction object.
func (s *Store) Tx(ctx context.Context) (*Tx, error) {
	switch t := s.db.(type) {
	case *sql.Tx:
		return &Tx{t}, nil
	case *sql.DB:
		tx, err := t.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		return &Tx{tx}, nil
	case *Tx:
		return t, nil
	default:
		panic(fmt.Sprintf("can't open transaction with unknown implementation of dbutil.DB: %T", t))
	}
}

// InTx returns true if current context is inside a transaction.
func (s *Store) InTx() bool {
	_, ok := s.db.(*Tx)
	return ok
}

func (s *Store) observe(ctx context.Context, family, title string) (context.Context, func(*error, ...otlog.Field)) {
	began := s.clock()
	tr, ctx := trace.New(ctx, "authz.permsstore.Store."+family, title)

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

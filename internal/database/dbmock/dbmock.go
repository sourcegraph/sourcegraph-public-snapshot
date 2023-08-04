// The dbmock package facilitates embedding mock stores directly in the
// datbase.DB object.
package dbmock

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

// mockedDB is a wrapper around a database.DB, and has an additional
// mockedStore. A specific mockedStore can be unwrapped using the
// get function.
type mockedDB struct {
	database.DB
	mockedStore any
}

// New embeds each mocked store in the provided DB.
func New(db database.DB, stores ...toEmbeddable) database.DB {
	for _, store := range stores {
		db = store.ToEmbeddable().Embed(db)
	}
	return db
}

func (mdb *mockedDB) WithTransact(ctx context.Context, f func(tx database.DB) error) error {
	return mdb.DB.WithTransact(ctx, func(tx database.DB) error {
		return f(&mockedDB{DB: tx, mockedStore: mdb.mockedStore})
	})
}

func (mdb *mockedDB) With(other basestore.ShareableStore) database.DB {
	return mdb.DB.With(other)
}

// Embeddable describes a store that can embed itself in a database.DB.
type Embeddable interface {
	Embed(database.DB) database.DB
}

// toEmbeddable describes a mock store that can be converted to an Embeddable store.
// The implementation of this interface should simply be:
//
//	func (store *MockStore) toEmbeddable() dbmock.Embeddable {
//		return dbmock.NewEmbeddable(store)
//	}
type toEmbeddable interface {
	ToEmbeddable() Embeddable
}

type embeddable[T toEmbeddable] struct {
	store any
}

// Embed embeds the wrapped store inside the database.DB by wrapping
// the db inside a mockedDB.
func (e *embeddable[T]) Embed(db database.DB) database.DB {
	return &mockedDB{
		DB:          db,
		mockedStore: e.store,
	}
}

func NewEmbeddable[T toEmbeddable](store T) *embeddable[T] {
	return &embeddable[T]{store}
}

// Get fetches the mocked interface T from the provided DB.
// If no mocked interface is found, nil is returned.
func get[T any](db database.DB) *T {
	switch v := db.(type) {
	case *mockedDB:
		if t, ok := v.mockedStore.(T); ok {
			return &t
		}
		return get[T](v.DB)
	}
	return nil
}

// MockableStore wraps basestore.Store. The basestore.Store is hidden behind
// a private field so that it cannot be interacted with.
// The basestore.Store can only be initialized by calling `.WithDB`, which
// will then return the embedded store.
type MockableStore[T any] struct {
	inner *basestore.Store
	store T
}

// NewMockableStore returns a new MockableStore instance from store.
// The generic parameter T should be the interface that you wish to be
// mockable.
func NewMockableStore[T any](store T) *MockableStore[T] {
	return &MockableStore[T]{store: store}
}

// WithDB initializes the inner basestore.Store and returns the underlying
// store.
//
// Any attempts to use the inner store before calling WithDB will result in
// a panic.
func (s *MockableStore[T]) WithDB(db database.DB) T {
	if i := get[T](db); i != nil {
		return *i
	}

	s.inner = basestore.NewWithHandle(db.Handle())
	return s.store
}

func (s *MockableStore[T]) Done(err error) error {
	if s.inner == nil {
		panic("Store not initialized. Did you forget to call .With(db)?")
	}

	return s.inner.Done(err)
}

func (s *MockableStore[T]) Exec(ctx context.Context, query *sqlf.Query) error {
	if s.inner == nil {
		panic("Store not initialized. Did you forget to call .With(db)?")
	}

	return s.inner.Exec(ctx, query)
}

func (s *MockableStore[T]) ExecResult(ctx context.Context, query *sqlf.Query) (sql.Result, error) {
	if s.inner == nil {
		panic("Store not initialized. Did you forget to call .With(db)?")
	}

	return s.inner.ExecResult(ctx, query)
}

func (s *MockableStore[T]) Handle() basestore.TransactableHandle {
	if s.inner == nil {
		panic("Store not initialized. Did you forget to call .With(db)?")
	}

	return s.inner.Handle()
}

func (s *MockableStore[T]) InTransaction() bool {
	if s.inner == nil {
		panic("Store not initialized. Did you forget to call .With(db)?")
	}

	return s.inner.InTransaction()
}

func (s *MockableStore[T]) Query(ctx context.Context, query *sqlf.Query) (*sql.Rows, error) {
	if s.inner == nil {
		panic("Store not initialized. Did you forget to call .With(db)?")
	}

	return s.inner.Query(ctx, query)
}

func (s *MockableStore[T]) QueryRow(ctx context.Context, query *sqlf.Query) *sql.Row {
	if s.inner == nil {
		panic("Store not initialized. Did you forget to call .With(db)?")
	}

	return s.inner.QueryRow(ctx, query)
}

func (s *MockableStore[T]) SetLocal(ctx context.Context, key string, value string) (func(context.Context) error, error) {
	if s.inner == nil {
		panic("Store not initialized. Did you forget to call .With(db)?")
	}

	return s.inner.SetLocal(ctx, key, value)
}

func (s *MockableStore[T]) Transact(ctx context.Context) (*basestore.Store, error) {
	if s.inner == nil {
		panic("Store not initialized. Did you forget to call .With(db)?")
	}

	return s.inner.Transact(ctx)
}

func (s *MockableStore[T]) WithTransact(ctx context.Context, f func(tx *basestore.Store) error) error {
	if s.inner == nil {
		panic("Store not initialized. Did you forget to call .With(db)?")
	}

	return s.inner.WithTransact(ctx, f)
}

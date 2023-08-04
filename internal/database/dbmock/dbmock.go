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

type mockedDB struct {
	database.DB
	embeddedInterface any
}

func (mdb *mockedDB) WithTransact(ctx context.Context, f func(tx database.DB) error) error {
	return mdb.DB.WithTransact(ctx, func(tx database.DB) error {
		return f(&mockedDB{DB: tx, embeddedInterface: mdb.embeddedInterface})
	})
}

func (mdb *mockedDB) With(other basestore.ShareableStore) database.DB {
	return mdb.DB.With(other)
}

type Embeddable interface {
	Embed(database.DB) database.DB
}

type embeddable struct {
	store any
}

func NewEmbeddable[T any](store RetrievableStore[T]) *embeddable {
	return &embeddable{store}
}

func (e *embeddable) Embed(db database.DB) database.DB {
	return &mockedDB{
		DB:                db,
		embeddedInterface: e.store,
	}
}

// Get fetches the mocked interface T from the provided DB.
// If no mocked interface is found, nil is returned.
func get[T any](db database.DB) *T {
	switch v := db.(type) {
	case *mockedDB:
		if t, ok := v.embeddedInterface.(T); ok {
			return &t
		}
		return get[T](v.DB)
	}
	return nil
}

type RetrievableStore[T any] interface {
	GetStoreFunc() NewStoreFunc[T]
}

type MockableStore[T RetrievableStore[T]] struct {
	inner *basestore.Store
	store T
}

func NewMockableStore[T RetrievableStore[T]](store T) *MockableStore[T] {
	return &MockableStore[T]{store: store}
}

func (s *MockableStore[T]) GetStoreFunc() NewStoreFunc[*MockableStore[T]] {
	return func(db database.DB) *MockableStore[T] {
		s.inner = basestore.NewWithHandle(db.Handle())
		return s
	}
}

func (s *MockableStore[T]) With(db database.DB) T {
	s.inner = basestore.NewWithHandle(db.Handle())

	return s.store.GetStoreFunc()(db)
}

type NewStoreFunc[T any] func(database.DB) T

func (f NewStoreFunc[T]) With(db database.DB) T {
	if i := get[T](db); i != nil {
		return *i
	}

	return f(db)
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

// New embeds each mock option in the provided DB.
func New(db database.DB, stores ...Embeddable) database.DB {
	for _, store := range stores {
		db = store.Embed(db)
	}
	return db
}

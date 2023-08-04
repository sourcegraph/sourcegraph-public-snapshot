// The dbmock package facilitates embedding mock stores directly in the
// datbase.DB object.
package dbmock

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

// BaseStore is a store without a database connection.
// It can be turned into a store with a database connection
// by calling .WithDB and providing a DB.
type BaseStore[T any] interface {
	WithDB(database.DB) T
}

// baseStore implements BaseStore.
// It wraps another BaseStore, but checks the provided database.DB
// for any mocks and returns a mock if found.
type baseStore[T any] struct {
	store BaseStore[T]
}

func NewBaseStore[T any](store BaseStore[T]) BaseStore[T] {
	return &baseStore[T]{
		store: store,
	}
}

func (b *baseStore[T]) WithDB(db database.DB) T {
	if i := get[T](db); i != nil {
		return *i
	}

	return b.store.WithDB(db)
}

// mockedDB is a wrapper around a database.DB, and has an additional
// mockedStore. A specific mockedStore can be unwrapped using the
// get function.
type mockedDB struct {
	database.DB
	mockedStore any
}

func (mdb *mockedDB) WithTransact(ctx context.Context, f func(tx database.DB) error) error {
	return mdb.DB.WithTransact(ctx, func(tx database.DB) error {
		return f(&mockedDB{DB: tx, mockedStore: mdb.mockedStore})
	})
}

// New embeds each mocked store in the provided DB.
func New(db database.DB, stores ...any) database.DB {
	for _, store := range stores {
		db = &mockedDB{
			DB:          db,
			mockedStore: store,
		}
	}
	return db
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

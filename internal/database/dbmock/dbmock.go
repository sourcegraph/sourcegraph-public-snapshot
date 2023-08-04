// The dbmock package facilitates embedding mock stores directly in the
// datbase.DB object.
package dbmock

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

// Configurable is any type that, when given database.DB, turns into T.
type Configurable[T any] interface {
	WithDB(database.DB) T
}

// BaseStore is a store without a database connection.
// It can be turned into a store with a database connection
// by calling .WithDB and providing a DB.
type BaseStore[K any, T Configurable[K]] interface {
	WithDB(database.DB) K
}

// baseStore implements BaseStore.
// It checks the provided database.DB for any mocks.
type baseStore[K any, T Configurable[K]] struct {
	store T
}

func NewBaseStore[K any, T Configurable[K]](store T) BaseStore[K, T] {
	return &baseStore[K, T]{
		store: store,
	}
}

func (b *baseStore[K, T]) WithDB(db database.DB) K {
	if i := get[K](db); i != nil {
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

func (mdb *mockedDB) With(other basestore.ShareableStore) database.DB {
	return mdb.DB.With(other)
}

// New embeds each mocked store in the provided DB.
func New(db database.DB, stores ...toEmbeddable) database.DB {
	for _, store := range stores {
		db = store.ToEmbeddable().Embed(db)
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

// The dbmock package facilitates embedding mock stores directly in the
// datbase.DB object.
package dbmock

import (
	"context"
	"reflect"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type mockedDB struct {
	database.DB
	mockedStore reflect.Value
}

func (mdb *mockedDB) WithTransact(ctx context.Context, f func(tx database.DB) error) error {
	return mdb.DB.WithTransact(ctx, func(tx database.DB) error {
		return f(&mockedDB{DB: tx, mockedStore: mdb.mockedStore})
	})
}

// GetMock fetches the mocked interface T from the provided DB.
// If no mocked interface is found, nil is returned.
func GetMock[T basestore.ShareableStore](db database.DB) (t T) {
	switch v := db.(type) {
	case *mockedDB:
		if v.mockedStore.Type().Implements(reflect.TypeOf((*T)(nil)).Elem()) {
			if mock, ok := v.mockedStore.Interface().(T); ok {
				return mock
			}
		}
		return GetMock[T](v.DB)
	}
	return t
}

type MockOption func(database.DB) database.DB

// WithMock creates a new MockOption from the provided store.
// Store must implement both the basestore.ShareableStore and
// the MockableStore interfaces.
func WithMock[T basestore.ShareableStore](val T) MockOption {
	return func(db database.DB) database.DB {
		return &mockedDB{
			DB:          db,
			mockedStore: reflect.ValueOf(val),
		}
	}
}

// NewMockedDB embeds each mock option in the provided DB.
func NewMockedDB(db database.DB, options ...MockOption) database.DB {
	for _, option := range options {
		db = option(db)
	}

	return db
}

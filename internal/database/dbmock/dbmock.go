// The dbmock package facilitates embedding mock stores directly in the
// datbase.DB object.
package dbmock

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type mockedDB struct {
	database.DB
	mockedStore basestore.ShareableStore
}

func (mdb *mockedDB) WithTransact(ctx context.Context, f func(tx database.DB) error) error {
	return mdb.DB.WithTransact(ctx, func(tx database.DB) error {
		return f(&mockedDB{DB: tx, mockedStore: mdb.mockedStore})
	})
}

// Get fetches the mocked interface T from the provided DB.
// If no mocked interface is found, nil is returned.
func Get[T basestore.ShareableStore](db database.DB) (t T) {
	switch v := db.(type) {
	case *mockedDB:
		if t, ok := v.mockedStore.(T); ok {
			return t
		}
		return Get[T](v.DB)
	}
	return t
}

// New embeds each mock option in the provided DB.
func New(db database.DB, mockStores ...basestore.ShareableStore) database.DB {
	for _, mockStore := range mockStores {
		db = &mockedDB{
			DB:          db,
			mockedStore: mockStore,
		}
	}

	return db
}

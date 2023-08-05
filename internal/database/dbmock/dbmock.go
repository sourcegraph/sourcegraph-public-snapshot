// The dbmock package facilitates embedding mock stores directly in the
// datbase.DB object.
package dbmock

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

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
func Get[T any](db database.DB) (t T) {
	switch v := db.(type) {
	case *mockedDB:
		if t, ok := v.mockedStore.(T); ok {
			return t
		}
		return Get[T](v.DB)
	}
	return t
}

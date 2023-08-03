// The dbmock package facilitates embedding mock stores directly in the
// datbase.DB object.
package dbmock

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/mockstore"
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

func (mdb *mockedDB) With(other basestore.ShareableStore) database.DB {
	return mdb.DB.With(other)
}

// New embeds each mock option in the provided DB.
func New(db database.DB, options ...mockstore.MockOption) database.DB {
	return &mockedDB{
		DB:          db,
		mockedStore: mockstore.NewMockableShareableStore(db, options...),
	}
}

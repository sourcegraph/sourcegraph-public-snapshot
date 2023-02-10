// Package fakedb contains in-memory, partial implementations of stores
// from the database package. This set of fakes is meant to be extended
// as needed.
package fakedb

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

// New creates a set of fakes currently available to database stores.
func New() Fakes {
	return Fakes{
		TeamStore: &Teams{},
		UserStore: &Users{},
	}
}

// Fakes aggregates together specific stores and makes them accessible
// to the test.
type Fakes struct {
	TeamStore *Teams
	UserStore *Users
}

// Wire injects fakes into a database.MockDB.
func (fs Fakes) Wire(db *database.MockDB) {
	db.TeamsFunc.SetDefaultReturn(fs.TeamStore)
	db.UsersFunc.SetDefaultReturn(fs.UserStore)
	db.WithTransactFunc.SetDefaultHook(func(_ context.Context, callback func(database.DB) error) error {
		return callback(db)
	})
}

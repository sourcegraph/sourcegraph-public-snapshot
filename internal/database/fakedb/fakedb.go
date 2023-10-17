// Package fakedb contains in-memory, partial implementations of stores
// from the database package. This set of fakes is meant to be extended
// as needed.
package fakedb

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
)

// New creates a set of fakes currently available to database stores.
func New() Fakes {
	teams := &Teams{}
	users := &Users{}
	teams.users = users
	return Fakes{
		TeamStore: teams,
		UserStore: users,
	}
}

// Fakes aggregates together specific stores and makes them accessible
// to the test. It also exposes methods useful for test setup
// or data validation for white-box testing. The methods that correspond
// to specific stores are implemented next to the specific fake store.
type Fakes struct {
	TeamStore *Teams
	UserStore *Users
}

// Wire injects fakes into a database.MockDB.
func (fs Fakes) Wire(db *dbmocks.MockDB) {
	db.TeamsFunc.SetDefaultReturn(fs.TeamStore)
	db.UsersFunc.SetDefaultReturn(fs.UserStore)
	db.WithTransactFunc.SetDefaultHook(func(_ context.Context, callback func(database.DB) error) error {
		return callback(db)
	})
}

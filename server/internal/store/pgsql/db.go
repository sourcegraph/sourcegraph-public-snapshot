package pgsql

import (
	"fmt"
	"sync"

	"gopkg.in/gorp.v1"

	"sourcegraph.com/sourcegraph/sourcegraph/util/dbutil2"
)

var (
	// Schema is the DB Schema for the database used by this package.
	Schema = dbutil2.Schema{
		CreateSQL: []string{
			`CREATE EXTENSION IF NOT EXISTS citext;`,
			`CREATE EXTENSION IF NOT EXISTS hstore;`,
		},
		Map: &gorp.DbMap{Dialect: gorp.PostgresDialect{}},
	}
)

var (
	globalDBH *dbutil2.Handle // global DB handle
	dbLock    sync.Mutex      // protects globalDBH
)

// globalDB opens the DB if it isn't already open, and returns
// it. Subsequent calls return the same DB handle.
func globalDB() (*dbutil2.Handle, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	if globalDBH != nil {
		return globalDBH, nil
	}

	dbh, err := OpenDB(0)
	if err != nil {
		return nil, err
	}

	globalDBH = dbh
	return globalDBH, nil
}

// OpenDB opens and returns the DB handle for the DB. Use DB unless
// you need access to the low-level DB handle or need to handle
// errors.
func OpenDB(mode dbutil2.Mode) (*dbutil2.Handle, error) {
	dbh, err := dbutil2.Open("", Schema, mode)
	if err != nil {
		return nil, fmt.Errorf("open DB: %s", err)
	}
	return dbh, nil
}

package pgsql

import (
	"fmt"
	"log"
	"sync"

	"github.com/sqs/modl"

	"src.sourcegraph.com/sourcegraph/util/dbutil2"
)

var (
	// Schema is the DB Schema for the database used by this package.
	Schema = dbutil2.Schema{
		CreateSQL: []string{
			`CREATE EXTENSION IF NOT EXISTS citext;`,
			`CREATE EXTENSION IF NOT EXISTS hstore;`,
		},
		Map: &modl.DbMap{Dialect: modl.PostgresDialect{}},
	}
)

var (
	globalDBH *dbutil2.Handle // global DB handle
	dbLock    sync.Mutex      // protects globalDBH
)

// DB opens the DB if it isn't already open, and returns
// it. Subsequent calls return the same DB handle.
func DB() *dbutil2.Handle {
	dbLock.Lock()
	defer dbLock.Unlock()

	if globalDBH != nil {
		return globalDBH
	}

	dbh, err := OpenDB(0)
	if err != nil {
		log.Fatal(err)
	}

	globalDBH = dbh
	return globalDBH
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

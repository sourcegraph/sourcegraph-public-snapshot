package dbutil2

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	_ "github.com/lib/pq"
)

var (
	opened     map[string]*sql.DB // cache of Open dataSource -> DB
	openedLock sync.Mutex         // protects opened
)

// open opens the DB identified by dataSource. If an existing *sql.DB
// already exists for the same dataSource, the existing one is
// returned instead of opening a new one.
func open(dataSource string) (*sql.DB, error) {
	openedLock.Lock()
	defer openedLock.Unlock()
	if db, present := opened[dataSource]; present {
		return db, nil
	}

	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Cache for next time.
	if opened == nil {
		opened = map[string]*sql.DB{}
	}
	opened[dataSource] = db
	return db, nil
}

// Open creates a new DB handle with the given schema by connecting to
// the database identified by dataSource (e.g., "dbname=mypgdb" or
// blank to use the PG* env vars).
//
// Open assumes that the database already exists.
func Open(dataSource string) (*sql.DB, error) {
	db, err := open(dataSource)
	if err != nil {
		return nil, fmt.Errorf("%s (datasource=%q)", err, dataSource)
	}

	// Ensure we're in UTC.
	var tz string
	if err := db.QueryRow("SELECT current_setting('TIMEZONE')").Scan(&tz); err != nil {
		return nil, fmt.Errorf("getting DB timezone: %s", err)
	}
	if tz != "UTC" {
		return nil, fmt.Errorf("PostgresQL timezone must be UTC, but it is set to %q. (Set it by specifying `timezone = 'UTC'` in postgresql.conf and then restart PostgreSQL.)", tz)
	}
	return db, nil
}

// IsAlreadyExistsError returns true if err is a PostgreSQL error that
// something "already exists" (such as a table).
func IsAlreadyExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already exists")
}

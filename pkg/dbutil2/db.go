package dbutil2

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	_ "github.com/lib/pq"
)

// A Schema describes a database schema.
type Schema struct {
	// CreateSQL contains SQL statements run immediately after
	// creating the DB-mapped tables in this schema.
	CreateSQL []string

	// DropSQL contains SQL statements run immediately before
	// dropping the DB-mapped tables in this schema.
	DropSQL []string
}

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
func Open(dataSource string, schema Schema) (*Handle, error) {
	db, err := open(dataSource)
	if err != nil {
		return nil, fmt.Errorf("%s (datasource=%q)", err, dataSource)
	}

	h := &Handle{
		DataSource: dataSource,
		schema:     schema,
		Db:         db,
	}
	if err := h.configure(); err != nil {
		return nil, fmt.Errorf("configuring DB handle %q: %s", dataSource, err)
	}

	return h, nil
}

// configureDB enables DB trace logging if the PGTRACE env var is
// set and checks that the DB timezone is UTC.
func (h *Handle) configure() error {
	// Ensure we're in UTC.
	var tz string
	if err := h.Db.QueryRow("SELECT current_setting('TIMEZONE')").Scan(&tz); err != nil {
		return fmt.Errorf("getting DB timezone: %s", err)
	}
	if tz != "UTC" {
		return fmt.Errorf("PostgresQL timezone must be UTC, but it is set to %q. (Set it by specifying `timezone = 'UTC'` in postgresql.conf and then restart PostgreSQL.)", tz)
	}
	return nil
}

// A Handle is the interface to a database. It can safely be used by
// concurrent goroutines.
type Handle struct {
	// DataSource is the data source string used to connect to this
	// handle's database.
	DataSource string

	// schema is the Schema that this handle was created from.
	schema Schema

	Db *sql.DB
}

// CreateUnloggedTables determines whether the PostgreSQL tables
// should be created as unlogged. It is set to true during tests
// because unlogged tables are faster to use and to truncate, and the
// WAL is not needed. See
// http://www.postgresql.org/docs/9.1/static/sql-createtable.html for
// more info.
var CreateUnloggedTables bool

// CreateSchema creates the schema for this handle in the database
// it's connected to.
func (h *Handle) CreateSchema() error {
	var errs []error
	for _, sql := range h.schema.CreateSQL {
		if _, err := h.Db.Exec(sql); err != nil && !IsAlreadyExistsError(err) {
			errs = append(errs, fmt.Errorf("%s (on SQL: %s)", err, sql))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%d errors creating schema: %v (data source is %q)", len(errs), errs, h.DataSource)
	}
	return nil
}

// DropSchema drops the schema for this handle in the database
// it's connected to.
func (h *Handle) DropSchema() error {
	var errs []error
	for _, sql := range h.schema.DropSQL {
		if _, err := h.Db.Exec(sql); err != nil {
			errs = append(errs, fmt.Errorf("%s (on SQL: %s)", err, sql))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%d errors dropping schema: %v (data source is %q)", len(errs), errs, h.DataSource)
	}
	return nil
}

// IsAlreadyExistsError returns true if err is a PostgreSQL error that
// something "already exists" (such as a table).
func IsAlreadyExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already exists")
}

package dbutil2

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/executil"

	"gopkg.in/gorp.v1"

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

	// Map is a DbMap without the Db/Dbx set (because a schema can be
	// used to construct several DB connections).
	Map *gorp.DbMap
}

type Mode uint

const (
	// CreateDBIfNotExists makes Open create the database
	// (using the PostgreSQL "createdb" program, which must be in your
	// $PATH) if it does not exist. It will be created with the
	// default options inherited from your PG* env vars (which likely
	// means the $PGUSER will be the DB's owner, etc.).
	CreateDBIfNotExists Mode = 1 << iota
)

var (
	opened     map[string]*sql.DB // cache of Open dataSource -> DB
	openedLock sync.Mutex         // protects opened
)

// open opens the DB identified by dataSource. If an existing *sql.DB
// already exists for the same dataSource, the existing one is
// returned instead of opening a new one.
func open(dataSource string, mode Mode) (*sql.DB, error) {
	openedLock.Lock()
	defer openedLock.Unlock()
	if db, present := opened[dataSource]; present {
		return db, nil
	}

	triedCreate := false
tryOpen:
	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		return nil, err
	}
	create := mode&CreateDBIfNotExists != 0
	if err := db.Ping(); err != nil {
		if !triedCreate && create && strings.Contains(err.Error(), "does not exist") {
			// DB likely doesn't exist; try creating it.
			if err2 := createdb(dataSource); err2 != nil {
				return nil, fmt.Errorf("creating DB %s failed: %s (tried to create DB because Ping failed: %s)", dataSource, err2, err)
			}
			triedCreate = true
			goto tryOpen
		} else {
			return nil, err
		}
	}

	// Cache for next time.
	if opened == nil {
		opened = map[string]*sql.DB{}
	}
	opened[dataSource] = db
	return db, nil
}

// createdb calls the PostgreSQL "createdb" program to create a new
// PostgreSQL database using the info from ds.
func createdb(ds string) error {
	cmd := exec.Command("createdb")
	if ds != "" {
		overrides := getPGEnvsFromDataSource(ds)
		executil.OverrideEnvAll(cmd, overrides)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("createdb %q failed (%s) with output:\n%s", ds, err, out)
	}
	return nil
}

// Open creates a new DB handle with the given schema by connecting to
// the database identified by dataSource (e.g., "dbname=mypgdb" or
// blank to use the PG* env vars).
//
// Open assumes that the database already exists.
func Open(dataSource string, schema Schema, mode Mode) (*Handle, error) {
	db, err := open(dataSource, mode)
	if err != nil {
		return nil, fmt.Errorf("%s (datasource=%q)", err, dataSource)
	}

	dbm := *schema.Map // copy
	dbm.Db = db
	h := &Handle{
		DataSource: dataSource,
		schema:     schema,
		DbMap:      &dbm,
	}
	if err := h.configure(); err != nil {
		return nil, fmt.Errorf("configuring DB handle %q: %s", dataSource, err)
	}

	return h, nil
}

// configureDB enables DB trace logging if the PGTRACE env var is
// set and checks that the DB timezone is UTC.
func (h *Handle) configure() error {
	if trace, err := strconv.ParseBool(os.Getenv("PGTRACE")); err == nil && trace {
		dbname, err := h.SelectStr("SELECT current_database()")
		if err != nil {
			return err
		}
		h.DbMap.TraceOn("["+dbname+"]", log.New(os.Stdout, "", log.Lmicroseconds))
	}

	// Ensure we're in UTC.
	tz, err := h.SelectStr("SELECT current_setting('TIMEZONE')")
	if err != nil {
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

	// DbMap is from the Schema that this handle was created
	// from. Don't modify the DB mapping (by calling AddTable, for
	// example) after init time because other goroutines might be
	// using this handle concurrently and because changes will not be
	// propagated to other handles built from the same underlying
	// schema.
	//
	// It is embedded (although it also exists underneath the schema
	// field) so that Handle exports DbMap's methods.
	*gorp.DbMap
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
	if err := h.DbMap.CreateTablesIfNotExists(); err != nil {
		return err
	}
	var errs []error
	for _, sql := range h.schema.CreateSQL {
		if _, err := h.Exec(sql); err != nil && !IsAlreadyExistsError(err) {
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
		if _, err := h.Exec(sql); err != nil {
			errs = append(errs, fmt.Errorf("%s (on SQL: %s)", err, sql))
		}
	}
	if err := h.DropTablesIfExists(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("%d errors dropping schema: %v (data source is %q)", len(errs), errs, h.DataSource)
	}
	return nil
}

// UnderlyingSQLExecutor implements dbutil.SQLExecutorWrapper so that
// other utility funcs can unwrap Handle to get to its DbMap without
// having to import package dbutil.
func (h *Handle) UnderlyingSQLExecutor() gorp.SqlExecutor { return h.DbMap }

// IsAlreadyExistsError returns true if err is a PostgreSQL error that
// something "already exists" (such as a table).
func IsAlreadyExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already exists")
}

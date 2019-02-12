package repos

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"net/url"
	"os"
	"strconv"
	"time"

	migr "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/migrations"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// A DB captures the essential methods of a sql.DB.
type DB interface {
	PrepareContext(ctx context.Context, q string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, q string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, q string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, q string, args ...interface{}) *sql.Row
}

// A Tx captures the essential methods of a sql.Tx.
type Tx interface {
	Rollback() error
	Commit() error
}

// A TxBeginner captures BeginTx method of a sql.DB
type TxBeginner interface {
	BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}

// NewDB returns a new *sql.DB from the given dsn (data source name).
func NewDB(dsn string) (*sql.DB, error) {
	// We want to configure the database client explicitly through the DSN.
	// lib/pq uses and gives precedence to these environment variables so we unset them.
	for _, v := range []string{
		"PGHOST", "PGHOSTADDR", "PGPORT",
		"PGDATABASE", "PGUSER", "PGPASSWORD",
		"PGSERVICE", "PGSERVICEFILE", "PGREALM",
		"PGOPTIONS", "PGAPPNAME", "PGSSLMODE",
		"PGSSLCERT", "PGSSLKEY", "PGSSLROOTCERT",
		"PGREQUIRESSL", "PGSSLCRL", "PGREQUIREPEER",
		"PGKRBSRVNAME", "PGGSSLIB", "PGCONNECT_TIMEOUT",
		"PGCLIENTENCODING", "PGDATESTYLE", "PGTZ",
		"PGGEQO", "PGSYSCONFDIR", "PGLOCALEDIR",
	} {
		os.Unsetenv(v)
	}

	cfg, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse dsn")
	}

	qry := cfg.Query()

	// Force PostgreSQL session timezone to UTC.
	qry.Set("timezone", "UTC")

	// Set max open and idle connections
	maxOpen, _ := strconv.Atoi(qry.Get("max_conns"))
	if maxOpen == 0 {
		maxOpen = 30
	}
	qry.Del("max_conns")

	cfg.RawQuery = qry.Encode()
	db, err := sql.Open("postgres", cfg.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	// TODO(tsenart): Instrument with Prometheus

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxOpen)

	return db, nil
}

// MigrateDB runs all migrations from github.com/sourcegraph/sourcegraph/migrations
// against the given sql.DB
func MigrateDB(db *sql.DB) error {
	var cfg postgres.Config
	driver, err := postgres.WithInstance(db, &cfg)
	if err != nil {
		return err
	}

	s := bindata.Resource(migrations.AssetNames(), migrations.Asset)
	d, err := bindata.WithInstance(s)
	if err != nil {
		return err
	}

	m, err := migr.NewWithInstance("go-bindata", d, "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err == nil || err == migr.ErrNoChange {
		return nil
	}

	if os.IsNotExist(err) {
		// This should only happen if the DB is ahead of the migrations available
		version, dirty, verr := m.Version()
		if verr != nil {
			return verr
		}
		if dirty { // this shouldn't happen, but checking anyways
			return err
		}
		log15.Warn("WARNING: Detected an old version of Sourcegraph. The database has migrated to a newer version. If you have applied a rollback, this is expected and you can ignore this warning. If not, please contact support@sourcegraph.com for further assistance.", "db_version", version)
		return nil
	}
	return err
}

// nullTime represents a time.Time that may be null. nullTime implements the
// sql.Scanner interface so it can be used as a scan destination, similar to
// sql.NullString. When the scanned value is null, Time is set to the zero value.
type nullTime struct{ *time.Time }

// Scan implements the Scanner interface.
func (nt *nullTime) Scan(value interface{}) error {
	*nt.Time, _ = value.(time.Time)
	return nil
}

// Value implements the driver Valuer interface.
func (nt nullTime) Value() (driver.Value, error) {
	if nt.Time == nil {
		return nil, nil
	}
	return *nt.Time, nil
}

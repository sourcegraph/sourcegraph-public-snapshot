package dbconn

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/migrations"
)

// Database describes one of our Postgres (or Postgres-like) databases.
type Database struct {
	// Name is the name of the database.
	Name string

	// MigrationsTable is the migrations SQL table name.
	MigrationsTable string

	// FS describes the raw migration assets to run to migrate the target schema to a new
	// version.
	FS fs.FS

	// TargetsTimescaleDB indicates if the database targets TimescaleDB. Otherwise, Postgres.
	TargetsTimescaleDB bool
}

var (
	Frontend = &Database{
		Name:            "frontend",
		MigrationsTable: "schema_migrations",
		FS:              migrations.Frontend,
	}

	CodeIntel = &Database{
		Name:            "codeintel",
		MigrationsTable: "codeintel_schema_migrations",
		FS:              migrations.CodeIntel,
	}

	CodeInsights = &Database{
		Name:               "codeinsights",
		TargetsTimescaleDB: true,
		MigrationsTable:    "codeinsights_schema_migrations",
		FS:                 migrations.CodeInsights,
	}
)

func MigrateDB(db *sql.DB, database *Database) error {
	m, err := NewMigrate(db, database)
	if err != nil {
		return err
	}
	if err := DoMigrate(m); err != nil {
		return errors.Wrap(err, "Failed to migrate the DB. Please contact support@sourcegraph.com for further assistance")
	}
	return nil
}

// NewMigrate returns a new configured migration object for the given database. The migration can
// be subsequently run by invoking `dbconn.DoMigrate`.
func NewMigrate(db *sql.DB, database *Database) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: database.MigrationsTable,
	})
	if err != nil {
		return nil, err
	}

	d, err := httpfs.New(http.FS(database.FS), ".")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("httpfs", d, "postgres", driver)
	if err != nil {
		return nil, err
	}

	// In case another process was faster and runs migrations, we will wait
	// this long
	m.LockTimeout = 5 * time.Minute
	if os.Getenv("LOG_MIGRATE_TO_STDOUT") != "" {
		m.Log = stdoutLogger{}
	}

	return m, nil
}

// DoMigrate runs all up migrations.
func DoMigrate(m *migrate.Migrate) (err error) {
	err = m.Up()
	if err == nil || err == migrate.ErrNoChange {
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

type stdoutLogger struct{}

func (stdoutLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}
func (logger stdoutLogger) Verbose() bool {
	return true
}

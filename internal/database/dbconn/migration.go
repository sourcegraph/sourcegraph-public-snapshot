package dbconn

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	codeinsightsMigrations "github.com/sourcegraph/sourcegraph/migrations/codeinsights"
	codeintelMigrations "github.com/sourcegraph/sourcegraph/migrations/codeintel"
	frontendMigrations "github.com/sourcegraph/sourcegraph/migrations/frontend"
)

// databases configures the migrations we want based on a database name. This
// configuration includes the name of the migration version table as well as
// the raw migration assets to run to migrate the target schema to a new version.
var databases = map[string]struct {
	MigrationsTable string
	TimescaleDB     bool
	Resource        *bindata.AssetSource
}{
	"frontend": {
		MigrationsTable: "schema_migrations",
		Resource:        bindata.Resource(frontendMigrations.AssetNames(), frontendMigrations.Asset),
	},
	"codeintel": {
		MigrationsTable: "codeintel_schema_migrations",
		Resource:        bindata.Resource(codeintelMigrations.AssetNames(), codeintelMigrations.Asset),
	},
	"codeinsights": {
		TimescaleDB:     true,
		MigrationsTable: "codeinsights_schema_migrations",
		Resource:        bindata.Resource(codeinsightsMigrations.AssetNames(), codeinsightsMigrations.Asset),
	},
}

// PostgresDatabaseNames is the list of database names (configured via `dbutil.databases`) that are
// vanilla Postgres (not TimescaleDB).
var PostgresDatabaseNames = func() []string {
	var names []string
	for databaseName, info := range databases {
		if !info.TimescaleDB {
			names = append(names, databaseName)
		}
	}
	return names
}()

// MigrationTables returns the list of migration table names (configured via `dbutil.databases`).
var MigrationTables = func() []string {
	var migrationTables []string
	for _, db := range databases {
		migrationTables = append(migrationTables, db.MigrationsTable)
	}

	return migrationTables
}()

func MigrateDB(db *sql.DB, databaseName string) error {
	m, err := NewMigrate(db, databaseName)
	if err != nil {
		return err
	}
	if err := DoMigrate(m); err != nil {
		return errors.Wrap(err, "Failed to migrate the DB. Please contact support@sourcegraph.com for further assistance")
	}
	return nil
}

// NewMigrate returns a new configured migration object for the given database name. This database
// name must be present in the `dbconn.databases` map. This migration can be subsequently run by
// invoking `dbconn.DoMigrate`.
func NewMigrate(db *sql.DB, databaseName string) (*migrate.Migrate, error) {
	schemaData, ok := databases[databaseName]
	if !ok {
		return nil, fmt.Errorf("unknown database '%s'", databaseName)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: schemaData.MigrationsTable,
	})
	if err != nil {
		return nil, err
	}

	d, err := bindata.WithInstance(schemaData.Resource)
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("go-bindata", d, "postgres", driver)
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

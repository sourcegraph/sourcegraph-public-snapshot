package dbconn

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func migrateDB(db *sql.DB, schema *schemas.Schema) (func(), error) {
	m, err := newMigrate(db, schema)
	if err != nil {
		return nil, err
	}

	if err := doMigrate(m); err != nil {
		return nil, errors.Wrap(err, "Failed to migrate the DB. Please contact support@sourcegraph.com for further assistance")
	}

	return func() { m.Close() }, nil
}

// newMigrate returns a new configured migration object for the given database. The migration can
// be subsequently run by invoking `dbconn.DoMigrate`.
func newMigrate(db *sql.DB, schema *schemas.Schema) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: schema.MigrationsTableName,
	})
	if err != nil {
		return nil, err
	}

	d, err := httpfs.New(http.FS(schema.FS), ".")
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

// doMigrate runs all up migrations.
func doMigrate(m *migrate.Migrate) (err error) {
	err = m.Up()
	if err == nil || err == migrate.ErrNoChange {
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
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

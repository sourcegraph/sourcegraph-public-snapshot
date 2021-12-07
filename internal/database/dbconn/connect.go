package dbconn

import (
	"database/sql"

	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

// ConnectInternal connects to the given data source and return the handle. After successful connection,
// the schema version of the database will be compared against an expected version and the supplied migrations
// may be run (taking an advisory lock to ensure exclusive access).
//
// This function returns a basestore-style callback that closes the database. This should be called
// instead of calling Close directly on the database handle as it also handles closing migration objects
// associated with the handle.
//
// If appName is supplied, then it overrides the application_name field in the DSN. This is a separate
// parameter needed because we have multiple apps connecting to the same database but have a single shared
// DSN configured.
//
// If dbName is supplied, then metrics will be reported for activity on the returned handle. This value is
// used for its Prometheus label value instead of whatever actual value is set in dataSource.
//
// Note: github.com/jackc/pgx parses the environment as well. This function will also use the value
// of PGDATASOURCE if supplied and dataSource is the empty string.
func ConnectInternal(dsn, appName, dbName string, schemas []*schemas.Schema) (*sql.DB, func(err error) error, error) {
	cfg, err := buildConfig(dsn, appName)
	if err != nil {
		return nil, nil, err
	}

	db, err := newWithConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	var closeFns []func()

	closeAll := func(err error) error {
		for i := len(closeFns) - 1; i >= 0; i-- {
			closeFns[i]()
		}

		if closeErr := db.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
		}

		return err
	}

	for _, schema := range schemas {
		close, err := migrateDB(db, schema)
		if err != nil {
			return nil, nil, closeAll(err)
		}

		closeFns = append(closeFns, close)
	}

	if dbName != "" {
		if err := prometheus.Register(newMetricsCollector(db, dbName, appName)); err != nil {
			return nil, nil, closeAll(err)
		}
	}

	return db, closeAll, nil
}

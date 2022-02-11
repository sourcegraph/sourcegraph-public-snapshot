package dbconn

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ConnectInternal connects to the given data source and return the handle.
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
func ConnectInternal(dsn, appName, dbName string) (_ *sql.DB, err error) {
	cfg, err := buildConfig(dsn, appName)
	if err != nil {
		return nil, err
	}

	db, err := newWithConfig(cfg)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if closeErr := db.Close(); closeErr != nil {
				err = errors.Append(err, closeErr)
			}
		}
	}()

	if dbName != "" {
		if err := prometheus.Register(newMetricsCollector(db, dbName, appName)); err != nil {
			return nil, err
		}
	}

	return db, nil
}

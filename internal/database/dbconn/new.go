package dbconn

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

// Opts contain arguments passed to database connection initialisation functions.
type Opts struct {
	// DSN (data source name) is a URI like string containing all data needed to connect to the database.
	DSN string

	// DBName is used only for Prometheus metrics instead of whatever actual database name is set in DSN.
	// This is needed because in our dev environment we use a single physical database (and DSN) for all our different
	// logical databases.
	DBName string

	// AppName overrides the application_name in the DSN. This separate parameter is needed
	// because we have multiple apps connecting to the same database, but have a single shared DSN configured.
	AppName string
}

// New connects to the given data source and returns the handle.
//
// If dbname is set then metric will be reported for the returned handle.
// dbname is used for its Prometheus label value instead of whatever actual value is set in dataSource.
// This is needed because in our dev environment we use a single physical database (and DSN) for all our different
// logical databases. app, however is set as the application_name in the connection string. This is needed
// because we have multiple apps connecting to the same database, but have a single shared DSN.
//
// Note: github.com/jackc/pgx parses the environment as well. This function will
// also use the value of PGDATASOURCE if supplied and dataSource is the empty
// string.
func New(opts Opts) (*sql.DB, error) {
	cfg, err := buildConfig(opts.DSN, opts.AppName)
	if err != nil {
		return nil, err
	}

	db, err := newWithConfig(cfg)
	if err != nil {
		return nil, err
	}

	if opts.DBName != "" {
		prometheus.MustRegister(newMetricsCollector(db, opts.DBName, opts.AppName))
	}

	return db, nil
}

package dbconn

import "database/sql"

var (
	// Global is the global DB connection.
	// Only use this after a call to SetupGlobalConnection.
	//
	// Soon to be replaced: Pass on DB interface as an argument instead.
	Global *sql.DB
)

// SetupGlobalConnection connects to the given data source and stores the handle
// globally.
//
// dbname is used for its Prometheus label value instead of whatever actual value is set in dataSource.
// This is needed because in our dev environment we use a single physical database (and DSN) for all our different
// logical databases. app, however is set as the application_name in the connection string. This is needed
// because we have multiple apps connecting to the same database, but have a single shared DSN.
//
// Note: github.com/jackc/pgx parses the environment as well. This function will
// also use the value of PGDATASOURCE if supplied and dataSource is the empty
// string.
//
// Soon to be replaced: Pass on DB interface as an argument instead.
func SetupGlobalConnection(opts Opts) (err error) {
	Global, err = New(opts)
	return err
}

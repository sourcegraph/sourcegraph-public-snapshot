package dbconn

import (
	"context"
	"database/sql"
	"log"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Restricted is the "sg_service" DB connection used to provide row level security.
	// See https://docs.sourcegraph.com/admin/repo/row_level_security for more information.
	// Only use this after a call to SetupRestrictedConnection.
	Restricted *sql.DB
)

// SetupRestrictedConnection connects to the given data source, ensures reduced
// privileges, and stores the handle globally.
//
// Each connection in the pool will adopt the "sg_service" role upon being checked
// out. This role has row security policies applied to it so that various authn/z
// checks are handled at the Postgres level instead of application code. This lowers
// the burden on developers, since you only have to opt-in to using the connection
// vs. invoking specific functions at specific times.
//
// SetupRestrictedConnection does not handle the various liveness checks
// (e.g. is the database up). Thus the restricted connection must be configured after the
//  base connection is set up in SetupGlobalConnection which will ensure that the
// liveness checks have already been set up.
func SetupRestrictedConnection(opts Opts) (err error) {
	Restricted, err = NewRestricted(opts)
	return err
}

// NewRestricted connects to the given data source, sets up the connection lifecycle
// hooks, and returns the handle to the pool.
func NewRestricted(opts Opts) (*sql.DB, error) {
	opts.AppName += "-restricted"
	config, err := buildConfig(opts.DSN, opts.AppName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to build restricted db config")
	}

	// Connect to the database and register Prometheus metrics for observability.
	db := stdlib.OpenDB(*config, stdlib.OptionAfterConnect(func(ctx context.Context, conn *pgx.Conn) error {
		if _, err := conn.Exec(ctx, "SET SESSION AUTHORIZATION 'sg_service';"); err != nil {
			log.Println(errors.Wrap(err, "unable to assume sg_service role"))
			return err
		}
		return nil
	}))
	if db == nil {
		return nil, errors.New("unable to open restricted db connection")
	}
	prometheus.MustRegister(newMetricsCollector(db, opts.DBName, opts.AppName))
	return db, nil
}

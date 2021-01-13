// Package dbconn provides functionality to connect to our DB and migrate it.
//
// Most services should connect to the frontend for DB access instead, using
// api.InternalClient.
package dbconn

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gchaincl/sqlhooks"
	"github.com/inconshreveable/log15"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var (
	// Global is the global DB connection.
	// Only use this after a call to SetupGlobalConnection.
	Global *sql.DB

	defaultDataSource      = env.Get("PGDATASOURCE", "", "Default dataSource to pass to Postgres. See https://godoc.org/github.com/jackc/pgx for more information.")
	defaultApplicationName = env.Get("PGAPPLICATIONNAME", "sourcegraph", "The value of application_name appended to dataSource")
)

// SetupGlobalConnection connects to the given data source and stores the handle
// globally.
//
// Note: github.com/jackc/pgx parses the environment as well. This function will
// also use the value of PGDATASOURCE if supplied and dataSource is the empty
// string.
func SetupGlobalConnection(dataSource string) (err error) {
	Global, err = New(dataSource, "_app")
	return err
}

// New connects to the given data source and returns the handle.
//
// Note: github.com/jackc/pgx parses the environment as well. This function will
// also use the value of PGDATASOURCE if supplied and dataSource is the empty
// string.
func New(dataSource, dbNameSuffix string) (*sql.DB, error) {
	// Force PostgreSQL session timezone to UTC.
	if v, ok := os.LookupEnv("PGTZ"); ok && v != "UTC" && v != "utc" {
		log15.Warn("Ignoring PGTZ environment variable; using PGTZ=UTC.", "ignoredPGTZ", v)
	}
	if err := os.Setenv("PGTZ", "UTC"); err != nil {
		return nil, errors.Wrap(err, "Error setting PGTZ=UTC")
	}

	cfg, err := buildConfig(dataSource)
	if err != nil {
		return nil, err
	}

	db, err := openDBWithStartupWait(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "DB not available")
	}
	registerPrometheusCollector(db, dbNameSuffix)
	configureConnectionPool(db)
	return db, nil
}

func MigrateDB(db *sql.DB, databaseName string) error {
	m, err := dbutil.NewMigrate(db, databaseName)
	if err != nil {
		return err
	}
	if err := dbutil.DoMigrate(m); err != nil {
		return errors.Wrap(err, "Failed to migrate the DB. Please contact support@sourcegraph.com for further assistance")
	}
	return nil
}

var startupTimeout = func() time.Duration {
	str := env.Get("DB_STARTUP_TIMEOUT", "10s", "keep trying for this long to connect to PostgreSQL database before failing")
	d, err := time.ParseDuration(str)
	if err != nil {
		log.Fatalln("DB_STARTUP_TIMEOUT:", err)
	}
	return d
}()

// buildConfig takes either a Postgres connection string or connection URI,
// parses it, and returns a config with additional parameters.
func buildConfig(dataSource string) (*pgx.ConnConfig, error) {
	if dataSource == "" {
		dataSource = defaultDataSource
	}

	cfg, err := pgx.ParseConfig(dataSource)
	if err != nil {
		return nil, err
	}

	if cfg.RuntimeParams == nil {
		cfg.RuntimeParams = make(map[string]string)
	}

	// pgx doesn't support fallback_application_name so we emulate it
	// by checking if application_name is set and setting a default
	// value if not.
	if _, ok := cfg.RuntimeParams["application_name"]; !ok {
		cfg.RuntimeParams["application_name"] = defaultApplicationName
	}

	return cfg, nil
}

func openDBWithStartupWait(cfg *pgx.ConnConfig) (db *sql.DB, err error) {
	// Allow the DB to take up to 10s while it reports "pq: the database system is starting up".
	startupDeadline := time.Now().Add(startupTimeout)
	for {
		if time.Now().After(startupDeadline) {
			return nil, fmt.Errorf("database did not start up within %s (%v)", startupTimeout, err)
		}
		db, err = open(cfg)
		if err == nil {
			err = db.Ping()
		}
		if err != nil && isDatabaseLikelyStartingUp(err) {
			time.Sleep(startupTimeout / 10)
			continue
		}
		return db, err
	}
}

// isDatabaseLikelyStartingUp returns whether the err likely just means the PostgreSQL database is
// starting up, and it should not be treated as a fatal error during program initialization.
func isDatabaseLikelyStartingUp(err error) bool {
	if strings.Contains(err.Error(), "pq: the database system is starting up") {
		// Wait for DB to start up.
		return true
	}
	if e, ok := errors.Cause(err).(net.Error); ok && strings.Contains(e.Error(), "connection refused") {
		// Wait for DB to start listening.
		return true
	}
	return false
}

var registerOnce sync.Once

// Open creates a new DB handle with the given schema by connecting to
// the database identified by dataSource (e.g., "dbname=mypgdb" or
// blank to use the PG* env vars).
//
// Open assumes that the database already exists.
func Open(dataSource string) (*sql.DB, error) {
	cfg, err := pgx.ParseConfig(dataSource)
	if err != nil {
		return nil, err
	}

	return open(cfg)
}

func open(cfg *pgx.ConnConfig) (*sql.DB, error) {
	cfgKey := stdlib.RegisterConnConfig(cfg)

	registerOnce.Do(func() {
		sql.Register("postgres-proxy", sqlhooks.Wrap(stdlib.GetDefaultDriver(), &hook{}))
	})
	db, err := sql.Open("postgres-proxy", cfgKey)
	if err != nil {
		return nil, errors.Wrap(err, "postgresql open")
	}
	return db, nil
}

// Ping attempts to contact the database and returns a non-nil error upon failure. It is intended to
// be used by health checks.
func Ping(ctx context.Context) error { return Global.PingContext(ctx) }

type key int

const bulkInsertionKey key = iota

// BulkInsertion returns true if the bulkInsertionKey context value is true.
func BulkInsertion(ctx context.Context) bool {
	v, ok := ctx.Value(bulkInsertionKey).(bool)
	if !ok {
		return false
	}
	return v
}

// WithBulkInsertion sets the bulkInsertionKey context value.
func WithBulkInsertion(ctx context.Context, bulkInsertion bool) context.Context {
	return context.WithValue(ctx, bulkInsertionKey, bulkInsertion)
}

type hook struct{}

// postgresBulkInsertRowsPattern matches `($1, $2, $3), ($4, $5, $6), ...` which
// we use to cut out the row payloads from bulk insertion tracing data. We don't
// need all the parameter data for such requests, which are too big to fit into
// Jaeger spans. Note that we don't just capture `($1.*`, as we want queries with
// a trailing ON CONFLICT clause not to be semantically mangled in the log output.
var postgresBulkInsertRowsPattern = lazyregexp.New(`(\([$\d,\s]+\)[,\s]*)+`)

// postgresBulkInsertRowsReplacement replaces the all-placeholder rows matched
// by the pattern defined above.
var postgresBulkInsertRowsReplacement = []byte("(...) ")

// Before implements sqlhooks.Hooks
func (h *hook) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	if BulkInsertion(ctx) {
		query = string(postgresBulkInsertRowsPattern.ReplaceAll([]byte(query), postgresBulkInsertRowsReplacement))
	}

	tr, ctx := trace.New(ctx, "sql", query,
		trace.Tag{Key: "span.kind", Value: "client"},
		trace.Tag{Key: "db.type", Value: "sql"},
	)

	if !BulkInsertion(ctx) {
		tr.LogFields(otlog.Lazy(func(fv otlog.Encoder) {
			for i, arg := range args {
				fv.EmitString(strconv.Itoa(i+1), fmt.Sprintf("%v", arg))
			}
		}))
	} else {
		tr.LogFields(otlog.Bool("bulk_insert", true), otlog.Int("num_args", len(args)))
	}

	return ctx, nil
}

// After implements sqlhooks.Hooks
func (h *hook) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	if tr := trace.TraceFromContext(ctx); tr != nil {
		tr.Finish()
	}
	return ctx, nil
}

// After implements sqlhooks.OnErroer
func (h *hook) OnError(ctx context.Context, err error, query string, args ...interface{}) error {
	if tr := trace.TraceFromContext(ctx); tr != nil {
		tr.SetError(err)
		tr.Finish()
	}
	return err
}

func registerPrometheusCollector(db *sql.DB, dbNameSuffix string) {
	c := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "src",
			Subsystem: "pgsql" + dbNameSuffix,
			Name:      "open_connections",
			Help:      "Number of open connections to pgsql DB, as reported by pgsql.DB.Stats()",
		},
		func() float64 {
			s := db.Stats()
			return float64(s.OpenConnections)
		},
	)
	prometheus.MustRegister(c)
}

// configureConnectionPool sets reasonable sizes on the built in DB queue. By
// default the connection pool is unbounded, which leads to the error `pq:
// sorry too many clients already`.
func configureConnectionPool(db *sql.DB) {
	var err error
	maxOpen := 30
	if e := os.Getenv("SRC_PGSQL_MAX_OPEN"); e != "" {
		maxOpen, err = strconv.Atoi(e)
		if err != nil {
			log.Fatalf("SRC_PGSQL_MAX_OPEN is not an int: %s", e)
		}
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxOpen)
	db.SetConnMaxLifetime(time.Minute)
}

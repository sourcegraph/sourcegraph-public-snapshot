package dbconn

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/qustavo/sqlhooks/v2"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var startupTimeout = func() time.Duration {
	str := env.Get("DB_STARTUP_TIMEOUT", "10s", "keep trying for this long to connect to PostgreSQL database before failing")
	d, err := time.ParseDuration(str)
	if err != nil {
		log.Fatalln("DB_STARTUP_TIMEOUT:", err)
	}
	return d
}()

var defaultMaxOpen = func() int {
	str := env.Get("SRC_PGSQL_MAX_OPEN", "30", "Maximum number of open connections to Postgres")
	v, err := strconv.Atoi(str)
	if err != nil {
		log.Fatalln("SRC_PGSQL_MAX_OPEN:", err)
	}
	return v
}()

var defaultMaxIdle = func() int {
	// For now, use the old default of max_idle == max_open
	str := env.Get("SRC_PGSQL_MAX_IDLE", "30", "Maximum number of idle connections to Postgres")
	v, err := strconv.Atoi(str)
	if err != nil {
		log.Fatalln("SRC_PGSQL_MAX_IDLE:", err)
	}
	return v
}()

func newWithConfig(cfg *pgx.ConnConfig) (*sql.DB, error) {
	db, err := openDBWithStartupWait(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "DB not available")
	}

	if err := ensureMinimumPostgresVersion(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func openDBWithStartupWait(cfg *pgx.ConnConfig) (db *sql.DB, err error) {
	// Allow the DB to take up to 10s while it reports "pq: the database system is starting up".
	startupDeadline := time.Now().Add(startupTimeout)
	for {
		if time.Now().After(startupDeadline) {
			return nil, errors.Errorf("database did not start up within %s (%v)", startupTimeout, err)
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

// extendedDriver wraps sqlHooks' driver to provide a conn that implements Ping, ResetSession
// and CheckNamedValue, which is mandatory as otelsql is instrumenting these methods.
// For all mandatory methods the sqlHooks driver is used. For the optional methods namely Ping, ResetSession and CheckNamedValue
// (which the sqlHooks driver does not implement), extendedConn goes to the original default driver.
//
//                             Ping()
//                             ResetSession()
//                             CheckNamedValue()
//                    ┌──────────────────────────────┐
//                    │                              │
//                    │                              │
//                    │                              │
// ┌───────┐   ┌──────┴─────┐   ┌────────┐     ┌─────▼───────┐
// │       │   │            │   │        │     │             │
// │otelsql├──►│extendedConn├──►│sqlhooks├────►│DefaultDriver│
// │       │   │            │   │        │     │             │
// └─┬─────┘   └─┬──────────┘   └─┬──────┘     └─┬───────────┘
//   │           │                │              │
//   │           │                │              │Implements all SQL driver methods
//   │           │                │
//   │           │                │Only implements mandatory ones
//   │           │                │Ping(), ResetSession() and CheckNamedValue() are missing.
//   │           │
//   │           │Implement all SQL driver methods
//   │
//   │Expects all SQL driver methods
//
// A sqlhooks.Driver must be used as a Driver otherwise errors will be raised.
type extendedDriver struct {
	driver.Driver
}

// extendedConn wraps sqlHooks' conn that does implement Ping, ResetSession and
// CheckNamedValue into one that does, by accessing the underlying conn from the
// original driver that does implement these methods.
type extendedConn struct {
	driver.ExecerContext
	driver.QueryerContext
	driver.Conn
	driver.ConnPrepareContext
	driver.ConnBeginTx
}

var _ driver.Pinger = &extendedConn{}
var _ driver.SessionResetter = &extendedConn{}
var _ driver.NamedValueChecker = &extendedConn{}

// Open returns a conn wrapped through extendedConn, implementing the
// Ping, ResetSession and CheckNamedValue optional methods that the
// otelsql.Conn expects to be implemented.
func (d *extendedDriver) Open(str string) (driver.Conn, error) {
	if _, ok := d.Driver.(*sqlhooks.Driver); !ok {
		return nil, errors.New("sql driver is not a sqlhooks.Driver")
	}
	c, err := d.Driver.Open(str)
	if err != nil {
		return nil, err
	}

	// Ensure we're not casting things blindly.
	if _, ok := c.(any).(driver.ExecerContext); !ok {
		return nil, errors.New("sql conn doen't implement driver.ExecerContext")
	}
	if _, ok := c.(any).(driver.QueryerContext); !ok {
		return nil, errors.New("sql conn doen't implement driver.QueryerContext")
	}
	if _, ok := c.(any).(driver.Conn); !ok {
		return nil, errors.New("sql conn doen't implement driver.Conn")
	}
	if _, ok := c.(any).(driver.ConnPrepareContext); !ok {
		return nil, errors.New("sql conn doen't implement driver.ConnPrepareContext")
	}
	if _, ok := c.(any).(driver.ConnBeginTx); !ok {
		return nil, errors.New("sql conn doen't implement driver.ConnBeginTx")
	}

	// Build the extended connection.
	return &extendedConn{
		ExecerContext:      c.(any).(driver.ExecerContext),
		QueryerContext:     c.(any).(driver.QueryerContext),
		Conn:               c.(any).(driver.Conn),
		ConnPrepareContext: c.(any).(driver.ConnPrepareContext),
		ConnBeginTx:        c.(any).(driver.ConnBeginTx),
	}, nil
}

// Access the underlying connection, so we can forward the methods that
// sqlhooks does not implement on its own.
func (n *extendedConn) rawConn() driver.Conn {
	c := n.Conn.(*sqlhooks.ExecerQueryerContextWithSessionResetter)
	return c.Conn.Conn
}

func (n *extendedConn) Ping(ctx context.Context) error {
	return n.rawConn().(driver.Pinger).Ping(ctx)
}

func (n *extendedConn) ResetSession(ctx context.Context) error {
	return n.rawConn().(driver.SessionResetter).ResetSession(ctx)
}

func (n *extendedConn) CheckNamedValue(namedValue *driver.NamedValue) error {
	return n.rawConn().(driver.NamedValueChecker).CheckNamedValue(namedValue)
}

func registerPostgresProxy() {
	m := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_pgsql_request_total",
		Help: "Total number of SQL requests to the database.",
	}, []string{"type"})

	dri := sqlhooks.Wrap(stdlib.GetDefaultDriver(), combineHooks(
		&metricHooks{
			metricSQLSuccessTotal: m.WithLabelValues("success"),
			metricSQLErrorTotal:   m.WithLabelValues("error"),
		},
	))
	sql.Register("postgres-proxy", &extendedDriver{dri})
}

var registerOnce sync.Once

func open(cfg *pgx.ConnConfig) (*sql.DB, error) {
	registerOnce.Do(registerPostgresProxy)

	db, err := otelsql.Open(
		"postgres-proxy",
		stdlib.RegisterConnConfig(cfg),
		otelsql.WithTracerProvider(otel.GetTracerProvider()),
		otelsql.WithSQLCommenter(true),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			OmitConnResetSession: true,
			ArgumentOptions: otelsql.ArgumentOptions{
				EnableAttributes: true,
				Skip: func(ctx context.Context, query string, args []any) bool {
					// Do not decorate span with args as attributes if that's a bulk insertion
					// or if we have too many args (it's unreadable anyway).
					return isBulkInsertion(ctx) || len(args) > 24
				}},
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "postgresql open")
	}

	// Set max open and idle connections
	maxOpen, _ := strconv.Atoi(cfg.RuntimeParams["max_conns"])
	if maxOpen == 0 {
		maxOpen = defaultMaxOpen
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(defaultMaxIdle)
	db.SetConnMaxIdleTime(time.Minute)

	return db, nil
}

// isDatabaseLikelyStartingUp returns whether the err likely just means the PostgreSQL database is
// starting up, and it should not be treated as a fatal error during program initialization.
func isDatabaseLikelyStartingUp(err error) bool {
	substrings := []string{
		// Wait for DB to start up.
		"the database system is starting up",
		// Wait for DB to start listening.
		"connection refused",
		"failed to receive message",
	}

	msg := err.Error()
	for _, substring := range substrings {
		if strings.Contains(msg, substring) {
			return true
		}
	}

	return false
}

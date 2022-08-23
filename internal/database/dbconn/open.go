package dbconn

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/qustavo/sqlhooks/v2"

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

func registerPostgresProxy() {
	m := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_pgsql_request_total",
		Help: "Total number of SQL requests to the database.",
	}, []string{"type"})

	sql.Register("postgres-proxy", sqlhooks.Wrap(stdlib.GetDefaultDriver(), combineHooks(
		&metricHooks{
			metricSQLSuccessTotal: m.WithLabelValues("success"),
			metricSQLErrorTotal:   m.WithLabelValues("error"),
		},
		&tracingHooks{},
	)))
}

var registerOnce sync.Once

func open(cfg *pgx.ConnConfig) (*sql.DB, error) {
	registerOnce.Do(registerPostgresProxy)

	db, err := sql.Open("postgres-proxy", stdlib.RegisterConnConfig(cfg))
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

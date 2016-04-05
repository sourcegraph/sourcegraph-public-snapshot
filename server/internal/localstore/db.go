package localstore

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/sourcegraph/util/dbutil2"
)

var (
	// Schema is the DB Schema for the database used by this package.
	Schema = dbutil2.Schema{
		CreateSQL: []string{
			`CREATE EXTENSION IF NOT EXISTS citext;`,
			`CREATE EXTENSION IF NOT EXISTS hstore;`,
		},
		Map: &gorp.DbMap{Dialect: gorp.PostgresDialect{}},
	}
)

var (
	globalDBH *dbutil2.Handle // global DB handle
	dbLock    sync.Mutex      // protects globalDBH
)

// globalDB opens the DB if it isn't already open, and returns
// it. Subsequent calls return the same DB handle.
func globalDB() (*dbutil2.Handle, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	if globalDBH != nil {
		return globalDBH, nil
	}

	dbh, err := OpenDB(0)
	if err != nil {
		return nil, err
	}
	registerPrometheusCollector(dbh.DbMap.Db)
	limitConnectionPool(dbh.DbMap.Db)

	globalDBH = dbh
	return globalDBH, nil
}

// OpenDB opens and returns the DB handle for the DB. Use DB unless
// you need access to the low-level DB handle or need to handle
// errors.
func OpenDB(mode dbutil2.Mode) (*dbutil2.Handle, error) {
	dbh, err := dbutil2.Open("", Schema, mode)
	if err != nil {
		return nil, fmt.Errorf("open DB: %s", err)
	}
	return dbh, nil
}

func registerPrometheusCollector(db *sql.DB) {
	c := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: "src",
			Subsystem: "pgsql",
			Name:      "open_connections",
			Help:      "Number of open connections to pgsql globalDB, as reported by pgsql.DB.Stats()",
		},
		func() float64 {
			s := db.Stats()
			return float64(s.OpenConnections)
		},
	)
	prometheus.MustRegister(c)
}

// limitConnectionPool sets reasonable sizes on the built in DB queue. By
// default the connection pool is unbounded, which leads to the error `pq:
// sorry too many clients already`.
func limitConnectionPool(db *sql.DB) {
	// The default value for max_connections is 100 in pgsql. Defaults
	// allow for roughly 3 servers to burst. We don't change idle
	// connection size, which defaults to 2
	var err error
	maxOpen := 30
	if e := os.Getenv("SRC_PGSQL_MAX_OPEN"); e != "" {
		maxOpen, err = strconv.Atoi(e)
		if err != nil {
			log.Fatalf("SRC_PGSQL_MAX_OPEN is not an int: %s", e)
		}
	}
	db.SetMaxOpenConns(maxOpen)
}

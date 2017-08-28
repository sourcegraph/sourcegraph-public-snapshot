package localstore

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil2"
)

var (
	globalAppDB *sql.DB
	dbLock      sync.Mutex
)

// globalDB opens the app DB if it isn't already open,
// and returns it. Subsequent calls return the same DB handle.
func globalDB() (*sql.DB, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	if globalAppDB == nil {
		var err error
		globalAppDB, err = dbutil2.Open("")
		if err != nil {
			return nil, err
		}
		registerPrometheusCollector(globalAppDB, "_app")
		configureConnectionPool(globalAppDB)

		// TODO migrate
	}

	return globalAppDB, nil
}

type key int

const dbKey key = 0

// appDBH returns the app DB handle.
func appDBH(ctx context.Context) *sql.DB {
	db, ok := ctx.Value(dbKey).(*sql.DB)
	if ok {
		return db
	}
	db, err := globalDB()
	if err != nil {
		panic("DB not available: " + err.Error())
	}
	return db
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
}

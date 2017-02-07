package localstore

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil2"
)

var (
	// Note: the reason why DB table creation is done here (using the
	// CreateTable / DropTable pattern) is because tables often rely on the
	// table creation order and we should encourage our DB to be relational
	// (i.e. tables referencing eachother). This is hard to do inside of
	// multiple package init's, so we register all table creators/destructors
	// here.

	// AppSchema is the DB Schema for the app database used by this package.
	// Currently, all db stores except GlobalRefs are grouped under AppSchema.
	AppSchema = dbutil2.Schema{
		CreateSQL: []string{
			`CREATE EXTENSION IF NOT EXISTS citext;`,
			`CREATE EXTENSION IF NOT EXISTS hstore;`,
			new(globalDeps).CreateTable(),
			new(pkgs).CreateTable(),
		},
		DropSQL: []string{
			new(globalDeps).DropTable(),
			new(pkgs).DropTable(),
		},
		Map: &gorp.DbMap{Dialect: gorp.PostgresDialect{}},
	}
)

var (
	globalAppDBH *dbutil2.Handle // global app DB handle
	dbLock       sync.Mutex      // protects globalDBH
)

// globalDB opens the app DB if it isn't already open,
// and returns it. Subsequent calls return the same DB handle.
func globalDB() (*dbutil2.Handle, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	if globalAppDBH == nil {
		var err error
		globalAppDBH, err = openDB("", AppSchema)
		if err != nil {
			return nil, err
		}
		registerPrometheusCollector(globalAppDBH.DbMap.Db, "_app")
		configureConnectionPool(globalAppDBH.DbMap.Db)

		if _, err := globalAppDBH.Db.Query("select id from repo limit 0;"); err != nil {
			if err := globalAppDBH.CreateSchema(); err != nil {
				return nil, err
			}
		}
	}

	return globalAppDBH, nil
}

type key int

const dbhKey key = 0

// appDBH returns the app DB handle.
func appDBH(ctx context.Context) *dbutil2.Handle {
	dbh, ok := ctx.Value(dbhKey).(*dbutil2.Handle)
	if ok {
		return dbh
	}
	dbh, err := globalDB()
	if err != nil {
		panic("DB not available: " + err.Error())
	}
	return dbh
}

// openDB opens and returns the DB handle for the DB. Use DB unless
// you need access to the low-level DB handle or need to handle
// errors.
func openDB(dataSource string, schema dbutil2.Schema) (*dbutil2.Handle, error) {
	dbh, err := dbutil2.Open(dataSource, schema)
	if err != nil {
		return nil, fmt.Errorf("open DB (%s): %s", dataSource, err)
	}
	return dbh, nil
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

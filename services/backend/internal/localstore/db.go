package localstore

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
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
		},
		Map: &gorp.DbMap{Dialect: gorp.PostgresDialect{}},
	}

	// GraphSchema is the DB Schema for the graphstore database used by this package.
	// Currently, only GlobalRefs store is grouped under GraphSchema.
	GraphSchema = dbutil2.Schema{
		CreateSQL: []string{
			`CREATE EXTENSION IF NOT EXISTS citext;`,
			`CREATE EXTENSION IF NOT EXISTS hstore;`,

			// global_ref_* table creation.
			new(dbGlobalRefSource).CreateTable(),
			new(dbGlobalRefVersion).CreateTable(),
			new(dbGlobalRefFile).CreateTable(),
			new(dbGlobalRefName).CreateTable(),
			new(dbGlobalRefContainer).CreateTable(),
			new(dbGlobalRefBySource).CreateTable(),
			new(dbGlobalRefByFile).CreateTable(),
		},
		DropSQL: []string{
			// global_ref_* table deletion.
			new(dbGlobalRefSource).DropTable(),
			new(dbGlobalRefVersion).DropTable(),
			new(dbGlobalRefFile).DropTable(),
			new(dbGlobalRefName).DropTable(),
			new(dbGlobalRefContainer).DropTable(),
			new(dbGlobalRefBySource).DropTable(),
			new(dbGlobalRefByFile).DropTable(),
		},
		Map: &gorp.DbMap{Dialect: gorp.PostgresDialect{}},
	}
)

var (
	globalAppDBH   *dbutil2.Handle // global app DB handle
	globalGraphDBH *dbutil2.Handle // global graph DB handle
	dbLock         sync.Mutex      // protects globalDBH
)

// GlobalDBs opens the app and graph DBs if they aren't already open,
// and returns them. Subsequent calls return the same DB handles.
func GlobalDBs() (*dbutil2.Handle, *dbutil2.Handle, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	if globalAppDBH == nil || globalGraphDBH == nil {
		var err error
		globalAppDBH, err = openDB(getAppDBDataSource(), AppSchema, 0)
		if err != nil {
			return nil, nil, err
		}
		registerPrometheusCollector(globalAppDBH.DbMap.Db, "_app")
		configureConnectionPool(globalAppDBH.DbMap.Db)

		globalGraphDBH, err = openDB(getGraphDBDataSource(), GraphSchema, 0)
		if err != nil {
			return nil, nil, err
		}
		// If graph db has the same data source as app db, they will share the
		// underlying *sql.Db handle, so we do not re-register the prometheus
		// metric or limit the connection pool again on the same db handle.
		if globalGraphDBH.DbMap.Db != globalAppDBH.DbMap.Db {
			registerPrometheusCollector(globalGraphDBH.DbMap.Db, "_graph")
			configureConnectionPool(globalGraphDBH.DbMap.Db)
		}
	}

	return globalAppDBH, globalGraphDBH, nil
}

// appDBH returns the app DB handle.
func appDBH(ctx context.Context) gorp.SqlExecutor {
	appDBH, _, err := GlobalDBs()
	if err != nil {
		panic("DB not available: " + err.Error())
	}
	return traceutil.SQLExecutor{
		SqlExecutor: appDBH,
		Context:     ctx,
	}
}

// graphDBH returns the graph DB handle.
func graphDBH(ctx context.Context) gorp.SqlExecutor {
	_, graphDBH, err := GlobalDBs()
	if err != nil {
		panic("DB not available: " + err.Error())
	}
	return traceutil.SQLExecutor{
		SqlExecutor: graphDBH,
		Context:     ctx,
	}
}

// openDB opens and returns the DB handle for the DB. Use DB unless
// you need access to the low-level DB handle or need to handle
// errors.
func openDB(dataSource string, schema dbutil2.Schema, mode dbutil2.Mode) (*dbutil2.Handle, error) {
	dbh, err := dbutil2.Open(dataSource, schema, mode)
	if err != nil {
		return nil, fmt.Errorf("open DB (%s): %s", dataSource, err)
	}
	return dbh, nil
}

func getAppDBDataSource() string {
	// libpq defaults to the PG* env variable values when
	// data source is empty.
	return ""
}

func getGraphDBDataSource() string {
	graphDBEnv := map[string]string{
		"host":     os.Getenv("SG_GRAPH_PGHOST"),
		"port":     os.Getenv("SG_GRAPH_PGPORT"),
		"user":     os.Getenv("SG_GRAPH_PGUSER"),
		"password": os.Getenv("SG_GRAPH_PGPASSWORD"),
		"dbname":   os.Getenv("SG_GRAPH_PGDATABASE"),
		"sslmode":  os.Getenv("SG_GRAPH_PGSSLMODE"),
	}

	var dataSource []string
	for k, v := range graphDBEnv {
		if v != "" {
			dataSource = append(dataSource, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return strings.Join(dataSource, " ")
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

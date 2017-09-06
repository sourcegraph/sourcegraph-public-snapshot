package localstore

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/postgres"
	bindata "github.com/mattes/migrate/source/go-bindata"
	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore/migrations"
)

var (
	globalDB      *sql.DB
	globalMigrate *migrate.Migrate
)

// ConnectToDB connects to the given DB and stores the handle globally.
func ConnectToDB(dataSource string) {
	var err error
	globalDB, err = dbutil2.Open(dataSource)
	if err != nil {
		log.Fatal("DB not available: " + err.Error())
	}
	registerPrometheusCollector(globalDB, "_app")
	configureConnectionPool(globalDB)

	globalMigrate = newMigrate(globalDB)

	// support for legacy tables
	if _, _, err := globalMigrate.Version(); err == migrate.ErrNilVersion {
		if _, err := globalDB.Query("select id from repo limit 0;"); err == nil {
			// no version in DB, but "repo" table exists
			if err := globalMigrate.Force(1503575588); err != nil {
				log.Fatal("error force migrating DB: " + err.Error())
			}
		}
	}

	if err := globalMigrate.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("error migrating DB: " + err.Error())
	}
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

func newMigrate(db *sql.DB) *migrate.Migrate {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	s := bindata.Resource(migrations.AssetNames(), migrations.Asset)
	d, err := bindata.WithInstance(s)
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithInstance("go-bindata", d, "postgres", driver)
	if err != nil {
		log.Fatal(err)
	}

	return m
}

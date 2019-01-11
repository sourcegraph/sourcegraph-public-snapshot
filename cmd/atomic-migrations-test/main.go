package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/lib/pq"
	"github.com/sourcegraph/sourcegraph/migrations"

	"github.com/pkg/errors"
)

func main() {
	fs := flag.NewFlagSet("atomic-migrations-test", flag.ExitOnError)
	concurrency := fs.Int("concurrency", 2, "Number of concurrent database migration processes")
	dsn := fs.String(
		"dsn",
		"postgres://sourcegraph:sourcegraph@127.0.0.1/postgres?sslmode=disable&timezone=UTC",
		"Database connection string to use",
	)

	_ = fs.Parse(os.Args[1:])

	if err := run(*concurrency, *dsn); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("No errors migrating with %d concurrent processes", *concurrency)
	}
}

func run(concurrency int, dsn string) error {
	db, cleanup, err := testDatabase(dsn)
	if err != nil {
		return err
	}
	defer cleanup()

	ch := make(chan error, concurrency)
	for i := 0; i < concurrency; i++ {
		go func() { ch <- migrateDB(db) }()
	}

	errs := make([]string, 0, concurrency)
	for i := 0; i < cap(errs); i++ {
		if err := <-ch; err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

func testDatabase(dsn string) (*sql.DB, func(), error) {
	config, err := url.Parse(dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse dsn %q: %s", dsn, err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	dbname := "sourcegraph-test-" + strconv.FormatUint(rng.Uint64(), 10)

	db, err := newDB(config.String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %s", err)
	}

	_, err = db.Exec(`CREATE DATABASE ` + pq.QuoteIdentifier(dbname))
	if err != nil {
		return nil, nil, err
	}

	config.Path = "/" + dbname
	testDB, err := newDB(config.String())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %s", err)
	}

	return testDB, func() {
		defer db.Close()

		if err := testDB.Close(); err != nil {
			log.Printf("failed to close test database: %s", err)
		}
		_, err = db.Exec(killClientConnsQuery, dbname)
		if err != nil {
			log.Printf("failed to close test database: %s", err)
		}
		_, err = db.Exec(`DROP DATABASE ` + pq.QuoteIdentifier(dbname))
		if err != nil {
			log.Printf("failed to close test database: %s", err)
		}
	}, nil
}

const killClientConnsQuery = `
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity WHERE datname = $1`

// newDB returns a new *sql.DB from the given dsn (data source name).
func newDB(dsn string) (*sql.DB, error) {
	// We want to configure the database client explicitly through the DSN.
	// lib/pq uses and gives precedence to these environment variables so we unset them.
	for _, v := range []string{
		"PGHOST", "PGHOSTADDR", "PGPORT",
		"PGDATABASE", "PGUSER", "PGPASSWORD",
		"PGSERVICE", "PGSERVICEFILE", "PGREALM",
		"PGOPTIONS", "PGAPPNAME", "PGSSLMODE",
		"PGSSLCERT", "PGSSLKEY", "PGSSLROOTCERT",
		"PGREQUIRESSL", "PGSSLCRL", "PGREQUIREPEER",
		"PGKRBSRVNAME", "PGGSSLIB", "PGCONNECT_TIMEOUT",
		"PGCLIENTENCODING", "PGDATESTYLE", "PGTZ",
		"PGGEQO", "PGSYSCONFDIR", "PGLOCALEDIR",
	} {
		os.Unsetenv(v)
	}

	cfg, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse dsn")
	}

	qry := cfg.Query()

	// Force PostgreSQL session timezone to UTC.
	qry.Set("timezone", "UTC")

	// Set max open and idle connections
	maxOpen, _ := strconv.Atoi(qry.Get("max_conns"))
	if maxOpen == 0 {
		maxOpen = 30
	}
	qry.Del("max_conns")

	cfg.RawQuery = qry.Encode()
	db, err := sql.Open("postgres", cfg.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxOpen)

	return db, nil
}

// migrateDB runs all migrations from github.com/sourcegraph/sourcegraph/migrations
// against the given sql.DB
func migrateDB(db *sql.DB) error {
	var cfg postgres.Config
	driver, err := postgres.WithInstance(db, &cfg)
	if err != nil {
		return err
	}

	s := bindata.Resource(migrations.AssetNames(), migrations.Asset)
	d, err := bindata.WithInstance(s)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("go-bindata", d, "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err == nil || err == migrate.ErrNoChange {
		return nil
	}

	return err
}

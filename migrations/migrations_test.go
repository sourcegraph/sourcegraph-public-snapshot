package migrations_test

import (
	"database/sql"
	"net/http"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func TestMigrations(t *testing.T) {
	if os.Getenv("SKIP_MIGRATION_TEST") != "" {
		t.Skip()
	}

	db := dbtest.NewDB(t)

	for _, tt := range []struct {
		name   string
		schema *schemas.Schema
	}{
		{"Frontend", schemas.Frontend},
		{"CodeIntel", schemas.CodeIntel},
	} {

		t.Logf("Running migrations in %s", tt.name)
		testMigrations(t, db, tt.schema)
	}
}

// testMigrations runs all migrations up, then the migrations for the given database
// all the way back down, then back up to check for syntax errors and reversibility.
func testMigrations(t *testing.T, db *sql.DB, schema *schemas.Schema) {
	m := makeMigration(t, db, schema)

	// All the way up
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("unexpected error migrating database: %s", err)
	}

	// All the way down
	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("unexpected error running down migrations: %s", err)
	}

	// All the way up again
	if _, err := db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		t.Fatalf("failed to recreate schema")
	}

	m = makeMigration(t, db, schema)
	if err := m.Up(); err != nil {
		t.Fatalf("unexpected error re-running up migrations: %s", err)
	}
}

func makeMigration(t *testing.T, db *sql.DB, schema *schemas.Schema) *migrate.Migrate {
	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: schema.MigrationsTableName,
	})
	if err != nil {
		t.Fatalf("unexpected error creating driver: %s", err)
	}

	d, err := httpfs.New(http.FS(schema.FS), ".")
	if err != nil {
		t.Fatalf("unexpected error creating migration source: %s", err)
	}

	m, err := migrate.NewWithInstance("httpfs", d, "postgres", driver)
	if err != nil {
		t.Fatalf("unexpected error creating migration: %s", err)
	}

	return m
}

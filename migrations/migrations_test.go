package migrations_test

import (
	"database/sql"
	"io/fs"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/migrations"
)

func TestIDConstraints(t *testing.T) {
	cases := []struct {
		Name string
		FS   fs.FS
	}{
		{Name: "frontend", FS: migrations.Frontend},
		{Name: "codeintel", FS: migrations.CodeIntel},
		{Name: "codeinsights", FS: migrations.CodeInsights},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ups, err := fs.Glob(c.FS, "*.up.sql")
			if err != nil {
				t.Fatal(err)
			}

			if len(ups) == 0 {
				t.Fatal("no up migrations found")
			}

			byID := map[int][]string{}
			for _, name := range ups {
				id, err := strconv.Atoi(name[:strings.IndexByte(name, '_')])
				if err != nil {
					t.Fatalf("failed to parse name %q: %v", name, err)
				}
				byID[id] = append(byID[id], name)
			}

			var ids []int
			for id, names := range byID {
				if len(names) > 1 {
					t.Errorf("multiple migrations with ID %d: %s", id, strings.Join(names, " "))
				}

				ids = append(ids, id)
			}
			sort.Ints(ids)

			for i, id := range ids {
				if i != 0 && ids[i-1]+1 != id {
					// Check if we are using sequential migrations.
					t.Errorf("gap in migrations between %s and %s", byID[ids[i-1]][0], byID[id][0])
				}
			}
		})
	}
}

func TestMigrations(t *testing.T) {
	if os.Getenv("SKIP_MIGRATION_TEST") != "" {
		t.Skip()
	}

	db := dbtesting.GetDB(t)

	for _, tt := range []struct {
		name     string
		database *dbconn.Database
	}{
		{"Frontend", dbconn.Frontend},
		{"CodeIntel", dbconn.CodeIntel},
	} {

		t.Logf("Running migrations in %s", tt.name)
		testMigrations(t, db, tt.database)
	}
}

// testMigrations runs all migrations up, then the migrations for the given database
// all the way back down, then back up to check for syntax errors and reversibility.
func testMigrations(t *testing.T, db *sql.DB, database *dbconn.Database) {
	m := makeMigration(t, db, database)

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

	m = makeMigration(t, db, database)
	if err := m.Up(); err != nil {
		t.Fatalf("unexpected error re-running up migrations: %s", err)
	}
}

func makeMigration(t *testing.T, db *sql.DB, database *dbconn.Database) *migrate.Migrate {
	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: database.MigrationsTable,
	})
	if err != nil {
		t.Fatalf("unexpected error creating driver: %s", err)
	}

	d, err := httpfs.New(http.FS(database.FS), ".")
	if err != nil {
		t.Fatalf("unexpected error creating migration source: %s", err)
	}

	m, err := migrate.NewWithInstance("httpfs", d, "postgres", driver)
	if err != nil {
		t.Fatalf("unexpected error creating migration: %s", err)
	}

	return m
}

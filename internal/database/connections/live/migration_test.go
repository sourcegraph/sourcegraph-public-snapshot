package connections

import (
	"context"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var testSchemas = []string{
	"frontend",
	"codeintel",
}

func TestMigrations(t *testing.T) {
	for _, name := range testSchemas {
		schema, ok := getSchema(name)
		if !ok {
			t.Fatalf("missing schema %s", name)
		}

		t.Run(name, func(t *testing.T) {
			testMigrations(t, name, schema)
		})
	}
}

func getSchema(name string) (*schemas.Schema, bool) {
	for _, schema := range schemas.Schemas {
		if schema.Name == name {
			return schema, true
		}
	}

	return nil, false
}

func testMigrations(t *testing.T, name string, schema *schemas.Schema) {
	t.Helper()

	ctx := context.Background()
	db := dbtest.NewRawDB(t)
	storeFactory := newStoreFactory(&observation.TestContext)
	migrationRunner := runnerFromDB(storeFactory, db, schema)

	t.Run("up", func(t *testing.T) {
		options := runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName: name,
					Type:       runner.MigrationOperationTypeUpgrade,
				},
			},
		}
		if err := migrationRunner.Run(ctx, options); err != nil {
			t.Fatalf("failed to perform initial upgrade: %s", err)
		}
	})
	t.Run("down", func(t *testing.T) {
		options := runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName: name,
					Type:       runner.MigrationOperationTypeTargetedDown,
					// Run down to the root "squashed commits" migration. We don't go
					// any farther than that because it would require a fresh database,
					// and that doesn't adequately test upgrade idempotency.
					TargetVersions: []int{schema.Definitions.Root().ID},
				},
			},
		}
		if err := migrationRunner.Run(ctx, options); err != nil {
			t.Fatalf("failed to perform downgrade: %s", err)
		}
	})

	all := schema.Definitions.All()

	t.Run("idempotent up", func(t *testing.T) {
		for i, definition := range all {
			options := runner.Options{
				Operations: []runner.MigrationOperation{
					{
						SchemaName:     name,
						Type:           runner.MigrationOperationTypeTargetedUp,
						TargetVersions: []int{definition.ID},
					},
				},
			}
			if err := migrationRunner.Run(ctx, options); err != nil {
				t.Fatalf("failed to perform upgrade to version %d: %s", definition.ID, err)
			}

			if i == 0 {
				// Skip root migration
				continue
			}

			if _, err := db.Exec(definition.UpQuery.Query(sqlf.PostgresBindVar)); err != nil {
				t.Errorf("migration %d is not idempotent%s: %s", definition.ID, formatHint(err), err)
			}
		}
	})

	t.Run("idempotent down", func(t *testing.T) {
		// Skip root migration
		for i := len(all) - 1; i > 0; i-- {
			definition := all[i]

			options := runner.Options{
				Operations: []runner.MigrationOperation{
					{
						SchemaName:     name,
						Type:           runner.MigrationOperationTypeTargetedDown,
						TargetVersions: definition.Parents,
					},
				},
			}
			if err := migrationRunner.Run(ctx, options); err != nil {
				t.Fatalf("failed to perform downgrade to versions %v: %s", definition.Parents, err)
			}
			if _, err := db.Exec(definition.DownQuery.Query(sqlf.PostgresBindVar)); err != nil {
				t.Errorf("migration %d is not idempotent%s: %s", definition.ID, formatHint(err), err)
			}
		}
	})
}

func formatHint(err error) string {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return ""
	}

	switch pgErr.Code {
	case
		// `undefined_*` error codes
		"42703", "42883", "42P01",
		"42P02", "42704":

		return ` (hint: use "IF EXISTS" in deletion statements)`

	case
		// `duplicate_*` error codes
		"42701", "42P03", "42P04",
		"42723", "42P05", "42P06",
		"42P07", "42712", "42710":

		return ` (hint: use "IF NOT EXISTS"/"CREATE OR REPLACE" in creation statements (e.g., tables, indexes, views, functions), or drop existing objects prior to creating them (e.g., user-defined types, constraints, triggers))`

	}

	return ""
}

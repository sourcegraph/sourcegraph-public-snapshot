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
	"codeinsights",
}

func TestMigrations(t *testing.T) {
	for _, name := range testSchemas {
		schema, ok := getSchema(name)
		if !ok {
			t.Fatalf("missing schema %s", name)
		}

		t.Run(name, func(t *testing.T) {
			testMigrations(t, name, schema)
			testMigrationIdempotency(t, name, schema)
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
	all := schema.Definitions.All()

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
		// Run down to the root "squashed commits" migration. For this, we need to select
		// the _last_ nonIdempotent migration in the prefix of the migration definitions.
		// This will ensure that we downgrade _to_ the squashed migrations, but do not try
		// to re-run them on the way back up.

		var target int
		for offset := 0; offset < len(all); offset++ {
			// This is the last definition _or_ the next migration is idempotent
			if offset+1 >= len(all) || !all[offset+1].NonIdempotent {
				target = all[offset].ID
				break
			}
		}
		if target == 0 {
			t.Fatalf("failed to locate last squashed migration definition")
		}

		options := runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName:     name,
					Type:           runner.MigrationOperationTypeTargetedDown,
					TargetVersions: []int{target},
				},
			},
		}
		if err := migrationRunner.Run(ctx, options); err != nil {
			t.Fatalf("failed to perform downgrade: %s", err)
		}
	})
	t.Run("up again", func(t *testing.T) {
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
}

func testMigrationIdempotency(t *testing.T, name string, schema *schemas.Schema) {
	t.Helper()

	ctx := context.Background()
	db := dbtest.NewRawDB(t)
	storeFactory := newStoreFactory(&observation.TestContext)
	migrationRunner := runnerFromDB(storeFactory, db, schema)
	all := schema.Definitions.All()

	t.Run("idempotent up", func(t *testing.T) {
		for _, definition := range all {
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

			if definition.NonIdempotent {
				// Some migrations are explicitly non-idempotent (squashed migrations)
				// Skip these here
				continue
			}

			if _, err := db.Exec(definition.UpQuery.Query(sqlf.PostgresBindVar)); err != nil {
				t.Errorf("migration %d is not idempotent%s: %s", definition.ID, formatHint(err), err)
			}
		}
	})

	t.Run("idempotent down", func(t *testing.T) {
		for i := len(all) - 1; i >= 0; i-- {
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

			if definition.NonIdempotent {
				// Some migrations are explicitly non-idempotent (squashed migrations)
				// Skip these here
				continue
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

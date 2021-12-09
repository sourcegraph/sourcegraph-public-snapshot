package connections

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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

	upOptions := runner.Options{
		Up:          true,
		SchemaNames: []string{name},
	}
	downOptions := runner.Options{
		Up:          false,
		SchemaNames: []string{name},
	}

	if err := migrationRunner.Run(ctx, upOptions); err != nil {
		t.Fatalf("failed to perform initial migration: %s", err)
	}
	if err := migrationRunner.Run(ctx, downOptions); err != nil {
		t.Fatalf("failed to perform down migration: %s", err)
	}

	// TEMPORARY
	if _, err := db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;"); err != nil {
		t.Fatalf("failed to drop schema: %s\n", err)
	}

	if err := migrationRunner.Run(ctx, upOptions); err != nil {
		t.Fatalf("failed to perform subsequent migration: %s", err)
	}
}

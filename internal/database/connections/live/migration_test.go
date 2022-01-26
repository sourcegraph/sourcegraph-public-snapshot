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
		Up: false,
		// Run down to the root "squashed commits" migration. We don't go
		// any farther than that because it would require a fresh database,
		// and that doesn't adequately test upgrade idempotency.
		TargetMigration: schema.Definitions.First(),
		SchemaNames:     []string{name},
	}

	//
	// Run migrations up -> down -> up

	if err := migrationRunner.Run(ctx, upOptions); err != nil {
		t.Fatalf("failed to perform initial upgrade: %s", err)
	}
	if err := migrationRunner.Run(ctx, downOptions); err != nil {
		t.Fatalf("failed to perform downgrade: %s", err)
	}
	if err := migrationRunner.Run(ctx, upOptions); err != nil {
		t.Fatalf("failed to perform subsequent upgrade: %s", err)
	}
}

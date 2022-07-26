package stitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Notable versions:
//
// v3.29.0 -> oldest supported
// v3.37.0 -> directories introduced
// v3.38.0 -> privileged migrations introduced

// Base migrations:
//
// frontend:
//   3.29.0 -> 1528395787
//   3.33.0 -> 1528395834
//   3.38.0 -> 1528395943
//   3.41.0 -> 1644868458
//
// codeintel:
//   3.29.0 -> 1000000005
//   3.33.0 -> 1000000015
//   3.38.0 -> 1000000029
//
// codeinsights:
//   3.41.0 -> 1000000026
//   3.38.0 -> 1000000020
//   3.29.0 -> 1000000000

func TestStitchFrontendDefinitions(t *testing.T) {
	if testing.Short() {
		return
	}
	t.Parallel()

	// Note: negative values imply a quashed migration split into a privileged and
	// unprivileged version. See `readMigrations` in this package for more details.
	testStitchGraphShape(t, "frontend", 41, 42, +1644868458, []int{1654770608, 1658174103, 1658225452, 1657663493})
	testStitchGraphShape(t, "frontend", 40, 42, +1528395943, []int{1654770608, 1658174103, 1658225452, 1657663493})
	testStitchGraphShape(t, "frontend", 38, 42, +1528395943, []int{1654770608, 1658174103, 1658225452, 1657663493})
	testStitchGraphShape(t, "frontend", 37, 42, -1528395834, []int{1654770608, 1658174103, 1658225452, 1657663493})
	testStitchGraphShape(t, "frontend", 35, 42, -1528395834, []int{1654770608, 1658174103, 1658225452, 1657663493})
	testStitchGraphShape(t, "frontend", 29, 42, -1528395787, []int{1654770608, 1658174103, 1658225452, 1657663493})

	// Test a different leaf
	testStitchGraphShape(t, "frontend", 35, 40, -1528395834, []int{1652228814, 1652707934})
}

func TestStitchCodeintelDefinitions(t *testing.T) {
	if testing.Short() {
		return
	}
	t.Parallel()

	// Note: negative values imply a quashed migration split into a privileged and
	// unprivileged version. See `readMigrations` in this package for more details.
	testStitchGraphShape(t, "codeintel", 41, 42, +1000000029, []int{1000000034})
	testStitchGraphShape(t, "codeintel", 40, 42, +1000000029, []int{1000000034})
	testStitchGraphShape(t, "codeintel", 38, 42, +1000000029, []int{1000000034})
	testStitchGraphShape(t, "codeintel", 37, 42, -1000000015, []int{1000000034})
	testStitchGraphShape(t, "codeintel", 35, 42, -1000000015, []int{1000000034})
	testStitchGraphShape(t, "codeintel", 29, 42, -1000000005, []int{1000000034})

	// Test a different leaf
	testStitchGraphShape(t, "codeintel", 32, 37, -1000000005, []int{1000000030})
}

func TestStitchCodeinsightsDefinitions(t *testing.T) {
	if testing.Short() {
		return
	}
	t.Parallel()

	// Note: negative values imply a quashed migration split into a privileged and
	// unprivileged version. See `readMigrations` in this package for more details.
	testStitchGraphShape(t, "codeinsights", 41, 42, +1000000026, []int{1656517037, 1656608833})
	testStitchGraphShape(t, "codeinsights", 40, 42, +1000000020, []int{1656517037, 1656608833})
	testStitchGraphShape(t, "codeinsights", 38, 42, +1000000020, []int{1656517037, 1656608833})
	testStitchGraphShape(t, "codeinsights", 37, 42, -1000000000, []int{1656517037, 1656608833})
	testStitchGraphShape(t, "codeinsights", 35, 42, -1000000000, []int{1656517037, 1656608833})
	testStitchGraphShape(t, "codeinsights", 29, 42, -1000000000, []int{1656517037, 1656608833})

	// Test a different leaf
	testStitchGraphShape(t, "codeinsights", 38, 39, +1000000020, []int{1649801281})
}

func TestStitchAndApplyFrontendDefinitions(t *testing.T) {
	if testing.Short() {
		return
	}
	t.Parallel()

	testStitchApplication(t, "frontend", 41, 42)
	testStitchApplication(t, "frontend", 40, 42)
	testStitchApplication(t, "frontend", 38, 42)
	testStitchApplication(t, "frontend", 37, 42)
	testStitchApplication(t, "frontend", 35, 42)
	testStitchApplication(t, "frontend", 29, 42)
}

func TestStitchAndApplyCodeintelDefinitions(t *testing.T) {
	if testing.Short() {
		return
	}
	t.Parallel()

	testStitchApplication(t, "codeintel", 41, 42)
	testStitchApplication(t, "codeintel", 40, 42)
	testStitchApplication(t, "codeintel", 38, 42)
	testStitchApplication(t, "codeintel", 37, 42)
	testStitchApplication(t, "codeintel", 35, 42)
	testStitchApplication(t, "codeintel", 29, 42)
}

func TestStitchAndApplyCodeinsightsDefinitions(t *testing.T) {
	if testing.Short() {
		return
	}
	t.Parallel()

	testStitchApplication(t, "codeinsights", 41, 42)
	testStitchApplication(t, "codeinsights", 40, 42)
	testStitchApplication(t, "codeinsights", 38, 42)
	testStitchApplication(t, "codeinsights", 37, 42)
	testStitchApplication(t, "codeinsights", 35, 42)
	testStitchApplication(t, "codeinsights", 29, 42)
}

// testStitchGraphShape stitches the migrations between the given minor version ranges, then
// asserts that the resulting graph has the expected root and leaf values.
func testStitchGraphShape(t *testing.T, schemaName string, from, to, expectedRoot int, expectedLeaves []int) {
	t.Run(fmt.Sprintf("stitch 3.%d -> 3.%d", from, to), func(t *testing.T) {
		definitions, err := StitchDefinitions(schemaName, repositoryRoot(t), makeRange(from, to))
		if err != nil {
			t.Fatalf("failed to stitch definitions: %s", err)
		}

		var leafIDs []int
		for _, migration := range definitions.Leaves() {
			leafIDs = append(leafIDs, migration.ID)
		}

		if rootID := definitions.Root().ID; rootID != expectedRoot {
			t.Fatalf("unexpected root migration. want=%d have=%d", expectedRoot, rootID)
		}
		if len(leafIDs) != len(expectedLeaves) || cmp.Diff(expectedLeaves, leafIDs) != "" {
			t.Fatalf("unexpected leaf migrations. want=%v have=%v", expectedLeaves, leafIDs)
		}
	})
}

// testStitchApplication stitches the migrations bewteen the given minor version ranges, then
// runs the resulting migrations over a test database instance. The resulting database is then
// compared against the target version's description (in the git-tree).
func testStitchApplication(t *testing.T, schemaName string, from, to int) {
	t.Run(fmt.Sprintf("upgrade 3.%d -> 3.%d", from, to), func(t *testing.T) {
		definitions, err := StitchDefinitions(schemaName, repositoryRoot(t), makeRange(from, to))
		if err != nil {
			t.Fatalf("failed to stitch definitions: %s", err)
		}

		ctx := context.Background()
		logger := logtest.Scoped(t)
		db := dbtest.NewRawDB(logger, t)
		migrationsTableName := "testing"

		store := connections.NewStoreShim(store.NewWithDB(db, migrationsTableName, store.NewOperations(&observation.TestContext)))
		if err := store.EnsureSchemaTable(ctx); err != nil {
			t.Fatalf("failed to prepare store: %s", err)
		}

		migrationRunner := runner.NewRunnerWithSchemas(logger, map[string]runner.StoreFactory{
			schemaName: func(ctx context.Context) (runner.Store, error) { return store, nil },
		}, []*schemas.Schema{
			{
				Name:                schemaName,
				MigrationsTableName: migrationsTableName,
				Definitions:         definitions,
			},
		})

		if err := migrationRunner.Run(ctx, runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName: schemaName,
					Type:       runner.MigrationOperationTypeUpgrade,
				},
			},
		}); err != nil {
			t.Fatalf("failed to upgrade: %s", err)
		}

		if err := migrationRunner.Validate(ctx, schemaName); err != nil {
			t.Fatalf("failed to validate: %s", err)
		}

		fileSuffix := ""
		if schemaName != "frontend" {
			fileSuffix = "." + schemaName
		}
		expectedSchema := expectedSchema(
			t,
			fmt.Sprintf("v3.%d.0", to),
			fmt.Sprintf("internal/database/schema%s.json", fileSuffix),
		)

		schemas, err := store.Describe(ctx)
		if err != nil {
			t.Fatalf("failed to describe database: %s", err)
		}
		schema := canonicalize(schemas["public"])

		if diff := cmp.Diff(expectedSchema, schema); diff != "" {
			t.Fatalf("unexpected schema (-want +got):\n%s", diff)
		}
	})
}

func repositoryRoot(t *testing.T) string {
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %s", err)
	}

	return strings.TrimSuffix(root, "/internal/database/migration/stitch")
}

func makeRange(from, to int) []string {
	revs := make([]string, 0, to-from)
	for v := from; v <= to; v++ {
		revs = append(revs, fmt.Sprintf("v3.%d.0", v))
	}

	return revs
}

func expectedSchema(t *testing.T, rev, filename string) (schemaDescription schemas.SchemaDescription) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", rev, filename))
	cmd.Dir = repositoryRoot(t)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}

	if err := json.NewDecoder(bytes.NewReader(out)).Decode(&schemaDescription); err != nil {
		t.Fatalf("failed to decode json: %s", err)
	}

	return canonicalize(schemaDescription)
}

// copied from the drift command
func canonicalize(schemaDescription schemas.SchemaDescription) schemas.SchemaDescription {
	schemas.Canonicalize(schemaDescription)

	for i, table := range schemaDescription.Tables {
		for j := range table.Columns {
			schemaDescription.Tables[i].Columns[j].Index = -1
		}
	}

	return schemaDescription
}

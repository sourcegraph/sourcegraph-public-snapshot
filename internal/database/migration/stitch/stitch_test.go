package stitch

import (
	"context"
	"fmt"
	"os"
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
	t.Parallel()

	testStitchGraphShape(t, "frontend", 41, 42, +1644868458, []int{1657635365})
	testStitchGraphShape(t, "frontend", 40, 42, +1528395943, []int{1657635365})
	testStitchGraphShape(t, "frontend", 38, 42, +1528395943, []int{1657635365})
	testStitchGraphShape(t, "frontend", 37, 42, -1528395834, []int{1657635365})
	testStitchGraphShape(t, "frontend", 35, 42, -1528395834, []int{1657635365})
	testStitchGraphShape(t, "frontend", 29, 42, -1528395787, []int{1657635365})
}

func TestStitchCodeintelDefinitions(t *testing.T) {
	t.Parallel()

	testStitchGraphShape(t, "codeintel", 41, 42, +1000000029, []int{1000000034})
	testStitchGraphShape(t, "codeintel", 40, 42, +1000000029, []int{1000000034})
	testStitchGraphShape(t, "codeintel", 38, 42, +1000000029, []int{1000000034})
	testStitchGraphShape(t, "codeintel", 37, 42, -1000000015, []int{1000000034})
	testStitchGraphShape(t, "codeintel", 35, 42, -1000000015, []int{1000000034})
	testStitchGraphShape(t, "codeintel", 29, 42, -1000000005, []int{1000000034})
}

func TestStitchCodeinsightsDefinitions(t *testing.T) {
	t.Parallel()

	testStitchGraphShape(t, "codeinsights", 41, 42, +1000000026, []int{1656517037, 1656608833})
	testStitchGraphShape(t, "codeinsights", 40, 42, +1000000020, []int{1656517037, 1656608833})
	testStitchGraphShape(t, "codeinsights", 38, 42, +1000000020, []int{1656517037, 1656608833})
	testStitchGraphShape(t, "codeinsights", 37, 42, -1000000000, []int{1656517037, 1656608833})
	testStitchGraphShape(t, "codeinsights", 35, 42, -1000000000, []int{1656517037, 1656608833})
	testStitchGraphShape(t, "codeinsights", 29, 42, -1000000000, []int{1656517037, 1656608833})
}

func TestStitchAndApplyFrontendDefinitions(t *testing.T) {
	if testing.Short() {
		return
	}
	t.Parallel()

	testStitchApplication(t, "frontend", 41, 42, nil)
	testStitchApplication(t, "frontend", 40, 42, nil)
	testStitchApplication(t, "frontend", 38, 42, nil)
	testStitchApplication(t, "frontend", 37, 42, nil)
	testStitchApplication(t, "frontend", 35, 42, nil)
	testStitchApplication(t, "frontend", 29, 42, nil)
}

func TestStitchAndApplyCodeintelDefinitions(t *testing.T) {
	if testing.Short() {
		return
	}
	t.Parallel()

	testStitchApplication(t, "codeintel", 41, 42, nil)
	testStitchApplication(t, "codeintel", 40, 42, nil)
	testStitchApplication(t, "codeintel", 38, 42, nil)
	testStitchApplication(t, "codeintel", 37, 42, nil)
	testStitchApplication(t, "codeintel", 35, 42, nil)
	testStitchApplication(t, "codeintel", 29, 42, nil)
}

func TestStitchAndApplyCodeinsightsDefinitions(t *testing.T) {
	if testing.Short() {
		return
	}
	t.Parallel()

	testStitchApplication(t, "codeinsights", 41, 42, nil)
	testStitchApplication(t, "codeinsights", 40, 42, nil)
	testStitchApplication(t, "codeinsights", 38, 42, nil)
	testStitchApplication(t, "codeinsights", 37, 42, nil)
	testStitchApplication(t, "codeinsights", 35, 42, nil)
	testStitchApplication(t, "codeinsights", 29, 42, nil)
}

func testStitchGraphShape(t *testing.T, schemaName string, from, to, expectedRoot int, expectedLeaves []int) {
	t.Run(fmt.Sprintf("stitch 3.%d -> 3.%d", from, to), func(t *testing.T) {
		t.Parallel()

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

func testStitchApplication(t *testing.T, schemaName string, from, to int, payload any) {
	t.Run(fmt.Sprintf("upgrade 3.%d -> 3.%d", from, to), func(t *testing.T) {
		t.Parallel()

		definitions, err := StitchDefinitions(schemaName, repositoryRoot(t), makeRange(from, to))
		if err != nil {
			t.Fatalf("failed to stitch definitions: %s", err)
		}

		ctx := context.Background()
		logger := logtest.Scoped(t)
		db := dbtest.NewRawDB(logger, t)

		migrationRunner := runner.NewRunnerWithSchemas(logger, map[string]runner.StoreFactory{
			schemaName: func(ctx context.Context) (runner.Store, error) {
				shim := connections.NewStoreShim(store.NewWithDB(db, "testing", store.NewOperations(&observation.TestContext)))
				if err := shim.EnsureSchemaTable(ctx); err != nil {
					return nil, err
				}

				return shim, nil
			},
		}, []*schemas.Schema{
			{
				Name:                schemaName,
				MigrationsTableName: "testing",
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

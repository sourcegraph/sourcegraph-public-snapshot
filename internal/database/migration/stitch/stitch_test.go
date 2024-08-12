package stitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestMain(m *testing.M) {
	logtest.Init(m)
	os.Exit(m.Run())
}

// Notable versions:
//
// v3.29.0 -> oldest supported
// v3.37.0 -> directories introduced
// v3.38.0 -> privileged migrations introduced

func TestStitchFrontendDefinitions(t *testing.T) {
	if testing.Short() || os.Getenv("BAZEL_TEST") == "1" {
		t.Skip()
		return
	}
	t.Parallel()

	boundsByRev := map[string]shared.MigrationBounds{
		"v3.29.0": {RootID: -1528395787, LeafIDs: []int{1528395834}},
		"v3.30.0": {RootID: -1528395787, LeafIDs: []int{1528395853}},
		"v3.31.0": {RootID: -1528395787, LeafIDs: []int{1528395871}},
		"v3.32.0": {RootID: -1528395787, LeafIDs: []int{1528395891}},
		"v3.33.0": {RootID: -1528395834, LeafIDs: []int{1528395918}},
		"v3.34.0": {RootID: -1528395834, LeafIDs: []int{1528395944}},
		"v3.35.0": {RootID: -1528395834, LeafIDs: []int{1528395964}},
		"v3.36.0": {RootID: -1528395834, LeafIDs: []int{1528395968}},
		"v3.37.0": {RootID: -1528395834, LeafIDs: []int{1645106226}},
		"v3.38.0": {RootID: +1528395943, LeafIDs: []int{1646652951, 1647282553}},
		"v3.39.0": {RootID: +1528395943, LeafIDs: []int{1649441222, 1649759318, 1649432863}},
		"v3.40.0": {RootID: +1528395943, LeafIDs: []int{1652228814, 1652707934}},
		"v3.41.0": {RootID: +1644868458, LeafIDs: []int{1655481894}},
		"v3.42.0": {RootID: +1646027072, LeafIDs: []int{1654770608, 1658174103, 1658225452, 1657663493}},
	}

	makeTest := func(from int, to int, expectedRoot int) {
		filteredLeavesByRev := make(map[string]shared.MigrationBounds, len(boundsByRev))
		for i := from; i <= to; i++ {
			v := fmt.Sprintf("v3.%d.0", i)
			filteredLeavesByRev[v] = boundsByRev[v]
		}

		v := fmt.Sprintf("v3.%d.0", to)
		testStitchGraphShape(t, "frontend", from, to, expectedRoot, boundsByRev[v].LeafIDs, filteredLeavesByRev)
	}

	// Note: negative values imply a quashed migration split into a privileged and
	// unprivileged version. See `readMigrations` in this package for more details.
	makeTest(41, 42, +1644868458)
	makeTest(40, 42, +1528395943)
	makeTest(38, 42, +1528395943)
	makeTest(37, 42, -1528395834)
	makeTest(35, 42, -1528395834)
	makeTest(29, 42, -1528395787)
	makeTest(35, 40, -1528395834) // Test a different leaf
}

func TestStitchCodeintelDefinitions(t *testing.T) {
	if testing.Short() || os.Getenv("BAZEL_TEST") == "1" {
		t.Skip()
		return
	}
	t.Parallel()

	boundsByRev := map[string]shared.MigrationBounds{
		"v3.29.0": {RootID: -1000000005, LeafIDs: []int{1000000015}},
		"v3.30.0": {RootID: -1000000005, LeafIDs: []int{1000000018}},
		"v3.31.0": {RootID: -1000000005, LeafIDs: []int{1000000019}},
		"v3.32.0": {RootID: -1000000005, LeafIDs: []int{1000000019}},
		"v3.33.0": {RootID: -1000000015, LeafIDs: []int{1000000025}},
		"v3.34.0": {RootID: -1000000015, LeafIDs: []int{1000000030}},
		"v3.35.0": {RootID: -1000000015, LeafIDs: []int{1000000030}},
		"v3.36.0": {RootID: -1000000015, LeafIDs: []int{1000000030}},
		"v3.37.0": {RootID: -1000000015, LeafIDs: []int{1000000030}},
		"v3.38.0": {RootID: +1000000029, LeafIDs: []int{1000000034}},
		"v3.39.0": {RootID: +1000000029, LeafIDs: []int{1000000034}},
		"v3.40.0": {RootID: +1000000029, LeafIDs: []int{1000000034}},
		"v3.41.0": {RootID: +1000000029, LeafIDs: []int{1000000034}},
		"v3.42.0": {RootID: +1000000033, LeafIDs: []int{1000000034}},
	}

	makeTest := func(from int, to int, expectedRoot int) {
		filteredLeavesByRev := make(map[string]shared.MigrationBounds, len(boundsByRev))
		for i := from; i <= to; i++ {
			v := fmt.Sprintf("v3.%d.0", i)
			filteredLeavesByRev[v] = boundsByRev[v]
		}

		v := fmt.Sprintf("v3.%d.0", to)
		testStitchGraphShape(t, "codeintel", from, to, expectedRoot, boundsByRev[v].LeafIDs, filteredLeavesByRev)
	}

	// Note: negative values imply a quashed migration split into a privileged and
	// unprivileged version. See `readMigrations` in this package for more details.
	makeTest(41, 42, +1000000029)
	makeTest(40, 42, +1000000029)
	makeTest(38, 42, +1000000029)
	makeTest(37, 42, -1000000015)
	makeTest(35, 42, -1000000015)
	makeTest(29, 42, -1000000005)
	makeTest(32, 37, -1000000005) // Test a different leaf
}

func TestStitchCodeinsightsDefinitions(t *testing.T) {
	if testing.Short() || os.Getenv("BAZEL_TEST") == "1" {
		t.Skip()
		return
	}
	t.Parallel()

	boundsByRev := map[string]shared.MigrationBounds{
		"v3.29.0": {RootID: -1000000000, LeafIDs: []int{1000000006}},
		"v3.30.0": {RootID: -1000000000, LeafIDs: []int{1000000008}},
		"v3.31.0": {RootID: -1000000000, LeafIDs: []int{1000000011}},
		"v3.32.0": {RootID: -1000000000, LeafIDs: []int{1000000013}},
		"v3.33.0": {RootID: -1000000000, LeafIDs: []int{1000000016}},
		"v3.34.0": {RootID: -1000000000, LeafIDs: []int{1000000021}},
		"v3.35.0": {RootID: -1000000000, LeafIDs: []int{1000000024}},
		"v3.36.0": {RootID: -1000000000, LeafIDs: []int{1000000025}},
		"v3.37.0": {RootID: -1000000000, LeafIDs: []int{1000000027}},
		"v3.38.0": {RootID: +1000000020, LeafIDs: []int{1646761143}},
		"v3.39.0": {RootID: +1000000020, LeafIDs: []int{1649801281}},
		"v3.40.0": {RootID: +1000000020, LeafIDs: []int{1652289966}},
		"v3.41.0": {RootID: +1000000026, LeafIDs: []int{1651021000, 1652289966}},
		"v3.42.0": {RootID: +1000000027, LeafIDs: []int{1656517037, 1656608833}},
	}

	makeTest := func(from int, to int, expectedRoot int) {
		filteredBoundsByRev := make(map[string]shared.MigrationBounds, len(boundsByRev))
		for i := from; i <= to; i++ {
			v := fmt.Sprintf("v3.%d.0", i)
			filteredBoundsByRev[v] = boundsByRev[v]
		}

		v := fmt.Sprintf("v3.%d.0", to)
		testStitchGraphShape(t, "codeinsights", from, to, expectedRoot, boundsByRev[v].LeafIDs, filteredBoundsByRev)
	}

	// Note: negative values imply a quashed migration split into a privileged and
	// unprivileged version. See `readMigrations` in this package for more details.
	makeTest(41, 42, +1000000026)
	makeTest(40, 42, +1000000020)
	makeTest(38, 42, +1000000020)
	makeTest(37, 42, -1000000000)
	makeTest(35, 42, -1000000000)
	makeTest(29, 42, -1000000000)
	makeTest(38, 39, +1000000020) // Test a different leaf
}

func TestStitchAndApplyFrontendDefinitions(t *testing.T) {
	if testing.Short() || os.Getenv("BAZEL_TEST") == "1" {
		t.Skip()
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
	if testing.Short() || os.Getenv("BAZEL_TEST") == "1" {
		t.Skip()
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
	if testing.Short() || os.Getenv("BAZEL_TEST") == "1" {
		t.Skip()
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
// asserts that the resulting graph has the expected root, leaf, and version boundary values.
func testStitchGraphShape(t *testing.T, schemaName string, from, to, expectedRoot int, expectedLeaves []int, expectedBoundsByRev map[string]shared.MigrationBounds) {
	t.Run(fmt.Sprintf("stitch 3.%d -> 3.%d", from, to), func(t *testing.T) {
		stitched, err := StitchDefinitions(testMigrationsReader, schemaName, makeRange(from, to))
		if err != nil {
			t.Fatalf("failed to stitch definitions: %s", err)
		}

		var leafIDs []int
		for _, migration := range stitched.Definitions.Leaves() {
			leafIDs = append(leafIDs, migration.ID)
		}

		if rootID := stitched.Definitions.Root().ID; rootID != expectedRoot {
			t.Fatalf("unexpected root migration. want=%d have=%d", expectedRoot, rootID)
		}
		if len(leafIDs) != len(expectedLeaves) || cmp.Diff(expectedLeaves, leafIDs) != "" {
			t.Fatalf("unexpected leaf migrations. want=%v have=%v", expectedLeaves, leafIDs)
		}
		if diff := cmp.Diff(expectedBoundsByRev, stitched.BoundsByRev); diff != "" {
			t.Fatalf("unexpected migration bounds (-want +got):\n%s", diff)
		}
	})
}

// testStitchApplication stitches the migrations bewteen the given minor version ranges, then
// runs the resulting migrations over a test database instance. The resulting database is then
// compared against the target version's description (in the git-tree).
func testStitchApplication(t *testing.T, schemaName string, from, to int) {
	t.Run(fmt.Sprintf("upgrade 3.%d -> 3.%d", from, to), func(t *testing.T) {
		stitched, err := StitchDefinitions(testMigrationsReader, schemaName, makeRange(from, to))
		if err != nil {
			t.Fatalf("failed to stitch definitions: %s", err)
		}

		ctx := context.Background()
		logger := logtest.Scoped(t)
		db := dbtest.NewRawDB(logger, t)
		migrationsTableName := "testing"

		storeShim := connections.NewStoreShim(store.NewWithDB(observation.TestContextTB(t), db, migrationsTableName))
		if err := storeShim.EnsureSchemaTable(ctx); err != nil {
			t.Fatalf("failed to prepare store: %s", err)
		}

		migrationRunner := runner.NewRunnerWithSchemas(logger, map[string]runner.StoreFactory{
			schemaName: func(ctx context.Context) (runner.Store, error) { return storeShim, nil },
		}, []*schemas.Schema{
			{
				Name:                schemaName,
				MigrationsTableName: migrationsTableName,
				Definitions:         stitched.Definitions,
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

		schemaDescriptions, err := storeShim.Describe(ctx)
		if err != nil {
			t.Fatalf("failed to describe database: %s", err)
		}
		schema := canonicalize(schemaDescriptions["public"])

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

	filtered := schemaDescription.Tables[:0]
	for i, table := range schemaDescription.Tables {
		if table.Name == "migration_logs" {
			continue
		}

		for j := range table.Columns {
			schemaDescription.Tables[i].Columns[j].Index = -1
		}

		filtered = append(filtered, schemaDescription.Tables[i])
	}
	schemaDescription.Tables = filtered

	return schemaDescription
}

var testMigrationsReader MigrationsReader = &cachedMigrationsReader{
	inner: NewLazyMigrationsReader(),
	m:     make(map[string]func() (map[string]string, error)),
}

type cachedMigrationsReader struct {
	inner MigrationsReader

	mu sync.Mutex
	m  map[string]func() (map[string]string, error)
}

func (c *cachedMigrationsReader) Get(version string) (map[string]string, error) {
	c.mu.Lock()
	get, ok := c.m[version]
	if !ok {
		// we haven't calculated the version, store it as a sync.OnceValues to
		// singleflight requests.
		get = sync.OnceValues(func() (map[string]string, error) {
			return c.inner.Get(version)
		})
		c.m[version] = get
	}
	c.mu.Unlock()

	return get()
}

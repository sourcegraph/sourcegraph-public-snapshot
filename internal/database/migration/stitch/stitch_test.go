package stitch

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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

	testSimple(t, "frontend", 41, 42, +1644868458, []int{1657635365})
	testSimple(t, "frontend", 40, 42, +1528395943, []int{1657635365})
	testSimple(t, "frontend", 35, 42, -1528395834, []int{1657635365})
	testSimple(t, "frontend", 29, 42, -1528395787, []int{1657635365})
}

func TestStitchCodeintelDefinitions(t *testing.T) {
	t.Parallel()

	testSimple(t, "codeintel", 41, 42, +1000000029, []int{1000000034})
	testSimple(t, "codeintel", 38, 42, +1000000029, []int{1000000034})
	testSimple(t, "codeintel", 37, 42, -1000000015, []int{1000000034})
	testSimple(t, "codeintel", 29, 42, -1000000005, []int{1000000034})
}

func TestStitchCodeinsightsDefinitions(t *testing.T) {
	t.Parallel()

	testSimple(t, "codeinsights", 41, 42, +1000000026, []int{1656517037, 1656608833})
	testSimple(t, "codeinsights", 40, 42, +1000000020, []int{1656517037, 1656608833})
	testSimple(t, "codeinsights", 38, 42, +1000000020, []int{1656517037, 1656608833})
	testSimple(t, "codeinsights", 37, 42, -1000000000, []int{1656517037, 1656608833})
	testSimple(t, "codeinsights", 29, 42, -1000000000, []int{1656517037, 1656608833})
}

func testSimple(t *testing.T, schemaName string, from, to, expectedRoot int, expectedLeaves []int) {
	t.Run(fmt.Sprintf("3.%d -> 3.%d", from, to), func(t *testing.T) {
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

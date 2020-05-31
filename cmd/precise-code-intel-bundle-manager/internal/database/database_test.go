package database

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	sqlitereader "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

func init() {
	sqliteutil.SetLocalLibpath()
	sqliteutil.MustRegisterSqlite3WithPcre()
}

func TestDatabaseExists(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{"cmd/lsif-go/main.go", true},
		{"internal/index/indexer.go", true},
		{"missing.go", false},
	}

	db := openTestDatabase(t)
	for _, testCase := range testCases {
		if exists, err := db.Exists(context.Background(), testCase.path); err != nil {
			t.Fatalf("unexpected error %s", err)
		} else if exists != testCase.expected {
			t.Errorf("unexpected exists result for %s. want=%v have=%v", testCase.path, testCase.expected, exists)
		}
	}
}

func TestDatabaseDefinitions(t *testing.T) {
	// `\ts, err := indexer.Index()` -> `\t Index() (*Stats, error)`
	//                      ^^^^^           ^^^^^

	db := openTestDatabase(t)
	if actual, err := db.Definitions(context.Background(), "cmd/lsif-go/main.go", 110, 22); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else {
		expected := []Location{
			{
				Path:  "internal/index/indexer.go",
				Range: newRange(20, 1, 20, 6),
			},
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected definitions locations (-want +got):\n%s", diff)
		}
	}
}

func TestDatabaseReferences(t *testing.T) {
	// `func (w *Writer) EmitRange(start, end Pos) (string, error) {`
	//                   ^^^^^^^^^
	//
	// -> `\t\trangeID, err := i.w.EmitRange(lspRange(ipos, ident.Name, isQuotedPkgName))`
	//                             ^^^^^^^^^
	//
	// -> `\t\t\trangeID, err = i.w.EmitRange(lspRange(ipos, ident.Name, false))`
	//                              ^^^^^^^^^

	db := openTestDatabase(t)
	if actual, err := db.References(context.Background(), "protocol/writer.go", 85, 20); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else {
		expected := []Location{
			{
				Path:  "internal/index/indexer.go",
				Range: newRange(529, 22, 529, 31),
			}, {
				Path:  "internal/index/indexer.go",
				Range: newRange(380, 22, 380, 31),
			},
			{
				Path:  "protocol/writer.go",
				Range: newRange(85, 17, 85, 26),
			},
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected reference locations (-want +got):\n%s", diff)
		}
	}
}

func TestDatabaseHover(t *testing.T) {
	// `\tcontents, err := findContents(pkgs, p, f, obj)`
	//                     ^^^^^^^^^^^^

	db := openTestDatabase(t)
	if actualText, actualRange, exists, err := db.Hover(context.Background(), "internal/index/indexer.go", 628, 20); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else if !exists {
		t.Errorf("no hover found")
	} else {
		docstring := "findContents returns contents used as hover info for given object."
		signature := "func findContents(pkgs []*Package, p *Package, f *File, obj Object) ([]MarkedString, error)"
		expectedText := "```go\n" + signature + "\n```\n\n---\n\n" + docstring
		expectedRange := newRange(628, 18, 628, 30)

		if actualText != expectedText {
			t.Errorf("unexpected hover text. want=%s have=%s", expectedText, actualText)
		}

		if diff := cmp.Diff(expectedRange, actualRange); diff != "" {
			t.Errorf("unexpected hover range (-want +got):\n%s", diff)
		}
	}
}

func TestDatabaseMonikersByPosition(t *testing.T) {
	// `func NewMetaData(id, root string, info ToolInfo) *MetaData {`
	//       ^^^^^^^^^^^

	db := openTestDatabase(t)
	if actual, err := db.MonikersByPosition(context.Background(), "protocol/protocol.go", 92, 10); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else {
		expected := [][]types.MonikerData{
			{
				{
					Kind:                 "export",
					Scheme:               "gomod",
					Identifier:           "github.com/sourcegraph/lsif-go/protocol:NewMetaData",
					PackageInformationID: types.ID("213"),
				},
			},
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected moniker result (-want +got):\n%s", diff)
		}
	}
}

func TestDatabaseMonikerResults(t *testing.T) {
	edgeLocations := []Location{
		{
			Path:  "protocol/protocol.go",
			Range: newRange(600, 1, 600, 5),
		},
		{
			Path:  "protocol/protocol.go",
			Range: newRange(644, 1, 644, 5),
		},
		{
			Path:  "protocol/protocol.go",
			Range: newRange(507, 1, 507, 5),
		},
		{
			Path:  "protocol/protocol.go",
			Range: newRange(553, 1, 553, 5),
		},
		{
			Path:  "protocol/protocol.go",
			Range: newRange(462, 1, 462, 5),
		},
		{
			Path:  "protocol/protocol.go",
			Range: newRange(484, 1, 484, 5),
		},
		{
			Path:  "protocol/protocol.go",
			Range: newRange(410, 5, 410, 9),
		},
		{
			Path:  "protocol/protocol.go",
			Range: newRange(622, 1, 622, 5),
		},
		{
			Path:  "protocol/protocol.go",
			Range: newRange(440, 1, 440, 5),
		},
		{
			Path:  "protocol/protocol.go",
			Range: newRange(530, 1, 530, 5),
		},
	}

	markdownLocations := []Location{
		{
			Path:  "internal/index/helper.go",
			Range: newRange(78, 6, 78, 16),
		},
	}

	testCases := []struct {
		tableName          string
		scheme             string
		identifier         string
		skip               int
		take               int
		expectedLocations  []Location
		expectedTotalCount int
	}{
		{"definitions", "gomod", "github.com/sourcegraph/lsif-go/protocol:Edge", 0, 100, edgeLocations, 10},
		{"definitions", "gomod", "github.com/sourcegraph/lsif-go/protocol:Edge", 3, 4, edgeLocations[3:7], 10},
		{"references", "gomod", "github.com/slimsag/godocmd:ToMarkdown", 0, 100, markdownLocations, 1},
	}

	db := openTestDatabase(t)
	for i, testCase := range testCases {
		if actual, totalCount, err := db.MonikerResults(context.Background(), testCase.tableName, testCase.scheme, testCase.identifier, testCase.skip, testCase.take); err != nil {
			t.Fatalf("unexpected error for test case #%d: %s", i, err)
		} else {
			if totalCount != testCase.expectedTotalCount {
				t.Errorf("unexpected moniker result total count for test case #%d. want=%d have=%d", i, testCase.expectedTotalCount, totalCount)
			}

			if diff := cmp.Diff(testCase.expectedLocations, actual); diff != "" {
				t.Errorf("unexpected moniker result locations for test case #%d (-want +got):\n%s", i, diff)
			}
		}
	}
}

func TestDatabasePackageInformation(t *testing.T) {
	db := openTestDatabase(t)
	if actual, exists, err := db.PackageInformation(context.Background(), "protocol/protocol.go", types.ID("213")); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else if !exists {
		t.Errorf("no package information")
	} else {
		expected := types.PackageInformationData{
			Name:    "github.com/sourcegraph/lsif-go",
			Version: "v0.0.0-ad3507cbeb18",
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected package information (-want +got):\n%s", diff)
		}
	}
}

func openTestDatabase(t *testing.T) Database {
	filename := "../../../../internal/codeintel/bundles/persistence/sqlite/testdata/lsif-go@ad3507cb.lsif.db"

	// TODO(efritz) - rewrite test not to require actual reader
	reader, err := sqlitereader.NewReader(context.Background(), filename)
	if err != nil {
		t.Fatalf("unexpected error creating reader: %s", err)
	}

	documentCache, _, err := NewDocumentCache(1)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %s", err)
	}

	resultChunkCache, _, err := NewResultChunkCache(1)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %s", err)
	}

	db, err := OpenDatabase(context.Background(), filename, reader, documentCache, resultChunkCache)
	if err != nil {
		t.Fatalf("unexpected error opening database: %s", err)
	}
	t.Cleanup(func() { _ = db.Close })

	// Wrap in observed, as that's how it's used in production
	return NewObserved(db, filename, &observation.TestContext)
}

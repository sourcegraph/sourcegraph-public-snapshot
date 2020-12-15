package lsifstore

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const testBundleID = 447

func TestDatabaseExists(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	populateTestStore(t)
	store := NewStore(dbconn.Global, &observation.TestContext)

	testCases := []struct {
		path     string
		expected bool
	}{
		{"cmd/lsif-go/main.go", true},
		{"internal/index/indexer.go", true},
		{"missing.go", false},
	}

	for _, testCase := range testCases {
		if exists, err := store.Exists(context.Background(), testBundleID, testCase.path); err != nil {
			t.Fatalf("unexpected error %s", err)
		} else if exists != testCase.expected {
			t.Errorf("unexpected exists result for %s. want=%v have=%v", testCase.path, testCase.expected, exists)
		}
	}
}

func TestDatabaseRanges(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	populateTestStore(t)
	store := NewStore(dbconn.Global, &observation.TestContext)

	//   20: // NewWriter creates a new Writer.
	//   21: func NewWriter(w io.Writer, addContents bool) *Writer {
	// > 22:     return &Writer{
	// > 23:         w:           w,
	// > 24:         addContents: addContents,
	//   25:     }
	//   26: }
	//   27:
	//   28: func (w *Writer) NumElements() int {
	//   29:     return w.numElements
	//   30: }

	if actual, err := store.Ranges(context.Background(), testBundleID, "protocol/writer.go", 21, 24); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else {
		expected := []CodeIntelligenceRange{
			{
				Range: newRange(21, 9, 21, 15),
				Definitions: []Location{
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(12, 5, 12, 11)},
				},
				References: []Location{
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(12, 5, 12, 11)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(20, 47, 20, 53)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(21, 9, 21, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(27, 9, 27, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(31, 9, 31, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(36, 9, 36, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(41, 9, 41, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(46, 9, 46, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(51, 9, 51, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(56, 9, 56, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(61, 9, 61, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(75, 9, 75, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(80, 9, 80, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(85, 9, 85, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(90, 9, 90, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(95, 9, 95, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(100, 9, 100, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(105, 9, 105, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(110, 9, 110, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(115, 9, 115, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(120, 9, 120, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(125, 9, 125, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(130, 9, 130, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(135, 9, 135, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(140, 9, 140, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(145, 9, 145, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(150, 9, 150, 15)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(155, 9, 155, 15)},
				},
				HoverText: "```go\ntype Writer struct\n```\n\n---\n\nWriter emits vertices and edges to the underlying writer. This struct will guarantee that unique identifiers are generated for each element.\n\n---\n\n```go\nstruct {\n    w Writer\n    addContents bool\n    id int\n    numElements int\n}\n```",
			},
			{
				Range: newRange(22, 2, 22, 3),
				Definitions: []Location{
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(13, 1, 13, 2)},
				},
				References: []Location{
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(13, 1, 13, 2)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(22, 2, 22, 3)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(38, 26, 38, 27)},
				},
				HoverText: "```go\nstruct field w io.Writer\n```",
			},
			{
				Range: newRange(22, 15, 22, 16),
				Definitions: []Location{
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(20, 15, 20, 16)},
				},
				References: []Location{
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(20, 15, 20, 16)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(22, 15, 22, 16)},
				},
				HoverText: "```go\nvar w Writer\n```",
			},
			{
				Range: newRange(23, 2, 23, 13),
				Definitions: []Location{
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(14, 1, 14, 12)},
				},
				References: []Location{
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(14, 1, 14, 12)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(23, 2, 23, 13)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(63, 6, 63, 17)},
				},
				HoverText: "```go\nstruct field addContents bool\n```",
			},
			{
				Range: newRange(23, 15, 23, 26),
				Definitions: []Location{
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(20, 28, 20, 39)},
				},
				References: []Location{
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(20, 28, 20, 39)},
					{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(23, 15, 23, 26)},
				},
				HoverText: "```go\nvar addContents bool\n```",
			},
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected definitions locations (-want +got):\n%s", diff)
		}
	}
}

func TestDatabaseDefinitions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	populateTestStore(t)
	store := NewStore(dbconn.Global, &observation.TestContext)

	// `\ts, err := indexer.Index()` -> `\t Index() (*Stats, error)`
	//                      ^^^^^           ^^^^^

	if actual, err := store.Definitions(context.Background(), testBundleID, "cmd/lsif-go/main.go", 110, 22); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else {
		expected := []Location{
			{DumpID: testBundleID, Path: "internal/index/indexer.go", Range: newRange(20, 1, 20, 6)},
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected definitions locations (-want +got):\n%s", diff)
		}
	}
}

func TestDatabaseReferences(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	populateTestStore(t)
	store := NewStore(dbconn.Global, &observation.TestContext)

	// `func (w *Writer) EmitRange(start, end Pos) (string, error) {`
	//                   ^^^^^^^^^
	//
	// -> `\t\trangeID, err := i.w.EmitRange(lspRange(ipos, ident.Name, isQuotedPkgName))`
	//                             ^^^^^^^^^
	//
	// -> `\t\t\trangeID, err = i.w.EmitRange(lspRange(ipos, ident.Name, false))`
	//                              ^^^^^^^^^

	if actual, err := store.References(context.Background(), testBundleID, "protocol/writer.go", 85, 20); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else {
		expected := []Location{
			{DumpID: testBundleID, Path: "internal/index/indexer.go", Range: newRange(380, 22, 380, 31)},
			{DumpID: testBundleID, Path: "internal/index/indexer.go", Range: newRange(529, 22, 529, 31)},
			{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(85, 17, 85, 26)},
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected reference locations (-want +got):\n%s", diff)
		}
	}
}

func TestDatabaseHover(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	populateTestStore(t)
	store := NewStore(dbconn.Global, &observation.TestContext)

	// `\tcontents, err := findContents(pkgs, p, f, obj)`
	//                     ^^^^^^^^^^^^

	if actualText, actualRange, exists, err := store.Hover(context.Background(), testBundleID, "internal/index/indexer.go", 628, 20); err != nil {
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
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	populateTestStore(t)
	store := NewStore(dbconn.Global, &observation.TestContext)

	// `func NewMetaData(id, root string, info ToolInfo) *MetaData {`
	//       ^^^^^^^^^^^

	if actual, err := store.MonikersByPosition(context.Background(), testBundleID, "protocol/protocol.go", 92, 10); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else {
		expected := [][]MonikerData{
			{
				{
					Kind:                 "export",
					Scheme:               "gomod",
					Identifier:           "github.com/sourcegraph/lsif-go/protocol:NewMetaData",
					PackageInformationID: "60",
				},
			},
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected moniker result (-want +got):\n%s", diff)
		}
	}
}

func TestDatabaseMonikerResults(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	populateTestStore(t)
	store := NewStore(dbconn.Global, &observation.TestContext)

	edgeDefinitionLocations := []Location{
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(410, 5, 410, 9)},
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(411, 1, 411, 8)},
	}

	edgeReferenceLocations := []Location{
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(507, 1, 507, 5)},
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(530, 1, 530, 5)},
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(516, 8, 516, 12)},
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(410, 5, 410, 9)},
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(470, 8, 470, 12)},
		{DumpID: testBundleID, Path: "internal/index/helper.go", Range: newRange(78, 8, 78, 12)},
	}

	markdownReferenceLocations := []Location{
		{DumpID: testBundleID, Path: "internal/index/helper.go", Range: newRange(78, 6, 78, 16)},
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
		{"definitions", "gomod", "github.com/sourcegraph/lsif-go/protocol:Edge", 0, 5, edgeDefinitionLocations, 2},
		{"definitions", "gomod", "github.com/sourcegraph/lsif-go/protocol:Edge", 0, 1, edgeDefinitionLocations[:1], 2},
		{"definitions", "gomod", "github.com/sourcegraph/lsif-go/protocol:Edge", 1, 5, edgeDefinitionLocations[1:], 2},
		{"references", "gomod", "github.com/sourcegraph/lsif-go/protocol:Edge", 0, 5, edgeReferenceLocations[:5], 29},
		{"references", "gomod", "github.com/sourcegraph/lsif-go/protocol:Edge", 2, 2, edgeReferenceLocations[2:4], 29},
		{"references", "gomod", "github.com/slimsag/godocmd:ToMarkdown", 0, 5, markdownReferenceLocations, 1},
	}

	for i, testCase := range testCases {
		if actual, totalCount, err := store.MonikerResults(context.Background(), testBundleID, testCase.tableName, testCase.scheme, testCase.identifier, testCase.skip, testCase.take); err != nil {
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
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	populateTestStore(t)
	store := NewStore(dbconn.Global, &observation.TestContext)

	if actual, exists, err := store.PackageInformation(context.Background(), testBundleID, "protocol/protocol.go", "60"); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else if !exists {
		t.Errorf("no package information")
	} else {
		expected := PackageInformationData{
			Name:    "github.com/sourcegraph/lsif-go",
			Version: "v0.0.0-ad3507cbeb18",
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected package information (-want +got):\n%s", diff)
		}
	}
}

func populateTestStore(t testing.TB) *Store {
	contents, err := ioutil.ReadFile("./testdata/lsif-go@ad3507cb.sql")
	if err != nil {
		t.Fatalf("unexpected error reading testdata: %s", err)
	}

	tx, err := dbconn.Global.Begin()
	if err != nil {
		t.Fatalf("unexpected error starting transaction: %s", err)
	}
	defer func() {
		if err := tx.Commit(); err != nil {
			t.Fatalf("unexpected error finishing transaction: %s", err)
		}
	}()

	for _, line := range strings.Split(string(contents), "\n") {
		if line == "" || strings.HasPrefix(line, "---") {
			continue
		}

		if _, err := tx.Exec(line); err != nil {
			t.Fatalf("unexpected error loading database data: %s", err)
		}
	}

	return NewStore(dbconn.Global, &observation.TestContext)
}

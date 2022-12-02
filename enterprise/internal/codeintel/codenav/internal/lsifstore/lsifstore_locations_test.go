package lsifstore

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const testBundleID = 1

func TestDatabaseReferences(t *testing.T) {
	store := populateTestStore(t)

	// `func (w *Writer) EmitRange(start, end Pos) (string, error) {`
	//                   ^^^^^^^^^
	//
	// -> `\t\trangeID, err := i.w.EmitRange(lspRange(ipos, ident.Name, isQuotedPkgName))`
	//                             ^^^^^^^^^
	//
	// -> `\t\t\trangeID, err = i.w.EmitRange(lspRange(ipos, ident.Name, false))`
	//                              ^^^^^^^^^

	expected := []shared.Location{
		{DumpID: testBundleID, Path: "internal/index/indexer.go", Range: newRange(380, 22, 380, 31)},
		{DumpID: testBundleID, Path: "internal/index/indexer.go", Range: newRange(529, 22, 529, 31)},
		{DumpID: testBundleID, Path: "protocol/writer.go", Range: newRange(85, 17, 85, 26)},
	}

	testCases := []struct {
		limit    int
		offset   int
		expected []shared.Location
	}{
		{5, 0, expected},
		{2, 0, expected[:2]},
		{2, 1, expected[1:]},
		{5, 5, expected[:0]},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
			if actual, totalCount, err := store.GetReferenceLocations(context.Background(), testBundleID, "protocol/writer.go", 85, 20, testCase.limit, testCase.offset); err != nil {
				t.Fatalf("unexpected error %s", err)
			} else {
				if totalCount != 3 {
					t.Errorf("unexpected count. want=%d have=%d", 3, totalCount)
				}

				if diff := cmp.Diff(testCase.expected, actual); diff != "" {
					t.Errorf("unexpected reference locations (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func populateTestStore(t testing.TB) LsifStore {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(dbtest.NewDB(logger, t))
	store := New(codeIntelDB, &observation.TestContext)

	contents, err := os.ReadFile("./testdata/lsif-go@ad3507cb.sql")
	if err != nil {
		t.Fatalf("unexpected error reading testdata: %s", err)
	}

	tx, err := codeIntelDB.Transact(context.Background())
	if err != nil {
		t.Fatalf("unexpected error starting transaction: %s", err)
	}
	defer func() {
		if err := tx.Done(nil); err != nil {
			t.Fatalf("unexpected error finishing transaction: %s", err)
		}
	}()

	// Remove comments from the lines.
	var withoutComments []byte
	for _, line := range bytes.Split(contents, []byte{'\n'}) {
		if string(line) == "" || bytes.HasPrefix(line, []byte("--")) {
			continue
		}
		withoutComments = append(withoutComments, line...)
		withoutComments = append(withoutComments, '\n')
	}

	// Execute each statement. Split on ";\n" because statements may have e.g. string literals that
	// span multiple lines.
	for _, statement := range strings.Split(string(withoutComments), ";\n") {
		if strings.Contains(statement, "_schema_versions") {
			// Statements which insert into lsif_data_*_schema_versions should not be executed, as
			// these are already inserted during regular DB up migrations.
			continue
		}
		if _, err := tx.ExecContext(context.Background(), statement); err != nil {
			t.Fatalf("unexpected error loading database data: %s", err)
		}
	}

	return store
}

func TestExtractOccurrenceData(t *testing.T) {
	t.Run("definitions", func(t *testing.T) {
		testCases := []struct {
			explanation    string
			document       *scip.Document
			occurrence     *scip.Occurrence
			expectedRanges []*scip.Range
		}{
			{
				explanation: "#1 happy path: symbol name matches and is definition",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{1, 100, 1, 200},
							Symbol:      "react 17.1 main.go func1",
							SymbolRoles: 1, // is definition
						},
					},
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react 17.1 main.go func3",
							Relationships: []*scip.Relationship{
								{IsReference: true},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react 17.1 main.go func1",
					SymbolRoles: 1, // is definition
				},
				expectedRanges: []*scip.Range{
					scip.NewRange([]int32{1, 100, 1, 200}),
				},
			},
			{
				explanation: "#2 no ranges available: symbol name does not match",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{1, 100, 1, 200},
							Symbol:      "react-test index.js func2",
							SymbolRoles: 1, // is definition
						},
					},
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react-test index.js func2",
							Relationships: []*scip.Relationship{
								{IsReference: false},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react-jest main.js func7",
					SymbolRoles: 1, // is definition
				},
				expectedRanges: []*scip.Range{},
			},
			{
				explanation: "#3 symbol name match but the SymbolRole is not a definition",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{1, 100, 1, 200},
							Symbol:      "react-test index.js func2",
							SymbolRoles: 0, // not a definition
						},
					},
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react-test index.js func2",
							Relationships: []*scip.Relationship{
								{IsReference: true},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react-test index.js func2",
					SymbolRoles: 0, // not a definition
				},
				expectedRanges: []*scip.Range{},
			},
		}

		for _, testCase := range testCases {
			res := extractOccurrenceData(testCase.document, testCase.occurrence).definitions
			if diff := cmp.Diff(testCase.expectedRanges, res); diff != "" {
				t.Errorf("unexpected ranges (-want +got):\n%s  -- %s", diff, testCase.explanation)
			}
		}
	})

	t.Run("references", func(t *testing.T) {
		testCases := []struct {
			explanation    string
			document       *scip.Document
			occurrence     *scip.Occurrence
			expectedRanges []*scip.Range
		}{
			{
				explanation: "#1 happy path: symbol name matches and it is a reference",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{1, 100, 1, 200},
							Symbol:      "react 17.1 main.go func1",
							SymbolRoles: 0, // not a definition so its a reference
						},
					},
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react 17.1 main.go func1",
							Relationships: []*scip.Relationship{
								{IsReference: true},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react 17.1 main.go func1",
					SymbolRoles: 0, // not a definition so its a reference
				},
				expectedRanges: []*scip.Range{
					scip.NewRange([]int32{1, 100, 1, 200}),
				},
			},
			{
				explanation: "#2 no ranges available: symbol name does not match",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{1, 100, 1, 200},
							Symbol:      "react-test index.js func2",
							SymbolRoles: 1, // is a definition
						},
					},
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react-test index.js func2",
							Relationships: []*scip.Relationship{
								{IsReference: true},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react-jest main.js func7",
					SymbolRoles: 0, // not a definition so its a reference
				},
				expectedRanges: []*scip.Range{},
			},
			{
				explanation: "#3 symbol name match but it is not a reference",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{5, 500, 7, 700},
							Symbol:      "react-test index.js func2",
							SymbolRoles: 1, // is a definition
						},
					},
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react-test index.js func2",
							Relationships: []*scip.Relationship{
								{IsReference: false},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react-test index.js func2",
					SymbolRoles: 0, // not a definition so its a reference
				},
				expectedRanges: []*scip.Range{},
			},
		}

		for _, testCase := range testCases {
			res := extractOccurrenceData(testCase.document, testCase.occurrence).references
			if diff := cmp.Diff(testCase.expectedRanges, res); diff != "" {
				t.Errorf("unexpected ranges (-want +got):\n%s  -- %s", diff, testCase.explanation)
			}
		}
	})

	t.Run("implementations", func(t *testing.T) {
		testCases := []struct {
			explanation    string
			document       *scip.Document
			occurrence     *scip.Occurrence
			expectedRanges []*scip.Range
		}{
			{
				explanation: "#1 happy path: symbol name match and it is a definition or an interface definition",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{1, 100, 1, 200},
							Symbol:      "react 17.1 main.go func1",
							SymbolRoles: 1, // Definition
						},
					},
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react 17.1 main.go func2",
							Relationships: []*scip.Relationship{
								{IsImplementation: true},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react 17.1 main.go func1",
					SymbolRoles: 1,
				},
				expectedRanges: []*scip.Range{
					scip.NewRange([]int32{1, 100, 1, 200}),
				},
			},
			{
				explanation: "#2 no ranges available: symbol name does not match",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{1, 100, 1, 200},
							Symbol:      "react-test index.js func2",
							SymbolRoles: 1,
						},
					},
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react-test index.js func2",
							Relationships: []*scip.Relationship{
								{IsImplementation: true},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react-jest main.js func7",
					SymbolRoles: 1,
				},
				expectedRanges: []*scip.Range{},
			},
			{
				explanation: "#3 symbol name match and occurrence is a definition",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{5, 500, 7, 700},
							Symbol:      "react-test index.js func2",
							SymbolRoles: 1, // is definition
						},
					},
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react-test index.js func2",
							Relationships: []*scip.Relationship{
								{IsTypeDefinition: true},
								{IsImplementation: false},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react-test index.js func2",
					SymbolRoles: 1, // is definition
				},
				expectedRanges: []*scip.Range{
					scip.NewRange([]int32{5, 500, 7, 700}),
				},
			},
			{
				explanation: "#4 neither occurrence nor document's occurrence are a definition",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{5, 500, 7, 700},
							Symbol:      "react-test index.js func2",
							SymbolRoles: 0,
						},
					},
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react-test index.js func2",
							Relationships: []*scip.Relationship{
								{IsTypeDefinition: true},
								{IsImplementation: false},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react-test index.js func2",
					SymbolRoles: 0,
				},
				expectedRanges: []*scip.Range{},
			},
		}

		for _, testCase := range testCases {
			res := extractOccurrenceData(testCase.document, testCase.occurrence).implementations
			if diff := cmp.Diff(testCase.expectedRanges, res); diff != "" {
				t.Errorf("unexpected ranges (-want +got):\n%s -- %s", diff, testCase.explanation)
			}
		}
	})
}

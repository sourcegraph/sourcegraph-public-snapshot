package lsifstore

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	codeintelshared "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const (
	testSCIPUploadID = 2408562
)

var p = core.NewUploadRelPathUnchecked

func posMatcher(line int, char int) shared.Matcher {
	return shared.NewStartPositionMatcher(scip.Position{Line: int32(line), Character: int32(char)})
}

func TestExtractDefinitionLocationsFromPosition(t *testing.T) {
	store := populateTestStore(t)

	// `const lru = new LRU<string, V>(cacheOptions)`
	//        ^^^
	// -> `    if (lru.has(key)) {`
	//             ^^^
	// -> `        return lru.get(key)!`
	//                    ^^^
	// -> `    lru.set(key, value)`
	//         ^^^

	occ := scip.Occurrence{Symbol: "local 12", SymbolRoles: int32(scip.SymbolRole_Definition), Range: []int32{7, 10, 13}}
	scipDefinitionLocations := []shared.UsageBuilder{
		shared.NewUsageBuilder(&occ),
	}

	testCases := []struct {
		key                 FindUsagesKey
		expectedLocations   []shared.UsageBuilder
		expectedSymbolNames []string
	}{
		{FindUsagesKey{testSCIPUploadID, p("template/src/lsif/util.ts"), posMatcher(7, 12)}, scipDefinitionLocations, nil},
		{FindUsagesKey{testSCIPUploadID, p("template/src/lsif/util.ts"), posMatcher(10, 13)}, scipDefinitionLocations, nil},
		{FindUsagesKey{testSCIPUploadID, p("template/src/lsif/util.ts"), posMatcher(12, 19)}, scipDefinitionLocations, nil},
		{FindUsagesKey{testSCIPUploadID, p("template/src/lsif/util.ts"), posMatcher(15, 10)}, scipDefinitionLocations, nil},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
			if locations, symbolNames, err := store.ExtractDefinitionLocationsFromPosition(context.Background(), testCase.key); err != nil {
				t.Fatalf("unexpected error %s", err)
			} else {
				if diff := cmp.Diff(testCase.expectedLocations, locations, cmp.Comparer(shared.UsageBuilder.Equal)); diff != "" {
					t.Errorf("unexpected locations (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(testCase.expectedSymbolNames, symbolNames, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("unexpected symbol names (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestExtractReferenceLocationsFromPosition(t *testing.T) {
	store := populateTestStore(t)

	// `const lru = new LRU<string, V>(cacheOptions)`
	//        ^^^
	// -> `    if (lru.has(key)) {`
	//             ^^^
	// -> `        return lru.get(key)!`
	//                    ^^^
	// -> `    lru.set(key, value)`
	//         ^^^

	occs := []*scip.Occurrence{
		{Symbol: "local 12", Range: []int32{10, 12, 15}, SymbolRoles: 0},
		{Symbol: "local 12", Range: []int32{12, 19, 22}, SymbolRoles: 0},
		{Symbol: "local 12", Range: []int32{15, 8, 11}, SymbolRoles: 0},
	}
	scipExpected := genslices.Map(occs, shared.NewUsageBuilder)

	testCases := []struct {
		key                 FindUsagesKey
		expectedLocations   []shared.UsageBuilder
		expectedSymbolNames []string
	}{
		{FindUsagesKey{testSCIPUploadID, p("template/src/lsif/util.ts"), posMatcher(12, 21)}, scipExpected, nil},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
			if locations, symbolNames, err := store.ExtractReferenceLocationsFromPosition(context.Background(), testCase.key); err != nil {
				t.Fatalf("unexpected error %s", err)
			} else {
				if diff := cmp.Diff(testCase.expectedLocations, locations, cmp.Comparer(shared.UsageBuilder.Equal)); diff != "" {
					t.Errorf("unexpected locations (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff(testCase.expectedSymbolNames, symbolNames, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("unexpected symbol names (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestGetSymbolUsages(t *testing.T) {
	usageKind := shared.UsageKindReference
	uploadIDs := []int{testSCIPUploadID}
	skipPaths := map[int]string{}
	lookupSymbols := []string{
		"github.com/sourcegraph/lsif-go/protocol:DefinitionResult.Vertex",
		"scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`helpers.ts`/asArray().",
	}

	store := populateTestStore(t)

	usages, totalCount, err := store.GetSymbolUsages(context.Background(), SymbolUsagesOptions{
		UsageKind:           usageKind,
		UploadIDs:           uploadIDs,
		SkipPathsByUploadID: skipPaths,
		LookupSymbols:       lookupSymbols,
		Limit:               100,
		Offset:              0,
	})
	require.NoError(t, err)
	if expected := 9; totalCount != expected {
		t.Fatalf("unexpected total count: want=%d have=%d\n", expected, totalCount)
	}

	expectedUsages := []shared.Usage{
		{Path: p("template/src/providers.ts"), Range: shared.NewRange(10, 9, 10, 16)},
		{Path: p("template/src/providers.ts"), Range: shared.NewRange(186, 43, 186, 50)},
		{Path: p("template/src/providers.ts"), Range: shared.NewRange(296, 34, 296, 41)},
		{Path: p("template/src/providers.ts"), Range: shared.NewRange(324, 38, 324, 45)},
		{Path: p("template/src/providers.ts"), Range: shared.NewRange(384, 30, 384, 37)},
		{Path: p("template/src/providers.ts"), Range: shared.NewRange(415, 8, 415, 15)},
		{Path: p("template/src/providers.ts"), Range: shared.NewRange(420, 27, 420, 34)},
		{Path: p("template/src/search/providers.ts"), Range: shared.NewRange(9, 9, 9, 16)},
		{Path: p("template/src/search/providers.ts"), Range: shared.NewRange(225, 20, 225, 27)},
	}
	for i := range expectedUsages {
		usage := &expectedUsages[i]
		usage.UploadID = testSCIPUploadID
		usage.Symbol = "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`helpers.ts`/asArray()."
		usage.Kind = shared.UsageKindReference
	}

	if diff := cmp.Diff(expectedUsages, usages); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}

func populateTestStore(t testing.TB) LsifStore {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(logger, dbtest.NewDB(t))
	store := New(observation.TestContextTB(t), codeIntelDB)

	loadTestFile(t, codeIntelDB, "./testdata/code-intel-extensions@7802976b.sql")
	return store
}

func loadTestFile(t testing.TB, codeIntelDB codeintelshared.CodeIntelDB, filepath string) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("unexpected error reading testdata from %q: %s", filepath, err)
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
}

func TestExtractOccurrenceData(t *testing.T) {
	t.Run("definitions", func(t *testing.T) {
		testCases := []struct {
			explanation    string
			document       *scip.Document
			lookup         *scip.Occurrence
			expectedRanges []shared.UsageBuilder
		}{
			{
				explanation: "#1 happy path: symbol name matches and is definition",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{1, 100, 1, 200},
							Symbol:      "react 17.1 main.go func1",
							SymbolRoles: int32(scip.SymbolRole_Definition),
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
				lookup: &scip.Occurrence{
					Symbol:      "react 17.1 main.go func1",
					SymbolRoles: 1, // is definition
				},
				expectedRanges: []shared.UsageBuilder{
					shared.NewUsageBuilder(&scip.Occurrence{
						Symbol:      "react 17.1 main.go func1",
						Range:       []int32{1, 100, 200},
						SymbolRoles: int32(scip.SymbolRole_Definition),
					}),
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
				lookup: &scip.Occurrence{
					Symbol:      "react-jest main.js func7",
					SymbolRoles: 1, // is definition
				},
				expectedRanges: []shared.UsageBuilder{},
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
				lookup: &scip.Occurrence{
					Symbol:      "react-test index.js func2",
					SymbolRoles: 0, // not a definition
				},
				expectedRanges: []shared.UsageBuilder{},
			},
		}

		for _, testCase := range testCases {
			gotDefs := extractOccurrenceData(testCase.document, testCase.lookup).definitions
			if diff := cmp.Diff(testCase.expectedRanges, gotDefs, cmp.Comparer(shared.UsageBuilder.Equal)); diff != "" {
				t.Errorf("unexpected ranges (-want +got):\n%s  -- %s", diff, testCase.explanation)
			}
		}
	})

	t.Run("references", func(t *testing.T) {
		testCases := []struct {
			explanation    string
			document       *scip.Document
			occurrence     *scip.Occurrence
			expectedRanges []shared.UsageBuilder
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
								{
									Symbol:      "react 17.1 main.go func1",
									IsReference: true,
								},
								{
									Symbol:       "react 17.1 main.go func2",
									IsDefinition: true,
								},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react 17.1 main.go func1",
					SymbolRoles: 0, // not a definition so its a reference
				},
				expectedRanges: []shared.UsageBuilder{
					shared.NewUsageBuilder(&scip.Occurrence{
						Symbol:      "react 17.1 main.go func1",
						Range:       []int32{1, 100, 200},
						SymbolRoles: 0, // reference
					}),
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
				expectedRanges: []shared.UsageBuilder{},
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
				expectedRanges: []shared.UsageBuilder{},
			},
		}

		for _, testCase := range testCases {
			gotRefs := extractOccurrenceData(testCase.document, testCase.occurrence).references
			if diff := cmp.Diff(testCase.expectedRanges, gotRefs, cmp.Comparer(shared.UsageBuilder.Equal)); diff != "" {
				t.Errorf("unexpected ranges (-want +got):\n%s  -- %s", diff, testCase.explanation)
			}
		}
	})

	t.Run("documentation and signature documentation", func(t *testing.T) {
		testCases := []struct {
			explanation string
			document    *scip.Document
			occurrence  *scip.Occurrence
			hoverText   []string
		}{
			{
				explanation: "#1 backwards compatibility: SignatureDocumentation is absent, Documentation is present",
				document: &scip.Document{
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react 17.1 main.go func1",
							Documentation: []string{
								"```go\nfunc1()\n```",
								"it does the thing",
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react 17.1 main.go func1",
					SymbolRoles: 1,
				},
				hoverText: []string{
					"```go\nfunc1()\n```",
					"it does the thing",
				},
			},
			{
				explanation: "#2: SignatureDocumentation is present",
				document: &scip.Document{
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react 17.1 main.go func1",
							SignatureDocumentation: &scip.Document{
								Language: "go",
								Text:     "func1()",
							},
							Documentation: []string{
								"it does the thing",
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react 17.1 main.go func1",
					SymbolRoles: 1,
				},
				hoverText: []string{
					"```go\nfunc1()\n```",
					"it does the thing",
				},
			},
			{
				explanation: "#3: SignatureDocumentation is present, but Text/Language are empty",
				document: &scip.Document{
					Symbols: []*scip.SymbolInformation{
						{
							Symbol:                 "react 17.1 main.go func1",
							SignatureDocumentation: &scip.Document{},
							Documentation: []string{
								"it does the thing",
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react 17.1 main.go func1",
					SymbolRoles: 1,
				},
				hoverText: []string{
					"it does the thing",
				},
			},
		}

		for _, testCase := range testCases {
			if diff := cmp.Diff(testCase.hoverText, extractOccurrenceData(testCase.document, testCase.occurrence).hoverText); diff != "" {
				t.Errorf("unexpected documentation (-want +got):\n%s -- %s", diff, testCase.explanation)
			}
		}
	})

	t.Run("implementations", func(t *testing.T) {
		testCases := []struct {
			explanation    string
			document       *scip.Document
			occurrence     *scip.Occurrence
			expectedRanges []shared.UsageBuilder
		}{
			{
				explanation: "#1 happy path: we have implementation",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{3, 300, 4, 400},
							Symbol:      "react 17.1 main.go func1A",
							SymbolRoles: 1, // a definition
						},
					},
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react 17.1 main.go func1A",
							Relationships: []*scip.Relationship{
								{
									Symbol:           "react 17.1 main.go func1",
									IsImplementation: true,
								},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react 17.1 main.go func1",
					SymbolRoles: 1,
				},
				expectedRanges: []shared.UsageBuilder{
					shared.NewUsageBuilder(&scip.Occurrence{
						Symbol:      "react 17.1 main.go func1A",
						SymbolRoles: int32(scip.SymbolRole_Definition),
						Range:       []int32{3, 300, 400},
					}),
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
				expectedRanges: []shared.UsageBuilder{},
			},
		}

		for _, testCase := range testCases {
			gotImpls := extractOccurrenceData(testCase.document, testCase.occurrence).implementations
			if diff := cmp.Diff(testCase.expectedRanges, gotImpls, cmp.Comparer(shared.UsageBuilder.Equal)); diff != "" {
				t.Errorf("unexpected ranges (-want +got):\n%s -- %s", diff, testCase.explanation)
			}
		}
	})

	t.Run("prototypes", func(t *testing.T) {
		testCases := []struct {
			explanation    string
			document       *scip.Document
			occurrence     *scip.Occurrence
			expectedRanges []shared.UsageBuilder
		}{
			{
				explanation: "#1 happy path: we have prototype",
				document: &scip.Document{
					Occurrences: []*scip.Occurrence{
						{
							Range:       []int32{3, 300, 4, 400},
							Symbol:      "react 17.1 main.go func1",
							SymbolRoles: 1, // a definition
						},
					},
					Symbols: []*scip.SymbolInformation{
						{
							Symbol: "react 17.1 main.go func1A",
							Relationships: []*scip.Relationship{
								{
									Symbol:           "react 17.1 main.go func1",
									IsImplementation: true,
								},
							},
						},
					},
				},
				occurrence: &scip.Occurrence{
					Symbol:      "react 17.1 main.go func1A",
					SymbolRoles: 1,
				},
				expectedRanges: []shared.UsageBuilder{
					shared.NewUsageBuilder(&scip.Occurrence{
						Symbol:      "react 17.1 main.go func1",
						Range:       []int32{3, 300, 400},
						SymbolRoles: int32(scip.SymbolRole_Definition),
					}),
				},
			},
		}

		for _, testCase := range testCases {
			gotPrototypes := extractOccurrenceData(testCase.document, testCase.occurrence).prototypes
			if diff := cmp.Diff(testCase.expectedRanges, gotPrototypes, cmp.Comparer(shared.UsageBuilder.Equal)); diff != "" {
				t.Errorf("unexpected ranges (-want +got):\n%s -- %s", diff, testCase.explanation)
			}
		}
	})
}

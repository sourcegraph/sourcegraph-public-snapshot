package lsifstore

import (
	"context"
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// Note: You can `go test ./pkg -update` to update the expected `want` values in these tests.
// See https://github.com/hexops/autogold for more information.

func TestDocumentationPage(t *testing.T) {
	store := populateTestStore(t)

	got, err := store.DocumentationPage(context.Background(), testBundleID, "/")
	if err != nil {
		t.Fatal(err)
	}

	// Just a snapshot of the output so we know when anything changes.
	autogold.Equal(t, got)
}

func TestDocumentationPathInfo(t *testing.T) {
	store := populateTestStore(t)

	testCases := []struct {
		pathID string
		want   autogold.Value
	}{
		{"/github.com/sourcegraph/lsif-go", autogold.Want("/github.com/sourcegraph/lsif-go", &precise.DocumentationPathInfoData{
			PathID:  "/github.com/sourcegraph/lsif-go",
			IsIndex: true,
			Children: []string{
				"/github.com/sourcegraph/lsif-go/cmd",
				"/github.com/sourcegraph/lsif-go/internal",
			},
		})},
		{"/github.com/sourcegraph/lsif-go/cmd", autogold.Want("/github.com/sourcegraph/lsif-go/cmd", &precise.DocumentationPathInfoData{
			PathID:  "/github.com/sourcegraph/lsif-go/cmd",
			IsIndex: true,
			Children: []string{
				"/github.com/sourcegraph/lsif-go/cmd/lsif-go",
			},
		})},
		{"/github.com/sourcegraph/lsif-go/internal", autogold.Want("/github.com/sourcegraph/lsif-go/internal", &precise.DocumentationPathInfoData{
			PathID:  "/github.com/sourcegraph/lsif-go/internal",
			IsIndex: true,
			Children: []string{
				"/github.com/sourcegraph/lsif-go/internal/gomod",
				"/github.com/sourcegraph/lsif-go/internal/index",
			},
		})},
		{"/github.com/sourcegraph/lsif-go/internal/index", autogold.Want("/github.com/sourcegraph/lsif-go/internal/index", &precise.DocumentationPathInfoData{
			PathID:   "/github.com/sourcegraph/lsif-go/internal/index",
			Children: []string{},
		})},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got, err := store.DocumentationPathInfo(context.Background(), testBundleID, tc.pathID)
			if err != nil {
				t.Fatal(err)
			}
			tc.want.Equal(t, got)
		})
	}
}

func TestDocumentationDefinitions(t *testing.T) {
	store := populateTestStore(t)

	testCases := []autogold.Value{
		autogold.Want("/github.com/sourcegraph/lsif-go/internal/index#NewIndexer", []Location{{
			DumpID: 1,
			Path:   "internal/index/indexer.go",
			Range: Range{
				Start: Position{
					Line:      62,
					Character: 5,
				},
				End: Position{
					Line:      62,
					Character: 15,
				},
			},
		}}),
		autogold.Want("/github.com/sourcegraph/lsif-go/internal/gomod#versionPattern", []Location{{
			DumpID: 1,
			Path:   "internal/gomod/module.go",
			Range: Range{
				Start: Position{
					Line:      84,
					Character: 4,
				},
				End: Position{
					Line:      84,
					Character: 18,
				},
			},
		}}),
		autogold.Want("/github.com/sourcegraph/lsif-go/cmd/lsif-go#main", []Location{{
			DumpID: 1,
			Path:   "cmd/lsif-go/main.go",
			Range: Range{
				Start: Position{
					Line:      26,
					Character: 5,
				},
				End: Position{
					Line:      26,
					Character: 9,
				},
			},
		}}),
	}
	for _, want := range testCases {
		t.Run(want.Name(), func(t *testing.T) {
			pathID := want.Name()
			got, _, err := store.DocumentationDefinitions(context.Background(), testBundleID, pathID, 10, 0)
			if err != nil {
				t.Fatal(err)
			}
			want.Equal(t, got)
		})
	}
}

// Confirms that the various data fields (detail strings, labels, tags, etc.) are not changing via a
// snapshot.
func TestDocumentationSearch_resultData(t *testing.T) {
	store := populateTestStore(t)

	got, err := store.DocumentationSearch(context.Background(), "public", "error", nil)
	if err != nil {
		t.Fatal(err)
	}
	autogold.Equal(t, got)
}

func TestDocumentationSearch(t *testing.T) {
	store := populateTestStore(t)

	testCases := []struct {
		query      string
		repos      []string
		showLabels bool
		want       autogold.Value
	}{
		{
			query: "",
			want:  autogold.Want("empty query does not error", []string{}),
		},
		{
			query: "EmitDocument",
			want:  autogold.Want("basic global search", []string{"protocol.Writer.EmitDocument"}),
		},
		{
			query: "Emit",
			want:  autogold.Want("prefix global search returns no results", []string{}),
		},
		{
			query: "sourcegraph/lsif-go: Emit",
			want: autogold.Want("prefix repo search returns results", []string{
				"protocol.Writer.EmitResultSet", "protocol.Writer.EmitReferenceResult",
				"protocol.Writer.EmitDefinitionResult",
				"protocol.Writer.EmitProject",
				"protocol.Writer.EmitTextDocumentReferences",
				"protocol.Writer.EmitTextDocumentHover",
				"protocol.Writer.EmitTextDocumentDefinition",
				"protocol.Writer.EmitRange",
				"protocol.Writer.EmitPackageInformationEdge",
				"protocol.Writer.EmitNext",
			}),
		},
		{
			query: "sourcegraph/lsif-go: MonikerEdge",
			want: autogold.Want("suffix repo search returns results", []string{
				"protocol.NextMonikerEdge", "protocol.MonikerEdge",
				"protocol.NewNextMonikerEdge",
				"protocol.NewMonikerEdge",
				"protocol.Writer.EmitMonikerEdge",
			}),
		},
		{
			query:      "error",
			showLabels: true,
			want: autogold.Want("global search generic term", []string{
				"func realMain() error", "func InferModuleVersion(projectRoot string) (string, error)",
				"func (i *indexer) emitImportMoniker(sourceID, identifier string) error",
				"func (i *indexer) emitExportMoniker(sourceID, identifier string) error",
				"func (w *Writer) emit(v interface{}) error",
				"func (w *Writer) EmitResultSet() (string, error)",
				"func (w *Writer) EmitReferenceResult() (string, error)",
				"func (w *Writer) EmitDefinitionResult() (string, error)",
				"func (m *MarkedString) UnmarshalJSON(data []byte) error",
				"func (i *indexer) Index() (*Stats, error)",
			}),
		},
		{
			query:      "function: error",
			showLabels: true,
			want: autogold.Want("functions only", []string{
				"func realMain() error", "func InferModuleVersion(projectRoot string) (string, error)",
				"func ListModules(projectRoot string) (string, map[string]string, error)",
				"func constructMarkedString(s, comments, extra string) ([]protocol.MarkedString, error)",
				"func run(dir, command string, args ...string) (string, error)",
				"func findComments(pkgs []*packages.Package, p *packages.Package, f *ast.File, o types.Object) (string, error)",
				"func findContents(pkgs []*packages.Package, p *packages.Package, f *ast.File, obj types.Object) ([]protocol.MarkedString, error)",
				"func externalHoverContents(pkgs []*packages.Package, p *packages.Package, obj types.Object, pkg *types.Package) ([]protocol.MarkedString, error)",
			}),
		},
		{
			query:      "variable: versionPattern",
			showLabels: true,
			want:       autogold.Want("variables only", []string{"var versionPattern"}),
		},
		{
			query:      "constant string private: version",
			showLabels: true,
			want:       autogold.Want("multiple tags in the correct order", []string{"const version"}),
		},
		{
			// TODO(apidocs): This is not ideal behavior, no tags match so they're effectively
			// ignored. Should switch tag matching query to be `(tsv @@ 'foo') AND (tsv @@ 'bar')`
			// instead of `(tsv @@ 'foo <-> bar')` probably.
			query:      "private constant string: version",
			showLabels: true,
			want: autogold.Want("multiple tags in incorrect order", []string{
				"const version", "func cleanVersion(version string) string",
				"func NewPackageInformation(id, name, manager, version string) *PackageInformation",
				"func (i *indexer) ensurePackageInformation(packageName, version string) (string, error)",
				"func (w *Writer) EmitPackageInformation(packageName, scheme, version string) (string, error)",
			}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			results, err := store.DocumentationSearch(context.Background(), "public", tc.query, tc.repos)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, result := range results {
				if tc.showLabels {
					got = append(got, result.Label)
					continue
				}
				got = append(got, result.SearchKey)
			}
			tc.want.Equal(t, got)
		})
	}
}

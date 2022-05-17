package squirrel

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func init() {
	if _, ok := os.LookupEnv("NO_COLOR"); !ok {
		color.NoColor = false
	}
}

func TestNonLocalDefinition(t *testing.T) {
	repoDirs, err := os.ReadDir("test_repos")
	fatalIfErrorLabel(t, err, "reading test_repos")

	annotations := []annotation{}

	for _, repoDir := range repoDirs {
		if !repoDir.IsDir() {
			t.Fatalf("unexpected file %s", repoDir.Name())
		}

		base := filepath.Join("test_repos", repoDir.Name())
		err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			contents, err := os.ReadFile(path)
			fatalIfErrorLabel(t, err, "reading annotations from a file")

			rel, err := filepath.Rel(base, path)
			fatalIfErrorLabel(t, err, "getting relative path")
			repoCommitPath := types.RepoCommitPath{Repo: repoDir.Name(), Commit: "abc", Path: rel}

			annotations = append(annotations, collectAnnotations(repoCommitPath, string(contents))...)

			return nil
		})
		fatalIfErrorLabel(t, err, "walking a repo dir")
	}

	readFile := func(ctx context.Context, path types.RepoCommitPath) ([]byte, error) {
		contents, err := os.ReadFile(filepath.Join("test_repos", path.Repo, path.Path))
		fatalIfErrorLabel(t, err, "reading a file")
		return contents, nil
	}

	squirrel := New(readFile, nil)
	squirrel.errorOnParseFailure = true
	defer squirrel.Close()

	symbolToKindToAnnotations := groupBySymbolAndKind(annotations)
	symbols := []string{}
	for symbol := range symbolToKindToAnnotations {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	for _, symbol := range symbols {
		m := symbolToKindToAnnotations[symbol]
		var wantDef *annotation
		for _, ann := range m["def"] {
			if wantDef != nil {
				t.Fatalf("multiple definitions for symbol %s", symbol)
			}

			annCopy := ann
			wantDef = &annCopy
		}

		if wantDef == nil {
			t.Fatalf("no matching \"def\" annotation for \"ref\" %s", symbol)
		}

		for _, ref := range m["ref"] {
			squirrel.breadcrumbs = Breadcrumbs{}
			got, err := squirrel.symbolInfo(context.Background(), ref.repoCommitPathPoint)
			fatalIfErrorLabel(t, err, "symbolInfo")

			if got == nil {
				squirrel.breadcrumbs.prettyPrint(squirrel.readFile)
				t.Fatalf("no symbolInfo for symbol %s", symbol)
			}

			if ref.symbol == "f2" {
				squirrel.breadcrumbs.prettyPrint(squirrel.readFile)
			}

			gotDef := types.RepoCommitPathPoint{
				RepoCommitPath: got.Definition.RepoCommitPath,
				Point: types.Point{
					Row:    got.Definition.Row,
					Column: got.Definition.Column,
				},
			}

			if diff := cmp.Diff(wantDef.repoCommitPathPoint, gotDef); diff != "" {
				t.Fatalf("wrong definition (-want +got):\n%s", diff)
			}
		}
	}
}

func groupBySymbolAndKind(annotations []annotation) map[string]map[string][]annotation {
	grouped := map[string]map[string][]annotation{}

	for _, a := range annotations {
		if _, ok := grouped[a.symbol]; !ok {
			grouped[a.symbol] = map[string][]annotation{}
		}

		if _, ok := grouped[a.symbol][a.kind]; !ok {
			grouped[a.symbol][a.kind] = []annotation{}
		}

		grouped[a.symbol][a.kind] = append(grouped[a.symbol][a.kind], a)
	}

	return grouped
}

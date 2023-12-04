package squirrel

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"testing"

	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
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

	readFile := func(ctx context.Context, path types.RepoCommitPath) ([]byte, error) {
		return os.ReadFile(filepath.Join("test_repos", path.Repo, path.Path))
	}

	tempSquirrel := New(readFile, nil)
	allSymbols := []result.Symbol{}

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

			symbols, err := tempSquirrel.getSymbols(context.Background(), repoCommitPath)
			fatalIfErrorLabel(t, err, "getSymbols")
			allSymbols = append(allSymbols, symbols...)

			return nil
		})
		fatalIfErrorLabel(t, err, "walking a repo dir")
	}

	ss := func(ctx context.Context, args search.SymbolsParameters) (result.Symbols, error) {
		results := result.Symbols{}
	nextSymbol:
		for _, s := range allSymbols {
			if args.IncludePatterns != nil {
				for _, p := range args.IncludePatterns {
					match, err := regexp.MatchString(p, s.Path)
					fatalIfErrorLabel(t, err, "matching a pattern")
					if !match {
						continue nextSymbol
					}
				}
			}
			match, err := regexp.MatchString(args.Query, s.Name)
			if err != nil {
				return nil, err
			}
			if match {
				results = append(results, s)
			}
		}
		return results, nil
	}

	squirrel := New(readFile, ss)
	squirrel.errorOnParseFailure = true
	defer squirrel.Close()

	cwd, err := os.Getwd()
	fatalIfErrorLabel(t, err, "getting cwd")

	solo := ""
	for _, a := range annotations {
		for _, tag := range a.tags {
			if tag == "solo" {
				solo = a.symbol
			}
		}
	}

	testCount := 0

	symbolToTagToAnnotations := groupBySymbolAndTag(annotations)
	symbols := []string{}
	for symbol := range symbolToTagToAnnotations {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)
	for _, symbol := range symbols {
		if solo != "" && symbol != solo {
			continue
		}
		m := symbolToTagToAnnotations[symbol]
		if m["def"] == nil && m["path"] != nil {
			// It's a path definition, which is checked separately
			continue
		}
		var defAnn *annotation
		for _, ann := range m["def"] {
			if defAnn != nil {
				t.Fatalf("multiple definitions for symbol %s", symbol)
			}

			annCopy := ann
			defAnn = &annCopy
		}

		for _, ref := range m["ref"] {
			squirrel.breadcrumbs = Breadcrumbs{}
			gotSymbolInfo, err := squirrel.SymbolInfo(context.Background(), ref.repoCommitPathPoint)
			fatalIfErrorLabel(t, err, "symbolInfo")

			if slices.Contains(ref.tags, "nodef") {
				if gotSymbolInfo != nil {
					t.Fatalf("unexpected definition for %s", ref.symbol)
				}
				testCount += 1
				continue
			}

			if gotSymbolInfo == nil {
				squirrel.breadcrumbs.prettyPrint(squirrel.readFile)
				t.Fatalf("no symbolInfo for symbol %s", symbol)
			}

			if defAnn == nil {
				t.Fatalf("no \"def\" for symbol %q", symbol)
			}

			if gotSymbolInfo.Definition.Range == nil {
				squirrel.breadcrumbs.prettyPrint(squirrel.readFile)
				t.Fatalf("no definition range for symbol %s", symbol)
			}

			if m["print"] != nil {
				squirrel.breadcrumbs.prettyPrint(squirrel.readFile)
			}

			got := types.RepoCommitPathPoint{
				RepoCommitPath: gotSymbolInfo.Definition.RepoCommitPath,
				Point: types.Point{
					Row:    gotSymbolInfo.Definition.Row,
					Column: gotSymbolInfo.Definition.Column,
				},
			}

			if diff := cmp.Diff(defAnn.repoCommitPathPoint, got); diff != "" {
				squirrel.breadcrumbs.prettyPrint(squirrel.readFile)

				t.Errorf("wrong symbolInfo for %q\n", symbol)
				want := defAnn.repoCommitPathPoint
				t.Errorf("want: %s%s/%s:%d:%d\n", itermSource(filepath.Join(cwd, "test_repos", want.Repo, want.Path), want.Point.Row), want.Repo, want.Path, want.Point.Row, want.Point.Column)
				t.Errorf("got : %s%s/%s:%d:%d\n", itermSource(filepath.Join(cwd, "test_repos", got.Repo, got.Path), got.Point.Row), got.Repo, got.Path, got.Point.Row, got.Point.Column)
			}

			testCount += 1
		}
	}

	// Also test path definitions
	for _, a := range annotations {
		if solo != "" && a.symbol != solo {
			continue
		}
		for _, tag := range a.tags {
			if tag == "path" {
				squirrel.breadcrumbs = Breadcrumbs{}
				gotSymbolInfo, err := squirrel.SymbolInfo(context.Background(), a.repoCommitPathPoint)
				fatalIfErrorLabel(t, err, "symbolInfo")

				if gotSymbolInfo == nil {
					squirrel.breadcrumbs.prettyPrint(squirrel.readFile)
					t.Fatalf("no symbolInfo for path %s", a.symbol)
				}

				if gotSymbolInfo.Definition.Range != nil {
					squirrel.breadcrumbs.prettyPrint(squirrel.readFile)
					t.Fatalf("symbolInfo returned a range for %s", a.symbol)
				}

				if gotSymbolInfo.Definition.RepoCommitPath.Path != a.symbol {
					squirrel.breadcrumbs.prettyPrint(squirrel.readFile)
					t.Fatalf("expected path %s, got %s", a.symbol, gotSymbolInfo.Definition.RepoCommitPath.Path)
				}

				testCount += 1
			}
		}
	}

	t.Logf("%d tests in total", testCount)
}

func groupBySymbolAndTag(annotations []annotation) map[string]map[string][]annotation {
	grouped := map[string]map[string][]annotation{}

	for _, a := range annotations {
		if _, ok := grouped[a.symbol]; !ok {
			grouped[a.symbol] = map[string][]annotation{}
		}

		for _, tag := range a.tags {
			grouped[a.symbol][tag] = append(grouped[a.symbol][tag], a)
		}
	}

	return grouped
}

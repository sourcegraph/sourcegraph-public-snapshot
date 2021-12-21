package search_based

import (
	"context"
	"flag"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/search-based/api"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "update golden files, removing unused if running all tests")

func TestIndexers(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory, err %v", err)
	}
	if *update {
		goldensDir := filepath.Join(cwd, "golden")
		err := os.RemoveAll(goldensDir)
		if err != nil {
			t.Fatalf("os.RemoveAll failed to delete golden dir %v", goldensDir)
		}
	}

	for _, indexer := range AllIndexers {
		t.Run(indexer.Name(), func(t *testing.T) {
			env := goldenEnv{
				cwd:     cwd,
				tests:   filepath.Join(cwd, "tests", indexer.Name()),
				golden:  filepath.Join(cwd, "golden", indexer.Name()),
				indexer: indexer,
			}
			err := filepath.Walk(env.tests, func(path string, info fs.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				if err != nil {
					t.Fatalf("filepath.Walk error %v", err)
				}
				relativePath, err := filepath.Rel(env.tests, path)
				if err != nil {
					t.Fatalf("filelath.Rel error %v", err)
				}
				t.Run(relativePath, func(t *testing.T) {
					testIndexer(env, relativePath, t)
				})
				return nil
			})
			if err != nil {
				t.Fatalf("filepath.Walk error %v", err)
			}
		})
	}
}

func goldenDocument(doc *lsif_typed.Document) (string, error) {
	b := strings.Builder{}
	uri, err := url.Parse(doc.Uri)
	if err != nil {
		return "", err
	}
	if uri.Scheme != "file" {
		return "", errors.New("expected url scheme 'file', obtained " + uri.Scheme)
	}
	data, err := os.ReadFile(uri.Path)
	if err != nil {
		return "", err
	}
	sort.SliceStable(doc.Occurrences, func(i, j int) bool {
		return isRangeLess(doc.Occurrences[i].Range, doc.Occurrences[j].Range)
	})
	i := 0
	for lineNumber, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSuffix(line, "\r")
		b.WriteString("  ")
		b.WriteString(strings.ReplaceAll(line, "\t", " "))
		b.WriteString("\n")
		for i < len(doc.Occurrences) && doc.Occurrences[i].Range.Start.Line == int32(lineNumber) {
			occ := doc.Occurrences[i]
			isSingleLine := occ.Range.Start.Line == occ.Range.End.Line
			if !isSingleLine {
				// TODO
				continue
			}
			b.WriteString("//")
			for indent := int32(0); indent < occ.Range.Start.Character; indent++ {
				b.WriteRune(' ')
			}
			length := int(occ.Range.End.Character - occ.Range.Start.Character)
			for caret := 0; caret < length; caret++ {
				b.WriteRune('^')
			}
			b.WriteRune(' ')
			b.WriteString(occ.MonikerId)
			b.WriteRune(' ')
			switch occ.Role {
			case lsif_typed.MonikerOccurrence_ROLE_DEFINITION:
				b.WriteString("definition")
			case lsif_typed.MonikerOccurrence_ROLE_REFERENCE:
				b.WriteString("reference")
			}
			b.WriteString("\n")
			i++
		}
	}
	return b.String(), nil
}

type goldenEnv struct {
	cwd     string
	tests   string
	golden  string
	indexer api.Indexer
}

func isRangeLess(a *lsif_typed.Range, b *lsif_typed.Range) bool {
	if a.Start.Line != b.Start.Line {
		return a.Start.Line < b.Start.Line
	}
	if a.Start.Character != b.Start.Character {
		return a.Start.Character < b.Start.Character
	}
	if a.End.Line != b.End.Line {
		return a.End.Line < b.End.Line
	}
	return a.End.Character < b.End.Character
}

func testIndexer(env goldenEnv, relativePath string, t *testing.T) {
	originalPath := filepath.Join(env.tests, relativePath)
	data, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatalf("failed to read file %v, err %v", originalPath, err)
	}
	doc, err := env.indexer.Index(context.Background(), api.NewInput(originalPath, data), &api.IndexingOptions{})
	if err != nil {
		t.Fatalf("failed to index doc %v, err %v", originalPath, err)
	}
	obtainedGolden, err := goldenDocument(doc)
	if err != nil {
		t.Fatalf("failed to golden, err %v", err)
	}
	goldenPath := filepath.Join(env.golden, relativePath)
	if *update {
		parentDir := filepath.Dir(goldenPath)
		err = os.MkdirAll(parentDir, 0755)
		if err != nil {
			t.Fatalf("failed to create directory %v", parentDir)
		}
		err := os.WriteFile(goldenPath, []byte(obtainedGolden), 0644)
		if err != nil {
			t.Fatalf("failed to write to path %v", goldenPath)
		}
	} else {
		expectedGolden, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Fatalf("os.ReadFile failed, relativePath %v, err %v", goldenPath, err)
		}
		edits := myers.ComputeEdits(span.URIFromPath(goldenPath), string(expectedGolden), obtainedGolden)
		if len(edits) > 0 {
			diff := fmt.Sprint(gotextdiff.ToUnified(
				goldenPath+" (obtained)",
				goldenPath+" (expected)",
				string(expectedGolden),
				edits,
			))
			t.Fatalf("\n" + diff)
		}
	}
}

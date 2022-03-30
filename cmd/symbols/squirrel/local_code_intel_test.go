package squirrel

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestLocalCodeIntel(t *testing.T) {
	path := types.RepoCommitPath{Repo: "foo", Commit: "bar", Path: "test.java"}
	contents := `
class Foo {

    //             v f1.p def
    //             v f1.p ref
    void f1(String p) {

        //     v f1.x def
        //     v f1.x ref
        //         v f1.p ref
        String x = p;
    }

    //             v f2.p def
    //             v f2.p ref
    void f2(String p) {

        //     v f2.x def
        //     v f2.x ref
        //         v f2.p ref
        String x = p;
    }
}
`

	want := collectAnnotations(path, contents)

	payload := getLocalCodeIntel(t, path, contents)
	got := []annotation{}
	for _, symbol := range payload.Symbols {
		got = append(got, annotation{
			repoCommitPathPoint: types.RepoCommitPathPoint{
				RepoCommitPath: path,
				Point: types.Point{
					Row:    symbol.Def.Row,
					Column: symbol.Def.Column,
				},
			},
			symbol: "(unused)",
			kind:   "def",
		})

		for _, ref := range symbol.Refs {
			got = append(got, annotation{
				repoCommitPathPoint: types.RepoCommitPathPoint{
					RepoCommitPath: path,
					Point: types.Point{
						Row:    ref.Row,
						Column: ref.Column,
					},
				},
				symbol: "(unused)",
				kind:   "ref",
			})
		}
	}

	sortAnnotations(want)
	sortAnnotations(got)

	if diff := cmp.Diff(want, got, compareAnnotations); diff != "" {
		t.Fatalf("unexpected annotations (-want +got):\n%s", diff)
	}
}

func getLocalCodeIntel(t *testing.T, path types.RepoCommitPath, contents string) *types.LocalCodeIntelPayload {
	readFile := func(ctx context.Context, path types.RepoCommitPath) ([]byte, error) {
		return []byte(contents), nil
	}

	squirrel := NewSquirrelService(readFile)
	defer squirrel.Close()

	payload, err := squirrel.localCodeIntel(context.Background(), path)
	fatalIfError(t, err)

	return payload
}

type annotation struct {
	repoCommitPathPoint types.RepoCommitPathPoint
	symbol              string
	kind                string
}

func collectAnnotations(repoCommitPath types.RepoCommitPath, contents string) []annotation {
	annotations := []annotation{}

	lines := strings.Split(contents, "\n")

	// Annotation at the end of the line
	for i, line := range lines {
		matches := regexp.MustCompile(`^([^<]+)< "([^"]+)" ([a-zA-Z0-9_.-]+) (def|ref)`).FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		substr, symbol, kind := matches[2], matches[3], matches[4]

		annotations = append(annotations, annotation{
			repoCommitPathPoint: types.RepoCommitPathPoint{
				RepoCommitPath: repoCommitPath,
				Point: types.Point{
					Row:    i,
					Column: strings.Index(line, substr),
				},
			},
			symbol: symbol,
			kind:   kind,
		})
	}

	// Annotations below source lines
nextSourceLine:
	for sourceLine := 0; ; {
		for annLine := sourceLine + 1; ; annLine++ {
			if annLine >= len(lines) {
				break nextSourceLine
			}

			matches := regexp.MustCompile(`([^^]*)\^+ ([a-zA-Z0-9_.-]+) (def|ref)`).FindStringSubmatch(lines[annLine])
			if matches == nil {
				sourceLine = annLine
				continue nextSourceLine
			}

			prefix, symbol, kind := matches[1], matches[2], matches[3]

			annotations = append(annotations, annotation{
				repoCommitPathPoint: types.RepoCommitPathPoint{
					RepoCommitPath: repoCommitPath,
					Point: types.Point{
						Row:    sourceLine,
						Column: spacesToColumn(lines[sourceLine], lengthInSpaces(prefix)),
					},
				},
				symbol: symbol,
				kind:   kind,
			})
		}
	}

	// Annotations above source lines
previousSourceLine:
	for sourceLine := len(lines) - 1; ; {
		for annLine := sourceLine - 1; ; annLine-- {
			if annLine < 0 {
				break previousSourceLine
			}

			matches := regexp.MustCompile(`([^v]*)v+ ([a-zA-Z0-9_.-]+) (def|ref)`).FindStringSubmatch(lines[annLine])
			if matches == nil {
				sourceLine = annLine
				continue previousSourceLine
			}

			prefix, symbol, kind := matches[1], matches[2], matches[3]

			annotations = append(annotations, annotation{
				repoCommitPathPoint: types.RepoCommitPathPoint{
					RepoCommitPath: repoCommitPath,
					Point: types.Point{
						Row:    sourceLine,
						Column: spacesToColumn(lines[sourceLine], lengthInSpaces(prefix)),
					},
				},
				symbol: symbol,
				kind:   kind,
			})
		}
	}

	return annotations
}

func sortAnnotations(annotations []annotation) {
	sort.Slice(annotations, func(i, j int) bool {
		rowi := annotations[i].repoCommitPathPoint.Point.Row
		rowj := annotations[j].repoCommitPathPoint.Point.Row
		coli := annotations[i].repoCommitPathPoint.Point.Column
		colj := annotations[j].repoCommitPathPoint.Point.Column
		kindi := annotations[i].kind
		kindj := annotations[j].kind
		if rowi != rowj {
			return rowi < rowj
		} else if coli != colj {
			return coli < colj
		} else {
			return kindi < kindj
		}
	})
}

func printRowColumnKind(annotations []annotation) string {
	sortAnnotations(annotations)

	lines := []string{}
	for _, annotation := range annotations {
		lines = append(lines, fmt.Sprintf(
			"%d:%d %s",
			annotation.repoCommitPathPoint.Row,
			annotation.repoCommitPathPoint.Column,
			annotation.kind,
		))
	}

	return strings.Join(lines, "\n")
}

var compareAnnotations = cmp.Comparer(func(a, b annotation) bool {
	if a.repoCommitPathPoint.RepoCommitPath != b.repoCommitPathPoint.RepoCommitPath {
		return false
	}
	if a.repoCommitPathPoint.Point.Row != b.repoCommitPathPoint.Point.Row {
		return false
	}
	if a.repoCommitPathPoint.Point.Column != b.repoCommitPathPoint.Point.Column {
		return false
	}
	if a.kind != b.kind {
		return false
	}
	return true
})

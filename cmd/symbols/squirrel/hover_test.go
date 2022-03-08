package squirrel

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestHover(t *testing.T) {
	golang := `
// not a comment line

// comment line 1
// comment line 2
func f() {}
`

	_ = golang

	tests := []struct {
		path     string
		contents string
		want     string
	}{
		{"test.go", golang, "comment line 1\ncomment line 2\n"},
	}

	readFile := func(ctx context.Context, path types.RepoCommitPath) ([]byte, error) {
		for _, test := range tests {
			if test.path == path.Path {
				return []byte(test.contents), nil
			}
		}
		return nil, fmt.Errorf("path %s not found", path.Path)
	}

	squirrel := NewSquirrelService(readFile, nil)
	defer squirrel.Close()

	for _, test := range tests {
		payload, err := squirrel.localCodeIntel(context.Background(), types.RepoCommitPath{Repo: "foo", Commit: "bar", Path: test.path}, readFile)
		fatalIfError(t, err)

		ok := false
		for _, symbol := range payload.Symbols {
			if symbol.Hover == nil {
				continue
			}

			got := *symbol.Hover

			if !strings.Contains(got, test.want) {
				continue
			} else {
				ok = true
				break
			}
		}

		if !ok {
			comments := []string{}
			for _, symbol := range payload.Symbols {
				if symbol.Hover == nil {
					continue
				}
				comments = append(comments, *symbol.Hover)
			}
			fmt.Printf("did not find comment %q. All comments:\n", test.want)
			for _, comment := range comments {
				fmt.Printf("%q\n", comment)
			}
			t.FailNow()
		}
	}
}

func fatalIfError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

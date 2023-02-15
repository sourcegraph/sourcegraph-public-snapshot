package squirrel

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestHover(t *testing.T) {
	java := `
class C {
	void m() {
		// not a comment line

		// comment line 1
		// comment line 2
		int x = 5;
	}
}
`

	golang := `
func main() {
	// not a comment line

	// comment line 1
	// comment line 2
	var x int
}
`

	csharp := `
namespace Foo {
    class Bar {
        static void Baz(int p) {
			// not a comment line

			// comment line 1
			// comment line 2
			var x = 5;
		}
	}
}
`

	tests := []struct {
		path     string
		contents string
		want     string
	}{
		{"test.java", java, "comment line 1\ncomment line 2\n"},
		{"test.go", golang, "comment line 1\ncomment line 2\n"},
		{"test.cs", csharp, "comment line 1\ncomment line 2\n"},
	}

	readFile := func(ctx context.Context, path types.RepoCommitPath) ([]byte, error) {
		for _, test := range tests {
			if test.path == path.Path {
				return []byte(test.contents), nil
			}
		}
		return nil, errors.Newf("path %s not found", path.Path)
	}

	squirrel := New(readFile, nil)
	defer squirrel.Close()

	for _, test := range tests {
		payload, err := squirrel.LocalCodeIntel(context.Background(), types.RepoCommitPath{Repo: "foo", Commit: "bar", Path: test.path})
		fatalIfError(t, err)

		ok := false
		for _, symbol := range payload.Symbols {
			got := symbol.Hover

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
				comments = append(comments, symbol.Hover)
			}
			t.Logf("did not find comment %q. All comments:\n", test.want)
			for _, comment := range comments {
				t.Logf("%q\n", comment)
			}
			t.FailNow()
		}
	}
}

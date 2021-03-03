package app

import (
	"context"
	"testing"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type symbolLocationArgs struct {
	vfs        map[string]string
	commitID   api.CommitID
	importPath string
	path       string
	receiver   *string
	symbol     string
}

type test struct {
	args symbolLocationArgs
	want *lsp.Location
}

func mkLocation(uri string, line, character int) *lsp.Location {
	return &lsp.Location{
		URI: "https://github.com/gorilla/mux?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#/mux.go",
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      line,
				Character: character,
			},
			End: lsp.Position{
				Line:      line,
				Character: character,
			},
		},
	}
}

func strptr(s string) *string {
	return &s
}

func TestSymbolLocation(t *testing.T) {
	vfs := map[string]string{
		"mux.go": "package mux\nconst Foo = 5\ntype Bar int\nfunc (b Bar) Quux() {}\nvar Floop = 6",
	}

	tests := []test{
		{
			args: symbolLocationArgs{
				vfs:        vfs,
				commitID:   "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
				importPath: "github.com/gorilla/mux",
				path:       "/",
				receiver:   nil,
				symbol:     "NonexistentSymbol",
			},
			want: nil,
		},
		{
			args: symbolLocationArgs{
				vfs:        vfs,
				commitID:   "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
				importPath: "github.com/gorilla/mux",
				path:       "/",
				receiver:   nil,
				symbol:     "Foo",
			},
			want: mkLocation("https://github.com/gorilla/mux?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#mux.go", 1, 6),
		},
		{
			args: symbolLocationArgs{
				vfs:        vfs,
				commitID:   "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
				importPath: "github.com/gorilla/mux",
				path:       "/",
				receiver:   nil,
				symbol:     "Bar",
			},
			want: mkLocation("https://github.com/gorilla/mux?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#mux.go", 2, 5),
		},
		{
			args: symbolLocationArgs{
				vfs:        vfs,
				commitID:   "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
				importPath: "github.com/gorilla/mux",
				path:       "/",
				receiver:   strptr("Bar"),
				symbol:     "Quux",
			},
			want: mkLocation("https://github.com/gorilla/mux?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#mux.go", 3, 13),
		},
		{
			args: symbolLocationArgs{
				vfs:        vfs,
				commitID:   "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
				importPath: "github.com/gorilla/mux",
				path:       "/",
				receiver:   nil,
				symbol:     "Floop",
			},
			want: mkLocation("https://github.com/gorilla/mux?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#mux.go", 4, 4),
		},
	}
	for i, test := range tests {
		got, _ := symbolLocation(context.Background(), mapFS(test.args.vfs), test.args.commitID, test.args.importPath, test.args.path, test.args.receiver, test.args.symbol)
		if got != test.want && (got == nil || test.want == nil || *got != *test.want) {
			t.Errorf("Test #%d:\ngot  %#v\nwant %#v", i, got, test.want)
		}
	}
}

// mapFS lets us easily instantiate a VFS with a map[string]string
// (which is less noisy than map[string][]byte in test fixtures).
func mapFS(m map[string]string) *stringMapFS {
	m2 := make(map[string][]byte, len(m))
	filenames := make([]string, 0, len(m))
	for k, v := range m {
		m2[k] = []byte(v)
		filenames = append(filenames, k)
	}
	return &stringMapFS{
		FileSystem: ctxvfs.Map(m2),
		filenames:  filenames,
	}
}

type stringMapFS struct {
	ctxvfs.FileSystem
	filenames []string
}

func (fs *stringMapFS) ListAllFiles(ctx context.Context) ([]string, error) {
	return fs.filenames, nil
}

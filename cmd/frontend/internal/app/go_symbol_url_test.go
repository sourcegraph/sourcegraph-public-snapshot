package app

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestParseGoSymbolURLPath(t *testing.T) {
	valid := map[string]*goSymbolSpec{
		"/go/github.com/gorilla/mux/-/Router/Match": {
			Pkg:      "github.com/gorilla/mux",
			Receiver: strptr("Router"),
			Symbol:   "Match",
		},

		"/go/github.com/gorilla/mux/-/Router": {
			Pkg:    "github.com/gorilla/mux",
			Symbol: "Router",
		},
	}

	invalid := map[string]string{
		"/ts/github.com/foo/bar/-/Bam": "invalid mode",
		"/go":                          "invalid symbol URL path",
		"/go/":                         "invalid symbol URL path",
		"/go/google.golang.org/api/cloudresourcemanager/v1": "invalid symbol URL path",
	}

	for path, want := range valid {
		t.Log(path)
		got, err := parseGoSymbolURLPath(path)
		if err != nil {
			t.Errorf("parseGoSymbolURLPath(%q) got error: %v", path, err)
		} else if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("parseGoSymbolURLPath(%q) mismatch (-want, +got):\n%s", path, diff)
		}
	}

	for path, errSub := range invalid {
		t.Log(path)
		got, err := parseGoSymbolURLPath(path)
		if err == nil {
			t.Errorf("parseGoSymbolURLPath(%q) expected error got: %v", path, *got)
		} else if !strings.Contains(err.Error(), errSub) {
			t.Errorf("parseGoSymbolURLPath(%q) expected error containing %q got: %v", path, errSub, err)
		}
	}
}

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
		spec := &goSymbolSpec{
			Pkg:      test.args.importPath,
			Receiver: test.args.receiver,
			Symbol:   test.args.symbol,
		}
		got, _ := symbolLocation(context.Background(), mapFS(test.args.vfs), test.args.commitID, spec, test.args.path)
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

package symbols

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/pkg/ctags"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	symbolsclient "github.com/sourcegraph/sourcegraph/pkg/symbols"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
)

func init() {
	if libSqlite3Pcre == "" {
        repositoryRoot, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			panic("can't find the libsqlite3-pcre library because LIBSQLITE3_PCRE was not set and you're not in the git repository, which is where the library is expected to be.")
		}
		if runtime.GOOS == "darwin" {
			libSqlite3Pcre = path.Join(strings.TrimSpace(string(repositoryRoot)), "libsqlite3-pcre.dylib")
		} else {
			libSqlite3Pcre = path.Join(strings.TrimSpace(string(repositoryRoot)), "libsqlite3-pcre.so")
		}
		if _, err := os.Stat(libSqlite3Pcre); os.IsNotExist(err) {
			panic(fmt.Errorf("can't find the libsqlite3-pcre library because LIBSQLITE3_PCRE was not set and %s doesn't exist at the root of the repository - try building it with `./dev/ts-script cmd/symbols/build.ts buildLibsqlite3Pcre`", libSqlite3Pcre))
		}
	}
}

func TestService(t *testing.T) {
	MustRegisterSqlite3WithPcre()

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { os.RemoveAll(tmpDir) }()

	files := map[string]string{"a.js": "var x = 1"}
	service := Service{
		FetchTar: func(ctx context.Context, repo gitserver.Repo, commit api.CommitID) (io.ReadCloser, error) {
			return createTar(files)
		},
		NewParser: func() (ctags.Parser, error) {
			return mockParser{"x", "y"}, nil
		},
		Path: tmpDir,
	}

	if err := service.Start(); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(service.Handler())
	defer server.Close()
	client := symbolsclient.Client{URL: server.URL}
	x := protocol.Symbol{Name: "x", Path: "a.js"}
	y := protocol.Symbol{Name: "y", Path: "a.js"}

	tests := map[string]struct {
		args protocol.SearchArgs
		want protocol.SearchResult
	}{
		"simple": {
			args: protocol.SearchArgs{First: 10},
			want: protocol.SearchResult{Symbols: []protocol.Symbol{x, y}},
		},
		"onematch": {
			args: protocol.SearchArgs{Query: "x", First: 10},
			want: protocol.SearchResult{Symbols: []protocol.Symbol{x}},
		},
		"nomatches": {
			args: protocol.SearchArgs{Query: "foo", First: 10},
			want: protocol.SearchResult{},
		},
		"caseinsensitiveexactmatch": {
			args: protocol.SearchArgs{Query: "^X$", First: 10},
			want: protocol.SearchResult{Symbols: []protocol.Symbol{x}},
		},
		"casesensitiveexactmatch": {
			args: protocol.SearchArgs{Query: "^x$", IsCaseSensitive: true, First: 10},
			want: protocol.SearchResult{Symbols: []protocol.Symbol{x}},
		},
		"casesensitivenoexactmatch": {
			args: protocol.SearchArgs{Query: "^X$", IsCaseSensitive: true, First: 10},
			want: protocol.SearchResult{},
		},
		"caseinsensitiveexactpathmatch": {
			args: protocol.SearchArgs{IncludePatterns: []string{"^A.js$"}, First: 10},
			want: protocol.SearchResult{Symbols: []protocol.Symbol{x, y}},
		},
		"casesensitiveexactpathmatch": {
			args: protocol.SearchArgs{IncludePatterns: []string{"^a.js$"}, IsCaseSensitive: true, First: 10},
			want: protocol.SearchResult{Symbols: []protocol.Symbol{x, y}},
		},
		"casesensitivenoexactpathmatch": {
			args: protocol.SearchArgs{IncludePatterns: []string{"^A.js$"}, IsCaseSensitive: true, First: 10},
			want: protocol.SearchResult{},
		},
		"exclude": {
			args: protocol.SearchArgs{ExcludePattern: "a.js", IsCaseSensitive: true, First: 10},
			want: protocol.SearchResult{},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			result, err := client.Search(context.Background(), test.args)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(*result, test.want) {
				t.Errorf("got %+v, want %+v", *result, test.want)
			}
		})
	}
}

func createTar(files map[string]string) (io.ReadCloser, error) {
	buf := new(bytes.Buffer)
	w := tar.NewWriter(buf)
	for name, body := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(body)),
		}
		if err := w.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(body)); err != nil {
			return nil, err
		}
	}

	err := w.Close()
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

type mockParser []string

func (m mockParser) Parse(name string, content []byte) ([]ctags.Entry, error) {
	entries := make([]ctags.Entry, len(m))
	for i, name := range m {
		entries[i] = ctags.Entry{Name: name, Path: "a.js"}
	}
	return entries, nil
}

func (mockParser) Close() {}

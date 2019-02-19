package symbols

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"database/sql"

	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/pkg/ctags"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	symbolsclient "github.com/sourcegraph/sourcegraph/pkg/symbols"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
)

func TestService(t *testing.T) {
	libSqlite3Pcre, found := os.LookupEnv("LIBSQLITE3_PCRE")
	if !found {
		rootPathOutput, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			panic(err)
		}
		libs, err := filepath.Glob(path.Join(strings.TrimSpace(string(rootPathOutput)), "libsqlite3-pcre.*"))
		if err != nil {
			panic(err)
		}
		if len(libs) == 0 {
			panic("Can't find the libsqlite3-pcre library because LIBSQLITE3_PCRE was not set and libsqlite3-pcre.* doesn't exist at the root of the repository. Try building it with `./dev/ts-script cmd/symbols/build.ts buildLibsqlite3Pcre`.")
		}
		libSqlite3Pcre = libs[0]
	}
	sql.Register("sqlite3_with_pcre", &sqlite3.SQLiteDriver{Extensions: []string{libSqlite3Pcre}})

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

	tests := map[string]struct {
		args protocol.SearchArgs
		want protocol.SearchResult
	}{
		"simple": {
			args: protocol.SearchArgs{First: 10},
			want: protocol.SearchResult{Symbols: []protocol.Symbol{{Name: "x"}, {Name: "y"}}},
		},
		"onematch": {
			args: protocol.SearchArgs{Query: "x", First: 10},
			want: protocol.SearchResult{Symbols: []protocol.Symbol{{Name: "x"}}},
		},
		"nomatches": {
			args: protocol.SearchArgs{Query: "foo", First: 10},
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
		entries[i] = ctags.Entry{Name: name}
	}
	return entries, nil
}

func (mockParser) Close() {}

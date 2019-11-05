package comby

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/store"
)

func TestMatchesUnmarshalling(t *testing.T) {
	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !exists() {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello foo")
}
`,
	}

	zipPath, cleanup, err := newZip(files)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	cases := []struct {
		args Args
		want string
	}{
		{
			args: Args{
				Input:         ZipPath(zipPath),
				MatchTemplate: "func",
				FilePatterns:  []string{".go"},
				Matcher:       ".go",
			},
			want: "func",
		},
	}

	for _, test := range cases {
		m, _ := Matches(test.args)
		if err != nil {
			t.Fatal(err)
		}
		got := m[0].Matches[0].Matched
		if got != test.want {
			t.Errorf("got %v, want %v", got, test.want)
			continue
		}
	}
}

func TestMatchesInZip(t *testing.T) {
	// If we are not on CI skip the test if comby is not installed.
	if os.Getenv("CI") == "" && !exists() {
		t.Skip("comby is not installed on the PATH. Try running 'bash <(curl -sL get.comby.dev)'.")
	}

	files := map[string]string{

		"README.md": `# Hello World

Hello world example in go`,
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello foo")
}
`,
	}

	zipPath, cleanup, err := newZip(files)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	cases := []struct {
		args Args
		want string
	}{
		{
			args: Args{
				Input:           ZipPath(zipPath),
				MatchTemplate:   "func",
				RewriteTemplate: "derp",
				FilePatterns:    []string{".go"},
				Matcher:         ".go",
			},
			want: `{"uri":"main.go","diff":"--- main.go\n+++ main.go\n@@ -2,6 +2,6 @@\n \n import \"fmt\"\n \n-func main() {\n+derp main() {\n \tfmt.Println(\"Hello foo\")\n }"}
`},
	}

	for _, test := range cases {
		b := new(bytes.Buffer)
		w := bufio.NewWriter(b)
		err := PipeTo(test.args, w)
		if err != nil {
			t.Fatal(err)
		}
		got := b.String()
		if got != test.want {
			t.Errorf("got %v, want %v", got, test.want)
			continue
		}
	}
}

func newZip(files map[string]string) (path string, cleanup func(), err error) {
	s, cleanup, err := newStore(files)
	if err != nil {
		return "", cleanup, err
	}

	ctx := context.Background()
	repo := gitserver.Repo{Name: "foo", URL: "u"}
	var commit api.CommitID = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
	path, err = s.PrepareZip(ctx, repo, commit)
	if err != nil {
		return "", cleanup, err
	}
	return path, cleanup, nil
}

func newStore(files map[string]string) (*store.Store, func(), error) {
	buf := new(bytes.Buffer)
	w := tar.NewWriter(buf)
	for name, body := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(body)),
		}
		if err := w.WriteHeader(hdr); err != nil {
			return nil, nil, err
		}
		if _, err := w.Write([]byte(body)); err != nil {
			return nil, nil, err
		}
	}
	// git-archive usually includes a pax header we should ignore.
	// use a body which matches a test case. Ensures we don't return this
	// false entry as a result.
	if err := addpaxheader(w, "Hello world\n"); err != nil {
		return nil, nil, err
	}

	err := w.Close()
	if err != nil {
		return nil, nil, err
	}
	d, err := ioutil.TempDir("", "comby_test")
	if err != nil {
		return nil, nil, err
	}
	return &store.Store{
		FetchTar: func(ctx context.Context, repo gitserver.Repo, commit api.CommitID) (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
		},
		Path: d,
	}, func() { os.RemoveAll(d) }, nil
}

func addpaxheader(w *tar.Writer, body string) error {
	hdr := &tar.Header{
		Name:       "pax_global_header",
		Typeflag:   tar.TypeXGlobalHeader,
		PAXRecords: map[string]string{"somekey": body},
	}
	return w.WriteHeader(hdr)
}

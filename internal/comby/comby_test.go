package comby

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/store"
)

func TestMatchesInZip(t *testing.T) {
	// If we are not on CI skip the test if comby is not installed.
	if _, err := exec.LookPath("comby"); os.Getenv("CI") == "" && err != nil {
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

	store, cleanup, err := newStore(files)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	cases := []struct {
		args Args
		want string
	}{
		{
			Args{
				Input:           Input{ZipPath: store.Path},
				MatchTemplate:   "func",
				RewriteTemplate: "derp",
				FilePatterns:    []string{".go"},
				Matcher:         ".go",
			}, `
{"uri":"main.go","diff":"--- main.go\n+++ main.go\n@@ -2,6 +2,6 @@\n \n import \"fmt\"\n \n-func main() {\n+derp main() {\n \tfmt.Println(\"Hello foo\")\n }"}
`},
	}

	for _, test := range cases {
		var b bytes.Buffer
		w := bufio.NewWriter(&b)
		err := Pipe(test.args, w)
		if err != nil {
			t.Fatal(err)
		}
		ifb.String()
	}

	/*
		testCases := []struct {
			name string
			want string
		}{
			{"case 1", "yes"},
		}

		for _, test := range testCases {
			t.Run(test.name, func(*testing.T) {
				got := "yes"
				if got != test.want {
					t.Errorf("failed %v, got %v, want %v", test.name, got, test.want)
				}
			})
		}
	*/
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
	d, err := ioutil.TempDir("", "search_test")
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

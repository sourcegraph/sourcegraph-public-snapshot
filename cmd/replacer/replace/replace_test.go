package replace_test

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/replacer/protocol"
	"github.com/sourcegraph/sourcegraph/cmd/replacer/replace"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/store"
)

func TestReplace(t *testing.T) {
	t.Skip("external tooling is not integrated yet.")

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

	cases := []struct {
		arg  protocol.RewriteSpecification
		want string
	}{
		{protocol.RewriteSpecification{
			MatchTemplate:   "foo",
			RewriteTemplate: "bar",
			FileExtension:   ".go",
		}, "bar"},
	}

	store, cleanup, err := newStore(files)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	ts := httptest.NewServer(&replace.Service{Store: store})
	defer ts.Close()

	for _, test := range cases {
		req := protocol.Request{
			Repo:                 "foo",
			URL:                  "u",
			Commit:               "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			RewriteSpecification: test.arg,
			FetchTimeout:         "500ms",
		}
		got, err := doReplace(ts.URL, &req)
		if err != nil {
			t.Errorf("%v failed: %s", test.arg, err)
			continue
		}

		// We have an extra newline to make expected readable
		if len(test.want) > 0 {
			test.want = test.want[1:]
		}

		if got != test.want {
			d, err := diff(test.want, got)
			if err != nil {
				t.Fatal(err)
			}
			t.Errorf("%v unexpected response:\n%s", test.arg, d)
			continue
		}

	}

}

func TestReplace_badrequest(t *testing.T) {
	cases := []protocol.Request{
		{
			Repo:   "foo",
			URL:    "u",
			Commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			// No MatchTemplate
		},
	}

	store, cleanup, err := newStore(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	ts := httptest.NewServer(&replace.Service{Store: store})
	defer ts.Close()

	for _, p := range cases {
		_, err := doReplace(ts.URL, &p)
		if err == nil {
			t.Fatalf("%v expected to fail", p)
		}
		if !strings.HasPrefix(err.Error(), "non-200 response: code=400 ") {
			t.Fatalf("%v expected to have HTTP 400 response. Got %s", p, err)
		}
	}

}

func doReplace(u string, p *protocol.Request) (string, error) {
	form := url.Values{
		"Repo":            []string{string(p.Repo)},
		"URL":             []string{string(p.URL)},
		"Commit":          []string{string(p.Commit)},
		"MatchTemplate":   []string{p.RewriteSpecification.MatchTemplate},
		"RewriteTemplate": []string{p.RewriteSpecification.RewriteTemplate},
		"FileExtension":   []string{p.RewriteSpecification.FileExtension},
	}
	resp, err := http.PostForm(u, form)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("non-200 response: code=%d body=%s", resp.StatusCode, string(body))
	}

	return string(body), err
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

func diff(b1, b2 string) (string, error) {
	f1, err := ioutil.TempFile("", "search_test")
	if err != nil {
		return "", err
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "search_test")
	if err != nil {
		return "", err
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	_, err = f1.WriteString(b1)
	if err != nil {
		return "", err
	}
	_, err = f2.WriteString(b2)
	if err != nil {
		return "", err
	}

	data, err := exec.Command("diff", "-u", "--label=want", f1.Name(), "--label=got", f2.Name()).CombinedOutput()
	if len(data) > 0 {
		err = nil
	}
	return string(data), err
}

func addpaxheader(w *tar.Writer, body string) error {
	hdr := &tar.Header{
		Name:       "pax_global_header",
		Typeflag:   tar.TypeXGlobalHeader,
		PAXRecords: map[string]string{"somekey": body},
	}
	return w.WriteHeader(hdr)
}

package search_test

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/searcher/search"
)

func TestSearch(t *testing.T) {
	files := map[string]string{
		"README.md": `# Hello World

Hello world example in go`,
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello world")
}
`,
	}

	cases := map[search.Params]string{
		search.Params{Pattern: "foo"}: "",

		search.Params{Pattern: "World", IsCaseSensitive: true}: `
README.md:1:# Hello World
`,

		search.Params{Pattern: "world", IsCaseSensitive: true}: `
README.md:3:Hello world example in go
main.go:6:	fmt.Println("Hello world")
`,

		search.Params{Pattern: "world"}: `
README.md:1:# Hello World
README.md:3:Hello world example in go
main.go:6:	fmt.Println("Hello world")
`,

		search.Params{Pattern: "func.*main"}: "",

		search.Params{Pattern: "func.*main", IsRegExp: true}: `
main.go:5:func main() {
`,

		search.Params{Pattern: "mai", IsWordMatch: true}: "",

		search.Params{Pattern: "main", IsWordMatch: true}: `
main.go:1:package main
main.go:5:func main() {
`,

		// Ensure we handle CaseInsensitive regexp searches with
		// special uppercase chars in pattern.
		search.Params{Pattern: `printL\B`, IsRegExp: true}: `
main.go:6:	fmt.Println("Hello world")
`,

		search.Params{Pattern: "world", ExcludePattern: "README.md"}: `
main.go:6:	fmt.Println("Hello world")
`,
		search.Params{Pattern: "world", IncludePattern: "*.md"}: `
README.md:1:# Hello World
README.md:3:Hello world example in go
`,

		search.Params{Pattern: "doesnotmatch"}: "",
	}

	store, cleanup, err := newStore(files)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	ts := httptest.NewServer(&search.Service{Store: store})
	defer ts.Close()

	for p, want := range cases {
		p2 := p
		p2.Repo = "foo"
		p2.Commit = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"
		m, err := doSearch(ts.URL, &p2)
		if err != nil {
			t.Errorf("%v failed: %s", p, err)
			continue
		}
		sort.Sort(sortByPath(m))
		got := toString(m)
		err = sanityCheckSorted(m)
		if err != nil {
			t.Errorf("%v malformed response: %s\n%s", p, err, got)
		}
		// We have an extra newline to make expected readable
		if len(want) > 0 {
			want = want[1:]
		}
		if got != want {
			d, err := diff(want, got)
			if err != nil {
				t.Fatal(err)
			}
			t.Errorf("%v unexpected response:\n%s", p, d)
		}
	}
}

func TestSearch_badrequest(t *testing.T) {
	cases := []search.Params{
		// Empty pattern
		{
			Repo:   "foo",
			Commit: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
		},

		// Bad regexp
		{
			Repo:     "foo",
			Commit:   "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			Pattern:  `\F`,
			IsRegExp: true,
		},

		// No repo
		{
			Commit:  "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			Pattern: "test",
		},

		// No commit
		{
			Repo:    "foo",
			Pattern: "test",
		},

		// Non-absolute commit
		{
			Repo:    "foo",
			Commit:  "HEAD",
			Pattern: "test",
		},

		// Bad include glob
		{
			Repo:           "foo",
			Commit:         "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			Pattern:        "test",
			IncludePattern: "[c-a]",
		},

		// Bad exclude glob
		{
			Repo:           "foo",
			Commit:         "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			Pattern:        "test",
			ExcludePattern: "[c-a]",
		},
	}

	store, cleanup, err := newStore(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	ts := httptest.NewServer(&search.Service{Store: store})
	defer ts.Close()

	for _, p := range cases {
		_, err := doSearch(ts.URL, &p)
		if err == nil {
			t.Fatalf("%v expected to fail", p)
		}
		if !strings.HasPrefix(err.Error(), "non-200 response: code=400 ") {
			t.Fatalf("%v expected to have HTTP 400 response. Got %s", p, err)
		}
	}
}

func doSearch(u string, p *search.Params) ([]search.FileMatch, error) {
	form := url.Values{
		"Repo":           []string{p.Repo},
		"Commit":         []string{p.Commit},
		"Pattern":        []string{p.Pattern},
		"IncludePattern": []string{p.IncludePattern},
		"ExcludePattern": []string{p.ExcludePattern},
	}
	if p.IsRegExp {
		form.Set("IsRegExp", "true")
	}
	if p.IsWordMatch {
		form.Set("IsWordMatch", "true")
	}
	if p.IsCaseSensitive {
		form.Set("IsCaseSensitive", "true")
	}
	resp, err := http.PostForm(u, form)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response: code=%d body=%s", resp.StatusCode, string(body))
	}

	var r search.Response
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	return r.Matches, err
}

func newStore(files map[string]string) (*search.Store, func(), error) {
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
	err := w.Close()
	if err != nil {
		return nil, nil, err
	}
	d, err := ioutil.TempDir("", "search_test")
	if err != nil {
		return nil, nil, err
	}
	return &search.Store{
		FetchTar: func(ctx context.Context, repo, commit string) (io.ReadCloser, error) {
			return ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
		},
		Path: d,
	}, func() { os.RemoveAll(d) }, nil
}

func toString(m []search.FileMatch) string {
	buf := new(bytes.Buffer)
	for _, f := range m {
		for _, l := range f.LineMatches {
			buf.WriteString(f.Path)
			buf.WriteByte(':')
			buf.WriteString(strconv.Itoa(l.LineNumber + 1))
			buf.WriteByte(':')
			buf.WriteString(l.Preview)
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

func sanityCheckSorted(m []search.FileMatch) error {
	if !sort.IsSorted(sortByPath(m)) {
		return errors.New("unsorted file matches, please sortByPath")
	}
	for i := range m {
		if i > 0 && m[i].Path == m[i-1].Path {
			return fmt.Errorf("duplicate FileMatch on %s", m[i].Path)
		}
		lm := m[i].LineMatches
		if !sort.IsSorted(sortByLineNumber(lm)) {
			return fmt.Errorf("unsorted LineMatches for %s", m[i].Path)
		}
		for j := range lm {
			if j > 0 && lm[j].LineNumber == lm[j-1].LineNumber {
				return fmt.Errorf("duplicate LineNumber on %s:%d", m[i].Path, lm[j].LineNumber)
			}
		}
	}
	return nil
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

	f1.WriteString(b1)
	f2.WriteString(b2)

	data, err := exec.Command("diff", "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) > 0 {
		err = nil
	}
	return string(data), err
}

type sortByPath []search.FileMatch

func (m sortByPath) Len() int           { return len(m) }
func (m sortByPath) Less(i, j int) bool { return m[i].Path < m[j].Path }
func (m sortByPath) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

type sortByLineNumber []search.LineMatch

func (m sortByLineNumber) Len() int           { return len(m) }
func (m sortByLineNumber) Less(i, j int) bool { return m[i].LineNumber < m[j].LineNumber }
func (m sortByLineNumber) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

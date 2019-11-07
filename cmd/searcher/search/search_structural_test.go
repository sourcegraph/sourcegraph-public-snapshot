package search

import (
	"context"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/store"
)

// Tests that includePatterns works. includePatterns serve a similar role in
// structural search, but are interpretted _differently_. includePatterns cannot
// be a regex expression (as in traditional search), but instead currently
// expect to be a list of patterns that assume a file path to be matched.
func TestIncludePatterns(t *testing.T) {
	input := map[string]string{
		"/a/b/c":         "",
		"/a/b/c/foo.go":  "",
		"c/foo.go":       "",
		"bar.go":         "",
		"/x/y/z/bar.go":  "",
		"/a/b/c/nope.go": "",
		"nope.go":        "",
	}

	want := []string{
		"/a/b/c/foo.go",
		"/x/y/z/bar.go",
		"bar.go",
	}

	zipData, err := createZip(input)
	if err != nil {
		t.Fatal(err)
	}
	zf, cleanup, err := MockZipFileOnDisk(zipData)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	p := &protocol.PatternInfo{
		Pattern:         "",
		IncludePatterns: []string{"a/b/c/foo.go", "bar.go"},
	}
	if err != nil {
		t.Fatal(err)
	}
	fileMatches, _, err := structuralSearch(context.Background(), zf, p.Pattern, p.IncludePatterns, "foo")
	if err != nil {
		t.Fatal(err)
	}

	got := make([]string, len(fileMatches))
	for i, fm := range fileMatches {
		got[i] = fm.Path
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got file matches %v, want %v", got, want)
	}
}

func MockZipFileOnDisk(data []byte) (string, func(), error) {
	z, err := store.MockZipFile(data)
	if err != nil {
		return "", nil, err
	}
	d, err := ioutil.TempDir("", "search_test")
	if err != nil {
		return "", nil, err
	}
	f, err := ioutil.TempFile(d, "search_zip")
	if err != nil {
		return "", nil, err
	}
	_, err = f.Write(z.Data)
	if err != nil {
		return "", nil, err
	}
	return f.Name(), func() { os.RemoveAll(d) }, nil
}

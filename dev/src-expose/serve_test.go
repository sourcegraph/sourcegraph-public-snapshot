package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

const testAddress = "test.local:3939"

func TestReposHandler(t *testing.T) {
	cases := []struct {
		name  string
		repos []string
	}{{
		name: "empty",
	}, {
		name:  "simple",
		repos: []string{"project1", "project2"},
	}, {
		name:  "nested",
		repos: []string{"project1", "project1/subproject", "project2", "dir/project3"},
	}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := gitInitRepos(t, tc.repos...)

			h, err := reposHandler(testLogger(t), testAddress, root)
			if err != nil {
				t.Fatal(err)
			}

			testReposHandler(t, h, tc.repos...)
		})
	}
}

func testReposHandler(t *testing.T, h http.Handler, names ...string) {
	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)

	get := func(path string) string {
		res, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		b, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if testing.Verbose() {
			t.Logf("GET %s:\n%s", path, b)
		}
		return string(b)
	}

	// Check we have some known strings on the index page
	index := get("/")
	for _, sub := range []string{"http://" + testAddress, "/v1/list-repos", "/repos/"} {
		if !strings.Contains(index, sub) {
			t.Errorf("index page does not contain substring %q", sub)
		}
	}

	// repos page will list the top-level dirs
	list := get("/repos/")
	for _, name := range names {
		if !strings.Contains(name, "/") && !strings.Contains(list, name) {
			t.Errorf("repos page does not contain substring %q", name)
		}
	}

	// check our API response
	type Repo struct {
		Name string
		URI  string
	}
	type Response struct{ Items []Repo }
	var want, got Response
	for _, name := range names {
		want.Items = append(want.Items, Repo{Name: name, URI: path.Join("/repos", name)})
	}
	if err := json.Unmarshal([]byte(get("/v1/list-repos")), &got); err != nil {
		t.Fatal(err)
	}
	opts := []cmp.Option{
		cmpopts.EquateEmpty(),
		cmpopts.SortSlices(func(a, b Repo) bool { return a.Name < b.Name }),
	}
	if !cmp.Equal(want, got, opts...) {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts...))
	}
}

func gitInitRepos(t *testing.T, names ...string) string {
	root, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(root) })
	root = filepath.Join(root, "repos-root")

	for _, name := range names {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
		p = filepath.Join(p, ".git")
		if err := exec.Command("git", "init", "--bare", p).Run(); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

func testLogger(t *testing.T) *log.Logger {
	return log.New(testWriter{t}, "testLogger ", log.LstdFlags)
}

type testWriter struct {
	*testing.T
}

func (tw testWriter) Write(p []byte) (n int, err error) {
	tw.T.Log(string(p))
	return len(p), nil
}

package servegit

import (
	"encoding/json"
	"io"
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

var discardLogger = log.New(io.Discard, "", log.LstdFlags)

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
		repos: []string{"project1", "project2", "dir/project3", "dir/project4.bare"},
	}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := gitInitRepos(t, tc.repos...)

			h := (&Serve{
				Info:  testLogger(t),
				Debug: discardLogger,
				Addr:  testAddress,
				Root:  root,
			}).handler()

			var want []Repo
			for _, name := range tc.repos {
				isBare := strings.HasSuffix(name, ".bare")
				uri := path.Join("/repos", name)
				clonePath := uri
				if !isBare {
					clonePath += "/.git"
				}
				want = append(want, Repo{Name: name, URI: uri, ClonePath: clonePath})

			}
			testReposHandler(t, h, want)
		})
	}
}

func testReposHandler(t *testing.T, h http.Handler, repos []Repo) {
	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)

	get := func(path string) string {
		res, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(res.Body)
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
	for _, repo := range repos {
		if path.Dir(repo.URI) != "/repos" {
			continue
		}
		if !strings.Contains(repo.Name, "/") && !strings.Contains(list, repo.Name) {
			t.Errorf("repos page does not contain substring %q", repo.Name)
		}
	}

	// check our API response
	type Response struct{ Items []Repo }
	var want, got Response
	want.Items = repos
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

func gitInitBare(t *testing.T, path string) {
	if err := exec.Command("git", "init", "--bare", path).Run(); err != nil {
		t.Fatal(err)
	}
}

func gitInit(t *testing.T, path string) {
	cmd := exec.Command("git", "init")
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
}

func gitInitRepos(t *testing.T, names ...string) string {
	root := t.TempDir()
	root = filepath.Join(root, "repos-root")

	for _, name := range names {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}

		if strings.HasSuffix(p, ".bare") {
			gitInitBare(t, p)
		} else {
			gitInit(t, p)
		}
	}

	return root
}

func TestIgnoreGitSubmodules(t *testing.T) {
	root := t.TempDir()

	if err := os.MkdirAll(filepath.Join(root, "dir"), os.ModePerm); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(root, "dir", ".git"), []byte("ignore me please"), os.ModePerm); err != nil {
		t.Fatal(err)
	}

	repos, err := (&Serve{
		Info:  testLogger(t),
		Debug: discardLogger,
		Root:  root,
	}).Repos()
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 0 {
		t.Fatalf("expected no repos, got %v", repos)
	}
}

func TestIsBareRepo(t *testing.T) {
	dir := t.TempDir()

	gitInitBare(t, dir)

	if !isBareRepo(dir) {
		t.Errorf("Path %s it not a bare repository", dir)
	}
}

func TestEmptyDirIsNotBareRepo(t *testing.T) {
	dir := t.TempDir()

	if isBareRepo(dir) {
		t.Errorf("Path %s it falsey detected as a bare repository", dir)
	}
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

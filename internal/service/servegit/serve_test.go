package servegit

import (
	"bytes"
	"encoding/json"
	"io"
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
	"github.com/sourcegraph/log/logtest"
)

const testAddress = "test.local:3939"

func testRepoWithPaths(fixedEndpoint string, root string, pathWithName string) Repo {
	var sb strings.Builder
	delimiter := "/"

	for _, str := range []string{fixedEndpoint, root, pathWithName} {
		sb.WriteString(delimiter)
		sb.WriteString(strings.Trim(str, delimiter))
	}

	uri := sb.String()

	clonePath := uri

	if !strings.HasSuffix(pathWithName, ".bare") {
		sb.WriteString(delimiter)
		sb.WriteString(".git")
		clonePath = sb.String()
	}

	return Repo{
		Name:        pathWithName,
		URI:         uri,
		ClonePath:   clonePath,
		AbsFilePath: filepath.Join(root, filepath.FromSlash(pathWithName)),
	}
}

func TestReposHandler(t *testing.T) {
	cases := []struct {
		name  string
		root  string
		repos []string
		want  []Repo
	}{{
		name: "empty",
	}, {
		name:  "simple",
		repos: []string{"project1", "project2"},
	}, {
		name:  "nested",
		repos: []string{"project1", "project2", "dir/project3", "dir/project4.bare"},
	}, {
		name:  "root-is-repo",
		root:  "parent",
		repos: []string{"parent"},
	}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			root := gitInitRepos(t, tc.repos...)
			if tc.repos != nil {
				root, err = filepath.EvalSymlinks(root)
				if err != nil {
					t.Fatalf("Error returned from filepath.EvalSymlinks(): %v", err)
				}
			}

			var want []Repo
			for _, path := range tc.repos {
				want = append(want, testRepoWithPaths("repos", root, path))
			}

			h := (&Serve{
				Logger: logtest.Scoped(t),
				ServeConfig: ServeConfig{
					Addr: testAddress,
				},
			}).handler()

			testReposHandler(t, h, want, root)
		})
	}
}

func TestReposHandler_EmptyResults(t *testing.T) {
	cases := []struct {
		name  string
		root  string
		repos []string
		want  []Repo
	}{{
		name:  "empty path",
		root:  "",
		repos: []string{"repo"},
	}, {
		name:  "whitespace path",
		root:  "  ",
		repos: []string{"repo"},
	}, {
		name:  "padded separator path",
		root:  " / ",
		repos: []string{"repo"},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := gitInitRepos(t, tc.repos...)
			depth := len(strings.Split(root, "/"))
			h := (&Serve{
				Logger: logtest.Scoped(t),
				ServeConfig: ServeConfig{
					Addr:     testAddress,
					MaxDepth: depth + 1,
				},
			}).handler()
			testReposHandler(t, h, tc.want, tc.root)
		})
	}

}

func testReposHandler(t *testing.T, h http.Handler, repos []Repo, root string) {
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
		return string(b)
	}

	post := func(path string, body []byte) string {
		res, err := http.Post(ts.URL+path, "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	// Check we have some known strings on the index page
	index := get("/")
	for _, sub := range []string{"http://" + testAddress, "/v1/list-repos-for-path", "/repos/"} {
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
	reqBody, err := json.Marshal(ListReposRequest{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(post("/v1/list-repos-for-path", reqBody)), &got); err != nil {
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
		Logger: logtest.Scoped(t),
	}).Repos(root)
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

func TestConvertGitCloneURLToCodebaseName(t *testing.T) {
	testCases := []struct {
		name     string
		cloneURL string
		expected string
	}{
		{
			name:     "GitHub SSH URL",
			cloneURL: "git@github.com:sourcegraph/sourcegraph.git",
			expected: "github.com/sourcegraph/sourcegraph",
		},
		{
			name:     "GitHub SSH URL without .git",
			cloneURL: "git@github.com:sourcegraph/sourcegraph",
			expected: "github.com/sourcegraph/sourcegraph",
		},
		{
			name:     "GitHub HTTPS URL",
			cloneURL: "https://github.com/sourcegraph/sourcegraph",
			expected: "github.com/sourcegraph/sourcegraph",
		},
		{
			name:     "Bitbucket SSH URL",
			cloneURL: "git@bitbucket.sgdev.org:sourcegraph/sourcegraph.git",
			expected: "bitbucket.sgdev.org/sourcegraph/sourcegraph",
		},
		{
			name:     "GitLab SSH URL",
			cloneURL: "git@gitlab.com:sourcegraph/sourcegraph.git",
			expected: "gitlab.com/sourcegraph/sourcegraph",
		},
		{
			name:     "GitLab HTTPS URL",
			cloneURL: "https://gitlab.com/sourcegraph/sourcegraph.git",
			expected: "gitlab.com/sourcegraph/sourcegraph",
		},
		{
			name:     "GitHub SSH URL",
			cloneURL: "git@github.com:sourcegraph/sourcegraph.git",
			expected: "github.com/sourcegraph/sourcegraph",
		},
		{
			name:     "SSH Alias URL",
			cloneURL: "github:sourcegraph/sourcegraph",
			expected: "github.com/sourcegraph/sourcegraph",
		},
		{
			name:     "GitHub HTTP URL",
			cloneURL: "http://github.com/sourcegraph/sourcegraph",
			expected: "github.com/sourcegraph/sourcegraph",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := convertGitCloneURLToCodebaseName(testCase.cloneURL)
			if actual != testCase.expected {
				t.Errorf("Expected %s but got %s", testCase.expected, actual)
			}
		})
	}
}

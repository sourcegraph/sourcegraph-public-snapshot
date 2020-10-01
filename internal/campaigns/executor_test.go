package campaigns

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

func TestExecutor_Integration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Test doesn't work on Windows because dummydocker is written in bash")
	}

	addToPath(t, "testdata/dummydocker")

	repo := &graphql.Repository{
		Name: "github.com/sourcegraph/src-cli",
		DefaultBranch: &graphql.Branch{
			Name:   "main",
			Target: struct{ OID string }{OID: "d34db33f"},
		},
	}

	filesInRepo := map[string]string{
		"README.md": "# Welcome to the README\n",
		"main.go":   "package main\n\nfunc main() {\n\tfmt.Println(     \"Hello World\")\n}\n",
	}

	steps := []Step{
		{Run: `echo -e "foobar\n" >> README.md`, Container: "alpine:13"},
		{Run: `go fmt main.go`, Container: "doesntmatter:13"},
	}

	h := newZipHandler(t, repo.Name, repo.DefaultBranch.Name, filesInRepo)
	ts := httptest.NewServer(h)
	defer ts.Close()

	var clientBuffer bytes.Buffer
	client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

	testTempDir, err := ioutil.TempDir("", "executor-integration-test-*")
	if err != nil {
		t.Fatal(err)
	}

	creator := &WorkspaceCreator{dir: testTempDir, client: client}
	opts := ExecutorOpts{
		Cache:       &ExecutionNoOpCache{},
		Creator:     creator,
		TempDir:     testTempDir,
		Parallelism: runtime.GOMAXPROCS(0),
		Timeout:     5 * time.Second,
	}

	called := false
	updateFn := func(task *Task, ts TaskStatus) { called = true }

	executor := newExecutor(opts, client, updateFn)

	template := &ChangesetTemplate{}
	executor.AddTask(repo, steps, template)

	executor.Start(context.Background())
	specs, err := executor.Wait()
	if err != nil {
		t.Fatal(err)
	}

	if !called {
		t.Fatalf("update was not called")
	}

	if have, want := len(specs), 1; have != want {
		t.Fatalf("wrong number of specs. want=%d, have=%d", want, have)
	}

	if have, want := len(specs[0].Commits), 1; have != want {
		t.Fatalf("wrong number of commits. want=%d, have=%d", want, have)
	}

	fileDiffs, err := diff.ParseMultiFileDiff([]byte(specs[0].Commits[0].Diff))
	if err != nil {
		t.Fatalf("failed to parse diff: %s", err)
	}

	diffsByName := map[string]*diff.FileDiff{}
	for _, fd := range fileDiffs {
		diffsByName[fd.OrigName] = fd
	}

	if have, want := len(diffsByName), 2; have != want {
		t.Fatalf("wrong number of diffsByName. want=%d, have=%d", want, have)
	}

	if _, ok := diffsByName["main.go"]; !ok {
		t.Errorf("main.go was not changed")
	}
	if _, ok := diffsByName["README.md"]; !ok {
		t.Errorf("README.md was not changed")
	}
}

func addToPath(t *testing.T, relPath string) {
	t.Helper()

	dummyDockerPath, err := filepath.Abs("testdata/dummydocker")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("PATH", fmt.Sprintf("%s%c%s", dummyDockerPath, os.PathListSeparator, os.Getenv("PATH")))
}

func newZipHandler(t *testing.T, repo, branch string, files map[string]string) http.HandlerFunc {
	wantPath := fmt.Sprintf("/%s@%s/-/raw", repo, branch)

	downloadName := filepath.Base(repo)
	mediaType := mime.FormatMediaType("Attachment", map[string]string{
		"filename": downloadName,
	})

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != wantPath {
			t.Errorf("request has wrong path. want=%q, have=%q", wantPath, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", mediaType)

		zipWriter := zip.NewWriter(w)

		for name, body := range files {
			f, err := zipWriter.Create(name)
			if err != nil {
				log.Fatal(err)
			}
			if _, err := f.Write([]byte(body)); err != nil {
				t.Errorf("failed to write body for %s to zip: %s", name, err)
			}
		}

		if err := zipWriter.Close(); err != nil {
			t.Fatalf("closing zipWriter failed: %s", err)
		}
	}
}

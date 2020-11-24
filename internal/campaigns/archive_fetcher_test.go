package campaigns

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

func TestWorkspaceCreator_Create(t *testing.T) {
	workspaceTmpDir := func(t *testing.T) string {
		testTempDir, err := ioutil.TempDir("", "executor-integration-test-*")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Remove(testTempDir) })

		return testTempDir
	}

	repo := &graphql.Repository{
		ID:            "src-cli",
		Name:          "github.com/sourcegraph/src-cli",
		DefaultBranch: &graphql.Branch{Name: "main", Target: graphql.Target{OID: "d34db33f"}},
	}

	archive := mockRepoArchive{
		repo: repo,
		files: map[string]string{
			"README.md": "# Welcome to the README\n",
		},
	}

	t.Run("success", func(t *testing.T) {
		requestsReceived := 0
		callback := func(_ http.ResponseWriter, _ *http.Request) {
			requestsReceived += 1
		}

		ts := httptest.NewServer(newZipArchivesMux(t, callback, archive))
		defer ts.Close()

		var clientBuffer bytes.Buffer
		client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

		testTempDir := workspaceTmpDir(t)

		creator := &WorkspaceCreator{dir: testTempDir, client: client}
		workspace, err := creator.Create(context.Background(), repo)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		wantZipFile := "github.com-sourcegraph-src-cli-d34db33f.zip"
		ok, err := dirContains(creator.dir, wantZipFile)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("temp dir doesnt contain zip file")
		}

		haveUnzippedFiles, err := readWorkspaceFiles(workspace)
		if err != nil {
			t.Fatalf("error walking workspace: %s", err)
		}

		if !cmp.Equal(archive.files, haveUnzippedFiles) {
			t.Fatalf("wrong files in workspace:\n%s", cmp.Diff(archive.files, haveUnzippedFiles))
		}

		// Create it a second time and make sure that the server wasn't called
		_, err = creator.Create(context.Background(), repo)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if requestsReceived != 1 {
			t.Fatalf("wrong number of requests received: %d", requestsReceived)
		}

		// Third time, but this time with cleanup, _after_ unzipping
		creator.deleteZips = true
		_, err = creator.Create(context.Background(), repo)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if requestsReceived != 1 {
			t.Fatalf("wrong number of requests received: %d", requestsReceived)
		}

		ok, err = dirContains(creator.dir, wantZipFile)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatalf("temp dir contains zip file but should not")
		}
	})

	t.Run("canceled", func(t *testing.T) {
		// We create a context that is canceled after the server sent down the
		// first file to simulate a slow download that's aborted by the user hitting Ctrl-C.

		firstFileWritten := make(chan struct{})
		callback := func(w http.ResponseWriter, r *http.Request) {
			// We flush the headers and the first file
			w.(http.Flusher).Flush()

			// Wait a bit for the client to start writing the file
			time.Sleep(50 * time.Millisecond)

			// Cancel the context to simulate the Ctrl-C
			firstFileWritten <- struct{}{}

			<-r.Context().Done()
		}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-firstFileWritten
			cancel()
		}()

		ts := httptest.NewServer(newZipArchivesMux(t, callback, archive))
		defer ts.Close()

		var clientBuffer bytes.Buffer
		client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

		testTempDir := workspaceTmpDir(t)

		creator := &WorkspaceCreator{dir: testTempDir, client: client}

		_, err := creator.Create(ctx, repo)
		if err == nil {
			t.Fatalf("error is nil")
		}

		zipFile := "github.com-sourcegraph-src-cli-d34db33f.zip"
		ok, err := dirContains(creator.dir, zipFile)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatalf("zip file in temp dir was not cleaned up")
		}
	})

	t.Run("non-default branch", func(t *testing.T) {
		otherBranchOID := "f00b4r"
		repo := &graphql.Repository{
			ID:            "src-cli-with-non-main-branch",
			Name:          "github.com/sourcegraph/src-cli",
			DefaultBranch: &graphql.Branch{Name: "main", Target: graphql.Target{OID: "d34db33f"}},

			Commit: graphql.Target{OID: otherBranchOID},
			Branch: graphql.Branch{Name: "other-branch", Target: graphql.Target{OID: otherBranchOID}},
		}

		archive := mockRepoArchive{repo: repo, files: map[string]string{}}

		ts := httptest.NewServer(newZipArchivesMux(t, nil, archive))
		defer ts.Close()

		var clientBuffer bytes.Buffer
		client := api.NewClient(api.ClientOpts{Endpoint: ts.URL, Out: &clientBuffer})

		testTempDir := workspaceTmpDir(t)

		creator := &WorkspaceCreator{dir: testTempDir, client: client}

		_, err := creator.Create(context.Background(), repo)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		wantZipFile := "github.com-sourcegraph-src-cli-" + otherBranchOID + ".zip"
		ok, err := dirContains(creator.dir, wantZipFile)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatalf("temp dir doesnt contain zip file")
		}
	})
}

func TestMkdirAll(t *testing.T) {
	// TestEnsureAll does most of the heavy lifting here; we're just testing the
	// MkdirAll scenarios here around whether the directory exists.

	// Create a shared workspace.
	base := mustCreateWorkspace(t)
	defer os.RemoveAll(base)

	t.Run("directory exists", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Join(base, "exist"), 0755); err != nil {
			t.Fatal(err)
		}

		if err := mkdirAll(base, "exist", 0750); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if err := mustHavePerm(t, filepath.Join(base, "exist"), 0750); err != nil {
			t.Error(err)
		}

		if !isDir(t, filepath.Join(base, "exist")) {
			t.Error("not a directory")
		}
	})

	t.Run("directory does not exist", func(t *testing.T) {
		if err := mkdirAll(base, "new", 0750); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if err := mustHavePerm(t, filepath.Join(base, "new"), 0750); err != nil {
			t.Error(err)
		}

		if !isDir(t, filepath.Join(base, "new")) {
			t.Error("not a directory")
		}
	})

	t.Run("directory exists, but is not a directory", func(t *testing.T) {
		f, err := os.Create(filepath.Join(base, "file"))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()

		err = mkdirAll(base, "file", 0750)
		if _, ok := err.(errPathExistsAsFile); !ok {
			t.Errorf("unexpected error of type %T: %v", err, err)
		}
	})
}

func TestEnsureAll(t *testing.T) {
	// Create a workspace.
	base := mustCreateWorkspace(t)
	defer os.RemoveAll(base)

	// Create three nested directories with 0700 permissions. We'll use Chmod
	// explicitly to avoid any umask issues.
	if err := os.MkdirAll(filepath.Join(base, "a", "b", "c"), 0700); err != nil {
		t.Fatal(err)
	}
	dirs := []string{
		filepath.Join(base, "a"),
		filepath.Join(base, "a", "b"),
		filepath.Join(base, "a", "b", "c"),
	}
	for _, dir := range dirs {
		if err := os.Chmod(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}

	// Now we'll set them to 0750 and see what happens.
	if err := ensureAll(base, "a/b/c", 0750); err != nil {
		t.Errorf("unexpected non-nil error: %v", err)
	}
	for _, dir := range dirs {
		if err := mustHavePerm(t, dir, 0750); err != nil {
			t.Error(err)
		}
	}
	if err := mustHavePerm(t, base, 0700); err != nil {
		t.Error(err)
	}

	// Finally, let's ensure we get an error when we try to ensure a directory
	// that doesn't exist.
	if err := ensureAll(base, "d", 0750); err == nil {
		t.Errorf("unexpected nil error")
	}
}

func mustCreateWorkspace(t *testing.T) string {
	base, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}

	// We'll explicitly set the base workspace to 0700 so we have a known
	// environment for testing.
	if err := os.Chmod(base, 0700); err != nil {
		t.Fatal(err)
	}

	return base
}

func mustGetPerm(t *testing.T, file string) os.FileMode {
	t.Helper()

	st, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}

	// We really only need the lower bits here.
	return st.Mode() & 0777
}

func isDir(t *testing.T, path string) bool {
	t.Helper()

	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	return st.IsDir()
}

func readWorkspaceFiles(workspace string) (map[string]string, error) {
	files := map[string]string{}

	err := filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(workspace, path)
		if err != nil {
			return err
		}

		files[rel] = string(content)
		return nil
	})

	return files, err
}

func dirContains(dir, filename string) (bool, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, f := range files {
		if f.Name() == filename {
			return true, nil
		}
	}

	return false, nil
}

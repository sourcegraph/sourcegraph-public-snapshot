package campaigns

import (
	"archive/zip"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

func TestDockerBindWorkspaceCreator_Create(t *testing.T) {
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

	// Create a zip file for all the other tests to use.
	f, err := ioutil.TempFile(workspaceTmpDir(t), "repo-zip-*")
	if err != nil {
		t.Fatal(err)
	}
	archivePath := f.Name()
	t.Cleanup(func() { os.Remove(archivePath) })

	zw := zip.NewWriter(f)
	for name, body := range archive.files {
		f, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	zw.Close()
	f.Close()

	t.Run("success", func(t *testing.T) {
		testTempDir := workspaceTmpDir(t)

		creator := &dockerBindWorkspaceCreator{dir: testTempDir}
		workspace, err := creator.Create(context.Background(), repo, archivePath)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		haveUnzippedFiles, err := readWorkspaceFiles(workspace)
		if err != nil {
			t.Fatalf("error walking workspace: %s", err)
		}

		if !cmp.Equal(archive.files, haveUnzippedFiles) {
			t.Fatalf("wrong files in workspace:\n%s", cmp.Diff(archive.files, haveUnzippedFiles))
		}
	})

	t.Run("failure", func(t *testing.T) {
		testTempDir := workspaceTmpDir(t)

		// Create an empty file (which is therefore a bad zip file).
		badZip, err := ioutil.TempFile(testTempDir, "bad-zip-*")
		if err != nil {
			t.Fatal(err)
		}
		badZipFile := badZip.Name()
		t.Cleanup(func() { os.Remove(badZipFile) })
		badZip.Close()

		creator := &dockerBindWorkspaceCreator{dir: testTempDir}
		if _, err := creator.Create(context.Background(), repo, badZipFile); err == nil {
			t.Error("unexpected nil error")
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

func readWorkspaceFiles(workspace Workspace) (map[string]string, error) {
	files := map[string]string{}
	wdir := workspace.WorkDir()
	err := filepath.Walk(*wdir, func(path string, info os.FileInfo, err error) error {
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

		rel, err := filepath.Rel(*wdir, path)
		if err != nil {
			return err
		}

		if strings.HasPrefix(rel, ".git") {
			return nil
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

package campaigns

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

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

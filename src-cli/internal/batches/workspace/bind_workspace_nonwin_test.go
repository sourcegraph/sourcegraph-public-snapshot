//go:build !windows
// +build !windows

package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestMkdirAllStatError(t *testing.T) {
	// This test can't be trivially reproduced on Windows, so we just won't run
	// it there.

	// Create a shared workspace.
	base := mustCreateWorkspace(t)

	// We'll create a directory and a file within it, remove the execute bit on
	// the directory, and then stat() the file to cause a failure.
	if err := os.MkdirAll(filepath.Join(base, "locked"), 0700); err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(filepath.Join(base, "locked", "file"))
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	if err := os.Chmod(filepath.Join(base, "locked"), 0600); err != nil {
		t.Fatal(err)
	}
	// Add the execute bit back to the directory so "base" can be cleaned up
	t.Cleanup(func() {
		if err := os.Chmod(filepath.Join(base, "locked"), 0700); err != nil {
			t.Fatal(err)
		}
	})

	err = mkdirAll(base, "locked/file", 0750)
	if err == nil {
		t.Errorf("unexpected nil error")
	} else if _, ok := err.(errPathExistsAsFile); ok {
		t.Errorf("unexpected error of type %T: %v", err, err)
	}
}

func mustHavePerm(t *testing.T, path string, want os.FileMode) error {
	t.Helper()

	if have := mustGetPerm(t, path); have != want {
		return errors.Errorf("unexpected permissions: have=%o want=%o", have, want)
	}
	return nil
}

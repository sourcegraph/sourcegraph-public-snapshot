package gitserver

import (
	"os"
	"testing"
)

// CreateRepoDir creates a repo directory for testing purposes.
// This includes creating a tmp dir and deleting it after test finishes running.
func CreateRepoDir(t *testing.T) string {
	return CreateRepoDirWithName(t, "")
}

// CreateRepoDirWithName creates a repo directory with a given name for testing purposes.
// This includes creating a tmp dir and deleting it after test finishes running.
func CreateRepoDirWithName(t *testing.T, name string) string {
	t.Helper()
	if name == "" {
		name = t.Name()
	}
	root, err := os.MkdirTemp("", name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll(root)
	})
	return root
}

package codeintel

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestInferRepo(t *testing.T) {
	cur, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cur)

	tempDir := t.TempDir()
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = runGitCommand("init")
	if err != nil {
		t.Fatal(err)
	}

	want := "github.com/a/b"

	_, err = runGitCommand("remote", "add", "origin", want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := InferRepo()
	if err != nil {
		t.Fatalf("unexpected error inferring repo: %s", err)
	}

	if got != want {
		t.Errorf("unexpected remote repo. want=%q have=%q", want, got)
	}
}

func TestParseRemote(t *testing.T) {
	testCases := map[string]string{
		"git@github.com:sourcegraph/src-cli.git": "github.com/sourcegraph/src-cli",
		"https://github.com/sourcegraph/src-cli": "github.com/sourcegraph/src-cli",
	}

	for input, expectedOutput := range testCases {
		t.Run(fmt.Sprintf("input=%q", input), func(t *testing.T) {
			output, err := parseRemote(input)
			if err != nil {
				t.Fatalf("unexpected error parsing remote: %s", err)
			}

			if output != expectedOutput {
				t.Errorf("unexpected repo name. want=%q have=%q", expectedOutput, output)
			}
		})
	}
}

func TestInferRoot(t *testing.T) {
	gitDir, err := os.MkdirTemp("", "temp-test-infer-root-rep")
	if err != nil {
		t.Fatalf("unexpected error creating temporary directory: %s", err)
	}

	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("git dir %s left intact for inspection", gitDir)
		} else {
			os.RemoveAll(gitDir)
		}
	})

	// Needed in CI
	_, err = runGitCommand("init")
	if err != nil {
		t.Fatalf("unexpected error initializing git repo: %v", err)
	}
	_, err = runGitCommand("config", "user.email", "test@sourcegraph.com")
	if err != nil {
		t.Fatalf("unexpected error configuring git repo: %v", err)
	}

	repoRootPath := "../.."

	testCases := map[string]string{
		"gitutil.go":            ".",
		"../../cmd/src/lsif.go": filepath.Join(repoRootPath, "cmd", "src"),
		"../../README.md":       repoRootPath,
	}

	for input, expectedOutput := range testCases {
		t.Run(fmt.Sprintf("input=%q", input), func(t *testing.T) {
			root, err := InferRoot(input)
			if err != nil {
				t.Fatalf("unexpected error inferring root: %s", err)
			}

			if root != expectedOutput {
				t.Errorf("unexpected remote root. want=%q have=%q", expectedOutput, root)
			}
		})
	}
}

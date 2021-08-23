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
	testCases := map[string]string{
		"gitutil.go":            filepath.Join("internal", "codeintel"),
		"../../cmd/src/lsif.go": filepath.Join("cmd", "src"),
		"../../README.md":       ".",
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

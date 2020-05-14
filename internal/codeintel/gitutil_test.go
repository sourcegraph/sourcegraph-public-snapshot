package codeintel

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestInferRepo(t *testing.T) {
	repo, err := InferRepo()
	if err != nil {
		t.Fatalf("unexpected error inferring repo: %s", err)
	}

	if repo != "github.com/sourcegraph/src-cli" {
		t.Errorf("unexpected remote repo. want=%q have=%q", "github.com/sourcegraph/src-cli", repo)
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

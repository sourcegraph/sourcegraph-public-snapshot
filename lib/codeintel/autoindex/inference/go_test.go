package inference

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestGoPatterns(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{"go.mod", true},
		{"subdir/go.mod", true},
		{"vendor/foo/go.mod", true},
		{"go.mod/subdir", false},
		{"foo.go", true},
		{"subdir/foo.go", false},
	}

	for _, testCase := range testCases {
		match := false
		for _, pattern := range GoPatterns() {
			if pattern.MatchString(testCase.path) {
				match = true
				break
			}
		}

		if match {
			if !testCase.expected {
				t.Error(fmt.Sprintf("did not expect match: %s", testCase.path))
			}

		} else if testCase.expected {
			t.Error(fmt.Sprintf("expected match: %s", testCase.path))
		}
	}
}

func TestInferGoIndexJobsGoModRoot(t *testing.T) {
	paths := []string{
		"go.mod",
	}

	expectedIndexJobs := []config.IndexJob{
		{
			Steps: []config.DockerStep{
				{
					Root:     "",
					Image:    lsifGoImage,
					Commands: []string{"go mod download"},
				},
			},
			Root:        "",
			Indexer:     lsifGoImage,
			IndexerArgs: []string{"lsif-go", "--no-animation"},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, InferGoIndexJobs(NewMockGitClient(), paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestInferGoIndexJobsGoModSubdirs(t *testing.T) {
	paths := []string{
		"a/go.mod",
		"b/go.mod",
		"c/go.mod",
	}

	expectedIndexJobs := []config.IndexJob{
		{
			Steps: []config.DockerStep{
				{
					Root:     "a",
					Image:    lsifGoImage,
					Commands: []string{"go mod download"},
				},
			},
			Root:        "a",
			Indexer:     lsifGoImage,
			IndexerArgs: []string{"lsif-go", "--no-animation"},
			Outfile:     "",
		},
		{
			Steps: []config.DockerStep{
				{
					Root:     "b",
					Image:    lsifGoImage,
					Commands: []string{"go mod download"},
				},
			},
			Root:        "b",
			Indexer:     lsifGoImage,
			IndexerArgs: []string{"lsif-go", "--no-animation"},
			Outfile:     "",
		},
		{
			Steps: []config.DockerStep{
				{
					Root:     "c",
					Image:    lsifGoImage,
					Commands: []string{"go mod download"},
				},
			},
			Root:        "c",
			Indexer:     lsifGoImage,
			IndexerArgs: []string{"lsif-go", "--no-animation"},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, InferGoIndexJobs(NewMockGitClient(), paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestInferGoIndexJobsNoGoModFile(t *testing.T) {
	paths := []string{
		"lib.go",
		"lib_test.go",
		"doc.go",
	}

	expectedIndexJobs := []config.IndexJob{
		{
			Steps:       nil,
			Root:        "",
			Indexer:     lsifGoImage,
			IndexerArgs: []string{"GO111MODULE=off", "lsif-go", "--no-animation"},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, InferGoIndexJobs(NewMockGitClient(), paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

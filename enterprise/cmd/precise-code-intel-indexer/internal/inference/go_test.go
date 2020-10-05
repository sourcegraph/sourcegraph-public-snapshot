package inference

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLSIFGoJobRecognizerCanIndex(t *testing.T) {
	testCases := []struct {
		paths    []string
		expected bool
	}{
		{paths: []string{"go.mod"}, expected: true},
		{paths: []string{"a/go.mod"}, expected: true},
		{paths: []string{"package.json"}, expected: false},
		{paths: []string{"foo/bar-go.mod"}, expected: false},
	}

	recognizer := lsifGoJobRecognizer{}

	for _, testCase := range testCases {
		if value := recognizer.CanIndex(testCase.paths); value != testCase.expected {
			t.Errorf("unexpected result from CanIndex. want=%v have=%v", testCase.expected, value)
		}
	}
}

func TestLsifGoJobRecognizerInferIndexJobsGoModRoot(t *testing.T) {
	recognizer := lsifGoJobRecognizer{}
	paths := []string{
		"go.mod",
	}

	expectedIndexJobs := []IndexJob{
		{
			DockerSteps: nil,
			Root:        "",
			Indexer:     "sourcegraph/lsif-go:latest",
			IndexerArgs: []string{"lsif-go", "--no-animation"},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, recognizer.InferIndexJobs(paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestLsifGoJobRecognizerInferIndexJobsGoModSubdirs(t *testing.T) {
	recognizer := lsifGoJobRecognizer{}
	paths := []string{
		"a/go.mod",
		"b/go.mod",
		"c/go.mod",
	}

	expectedIndexJobs := []IndexJob{
		{
			DockerSteps: nil,
			Root:        "a",
			Indexer:     "sourcegraph/lsif-go:latest",
			IndexerArgs: []string{"lsif-go", "--no-animation"},
			Outfile:     "",
		},
		{
			DockerSteps: nil,
			Root:        "b",
			Indexer:     "sourcegraph/lsif-go:latest",
			IndexerArgs: []string{"lsif-go", "--no-animation"},
			Outfile:     "",
		},
		{
			DockerSteps: nil,
			Root:        "c",
			Indexer:     "sourcegraph/lsif-go:latest",
			IndexerArgs: []string{"lsif-go", "--no-animation"},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, recognizer.InferIndexJobs(paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestLSIFGoJobRecognizerPatterns(t *testing.T) {
	recognizer := lsifGoJobRecognizer{}
	paths := []string{
		"go.mod",
		"subdir/go.mod",
	}

	for _, path := range paths {
		match := false
		for _, pattern := range recognizer.Patterns() {
			if pattern.MatchString(path) {
				match = true
				break
			}
		}

		if !match {
			t.Error(fmt.Sprintf("failed to match %s", path))
		}
	}
}

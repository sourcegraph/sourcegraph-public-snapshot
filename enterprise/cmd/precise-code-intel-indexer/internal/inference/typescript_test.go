package inference

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLSIFTscJobRecognizerCanIndex(t *testing.T) {
	recognizer := lsifTscJobRecognizer{}
	testCases := []struct {
		paths    []string
		expected bool
	}{
		{paths: []string{"tsconfig.json"}, expected: true},
		{paths: []string{"a/tsconfig.json"}, expected: true},
		{paths: []string{"package.json"}, expected: false},
		{paths: []string{"node_modules/foo/bar/package.json"}, expected: false},
		{paths: []string{"foo/bar-tsconfig.json"}, expected: false},
	}

	for _, testCase := range testCases {
		name := strings.Join(testCase.paths, ", ")

		t.Run(name, func(t *testing.T) {
			if value := recognizer.CanIndex(testCase.paths); value != testCase.expected {
				t.Errorf("unexpected result from CanIndex. want=%v have=%v", testCase.expected, value)
			}
		})
	}
}

func TestLsifTscJobRecognizerInferIndexJobsTsConfigRoot(t *testing.T) {
	recognizer := lsifTscJobRecognizer{}
	paths := []string{
		"tsconfig.json",
	}

	expectedIndexJobs := []IndexJob{
		{
			DockerSteps: nil,
			Root:        "",
			Indexer:     "sourcegraph/lsif-node:latest",
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, recognizer.InferIndexJobs(paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestLsifTscJobRecognizerInferIndexJobsTsConfigSubdirs(t *testing.T) {
	recognizer := lsifTscJobRecognizer{}
	paths := []string{
		"a/tsconfig.json",
		"b/tsconfig.json",
		"c/tsconfig.json",
	}

	expectedIndexJobs := []IndexJob{
		{
			DockerSteps: nil,
			Root:        "a",
			Indexer:     "sourcegraph/lsif-node:latest",
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			DockerSteps: nil,
			Root:        "b",
			Indexer:     "sourcegraph/lsif-node:latest",
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			DockerSteps: nil,
			Root:        "c",
			Indexer:     "sourcegraph/lsif-node:latest",
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, recognizer.InferIndexJobs(paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestLsifTscJobRecognizerInferIndexJobsInstallSteps(t *testing.T) {
	recognizer := lsifTscJobRecognizer{}
	paths := []string{
		"tsconfig.json",
		"package.json",
		"foo/baz/tsconfig.json",
		"foo/bar/baz/tsconfig.json",
		"foo/bar/bonk/tsconfig.json",
		"foo/bar/bonk/package.json",
		"foo/bar/package.json",
		"foo/bar/yarn.lock",
	}

	expectedIndexJobs := []IndexJob{
		{
			DockerSteps: []DockerStep{
				{
					Root:     "",
					Image:    "node:alpine3.12",
					Commands: []string{"npm", "install"},
				},
			},
			Root:        "",
			Indexer:     "sourcegraph/lsif-node:latest",
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			DockerSteps: []DockerStep{
				{
					Root:     "",
					Image:    "node:alpine3.12",
					Commands: []string{"npm", "install"},
				},
			},
			Root:        "foo/baz",
			Indexer:     "sourcegraph/lsif-node:latest",
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			DockerSteps: []DockerStep{
				{
					Root:     "",
					Image:    "node:alpine3.12",
					Commands: []string{"npm", "install"},
				},
				{
					Root:     "foo/bar",
					Image:    "node:alpine3.12",
					Commands: []string{"yarn", "--ignore-engines"},
				},
			},
			Root:        "foo/bar/baz",
			Indexer:     "sourcegraph/lsif-node:latest",
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			DockerSteps: []DockerStep{
				{
					Root:     "",
					Image:    "node:alpine3.12",
					Commands: []string{"npm", "install"},
				},
				{
					Root:     "foo/bar",
					Image:    "node:alpine3.12",
					Commands: []string{"yarn", "--ignore-engines"},
				},
				{
					Root:     "foo/bar/bonk",
					Image:    "node:alpine3.12",
					Commands: []string{"npm", "install"},
				},
			},
			Root:        "foo/bar/bonk",
			Indexer:     "sourcegraph/lsif-node:latest",
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, recognizer.InferIndexJobs(paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestLSIFTscJobRecognizerPatterns(t *testing.T) {
	recognizer := lsifTscJobRecognizer{}
	paths := []string{
		"tsconfig.json",
		"subdir/tsconfig.json",
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

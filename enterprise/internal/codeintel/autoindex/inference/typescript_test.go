package inference

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/config"
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
		{paths: []string{"node_modules/foo/bar/tsconfig.json"}, expected: false},
		{paths: []string{"foo/bar-tsconfig.json"}, expected: false},
	}

	for _, testCase := range testCases {
		name := strings.Join(testCase.paths, ", ")

		t.Run(name, func(t *testing.T) {
			if value := recognizer.CanIndex(testCase.paths, NewMockGitserverClientWrapper()); value != testCase.expected {
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

	expectedIndexJobs := []config.IndexJob{
		{
			Steps:       nil,
			Root:        "",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, recognizer.InferIndexJobs(paths, NewMockGitserverClientWrapper())); diff != "" {
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

	expectedIndexJobs := []config.IndexJob{
		{
			Steps:       nil,
			Root:        "a",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			Steps:       nil,
			Root:        "b",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			Steps:       nil,
			Root:        "c",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, recognizer.InferIndexJobs(paths, NewMockGitserverClientWrapper())); diff != "" {
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

	expectedIndexJobs := []config.IndexJob{
		{
			Steps: []config.DockerStep{
				{
					Root:     "",
					Image:    lsifTscImage,
					Commands: []string{"npm install"},
				},
			},
			Root:        "",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			Steps: []config.DockerStep{
				{
					Root:     "",
					Image:    lsifTscImage,
					Commands: []string{"npm install"},
				},
			},
			Root:        "foo/baz",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			Steps: []config.DockerStep{
				{
					Root:     "",
					Image:    lsifTscImage,
					Commands: []string{"npm install"},
				},
				{
					Root:     "foo/bar",
					Image:    lsifTscImage,
					Commands: []string{"yarn --ignore-engines"},
				},
			},
			Root:        "foo/bar/baz",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			Steps: []config.DockerStep{
				{
					Root:     "",
					Image:    lsifTscImage,
					Commands: []string{"npm install"},
				},
				{
					Root:     "foo/bar",
					Image:    lsifTscImage,
					Commands: []string{"yarn --ignore-engines"},
				},
				{
					Root:     "foo/bar/bonk",
					Image:    lsifTscImage,
					Commands: []string{"npm install"},
				},
			},
			Root:        "foo/bar/bonk",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, recognizer.InferIndexJobs(paths, NewMockGitserverClientWrapper())); diff != "" {
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

func TestLSIFTscLernaConfig(t *testing.T) {
	mockGit := NewMockGitserverClientWrapper()
	// this is kinda tied to the order in which the impl calls them :(
	{
		mockGit.RawContentsFunc.PushReturn([]byte(`{"npmClient": "yarn"}`), nil)
		mockGit.RawContentsFunc.PushReturn([]byte(`{}`), nil)
	}
	{
		mockGit.RawContentsFunc.PushReturn([]byte(`{}`), nil)
		mockGit.RawContentsFunc.PushReturn([]byte(`{"npmClient": "npm"}`), nil)
		mockGit.RawContentsFunc.PushReturn([]byte(`{}`), nil)
	}
	{
		mockGit.RawContentsFunc.PushReturn([]byte(`{"npmClient": "yarn"}`), nil)
		mockGit.RawContentsFunc.PushReturn([]byte(`{}`), nil)
	}

	recognizer := lsifTscJobRecognizer{}

	paths := [][]string{
		{
			"package.json",
			"lerna.json",
			"tsconfig.json",
		},
		{
			"package.json",
			"lerna.json",
			"tsconfig.json",
		},
		{
			"package.json",
			"tsconfig.json",
		},
		{
			"foo/package.json",
			"yarn.lock",
			"lerna.json",
			"package.json",
			"foo/bar/tsconfig.json",
		},
	}

	expectedJobs := [][]config.IndexJob{
		{
			{
				Steps: []config.DockerStep{
					{
						Root:     "",
						Image:    lsifTscImage,
						Commands: []string{"yarn --ignore-engines"},
					},
				},
				LocalSteps:  nil,
				Root:        "",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			},
		},
		{
			{
				Steps: []config.DockerStep{
					{
						Root:     "",
						Image:    lsifTscImage,
						Commands: []string{"npm install"},
					},
				},
				LocalSteps:  nil,
				Root:        "",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			},
		},
		{
			{
				Steps: []config.DockerStep{
					{
						Root:     "",
						Image:    lsifTscImage,
						Commands: []string{"npm install"},
					},
				},
				LocalSteps:  nil,
				Root:        "",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			},
		},
		{
			{
				Steps: []config.DockerStep{
					{
						Root:     "",
						Image:    lsifTscImage,
						Commands: []string{"yarn --ignore-engines"},
					},
					{
						Root:     "foo",
						Image:    lsifTscImage,
						Commands: []string{"yarn --ignore-engines"},
					},
				},
				LocalSteps:  nil,
				Root:        "foo/bar",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			},
		},
	}

	for i, paths := range paths {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if diff := cmp.Diff(expectedJobs[i], recognizer.InferIndexJobs(paths, mockGit)); diff != "" {
				t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLSIFTscNodeVersionInferrence(t *testing.T) {
	mockGit := NewMockGitserverClientWrapper()
	mockGit.RawContentsFunc.PushReturn([]byte(""), nil)
	mockGit.RawContentsFunc.PushReturn([]byte(`{"engines":{"node":"420"}}`), nil)

	recognizer := lsifTscJobRecognizer{}

	paths := [][]string{
		{
			"package.json",
			"tsconfig.json",
			".nvmrc",
		},
		{
			"tsconfig.json",
			"package.json",
		},
	}

	expectedJobs := [][]config.IndexJob{
		{
			{
				Steps: []config.DockerStep{
					{
						Root:     "",
						Image:    lsifTscImage,
						Commands: []string{nMuslCommand, "npm install"},
					},
				},
				LocalSteps: []string{
					nMuslCommand,
				},
				Root:        "",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			},
		},
		{
			{
				Steps: []config.DockerStep{
					{
						Root:     "",
						Image:    lsifTscImage,
						Commands: []string{nMuslCommand, "npm install"},
					},
				},
				LocalSteps: []string{
					nMuslCommand,
				},
				Root:        "",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			},
		},
	}

	for i, paths := range paths {
		if diff := cmp.Diff(expectedJobs[i], recognizer.InferIndexJobs(paths, mockGit)); diff != "" {
			t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
		}
	}
}

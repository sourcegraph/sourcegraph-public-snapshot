package inference

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

func TestLSIFGoJobRecognizerCanIndex(t *testing.T) {
	recognizer := lsifGoJobRecognizer{}
	testCases := []struct {
		paths    []string
		expected bool
	}{
		{paths: []string{"go.mod"}, expected: true},
		{paths: []string{"a/go.mod"}, expected: true},
		{paths: []string{"package.json"}, expected: false},
		{paths: []string{"vendor/foo/bar/go.mod"}, expected: false},
		{paths: []string{"foo/bar-go.mod"}, expected: false},
	}

	for _, testCase := range testCases {
		name := strings.Join(testCase.paths, ", ")

		t.Run(name, func(t *testing.T) {
			if value := recognizer.CanIndexRepo(testCase.paths, NewMockGitserverClientWrapper()); value != testCase.expected {
				t.Errorf("unexpected result from CanIndex. want=%v have=%v", testCase.expected, value)
			}
		})
	}
}

func TestLsifGoJobRecognizerInferIndexJobsGoModRoot(t *testing.T) {
	recognizer := lsifGoJobRecognizer{}
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
	if diff := cmp.Diff(expectedIndexJobs, recognizer.InferIndexJobs(paths, NewMockGitserverClientWrapper())); diff != "" {
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
	if diff := cmp.Diff(expectedIndexJobs, recognizer.InferIndexJobs(paths, NewMockGitserverClientWrapper())); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestLSIFGoJobRecognizerPatterns(t *testing.T) {
	recognizer := lsifGoJobRecognizer{}
	paths := []string{
		"go.mod",
		"subdir/go.mod",
		"vendor/foo/go.mod",
		"foo/vendor/go.mod",
		"foo/go.mod/vendor",
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

func TestLSIFGoPackages(t *testing.T) {
	recognizer := lsifGoJobRecognizer{}
	packages := []semantic.Package{
		{
			Scheme:  "gomod",
			Name:    "github.com/sourcegraph/sourcegraph",
			Version: "v2.3.2",
		},
		{
			Scheme:  "gomod",
			Name:    "github.com/aws/aws-sdk-go-v2/credentials",
			Version: "v0.1.0",
		},
		{
			Scheme:  "gomod",
			Name:    "github.com/sourcegraph/sourcegraph",
			Version: "v0.0.0-de0123456789",
		},
	}

	repoUpdater := NewMockRepoUpdaterClient()
	repoUpdater.EnqueueRepoUpdateFunc.SetDefaultReturn(&protocol.RepoUpdateResponse{ID: 42}, nil)
	gitserver := NewMockGitserverClient()
	gitserver.ResolveRevisionFunc.SetDefaultReturn("c42", nil)

	for _, pkg := range packages {
		recognizer.EnsurePackageRepo(context.Background(), pkg, repoUpdater, gitserver)
	}

	if len(repoUpdater.EnqueueRepoUpdateFunc.history) != 3 {
		t.Errorf("unexpected number of calls to EnqueueRepoUpdate, want %v, got %v", 2, len(repoUpdater.EnqueueRepoUpdateFunc.history))
	} else {
		expectedRepoNames := []string{"github.com/sourcegraph/sourcegraph", "github.com/aws/aws-sdk-go-v2", "github.com/sourcegraph/sourcegraph"}
		var calledRepoNames []string
		for _, hist := range repoUpdater.EnqueueRepoUpdateFunc.history {
			calledRepoNames = append(calledRepoNames, string(hist.Arg1))
		}
		if diff := cmp.Diff(expectedRepoNames, calledRepoNames); diff != "" {
			t.Errorf("unexpected repo names (-want +got):\n%s", diff)
		}
	}

	if len(gitserver.ResolveRevisionFunc.history) != 3 {
		t.Errorf("unexpected number of calls to ResolveRevision, want %v, got %v", 2, len(gitserver.ResolveRevisionFunc.history))
	} else {
		expectedCommitNames := []string{"v2.3.2", "v0.1.0", "de0123456789"}
		var calledCommitNames []string
		for _, hist := range gitserver.ResolveRevisionFunc.history {
			calledCommitNames = append(calledCommitNames, hist.Arg2)
		}
		if diff := cmp.Diff(expectedCommitNames, calledCommitNames); diff != "" {
			t.Errorf("unexpected commit names (-want +got):\n%s", diff)
		}
	}
}

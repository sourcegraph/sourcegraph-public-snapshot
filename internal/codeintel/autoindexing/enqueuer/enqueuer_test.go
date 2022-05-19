package enqueuer

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/api"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

var testConfig = Config{
	MaximumRepositoriesInspectedPerSecond:    rate.Inf,
	MaximumIndexJobsPerInferredConfiguration: 50,
}

func TestQueueIndexesExplicit(t *testing.T) {
	config := `{
		"shared_steps": [
			{
				"root": "/",
				"image": "node:12",
				"commands": [
					"yarn install --frozen-lockfile --non-interactive",
				],
			}
		],
		"index_jobs": [
			{
				"steps": [
					{
						// Comments are the future
						"image": "go:latest",
						"commands": ["go mod vendor"],
					}
				],
				"indexer": "lsif-go",
				"indexer_args": ["--no-animation"],
			},
			{
				"root": "web/",
				"indexer": "lsif-tsc",
				"indexer_args": ["-p", "."],
				"outfile": "lsif.dump",
			},
		]
	}`

	mockDBStore := NewMockDBStore()
	mockDBStore.TransactFunc.SetDefaultReturn(mockDBStore, nil)
	mockDBStore.DoneFunc.SetDefaultHook(func(err error) error { return err })
	mockDBStore.InsertIndexesFunc.SetDefaultHook(func(ctx context.Context, indexes []store.Index) ([]store.Index, error) { return indexes, nil })
	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, rev string) (api.CommitID, error) {
		return api.CommitID(fmt.Sprintf("c%d", repositoryID)), nil
	})

	inferenceService := NewMockInferenceService()

	scheduler := newIndexEnqueuer(mockDBStore, mockGitserverClient, nil, inferenceService, &testConfig, &observation.TestContext)
	_, _ = scheduler.QueueIndexes(context.Background(), 42, "HEAD", config, false)

	if len(mockDBStore.IsQueuedFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to IsQueued. want=%d have=%d", 1, len(mockDBStore.IsQueuedFunc.History()))
	} else {
		var commits []string
		for _, call := range mockDBStore.IsQueuedFunc.History() {
			commits = append(commits, call.Arg2)
		}
		sort.Strings(commits)

		if diff := cmp.Diff([]string{"c42"}, commits); diff != "" {
			t.Errorf("unexpected commits (-want +got):\n%s", diff)
		}
	}

	var indexes []store.Index
	for _, call := range mockDBStore.InsertIndexesFunc.History() {
		indexes = append(indexes, call.Result0...)
	}

	expectedIndexes := []store.Index{
		{
			RepositoryID: 42,
			Commit:       "c42",
			State:        "queued",
			DockerSteps: []store.DockerStep{
				{
					Root:     "/",
					Image:    "node:12",
					Commands: []string{"yarn install --frozen-lockfile --non-interactive"},
				},
				{
					Image:    "go:latest",
					Commands: []string{"go mod vendor"},
				},
			},
			Indexer:     "lsif-go",
			IndexerArgs: []string{"--no-animation"},
		},
		{
			RepositoryID: 42,
			Commit:       "c42",
			State:        "queued",
			DockerSteps: []store.DockerStep{
				{
					Root:     "/",
					Image:    "node:12",
					Commands: []string{"yarn install --frozen-lockfile --non-interactive"},
				},
			},
			Root:        "web/",
			Indexer:     "lsif-tsc",
			IndexerArgs: []string{"-p", "."},
			Outfile:     "lsif.dump",
		},
	}
	if diff := cmp.Diff(expectedIndexes, indexes); diff != "" {
		t.Errorf("unexpected indexes (-want +got):\n%s", diff)
	}
}

func TestQueueIndexesInDatabase(t *testing.T) {
	indexConfiguration := store.IndexConfiguration{
		ID:           1,
		RepositoryID: 42,
		Data: []byte(`{
			"shared_steps": [
				{
					"root": "/",
					"image": "node:12",
					"commands": [
						"yarn install --frozen-lockfile --non-interactive",
					],
				}
			],
			"index_jobs": [
				{
					"steps": [
						{
							// Comments are the future
							"image": "go:latest",
							"commands": ["go mod vendor"],
						}
					],
					"indexer": "lsif-go",
					"indexer_args": ["--no-animation"],
				},
				{
					"root": "web/",
					"indexer": "lsif-tsc",
					"indexer_args": ["-p", "."],
					"outfile": "lsif.dump",
				},
			]
		}`),
	}

	mockDBStore := NewMockDBStore()
	mockDBStore.TransactFunc.SetDefaultReturn(mockDBStore, nil)
	mockDBStore.DoneFunc.SetDefaultHook(func(err error) error { return err })
	mockDBStore.InsertIndexesFunc.SetDefaultHook(func(ctx context.Context, indexes []store.Index) ([]store.Index, error) { return indexes, nil })
	mockDBStore.GetIndexConfigurationByRepositoryIDFunc.SetDefaultReturn(indexConfiguration, true, nil)

	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, rev string) (api.CommitID, error) {
		return api.CommitID(fmt.Sprintf("c%d", repositoryID)), nil
	})

	inferenceService := NewMockInferenceService()

	scheduler := newIndexEnqueuer(mockDBStore, mockGitserverClient, nil, inferenceService, &testConfig, &observation.TestContext)
	_, _ = scheduler.QueueIndexes(context.Background(), 42, "HEAD", "", false)

	if len(mockDBStore.GetIndexConfigurationByRepositoryIDFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to GetIndexConfigurationByRepositoryID. want=%d have=%d", 1, len(mockDBStore.GetIndexConfigurationByRepositoryIDFunc.History()))
	} else {
		var repositoryIDs []int
		for _, call := range mockDBStore.GetIndexConfigurationByRepositoryIDFunc.History() {
			repositoryIDs = append(repositoryIDs, call.Arg1)
		}
		sort.Ints(repositoryIDs)

		if diff := cmp.Diff([]int{42}, repositoryIDs); diff != "" {
			t.Errorf("unexpected repository identifiers (-want +got):\n%s", diff)
		}
	}

	if len(mockDBStore.IsQueuedFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to IsQueued. want=%d have=%d", 1, len(mockDBStore.IsQueuedFunc.History()))
	} else {
		var commits []string
		for _, call := range mockDBStore.IsQueuedFunc.History() {
			commits = append(commits, call.Arg2)
		}
		sort.Strings(commits)

		if diff := cmp.Diff([]string{"c42"}, commits); diff != "" {
			t.Errorf("unexpected commits (-want +got):\n%s", diff)
		}
	}

	var indexes []store.Index
	for _, call := range mockDBStore.InsertIndexesFunc.History() {
		indexes = append(indexes, call.Result0...)
	}

	expectedIndexes := []store.Index{
		{
			RepositoryID: 42,
			Commit:       "c42",
			State:        "queued",
			DockerSteps: []store.DockerStep{
				{
					Root:     "/",
					Image:    "node:12",
					Commands: []string{"yarn install --frozen-lockfile --non-interactive"},
				},
				{
					Image:    "go:latest",
					Commands: []string{"go mod vendor"},
				},
			},
			Indexer:     "lsif-go",
			IndexerArgs: []string{"--no-animation"},
		},
		{
			RepositoryID: 42,
			Commit:       "c42",
			State:        "queued",
			DockerSteps: []store.DockerStep{
				{
					Root:     "/",
					Image:    "node:12",
					Commands: []string{"yarn install --frozen-lockfile --non-interactive"},
				},
			},
			Root:        "web/",
			Indexer:     "lsif-tsc",
			IndexerArgs: []string{"-p", "."},
			Outfile:     "lsif.dump",
		},
	}
	if diff := cmp.Diff(expectedIndexes, indexes); diff != "" {
		t.Errorf("unexpected indexes (-want +got):\n%s", diff)
	}
}

var yamlIndexConfiguration = []byte(`
shared_steps:
  - root: /
    image: node:12
    commands:
      - yarn install --frozen-lockfile --non-interactive

index_jobs:
  -
    steps:
      - image: go:latest
        commands:
          - go mod vendor
    indexer: lsif-go
    indexer_args:
      - --no-animation
  -
    root: web/
    indexer: lsif-tsc
    indexer_args: ['-p', '.']
    outfile: lsif.dump
`)

func TestQueueIndexesInRepository(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockDBStore.TransactFunc.SetDefaultReturn(mockDBStore, nil)
	mockDBStore.DoneFunc.SetDefaultHook(func(err error) error { return err })
	mockDBStore.InsertIndexesFunc.SetDefaultHook(func(ctx context.Context, indexes []store.Index) ([]store.Index, error) { return indexes, nil })

	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, rev string) (api.CommitID, error) {
		return api.CommitID(fmt.Sprintf("c%d", repositoryID)), nil
	})
	mockGitserverClient.FileExistsFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit, file string) (bool, error) {
		return file == "sourcegraph.yaml", nil
	})
	mockGitserverClient.RawContentsFunc.SetDefaultReturn(yamlIndexConfiguration, nil)

	inferenceService := NewMockInferenceService()

	scheduler := newIndexEnqueuer(mockDBStore, mockGitserverClient, nil, inferenceService, &testConfig, &observation.TestContext)

	if _, err := scheduler.QueueIndexes(context.Background(), 42, "HEAD", "", false); err != nil {
		t.Fatalf("unexpected error performing update: %s", err)
	}

	if len(mockDBStore.IsQueuedFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to IsQueued. want=%d have=%d", 1, len(mockDBStore.IsQueuedFunc.History()))
	} else {
		var commits []string
		for _, call := range mockDBStore.IsQueuedFunc.History() {
			commits = append(commits, call.Arg2)
		}
		sort.Strings(commits)

		if diff := cmp.Diff([]string{"c42"}, commits); diff != "" {
			t.Errorf("unexpected commits (-want +got):\n%s", diff)
		}
	}

	var indexes []store.Index
	for _, call := range mockDBStore.InsertIndexesFunc.History() {
		indexes = append(indexes, call.Result0...)
	}

	expectedIndexes := []store.Index{
		{
			RepositoryID: 42,
			Commit:       "c42",
			State:        "queued",
			DockerSteps: []store.DockerStep{
				{
					Root:     "/",
					Image:    "node:12",
					Commands: []string{"yarn install --frozen-lockfile --non-interactive"},
				},
				{
					Image:    "go:latest",
					Commands: []string{"go mod vendor"},
				},
			},
			Indexer:     "lsif-go",
			IndexerArgs: []string{"--no-animation"},
		},
		{
			RepositoryID: 42,
			Commit:       "c42",
			State:        "queued",
			DockerSteps: []store.DockerStep{
				{
					Root:     "/",
					Image:    "node:12",
					Commands: []string{"yarn install --frozen-lockfile --non-interactive"},
				},
			},
			Root:        "web/",
			Indexer:     "lsif-tsc",
			IndexerArgs: []string{"-p", "."},
			Outfile:     "lsif.dump",
		},
	}
	if diff := cmp.Diff(expectedIndexes, indexes); diff != "" {
		t.Errorf("unexpected indexes (-want +got):\n%s", diff)
	}
}

func TestQueueIndexesInferred(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockDBStore.TransactFunc.SetDefaultReturn(mockDBStore, nil)
	mockDBStore.DoneFunc.SetDefaultHook(func(err error) error { return err })
	mockDBStore.InsertIndexesFunc.SetDefaultHook(func(ctx context.Context, indexes []store.Index) ([]store.Index, error) { return indexes, nil })
	mockDBStore.RepoNameFunc.SetDefaultHook(func(ctx context.Context, i int) (string, error) { return fmt.Sprintf("%d", i), nil })

	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, rev string) (api.CommitID, error) {
		return api.CommitID(fmt.Sprintf("c%d", repositoryID)), nil
	})
	mockGitserverClient.ListFilesFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error) {
		switch repositoryID {
		case 42:
			return []string{"go.mod"}, nil
		case 44:
			return []string{"a/go.mod", "b/go.mod"}, nil
		default:
			return nil, nil
		}
	})

	inferenceService := NewMockInferenceService()
	inferenceService.InferIndexJobsFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, s1, s2 string) ([]config.IndexJob, error) {
		switch rn {
		case "42":
			return []config.IndexJob{{Root: ""}}, nil
		case "44":
			return []config.IndexJob{{Root: "a"}, {Root: "b"}}, nil
		default:
			return nil, nil
		}
	})

	scheduler := newIndexEnqueuer(mockDBStore, mockGitserverClient, nil, inferenceService, &testConfig, &observation.TestContext)

	for _, id := range []int{41, 42, 43, 44} {
		if _, err := scheduler.QueueIndexes(context.Background(), id, "HEAD", "", false); err != nil {
			t.Fatalf("unexpected error performing update: %s", err)
		}
	}

	indexRoots := map[int][]string{}
	for _, call := range mockDBStore.InsertIndexesFunc.History() {
		for _, index := range call.Result0 {
			indexRoots[index.RepositoryID] = append(indexRoots[index.RepositoryID], index.Root)
		}
	}

	expectedIndexRoots := map[int][]string{
		42: {""},
		44: {"a", "b"},
	}
	if diff := cmp.Diff(expectedIndexRoots, indexRoots); diff != "" {
		t.Errorf("unexpected indexes (-want +got):\n%s", diff)
	}

	if len(mockDBStore.IsQueuedFunc.History()) != 4 {
		t.Errorf("unexpected number of calls to IsQueued. want=%d have=%d", 4, len(mockDBStore.IsQueuedFunc.History()))
	} else {
		var commits []string
		for _, call := range mockDBStore.IsQueuedFunc.History() {
			commits = append(commits, call.Arg2)
		}
		sort.Strings(commits)

		if diff := cmp.Diff([]string{"c41", "c42", "c43", "c44"}, commits); diff != "" {
			t.Errorf("unexpected commits (-want +got):\n%s", diff)
		}
	}
}

func TestQueueIndexesInferredTooLarge(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockDBStore.TransactFunc.SetDefaultReturn(mockDBStore, nil)
	mockDBStore.DoneFunc.SetDefaultHook(func(err error) error { return err })
	mockDBStore.InsertIndexesFunc.SetDefaultHook(func(ctx context.Context, indexes []store.Index) ([]store.Index, error) { return indexes, nil })

	var paths []string
	for i := 0; i < 25; i++ {
		paths = append(paths, fmt.Sprintf("s%d/go.mod", i+1))
	}

	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, rev string) (api.CommitID, error) {
		return api.CommitID(fmt.Sprintf("c%d", repositoryID)), nil
	})
	mockGitserverClient.ListFilesFunc.SetDefaultHook(func(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error) {
		if repositoryID == 42 {
			return paths, nil
		}

		return nil, nil
	})

	inferenceService := NewMockInferenceService()

	config := testConfig
	config.MaximumIndexJobsPerInferredConfiguration = 20
	scheduler := newIndexEnqueuer(mockDBStore, mockGitserverClient, nil, inferenceService, &config, &observation.TestContext)

	if _, err := scheduler.QueueIndexes(context.Background(), 42, "HEAD", "", false); err != nil {
		t.Fatalf("unexpected error performing update: %s", err)
	}

	if len(mockDBStore.InsertIndexesFunc.History()) != 0 {
		t.Errorf("unexpected number of calls to InsertIndexes. want=%d have=%d", 0, len(mockDBStore.InsertIndexesFunc.History()))
	}
}

func TestQueueIndexesForPackage(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockDBStore.TransactFunc.SetDefaultReturn(mockDBStore, nil)
	mockDBStore.DoneFunc.SetDefaultHook(func(err error) error { return err })
	mockDBStore.InsertIndexesFunc.SetDefaultHook(func(ctx context.Context, indexes []store.Index) ([]store.Index, error) { return indexes, nil })
	mockDBStore.IsQueuedFunc.SetDefaultReturn(false, nil)
	mockDBStore.RepoNameFunc.SetDefaultHook(func(ctx context.Context, i int) (string, error) { return fmt.Sprintf("%d", i), nil })

	mockGitserverClient := NewMockGitserverClient()
	mockGitserverClient.ResolveRevisionFunc.SetDefaultHook(func(ctx context.Context, repoID int, versionString string) (api.CommitID, error) {
		if repoID != 42 || versionString != "4e7eeb0f8a96" {

			t.Errorf("unexpected (repoID, versionString) (%v, %v) supplied to EnqueueRepoUpdate", repoID, versionString)
		}
		return "c42", nil
	})
	mockGitserverClient.ListFilesFunc.SetDefaultReturn([]string{"go.mod"}, nil)

	mockRepoUpdater := NewMockRepoUpdaterClient()
	mockRepoUpdater.EnqueueRepoUpdateFunc.SetDefaultHook(func(ctx context.Context, repoName api.RepoName) (*protocol.RepoUpdateResponse, error) {
		if repoName != "github.com/sourcegraph/sourcegraph" {
			t.Errorf("unexpected repo %v supplied to EnqueueRepoUpdate", repoName)
		}
		return &protocol.RepoUpdateResponse{ID: 42}, nil
	})

	inferenceService := NewMockInferenceService()
	inferenceService.InferIndexJobsFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, s1, s2 string) ([]config.IndexJob, error) {
		return []config.IndexJob{
			{
				Root: "",
				Steps: []config.DockerStep{
					{
						Image:    "sourcegraph/lsif-go:latest",
						Commands: []string{"go mod download"},
					},
				},
				Indexer:     "sourcegraph/lsif-go:latest",
				IndexerArgs: []string{"lsif-go", "--no-animation"},
			},
		}, nil
	})

	scheduler := newIndexEnqueuer(mockDBStore, mockGitserverClient, mockRepoUpdater, inferenceService, &testConfig, &observation.TestContext)

	_ = scheduler.QueueIndexesForPackage(context.Background(), precise.Package{
		Scheme:  "gomod",
		Name:    "https://github.com/sourcegraph/sourcegraph",
		Version: "v3.26.0-4e7eeb0f8a96",
	})

	if len(mockDBStore.IsQueuedFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to IsQueued. want=%d have=%d", 1, len(mockDBStore.IsQueuedFunc.History()))
	} else {
		var commits []string
		for _, call := range mockDBStore.IsQueuedFunc.History() {
			commits = append(commits, call.Arg2)
		}
		sort.Strings(commits)

		if diff := cmp.Diff([]string{"c42"}, commits); diff != "" {
			t.Errorf("unexpected commits (-want +got):\n%s", diff)
		}
	}

	if len(mockDBStore.InsertIndexesFunc.History()) != 1 {
		t.Errorf("unexpected number of calls to InsertIndexes. want=%d have=%d", 1, len(mockDBStore.InsertIndexesFunc.History()))
	} else {
		var indexes []store.Index
		for _, call := range mockDBStore.InsertIndexesFunc.History() {
			indexes = append(indexes, call.Result0...)
		}

		expectedIndexes := []store.Index{
			{
				RepositoryID: 42,
				Commit:       "c42",
				State:        "queued",
				DockerSteps: []store.DockerStep{
					{
						Image:    "sourcegraph/lsif-go:latest",
						Commands: []string{"go mod download"},
					},
				},
				Indexer:     "sourcegraph/lsif-go:latest",
				IndexerArgs: []string{"lsif-go", "--no-animation"},
			},
		}
		if diff := cmp.Diff(expectedIndexes, indexes); diff != "" {
			t.Errorf("unexpected indexes (-want +got):\n%s", diff)
		}
	}
}

package batches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestTransformRecord(t *testing.T) {
	accessToken := "thisissecret-dont-tell-anyone"
	var accessTokenID int64 = 1234

	accessTokens := database.NewMockAccessTokenStore()
	accessTokens.CreateInternalFunc.SetDefaultReturn(accessTokenID, accessToken, nil)

	db := database.NewMockDB()
	db.AccessTokensFunc.SetDefaultReturn(accessTokens)
	repos := database.NewMockRepoStore()
	repos.GetFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: id, Name: "github.com/sourcegraph/sourcegraph"}, nil
	})
	db.ReposFunc.SetDefaultReturn(repos)

	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "https://test.io"}})
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	batchSpec := &btypes.BatchSpec{
		UserID:          123,
		NamespaceUserID: 123,
		RawSpec:         "horse",
		Spec:            &batcheslib.BatchSpec{},
	}

	workspace := &btypes.BatchSpecWorkspace{
		BatchSpecID:      batchSpec.ID,
		ChangesetSpecIDs: []int64{},
		RepoID:           5678,
		Branch:           "refs/heads/base-branch",
		Commit:           "d34db33f",
		Path:             "a/b/c",
		Steps: []batcheslib.Step{
			{Run: "echo lol >> readme.md", Container: "alpine:3"},
			{Run: "echo more lol >> readme.md", Container: "alpine:3"},
		},
		FileMatches:        []string{"a/b/c/foobar.go"},
		OnlyFetchWorkspace: true,
		StepCacheResults: map[int]btypes.StepCacheResult{
			1: {
				Key: "testcachekey",
				Value: &execution.AfterStepResult{
					Diff: "123",
				},
			},
		},
	}

	workspaceExecutionJob := &btypes.BatchSpecWorkspaceExecutionJob{
		ID:                   42,
		BatchSpecWorkspaceID: workspace.ID,
	}

	store := NewMockBatchesStore()
	store.GetBatchSpecFunc.SetDefaultReturn(batchSpec, nil)
	store.GetBatchSpecWorkspaceFunc.SetDefaultReturn(workspace, nil)
	var storedAccessToken int64
	store.SetBatchSpecWorkspaceExecutionJobAccessTokenFunc.SetDefaultHook(func(c context.Context, jobID, accessTokenID int64) error {
		storedAccessToken = accessTokenID
		return nil
	})
	store.DatabaseDBFunc.SetDefaultReturn(db)

	wantInput := batcheslib.WorkspacesExecutionInput{
		Spec: batchSpec.Spec,
		Workspace: batcheslib.Workspace{
			Repository: batcheslib.WorkspaceRepo{
				ID:   string(graphqlbackend.MarshalRepositoryID(workspace.RepoID)),
				Name: "github.com/sourcegraph/sourcegraph",
			},
			Branch: batcheslib.WorkspaceBranch{
				Name:   workspace.Branch,
				Target: batcheslib.Commit{OID: workspace.Commit},
			},
			Path:               workspace.Path,
			OnlyFetchWorkspace: workspace.OnlyFetchWorkspace,
			Steps:              workspace.Steps,
			SearchResultPaths:  workspace.FileMatches,
		},
	}

	marshaledInput, err := json.Marshal(&wantInput)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("with cache entry", func(t *testing.T) {
		job, err := transformRecord(context.Background(), store, workspaceExecutionJob, "hunter2")
		if err != nil {
			t.Fatalf("unexpected error transforming record: %s", err)
		}

		expected := apiclient.Job{
			ID: int(workspaceExecutionJob.ID),
			VirtualMachineFiles: map[string]string{
				"input.json":        string(marshaledInput),
				"testcachekey.json": `{"stepIndex":0,"diff":"123","outputs":null,"previousStepResult":{"Files":null,"Stdout":null,"Stderr":null}}`,
			},
			CliSteps: []apiclient.CliStep{
				{
					Commands: []string{"batch", "exec", "-f", "input.json"},
					Dir:      ".",
					Env: []string{
						"SRC_ENDPOINT=https://sourcegraph:hunter2@test.io",
						"SRC_ACCESS_TOKEN=" + accessToken,
					},
				},
			},
			RedactedValues: map[string]string{
				"https://sourcegraph:hunter2@test.io": "https://sourcegraph:PASSWORD_REMOVED@test.io",
				"hunter2":                             "PASSWORD_REMOVED",
				accessToken:                           "SRC_ACCESS_TOKEN_REMOVED",
			},
		}
		if diff := cmp.Diff(expected, job); diff != "" {
			t.Errorf("unexpected job (-want +got):\n%s", diff)
		}

		if storedAccessToken != accessTokenID {
			t.Errorf("wrong access token ID set on execution job: %d", storedAccessToken)
		}
	})

	t.Run("with cache disabled", func(t *testing.T) {
		// Set the no cache flag on the batch spec.
		batchSpec.NoCache = true

		job, err := transformRecord(context.Background(), store, workspaceExecutionJob, "hunter2")
		if err != nil {
			t.Fatalf("unexpected error transforming record: %s", err)
		}

		expected := apiclient.Job{
			ID: int(workspaceExecutionJob.ID),
			VirtualMachineFiles: map[string]string{
				"input.json": string(marshaledInput),
			},
			CliSteps: []apiclient.CliStep{
				{
					Commands: []string{"batch", "exec", "-f", "input.json"},
					Dir:      ".",
					Env: []string{
						"SRC_ENDPOINT=https://sourcegraph:hunter2@test.io",
						"SRC_ACCESS_TOKEN=" + accessToken,
					},
				},
			},
			RedactedValues: map[string]string{
				"https://sourcegraph:hunter2@test.io": "https://sourcegraph:PASSWORD_REMOVED@test.io",
				"hunter2":                             "PASSWORD_REMOVED",
				accessToken:                           "SRC_ACCESS_TOKEN_REMOVED",
			},
		}
		if diff := cmp.Diff(expected, job); diff != "" {
			t.Errorf("unexpected job (-want +got):\n%s", diff)
		}
	})
}

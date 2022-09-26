package batches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestTransformRecord(t *testing.T) {
	db := database.NewMockDB()
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
		Spec: &batcheslib.BatchSpec{
			Steps: []batcheslib.Step{
				{Run: "echo lol >> readme.md", Container: "alpine:3"},
				{Run: "echo more lol >> readme.md", Container: "alpine:3"},
			},
		},
	}

	workspace := &btypes.BatchSpecWorkspace{
		BatchSpecID:        batchSpec.ID,
		ChangesetSpecIDs:   []int64{},
		RepoID:             5678,
		Branch:             "refs/heads/base-branch",
		Commit:             "d34db33f",
		Path:               "a/b/c",
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
		UserID:               123,
	}

	store := NewMockBatchesStore()
	store.GetBatchSpecFunc.SetDefaultReturn(batchSpec, nil)
	store.GetBatchSpecWorkspaceFunc.SetDefaultReturn(workspace, nil)
	store.DatabaseDBFunc.SetDefaultReturn(db)

	wantInput := func(cachedStepResultFound bool, cachedStepResult execution.AfterStepResult) batcheslib.WorkspacesExecutionInput {
		return batcheslib.WorkspacesExecutionInput{
			BatchChangeAttributes: template.BatchChangeAttributes{
				Name:        batchSpec.Spec.Name,
				Description: batchSpec.Spec.Description,
			},
			Repository: batcheslib.WorkspaceRepo{
				ID:   string(graphqlbackend.MarshalRepositoryID(workspace.RepoID)),
				Name: "github.com/sourcegraph/sourcegraph",
			},
			Branch: batcheslib.WorkspaceBranch{
				Name:   workspace.Branch,
				Target: batcheslib.Commit{OID: workspace.Commit},
			},
			Path:                  workspace.Path,
			OnlyFetchWorkspace:    workspace.OnlyFetchWorkspace,
			Steps:                 batchSpec.Spec.Steps,
			SearchResultPaths:     workspace.FileMatches,
			CachedStepResultFound: cachedStepResultFound,
			CachedStepResult:      cachedStepResult,
		}
	}

	t.Run("with cache entry", func(t *testing.T) {
		job, err := transformRecord(context.Background(), logtest.Scoped(t), store, workspaceExecutionJob)
		if err != nil {
			t.Fatalf("unexpected error transforming record: %s", err)
		}

		marshaledInput, err := json.Marshal(wantInput(true, execution.AfterStepResult{Diff: "123"}))
		if err != nil {
			t.Fatal(err)
		}

		expected := apiclient.Job{
			ID:                  int(workspaceExecutionJob.ID),
			RepositoryName:      "github.com/sourcegraph/sourcegraph",
			RepositoryDirectory: "repository",
			Commit:              workspace.Commit,
			ShallowClone:        true,
			SparseCheckout:      []string{"a/b/c/*"},
			VirtualMachineFiles: map[string]string{
				"input.json": string(marshaledInput),
			},
			CliSteps: []apiclient.CliStep{
				{
					Commands: []string{"batch", "exec", "-f", "input.json", "-repo", "repository", "-tmp", ".src-tmp"},
					Dir:      ".",
					Env:      []string{},
				},
			},
			RedactedValues: map[string]string{},
		}
		if diff := cmp.Diff(expected, job); diff != "" {
			t.Errorf("unexpected job (-want +got):\n%s", diff)
		}
	})

	t.Run("with cache disabled", func(t *testing.T) {
		// Set the no cache flag on the batch spec.
		batchSpec.NoCache = true

		job, err := transformRecord(context.Background(), logtest.Scoped(t), store, workspaceExecutionJob)
		if err != nil {
			t.Fatalf("unexpected error transforming record: %s", err)
		}

		marshaledInput, err := json.Marshal(wantInput(false, execution.AfterStepResult{}))
		if err != nil {
			t.Fatal(err)
		}

		expected := apiclient.Job{
			ID:                  int(workspaceExecutionJob.ID),
			RepositoryName:      "github.com/sourcegraph/sourcegraph",
			RepositoryDirectory: "repository",
			Commit:              workspace.Commit,
			ShallowClone:        true,
			SparseCheckout:      []string{"a/b/c/*"},
			VirtualMachineFiles: map[string]string{
				"input.json": string(marshaledInput),
			},
			CliSteps: []apiclient.CliStep{
				{
					Commands: []string{"batch", "exec", "-f", "input.json", "-repo", "repository", "-tmp", ".src-tmp"},
					Dir:      ".",
					Env:      []string{},
				},
			},
			RedactedValues: map[string]string{},
		}
		if diff := cmp.Diff(expected, job); diff != "" {
			t.Errorf("unexpected job (-want +got):\n%s", diff)
		}
	})
}

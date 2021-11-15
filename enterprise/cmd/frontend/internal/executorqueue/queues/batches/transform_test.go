package batches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmock"
	"github.com/sourcegraph/sourcegraph/internal/types"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestTransformRecord(t *testing.T) {
	accessToken := "thisissecret-dont-tell-anyone"
	var accessTokenID int64 = 1234

	accessTokens := dbmock.NewMockAccessTokenStore()
	accessTokens.CreateInternalFunc.SetDefaultReturn(accessTokenID, accessToken, nil)

	repos := dbmock.NewMockRepoStore()
	repos.GetFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: id, Name: "github.com/sourcegraph/sourcegraph"}, nil
	})

	db := dbmock.NewMockDB()
	db.AccessTokensFunc.SetDefaultReturn(accessTokens)
	db.ReposFunc.SetDefaultReturn(repos)

	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "https://test.io"}})
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	batchSpec := &btypes.BatchSpec{UserID: 123, NamespaceUserID: 123, RawSpec: "horse"}

	workspace := &btypes.BatchSpecWorkspace{
		BatchSpecID:        batchSpec.ID,
		ChangesetSpecIDs:   []int64{},
		RepoID:             5678,
		Branch:             "refs/heads/base-branch",
		Commit:             "d34db33f",
		Path:               "a/b/c",
		Steps:              []batcheslib.Step{{Run: "echo lol >> readme.md", Container: "alpine:3"}},
		FileMatches:        []string{"a/b/c/foobar.go"},
		OnlyFetchWorkspace: true,
	}

	workspaceExecutionJob := &btypes.BatchSpecWorkspaceExecutionJob{
		ID:                   42,
		BatchSpecWorkspaceID: workspace.ID,
	}

	store := &dummyBatchesStore{mockDB: db, batchSpec: batchSpec, batchSpecWorkspace: workspace}

	wantInput := batcheslib.WorkspacesExecutionInput{
		RawSpec: batchSpec.RawSpec,
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

	job, err := transformRecord(context.Background(), store, workspaceExecutionJob, "hunter2")
	if err != nil {
		t.Fatalf("unexpected error transforming record: %s", err)
	}

	expected := apiclient.Job{
		ID:                  int(workspaceExecutionJob.ID),
		VirtualMachineFiles: map[string]string{"input.json": string(marshaledInput)},
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

	if store.accessTokenID != accessTokenID {
		t.Errorf("wrong access token ID set on execution job: %d", store.accessTokenID)
	}
}

type dummyBatchesStore struct {
	mockDB             database.DB
	batchSpec          *btypes.BatchSpec
	batchSpecWorkspace *btypes.BatchSpecWorkspace

	accessTokenID int64
}

func (db *dummyBatchesStore) GetBatchSpecWorkspace(context.Context, store.GetBatchSpecWorkspaceOpts) (*btypes.BatchSpecWorkspace, error) {
	return db.batchSpecWorkspace, nil
}
func (db *dummyBatchesStore) GetBatchSpec(context.Context, store.GetBatchSpecOpts) (*btypes.BatchSpec, error) {
	return db.batchSpec, nil
}
func (db *dummyBatchesStore) DatabaseDB() database.DB { return db.mockDB }
func (db *dummyBatchesStore) SetBatchSpecWorkspaceExecutionJobAccessToken(ctx context.Context, jobID, tokenID int64) (err error) {
	db.accessTokenID = tokenID
	return nil
}

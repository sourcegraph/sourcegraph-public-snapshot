package batches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/config"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestTransformRecord(t *testing.T) {
	accessToken := "thisissecret-dont-tell-anyone"
	database.Mocks.AccessTokens.Create = func(subjectUserID int32, scopes []string, note string, creatorID int32) (int64, string, error) {
		return 1234, accessToken, nil
	}
	t.Cleanup(func() { database.Mocks.AccessTokens.Create = nil })

	testBatchSpec := `batchSpec: yeah`
	index := &btypes.BatchSpecExecution{
		ID:              42,
		UserID:          1,
		NamespaceUserID: 1,
		BatchSpec:       testBatchSpec,
	}
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "https://test.io"}})
	t.Cleanup(func() {
		conf.Mock(nil)
	})
	config := &Config{
		Shared: &config.SharedConfig{
			FrontendUsername: "test*",
			FrontendPassword: "hunter2",
		},
	}

	database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{
			Username: "john_namespace",
		}, nil
	}
	t.Cleanup(func() {
		database.Mocks.Users.GetByID = nil
	})

	job, err := transformRecord(context.Background(), &dbtesting.MockDB{}, index, config)
	if err != nil {
		t.Fatalf("unexpected error transforming record: %s", err)
	}

	expected := apiclient.Job{
		ID:                  42,
		VirtualMachineFiles: map[string]string{"spec.yml": testBatchSpec},
		CliSteps: []apiclient.CliStep{
			{
				Commands: []string{
					"batch", "preview",
					"-f", "spec.yml",
					"-text-only",
					"-skip-errors",
					"-n", "john_namespace",
				},
				Dir: ".",
				Env: []string{
					"SRC_ENDPOINT=https://test%2A:hunter2@test.io",
					"SRC_ACCESS_TOKEN=" + accessToken,
				},
			},
		},
		RedactedValues: map[string]string{
			"https://test%2A:hunter2@test.io": "https://USERNAME_REMOVED:PASSWORD_REMOVED@test.io",
			"test*":                           "USERNAME_REMOVED",
			"hunter2":                         "PASSWORD_REMOVED",
			accessToken:                       "SRC_ACCESS_TOKEN_REMOVED",
		},
	}
	if diff := cmp.Diff(expected, job); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}
}

func TestTransformBatchSpecWorkspaceExecutionJobRecord(t *testing.T) {
	accessToken := "thisissecret-dont-tell-anyone"
	database.Mocks.AccessTokens.Create = func(subjectUserID int32, scopes []string, note string, creatorID int32) (int64, string, error) {
		return 1234, accessToken, nil
	}
	t.Cleanup(func() { database.Mocks.AccessTokens.Create = nil })

	database.Mocks.Repos.Get = func(ctx context.Context, id api.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: id, Name: "github.com/sourcegraph/sourcegraph"}, nil
	}
	t.Cleanup(func() { database.Mocks.Repos.Get = nil })

	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "https://test.io"}})
	t.Cleanup(func() {
		conf.Mock(nil)
	})
	config := &Config{
		Shared: &config.SharedConfig{
			FrontendUsername: "test*",
			FrontendPassword: "hunter2",
		},
	}

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

	store := &dummyBatchesStore{dbHandle: &dbtesting.MockDB{}, batchSpec: batchSpec, batchSpecWorkspace: workspace}

	wantInput := batcheslib.WorkspacesExecutionInput{
		RawSpec: batchSpec.RawSpec,
		Workspaces: []*batcheslib.Workspace{
			{
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
		},
	}

	marshaledInput, err := json.Marshal(&wantInput)
	if err != nil {
		t.Fatal(err)
	}

	job, err := transformBatchSpecWorkspaceExecutionJobRecord(context.Background(), store, workspaceExecutionJob, config)
	if err != nil {
		t.Fatalf("unexpected error transforming record: %s", err)
	}

	expected := apiclient.Job{
		ID:                  int(workspaceExecutionJob.ID),
		VirtualMachineFiles: map[string]string{"input.json": string(marshaledInput)},
		CliSteps: []apiclient.CliStep{
			{
				Commands: []string{
					"batch", "exec",
					"-f", "input.json",
					"-skip-errors",
				},
				Dir: ".",
				Env: []string{
					"SRC_ENDPOINT=https://test%2A:hunter2@test.io",
					"SRC_ACCESS_TOKEN=" + accessToken,
				},
			},
		},
		RedactedValues: map[string]string{
			"https://test%2A:hunter2@test.io": "https://USERNAME_REMOVED:PASSWORD_REMOVED@test.io",
			"test*":                           "USERNAME_REMOVED",
			"hunter2":                         "PASSWORD_REMOVED",
			accessToken:                       "SRC_ACCESS_TOKEN_REMOVED",
		},
	}
	if diff := cmp.Diff(expected, job); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}
}

type dummyBatchesStore struct {
	dbHandle           dbutil.DB
	batchSpec          *btypes.BatchSpec
	batchSpecWorkspace *btypes.BatchSpecWorkspace
}

func (db *dummyBatchesStore) GetBatchSpecWorkspace(context.Context, store.GetBatchSpecWorkspaceOpts) (*btypes.BatchSpecWorkspace, error) {
	return db.batchSpecWorkspace, nil
}
func (db *dummyBatchesStore) GetBatchSpec(context.Context, store.GetBatchSpecOpts) (*btypes.BatchSpec, error) {
	return db.batchSpec, nil
}
func (db *dummyBatchesStore) DB() dbutil.DB { return db.dbHandle }

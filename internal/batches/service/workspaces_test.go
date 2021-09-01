package service

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

func TestFindWorkspaces(t *testing.T) {
	repos := []*graphql.Repository{
		{ID: "repo-id-0", Name: "github.com/sourcegraph/automation-testing"},
		{ID: "repo-id-1", Name: "github.com/sourcegraph/sourcegraph"},
		{ID: "repo-id-2", Name: "bitbucket.sgdev.org/SOUR/automation-testing"},
	}
	steps := []batcheslib.Step{{Run: "echo 1"}}

	type finderResults map[*graphql.Repository][]string

	tests := map[string]struct {
		spec          *batcheslib.BatchSpec
		finderResults map[*graphql.Repository][]string

		// workspaces in which repo/path they are executed
		wantWorkspaces []RepoWorkspace
	}{
		"no workspace configuration": {
			spec:          &batcheslib.BatchSpec{Steps: steps},
			finderResults: finderResults{},
			wantWorkspaces: []RepoWorkspace{
				{RepoID: repos[0].ID, Path: ""},
				{RepoID: repos[1].ID, Path: ""},
				{RepoID: repos[2].ID, Path: ""},
			},
		},

		"workspace configuration matching no repos": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{In: "this-does-not-match", RootAtLocationOf: "package.json"},
				},
			},
			finderResults: finderResults{},
			wantWorkspaces: []RepoWorkspace{
				{RepoID: repos[0].ID, Path: ""},
				{RepoID: repos[1].ID, Path: ""},
				{RepoID: repos[2].ID, Path: ""},
			},
		},

		"workspace configuration matching 2 repos with no results": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{In: "*automation-testing", RootAtLocationOf: "package.json"},
				},
			},
			finderResults: finderResults{
				repos[0]: []string{},
				repos[2]: []string{},
			},
			wantWorkspaces: []RepoWorkspace{
				{RepoID: repos[1].ID, Path: ""},
			},
		},

		"workspace configuration matching 2 repos with 3 results each": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{In: "*automation-testing", RootAtLocationOf: "package.json"},
				},
			},
			finderResults: finderResults{
				repos[0]: {"a/b", "a/b/c", "d/e/f"},
				repos[2]: {"a/b", "a/b/c", "d/e/f"},
			},
			wantWorkspaces: []RepoWorkspace{
				{RepoID: repos[0].ID, Path: "a/b"},
				{RepoID: repos[0].ID, Path: "a/b/c"},
				{RepoID: repos[0].ID, Path: "d/e/f"},
				{RepoID: repos[1].ID, Path: ""},
				{RepoID: repos[2].ID, Path: "a/b"},
				{RepoID: repos[2].ID, Path: "a/b/c"},
				{RepoID: repos[2].ID, Path: "d/e/f"},
			},
		},

		"workspace configuration matches repo with OnlyFetchWorkspace": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{
						OnlyFetchWorkspace: true,
						In:                 "*automation-testing",
						RootAtLocationOf:   "package.json",
					},
				},
			},
			finderResults: finderResults{
				repos[0]: {"a/b", "a/b/c", "d/e/f"},
				repos[2]: {"a/b", "a/b/c", "d/e/f"},
			},
			wantWorkspaces: []RepoWorkspace{
				{RepoID: repos[0].ID, Path: "a/b", OnlyFetchWorkspace: true},
				{RepoID: repos[0].ID, Path: "a/b/c", OnlyFetchWorkspace: true},
				{RepoID: repos[0].ID, Path: "d/e/f", OnlyFetchWorkspace: true},
				{RepoID: repos[1].ID, Path: ""},
				{RepoID: repos[2].ID, Path: "a/b", OnlyFetchWorkspace: true},
				{RepoID: repos[2].ID, Path: "a/b/c", OnlyFetchWorkspace: true},
				{RepoID: repos[2].ID, Path: "d/e/f", OnlyFetchWorkspace: true},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			finder := &mockDirectoryFinder{results: tt.finderResults}
			workspaces, err := findWorkspaces(context.Background(), tt.spec, finder, repos)
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			sort.Slice(workspaces, func(i, j int) bool {
				if workspaces[i].RepoID == workspaces[j].RepoID {
					return workspaces[i].Path < workspaces[j].Path
				}
				return workspaces[i].RepoID < workspaces[j].RepoID
			})

			if diff := cmp.Diff(tt.wantWorkspaces, workspaces); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

type mockDirectoryFinder struct {
	results map[*graphql.Repository][]string
}

func (m *mockDirectoryFinder) FindDirectoriesInRepos(ctx context.Context, fileName string, repos ...*graphql.Repository) (map[*graphql.Repository][]string, error) {
	return m.results, nil
}

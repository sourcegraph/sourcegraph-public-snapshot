package service

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

func TestTaskBuilder_BuildAll_Workspaces(t *testing.T) {
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

		// tasks per repository ID and in which path they are executed
		wantTasks []RepoWorkspaces
	}{
		"no workspace configuration": {
			spec:          &batcheslib.BatchSpec{Steps: steps},
			finderResults: finderResults{},
			wantTasks: []RepoWorkspaces{
				{RepoID: repos[0].ID, Paths: []string{""}},
				{RepoID: repos[1].ID, Paths: []string{""}},
				{RepoID: repos[2].ID, Paths: []string{""}},
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
			wantTasks: []RepoWorkspaces{
				{RepoID: repos[0].ID, Paths: []string{""}},
				{RepoID: repos[1].ID, Paths: []string{""}},
				{RepoID: repos[2].ID, Paths: []string{""}},
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
			wantTasks: []RepoWorkspaces{
				{RepoID: repos[1].ID, Paths: []string{""}},
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
			wantTasks: []RepoWorkspaces{
				{RepoID: repos[0].ID, Paths: []string{"a/b", "a/b/c", "d/e/f"}},
				{RepoID: repos[1].ID, Paths: []string{""}},
				{RepoID: repos[2].ID, Paths: []string{"a/b", "a/b/c", "d/e/f"}},
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
			wantTasks: []RepoWorkspaces{
				{RepoID: repos[0].ID, Paths: []string{"a/b", "a/b/c", "d/e/f"}, OnlyFetchWorkspace: true},
				{RepoID: repos[1].ID, Paths: []string{""}},
				{RepoID: repos[2].ID, Paths: []string{"a/b", "a/b/c", "d/e/f"}, OnlyFetchWorkspace: true},
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

			sort.Slice(workspaces, func(i, j int) bool { return workspaces[i].RepoID < workspaces[j].RepoID })

			for _, workspace := range workspaces {
				sort.Slice(workspace.Paths, func(i, j int) bool { return workspace.Paths[i] < workspace.Paths[j] })
			}

			if diff := cmp.Diff(tt.wantTasks, workspaces); diff != "" {
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

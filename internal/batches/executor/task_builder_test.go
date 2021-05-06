package executor

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

func TestTaskBuilder_BuildTask_IfConditions(t *testing.T) {
	repo := &graphql.Repository{Name: "github.com/sourcegraph/automation-testing"}

	tests := map[string]struct {
		spec *batches.BatchSpec

		wantSteps []batches.Step
	}{
		"no if": {
			spec: &batches.BatchSpec{
				Steps: []batches.Step{
					{Run: "echo 1"},
				},
			},
			wantSteps: []batches.Step{
				{Run: "echo 1"},
			},
		},

		"if has static true value": {
			spec: &batches.BatchSpec{
				Steps: []batches.Step{
					{Run: "echo 1", If: "true"},
				},
			},
			wantSteps: []batches.Step{
				{Run: "echo 1", If: "true"},
			},
		},

		"one of many steps has if with static true value": {
			spec: &batches.BatchSpec{
				Steps: []batches.Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "true"},
					{Run: "echo 3"},
				},
			},
			wantSteps: []batches.Step{
				{Run: "echo 1"},
				{Run: "echo 2", If: "true"},
				{Run: "echo 3"},
			},
		},

		"if has static non-true value": {
			spec: &batches.BatchSpec{
				Steps: []batches.Step{
					{Run: "echo 1", If: "this is not true"},
				},
			},
			wantSteps: nil,
		},

		"one of many steps has if with static non-true value": {
			spec: &batches.BatchSpec{
				Steps: []batches.Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "every type system needs generics"},
					{Run: "echo 3"},
				},
			},
			wantSteps: []batches.Step{
				{Run: "echo 1"},
				{Run: "echo 3"},
			},
		},

		"if expression that can be partially evaluated to true": {
			spec: &batches.BatchSpec{
				Steps: []batches.Step{
					{Run: "echo 1", If: `${{ matches repository.name "github.com/sourcegraph/automation*" }}`},
				},
			},
			wantSteps: []batches.Step{
				{Run: "echo 1", If: `${{ matches repository.name "github.com/sourcegraph/automation*" }}`},
			},
		},

		"if expression that can be partially evaluated to false": {
			spec: &batches.BatchSpec{
				Steps: []batches.Step{
					{Run: "echo 1", If: `${{ matches repository.name "horse" }}`},
				},
			},
			wantSteps: nil,
		},

		"one of many steps has if expression that can be evaluated to true": {
			spec: &batches.BatchSpec{
				Steps: []batches.Step{
					{Run: "echo 1", If: `${{ matches repository.name "horse" }}`},
				},
			},
			wantSteps: nil,
		},

		"if expression that can NOT be partially evaluated": {
			spec: &batches.BatchSpec{
				Steps: []batches.Step{
					{Run: "echo 1", If: `${{ eq outputs.value "foobar" }}`},
				},
			},
			wantSteps: []batches.Step{
				{Run: "echo 1", If: `${{ eq outputs.value "foobar" }}`},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			builder, err := NewTaskBuilder(tt.spec, nil)
			if err != nil {
				t.Fatal(err)
			}
			task, ok, err := builder.buildTask(repo, "", false)
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			if !ok {
				if tt.wantSteps != nil {
					t.Fatalf("no task built, but steps expected")
				}
				return
			}

			opts := cmpopts.IgnoreUnexported(batches.Step{})
			if diff := cmp.Diff(tt.wantSteps, task.Steps, opts); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTaskBuilder_BuildAll_Workspaces(t *testing.T) {
	repos := []*graphql.Repository{
		{ID: "repo-id-0", Name: "github.com/sourcegraph/automation-testing"},
		{ID: "repo-id-1", Name: "github.com/sourcegraph/sourcegraph"},
		{ID: "repo-id-2", Name: "bitbucket.sgdev.org/SOUR/automation-testing"},
	}
	steps := []batches.Step{{Run: "echo 1"}}

	type finderResults map[*graphql.Repository][]string

	type wantTask struct {
		Path               string
		ArchivePathToFetch string
	}

	tests := map[string]struct {
		spec          *batches.BatchSpec
		finderResults map[*graphql.Repository][]string

		wantNumTasks int

		// tasks per repository ID and in which path they are executed
		wantTasks map[string][]wantTask
	}{
		"no workspace configuration": {
			spec:          &batches.BatchSpec{Steps: steps},
			finderResults: finderResults{},
			wantNumTasks:  len(repos),
			wantTasks: map[string][]wantTask{
				repos[0].ID: {{Path: ""}},
				repos[1].ID: {{Path: ""}},
				repos[2].ID: {{Path: ""}},
			},
		},

		"workspace configuration matching no repos": {
			spec: &batches.BatchSpec{
				Steps: steps,
				Workspaces: []batches.WorkspaceConfiguration{
					{In: "this-does-not-match", RootAtLocationOf: "package.json"},
				},
			},
			finderResults: finderResults{},
			wantNumTasks:  len(repos),
			wantTasks: map[string][]wantTask{
				repos[0].ID: {{Path: ""}},
				repos[1].ID: {{Path: ""}},
				repos[2].ID: {{Path: ""}},
			},
		},

		"workspace configuration matching 2 repos with no results": {
			spec: &batches.BatchSpec{
				Steps: steps,
				Workspaces: []batches.WorkspaceConfiguration{
					{In: "*automation-testing", RootAtLocationOf: "package.json"},
				},
			},
			finderResults: finderResults{
				repos[0]: []string{},
				repos[2]: []string{},
			},
			wantNumTasks: 1,
			wantTasks: map[string][]wantTask{
				repos[1].ID: {{Path: ""}},
			},
		},

		"workspace configuration matching 2 repos with 3 results each": {
			spec: &batches.BatchSpec{
				Steps: steps,
				Workspaces: []batches.WorkspaceConfiguration{
					{In: "*automation-testing", RootAtLocationOf: "package.json"},
				},
			},
			finderResults: finderResults{
				repos[0]: {"a/b", "a/b/c", "d/e/f"},
				repos[2]: {"a/b", "a/b/c", "d/e/f"},
			},
			wantNumTasks: 7,
			wantTasks: map[string][]wantTask{
				repos[0].ID: {{Path: "a/b"}, {Path: "a/b/c"}, {Path: "d/e/f"}},
				repos[1].ID: {{Path: ""}},
				repos[2].ID: {{Path: "a/b"}, {Path: "a/b/c"}, {Path: "d/e/f"}},
			},
		},

		"workspace configuration matches repo with OnlyFetchWorkspace": {
			spec: &batches.BatchSpec{
				Steps: steps,
				Workspaces: []batches.WorkspaceConfiguration{
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
			wantNumTasks: 7,
			wantTasks: map[string][]wantTask{
				repos[0].ID: {
					{Path: "a/b", ArchivePathToFetch: "a/b"},
					{Path: "a/b/c", ArchivePathToFetch: "a/b/c"},
					{Path: "d/e/f", ArchivePathToFetch: "d/e/f"},
				},
				repos[1].ID: {{Path: ""}},
				repos[2].ID: {
					{Path: "a/b", ArchivePathToFetch: "a/b"},
					{Path: "a/b/c", ArchivePathToFetch: "a/b/c"},
					{Path: "d/e/f", ArchivePathToFetch: "d/e/f"},
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			finder := &mockDirectoryFinder{results: tt.finderResults}
			tb, err := NewTaskBuilder(tt.spec, finder)
			if err != nil {
				t.Fatal(err)
			}

			tasks, err := tb.BuildAll(context.Background(), repos)
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			if have := len(tasks); have != tt.wantNumTasks {
				t.Fatalf("wrong number of tasks. want=%d, got=%d", tt.wantNumTasks, have)
			}

			haveTasks := map[string][]wantTask{}
			for _, task := range tasks {
				haveTasks[task.Repository.ID] = append(haveTasks[task.Repository.ID], wantTask{
					Path:               task.Path,
					ArchivePathToFetch: task.ArchivePathToFetch(),
				})
			}

			for _, tasks := range haveTasks {
				sort.Slice(tasks, func(i, j int) bool { return tasks[i].Path < tasks[j].Path })
			}

			if diff := cmp.Diff(tt.wantTasks, haveTasks); diff != "" {
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

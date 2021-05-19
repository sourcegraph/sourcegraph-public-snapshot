package executor

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/batch-change-utils/overridable"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/git"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

func TestCreateChangesetSpecs(t *testing.T) {
	srcCLI := &graphql.Repository{
		ID:            "src-cli",
		Name:          "github.com/sourcegraph/src-cli",
		DefaultBranch: &graphql.Branch{Name: "main", Target: graphql.Target{OID: "d34db33f"}},
	}

	defaultChangesetSpec := &batches.ChangesetSpec{
		BaseRepository: srcCLI.ID,
		CreatedChangeset: &batches.CreatedChangeset{
			BaseRef:        srcCLI.DefaultBranch.Name,
			BaseRev:        srcCLI.DefaultBranch.Target.OID,
			HeadRepository: srcCLI.ID,
			HeadRef:        "refs/heads/my-branch",
			Title:          "The title",
			Body:           "The body",
			Commits: []batches.GitCommitDescription{
				{
					Message:     "git commit message",
					Diff:        "cool diff",
					AuthorName:  "Sourcegraph",
					AuthorEmail: "batch-changes@sourcegraph.com",
				},
			},
			Published: false,
		},
	}

	specWith := func(s *batches.ChangesetSpec, f func(s *batches.ChangesetSpec)) *batches.ChangesetSpec {
		f(s)
		return s
	}

	defaultTask := &Task{
		BatchChangeAttributes: &BatchChangeAttributes{
			Name:        "the name",
			Description: "The description",
		},
		Template: &batches.ChangesetTemplate{
			Title:  "The title",
			Body:   "The body",
			Branch: "my-branch",
			Commit: batches.ExpandedGitCommitDescription{
				Message: "git commit message",
			},
			Published: parsePublishedFieldString(t, "false"),
		},
		Repository: srcCLI,
	}

	taskWith := func(t *Task, f func(t *Task)) *Task {
		f(t)
		return t
	}

	defaultResult := executionResult{
		Diff: "cool diff",
		ChangedFiles: &git.Changes{
			Modified: []string{"README.md"},
		},
		Outputs: map[string]interface{}{},
	}

	tests := []struct {
		name   string
		task   *Task
		result executionResult

		want    []*batches.ChangesetSpec
		wantErr string
	}{
		{
			name:   "success",
			task:   defaultTask,
			result: defaultResult,
			want: []*batches.ChangesetSpec{
				defaultChangesetSpec,
			},
			wantErr: "",
		},
		{
			name: "publish by branch",
			task: taskWith(defaultTask, func(task *Task) {
				published := `[{"github.com/sourcegraph/*@my-branch": true}]`
				task.Template.Published = parsePublishedFieldString(t, published)
			}),
			result: defaultResult,
			want: []*batches.ChangesetSpec{
				specWith(defaultChangesetSpec, func(s *batches.ChangesetSpec) {
					s.Published = true
				}),
			},
			wantErr: "",
		},
		{
			name: "publish by branch not matching",
			task: taskWith(defaultTask, func(task *Task) {
				published := `[{"github.com/sourcegraph/*@another-branch-name": true}]`
				task.Template.Published = parsePublishedFieldString(t, published)
			}),
			result: defaultResult,
			want: []*batches.ChangesetSpec{
				specWith(defaultChangesetSpec, func(s *batches.ChangesetSpec) {
					s.Published = false
				}),
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			have, err := createChangesetSpecs(tt.task, tt.result, true)
			if err != nil {
				if tt.wantErr != "" {
					if err.Error() != tt.wantErr {
						t.Fatalf("wrong error. want=%q, got=%q", tt.wantErr, err.Error())
					}
					return
				} else {
					t.Fatalf("unexpected error: %s", err)
				}
			}

			if !cmp.Equal(tt.want, have) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.want, have))
			}
		})
	}
}

func TestGroupFileDiffs(t *testing.T) {
	diff1 := `diff --git 1/1.txt 1/1.txt
new file mode 100644
index 0000000..19d6416
--- /dev/null
+++ 1/1.txt
@@ -0,0 +1,1 @@
+this is 1
`
	diff2 := `diff --git 1/2/2.txt 1/2/2.txt
new file mode 100644
index 0000000..c825d65
--- /dev/null
+++ 1/2/2.txt
@@ -0,0 +1,1 @@
+this is 2
`
	diff3 := `diff --git 1/2/3/3.txt 1/2/3/3.txt
new file mode 100644
index 0000000..1bd79fb
--- /dev/null
+++ 1/2/3/3.txt
@@ -0,0 +1,1 @@
+this is 3
`

	defaultBranch := "my-default-branch"
	allDiffs := diff1 + diff2 + diff3

	tests := []struct {
		diff          string
		defaultBranch string
		groups        []batches.Group
		want          map[string]string
	}{
		{
			diff: allDiffs,
			groups: []batches.Group{
				{Directory: "1/2/3", Branch: "everything-in-3"},
			},
			want: map[string]string{
				"my-default-branch": diff1 + diff2,
				"everything-in-3":   diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []batches.Group{
				{Directory: "1/2", Branch: "everything-in-2-and-3"},
			},
			want: map[string]string{
				"my-default-branch":     diff1,
				"everything-in-2-and-3": diff2 + diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []batches.Group{
				{Directory: "1", Branch: "everything-in-1-and-2-and-3"},
			},
			want: map[string]string{
				"my-default-branch":           "",
				"everything-in-1-and-2-and-3": diff1 + diff2 + diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []batches.Group{
				// Each diff is matched against each directory, last match wins
				{Directory: "1", Branch: "only-in-1"},
				{Directory: "1/2", Branch: "only-in-2"},
				{Directory: "1/2/3", Branch: "only-in-3"},
			},
			want: map[string]string{
				"my-default-branch": "",
				"only-in-3":         diff3,
				"only-in-2":         diff2,
				"only-in-1":         diff1,
			},
		},
		{
			diff: allDiffs,
			groups: []batches.Group{
				// Last one wins here, because it matches every diff
				{Directory: "1/2/3", Branch: "only-in-3"},
				{Directory: "1/2", Branch: "only-in-2"},
				{Directory: "1", Branch: "only-in-1"},
			},
			want: map[string]string{
				"my-default-branch": "",
				"only-in-1":         diff1 + diff2 + diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []batches.Group{
				{Directory: "", Branch: "everything"},
			},
			want: map[string]string{
				"my-default-branch": diff1 + diff2 + diff3,
			},
		},
	}

	for _, tc := range tests {
		have, err := groupFileDiffs(tc.diff, defaultBranch, tc.groups)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !cmp.Equal(tc.want, have) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tc.want, have))
		}
	}
}

func TestValidateGroups(t *testing.T) {
	repoName := "github.com/sourcegraph/src-cli"
	defaultBranch := "my-batch-change"

	tests := []struct {
		defaultBranch string
		groups        []batches.Group
		wantErr       string
	}{
		{
			groups: []batches.Group{
				{Directory: "a", Branch: "my-batch-change-a"},
				{Directory: "b", Branch: "my-batch-change-b"},
			},
			wantErr: "",
		},
		{
			groups: []batches.Group{
				{Directory: "a", Branch: "my-batch-change-SAME"},
				{Directory: "b", Branch: "my-batch-change-SAME"},
			},
			wantErr: "transformChanges would lead to multiple changesets in repository github.com/sourcegraph/src-cli to have the same branch \"my-batch-change-SAME\"",
		},
		{
			groups: []batches.Group{
				{Directory: "a", Branch: "my-batch-change-SAME"},
				{Directory: "b", Branch: defaultBranch},
			},
			wantErr: "transformChanges group branch for repository github.com/sourcegraph/src-cli is the same as branch \"my-batch-change\" in changesetTemplate",
		},
	}

	for _, tc := range tests {
		err := validateGroups(repoName, defaultBranch, tc.groups)
		var haveErr string
		if err != nil {
			haveErr = err.Error()
		}

		if haveErr != tc.wantErr {
			t.Fatalf("wrong error:\nwant=%q\nhave=%q", tc.wantErr, haveErr)
		}
	}
}

func parsePublishedFieldString(t *testing.T, input string) overridable.BoolOrString {
	t.Helper()

	var result overridable.BoolOrString
	if err := json.Unmarshal([]byte(input), &result); err != nil {
		t.Fatalf("failed to parse %q as overridable.BoolOrString: %s", input, err)
	}
	return result
}

package executor

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/copystructure"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
	"github.com/sourcegraph/sourcegraph/lib/batches/overridable"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"

	"github.com/sourcegraph/src-cli/internal/batches"
)

func TestCreateChangesetSpecs(t *testing.T) {
	defaultChangesetSpec := &batcheslib.ChangesetSpec{
		BaseRepository: testRepo1.ID,

		BaseRef:        testRepo1.BaseRef(),
		BaseRev:        testRepo1.Rev(),
		HeadRepository: testRepo1.ID,
		HeadRef:        "refs/heads/my-branch",
		Title:          "The title",
		Body:           "The body",
		Commits: []batcheslib.GitCommitDescription{
			{
				Message:     "git commit message",
				Diff:        "cool diff",
				AuthorName:  "Sourcegraph",
				AuthorEmail: "batch-changes@sourcegraph.com",
			},
		},
		Published: batcheslib.PublishedValue{Val: false},
	}

	specWith := func(s *batcheslib.ChangesetSpec, f func(s *batcheslib.ChangesetSpec)) *batcheslib.ChangesetSpec {
		copy, err := copystructure.Copy(s)
		if err != nil {
			t.Fatalf("deep copying spec: %+v", err)
		}

		s = copy.(*batcheslib.ChangesetSpec)
		f(s)
		return s
	}

	defaultTask := &Task{
		BatchChangeAttributes: &template.BatchChangeAttributes{
			Name:        "the name",
			Description: "The description",
		},
		Template: &batcheslib.ChangesetTemplate{
			Title:  "The title",
			Body:   "The body",
			Branch: "my-branch",
			Commit: batcheslib.ExpandedGitCommitDescription{
				Message: "git commit message",
			},
			Published: parsePublishedFieldString(t, "false"),
		},
		Repository: testRepo1,
	}

	taskWith := func(task *Task, f func(task *Task)) *Task {
		copy, err := copystructure.Copy(task)
		if err != nil {
			t.Fatalf("deep copying task: %+v", err)
		}

		task = copy.(*Task)
		f(task)
		return task
	}

	defaultResult := executionResult{
		Diff: "cool diff",
		ChangedFiles: &git.Changes{
			Modified: []string{"README.md"},
		},
		Outputs: map[string]interface{}{},
	}

	featuresWithoutOptionalPublished := featuresAllEnabled()
	featuresWithoutOptionalPublished.AllowOptionalPublished = false

	tests := []struct {
		name     string
		task     *Task
		features batches.FeatureFlags
		result   executionResult

		want    []*batcheslib.ChangesetSpec
		wantErr string
	}{
		{
			name:     "success",
			task:     defaultTask,
			features: featuresAllEnabled(),
			result:   defaultResult,
			want: []*batcheslib.ChangesetSpec{
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
			features: featuresAllEnabled(),
			result:   defaultResult,
			want: []*batcheslib.ChangesetSpec{
				specWith(defaultChangesetSpec, func(s *batcheslib.ChangesetSpec) {
					s.Published = batcheslib.PublishedValue{Val: true}
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
			features: featuresAllEnabled(),
			result:   defaultResult,
			want: []*batcheslib.ChangesetSpec{
				specWith(defaultChangesetSpec, func(s *batcheslib.ChangesetSpec) {
					s.Published = batcheslib.PublishedValue{Val: nil}
				}),
			},
			wantErr: "",
		},
		{
			name: "publish by branch not matching on an old Sourcegraph version",
			task: taskWith(defaultTask, func(task *Task) {
				published := `[{"github.com/sourcegraph/*@another-branch-name": true}]`
				task.Template.Published = parsePublishedFieldString(t, published)
			}),
			features: featuresWithoutOptionalPublished,
			result:   defaultResult,
			want: []*batcheslib.ChangesetSpec{
				specWith(defaultChangesetSpec, func(s *batcheslib.ChangesetSpec) {
					s.Published = batcheslib.PublishedValue{Val: false}
				}),
			},
			wantErr: "",
		},
		{
			name: "publish in UI on a supported version",
			task: taskWith(defaultTask, func(task *Task) {
				task.Template.Published = nil
			}),
			features: featuresAllEnabled(),
			result:   defaultResult,
			want: []*batcheslib.ChangesetSpec{
				specWith(defaultChangesetSpec, func(s *batcheslib.ChangesetSpec) {
					s.Published = batcheslib.PublishedValue{Val: nil}
				}),
			},
			wantErr: "",
		},
		{
			name: "publish in UI on an unsupported version",
			task: taskWith(defaultTask, func(task *Task) {
				task.Template.Published = nil
			}),
			features: featuresWithoutOptionalPublished,
			result:   defaultResult,
			want:     nil,
			wantErr:  errOptionalPublishedUnsupported.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			have, err := createChangesetSpecs(tt.task, tt.result, tt.features)
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
		groups        []batcheslib.Group
		want          map[string]string
	}{
		{
			diff: allDiffs,
			groups: []batcheslib.Group{
				{Directory: "1/2/3", Branch: "everything-in-3"},
			},
			want: map[string]string{
				"my-default-branch": diff1 + diff2,
				"everything-in-3":   diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []batcheslib.Group{
				{Directory: "1/2", Branch: "everything-in-2-and-3"},
			},
			want: map[string]string{
				"my-default-branch":     diff1,
				"everything-in-2-and-3": diff2 + diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []batcheslib.Group{
				{Directory: "1", Branch: "everything-in-1-and-2-and-3"},
			},
			want: map[string]string{
				"my-default-branch":           "",
				"everything-in-1-and-2-and-3": diff1 + diff2 + diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []batcheslib.Group{
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
			groups: []batcheslib.Group{
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
			groups: []batcheslib.Group{
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
		groups        []batcheslib.Group
		wantErr       string
	}{
		{
			groups: []batcheslib.Group{
				{Directory: "a", Branch: "my-batch-change-a"},
				{Directory: "b", Branch: "my-batch-change-b"},
			},
			wantErr: "",
		},
		{
			groups: []batcheslib.Group{
				{Directory: "a", Branch: "my-batch-change-SAME"},
				{Directory: "b", Branch: "my-batch-change-SAME"},
			},
			wantErr: "transformChanges would lead to multiple changesets in repository github.com/sourcegraph/src-cli to have the same branch \"my-batch-change-SAME\"",
		},
		{
			groups: []batcheslib.Group{
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

func parsePublishedFieldString(t *testing.T, input string) *overridable.BoolOrString {
	t.Helper()

	var result overridable.BoolOrString
	if err := json.Unmarshal([]byte(input), &result); err != nil {
		t.Fatalf("failed to parse %q as overridable.BoolOrString: %s", input, err)
	}
	return &result
}

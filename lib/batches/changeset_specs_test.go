package batches

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/copystructure"

	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
	"github.com/sourcegraph/sourcegraph/lib/batches/overridable"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
)

func TestCreateChangesetSpecs(t *testing.T) {
	defaultChangesetSpec := &ChangesetSpec{
		BaseRepository: "base-repo-id",
		BaseRef:        "refs/heads/my-cool-base-ref",
		BaseRev:        "f00b4r",
		// This field is deprecated and should always match BaseRepository.
		HeadRepository: "base-repo-id",
		HeadRef:        "refs/heads/my-branch",

		Title: "The title",
		Body:  "The body",
		Commits: []GitCommitDescription{
			{
				Message:     "git commit message",
				Diff:        "cool diff",
				AuthorName:  "Sourcegraph",
				AuthorEmail: "batch-changes@sourcegraph.com",
			},
		},
		Published: PublishedValue{Val: false},
	}

	specWith := func(s *ChangesetSpec, f func(s *ChangesetSpec)) *ChangesetSpec {
		copy, err := copystructure.Copy(s)
		if err != nil {
			t.Fatalf("deep copying spec: %+v", err)
		}

		s = copy.(*ChangesetSpec)
		f(s)
		return s
	}

	defaultInput := &ChangesetSpecInput{
		Repository: Repository{
			ID:          "base-repo-id",
			Name:        "github.com/sourcegraph/src-cli",
			FileMatches: []string{"go.mod", "README"},
			BaseRef:     "refs/heads/my-cool-base-ref",
			BaseRev:     "f00b4r",
		},

		BatchChangeAttributes: &template.BatchChangeAttributes{
			Name:        "the name",
			Description: "The description",
		},

		Template: &ChangesetTemplate{
			Title:  "The title",
			Body:   "The body",
			Branch: "my-branch",
			Commit: ExpandedGitCommitDescription{
				Message: "git commit message",
			},
			Published: parsePublishedFieldString(t, "false"),
		},

		Result: execution.AfterStepResult{
			Diff: "cool diff",
			ChangedFiles: git.Changes{
				Modified: []string{"README.md"},
			},
			Outputs: map[string]any{},
		},
	}

	inputWith := func(task *ChangesetSpecInput, f func(task *ChangesetSpecInput)) *ChangesetSpecInput {
		copy, err := copystructure.Copy(task)
		if err != nil {
			t.Fatalf("deep copying task: %+v", err)
		}

		task = copy.(*ChangesetSpecInput)
		f(task)
		return task
	}

	tests := []struct {
		name string

		input *ChangesetSpecInput

		want    []*ChangesetSpec
		wantErr string
	}{
		{
			name:  "success",
			input: defaultInput,
			want: []*ChangesetSpec{
				defaultChangesetSpec,
			},
			wantErr: "",
		},
		{
			name: "publish by branch",
			input: inputWith(defaultInput, func(input *ChangesetSpecInput) {
				published := `[{"github.com/sourcegraph/*@my-branch": true}]`
				input.Template.Published = parsePublishedFieldString(t, published)
			}),
			want: []*ChangesetSpec{
				specWith(defaultChangesetSpec, func(s *ChangesetSpec) {
					s.Published = PublishedValue{Val: true}
				}),
			},
			wantErr: "",
		},
		{
			name: "publish by branch not matching",
			input: inputWith(defaultInput, func(input *ChangesetSpecInput) {
				published := `[{"github.com/sourcegraph/*@another-branch-name": true}]`
				input.Template.Published = parsePublishedFieldString(t, published)
			}),
			want: []*ChangesetSpec{
				specWith(defaultChangesetSpec, func(s *ChangesetSpec) {
					s.Published = PublishedValue{Val: nil}
				}),
			},
			wantErr: "",
		},
		{
			name: "publish in UI",
			input: inputWith(defaultInput, func(input *ChangesetSpecInput) {
				input.Template.Published = nil
			}),
			want: []*ChangesetSpec{
				specWith(defaultChangesetSpec, func(s *ChangesetSpec) {
					s.Published = PublishedValue{Val: nil}
				}),
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			have, err := BuildChangesetSpecs(tt.input)
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
		groups        []Group
		want          map[string]string
	}{
		{
			diff: allDiffs,
			groups: []Group{
				{Directory: "1/2/3", Branch: "everything-in-3"},
			},
			want: map[string]string{
				"my-default-branch": diff1 + diff2,
				"everything-in-3":   diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []Group{
				{Directory: "1/2", Branch: "everything-in-2-and-3"},
			},
			want: map[string]string{
				"my-default-branch":     diff1,
				"everything-in-2-and-3": diff2 + diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []Group{
				{Directory: "1", Branch: "everything-in-1-and-2-and-3"},
			},
			want: map[string]string{
				"my-default-branch":           "",
				"everything-in-1-and-2-and-3": diff1 + diff2 + diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []Group{
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
			groups: []Group{
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
			groups: []Group{
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
		groups        []Group
		wantErr       string
	}{
		{
			groups: []Group{
				{Directory: "a", Branch: "my-batch-change-a"},
				{Directory: "b", Branch: "my-batch-change-b"},
			},
			wantErr: "",
		},
		{
			groups: []Group{
				{Directory: "a", Branch: "my-batch-change-SAME"},
				{Directory: "b", Branch: "my-batch-change-SAME"},
			},
			wantErr: "transformChanges would lead to multiple changesets in repository github.com/sourcegraph/src-cli to have the same branch \"my-batch-change-SAME\"",
		},
		{
			groups: []Group{
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

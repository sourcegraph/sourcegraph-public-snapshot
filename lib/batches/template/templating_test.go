package template

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TODO: Is renamed_files intentionally omitted from the docs?
func TestValidateBatchSpecTemplate(t *testing.T) {
	tests := []struct {
		name      string
		batchSpec string
		wantValid bool
		wantErr   error
	}{
		{
			name: "full batch spec, all valid template variables",
			batchSpec: `name: valid-batch-spec
				on:
				- repository: github.com/fake/fake

				steps:
				- run: |
						${{ repository.search_result_paths }}
						${{ repository.name }}
						${{ batch_change.name }}
						${{ batch_change.description }}
						${{ previous_step.modified_files }}
						${{ previous_step.added_files }}
						${{ previous_step.deleted_files }}
						${{ previous_step.renamed_files }}
						${{ previous_step.stdout }}
						${{ previous_step.stderr}}
						${{ step.modified_files }}
						${{ step.added_files }}
						${{ step.deleted_files }}
						${{ step.renamed_files }}
						${{ step.stdout}}
						${{ step.stderr}}
						${{ steps.modified_files }}
						${{ steps.added_files }}
						${{ steps.deleted_files }}
						${{ steps.renamed_files }}
						${{ steps.path }}
					container: my-container

				changesetTemplate:
				title: |
					${{ repository.search_result_paths }}
					${{ repository.name }}
					${{ repository.branch }}
					${{ batch_change.name }}
					${{ batch_change.description }}
					${{ steps.modified_files }}
					${{ steps.added_files }}
					${{ steps.deleted_files }}
					${{ steps.renamed_files }}
					${{ steps.path }}
					${{ batch_change_link }}
					body: I'm a changeset yay!
					branch: my-branch
					commit:
						message: I'm a changeset yay!
					`,
			wantValid: true,
		},
		{
			name: "valid template helpers",
			batchSpec: `${{ join repository.search_result_paths "\n" }}
				${{ join_if "---" "a" "b" "" "d" }}
				${{ replace "a/b/c/d" "/" "-" }}
				${{ split repository.name "/" }}
				${{ matches repository.name "github.com/my-org/terra*" }}
				${{ index steps.modified_files 1 }}`,
			wantValid: true,
		},
		{
			name:      "invalid step template variable",
			batchSpec: `${{ resipotory.search_result_paths }}`,
			wantValid: false,
			wantErr:   errors.New("validating batch spec template: unknown templating variable: 'resipotory'"),
		},
		{
			name:      "invalid step template variable, 1 level nested",
			batchSpec: `${{ repository.search_resalt_paths }}`,
			wantValid: false,
			wantErr:   errors.New("validating batch spec template: unknown templating variable: 'repository.search_resalt_paths'"),
		},
		{
			name:      "invalid changeset template variable",
			batchSpec: `${{ batch_chang_link }}`,
			wantValid: false,
			wantErr:   errors.New("validating batch spec template: unknown templating variable: 'batch_chang_link'"),
		},
		{
			name:      "invalid changeset template variable, 1 level nested",
			batchSpec: `${{ steps.mofidied_files }}`,
			wantValid: false,
			wantErr:   errors.New("validating batch spec template: unknown templating variable: 'steps.mofidied_files'"),
		},
		{
			name:      "escaped templating (github expression syntax) is ignored",
			batchSpec: `${{ "${{ ignore_me }}" }}`,
			wantValid: true,
		},
		{
			name: "output variables are ignored",
			batchSpec: `${{ outputs.IDontExist }}
						${{OUTPUTS.anotherOne}}
						${{ join outputs.myArray "," }}
						${{ index outputs.env.something 1 }}`,
			wantValid: true,
		},
		{
			name:      "output variables are ignored, but invalid step template variable still fails",
			batchSpec: `${{ outputs.unknown }} ${{ outputz.unknown }}`,
			wantValid: false,
			wantErr:   errors.New("validating batch spec template: unknown templating variable: 'outputz'"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotValid, gotErr := ValidateBatchSpecTemplate(tc.batchSpec)

			if tc.wantValid != gotValid {
				t.Fatalf("unexpected valid status. want valid=%t, got valid=%t\nerror message: %s", tc.wantValid, gotValid, gotErr)
			}

			if tc.wantErr == nil && gotErr != nil {
				t.Fatalf("unexpected non-nil error.\nwant=nil\n---\ngot=%s", gotErr)
			}

			if tc.wantErr != nil && gotErr == nil {
				t.Fatalf("unexpected nil error.\nwant=%s\n---\ngot=nil", tc.wantErr)
			}

			if tc.wantErr != nil && gotErr != nil && tc.wantErr.Error() != gotErr.Error() {
				t.Fatalf("unexpected error message\nwant=%s\n---\ngot=%s", tc.wantErr, gotErr)
			}
		})
	}
}

var testChanges = git.Changes{
	Modified: []string{"go.mod"},
	Added:    []string{"main.go.swp"},
	Deleted:  []string{".DS_Store"},
	Renamed:  []string{"new-filename.txt"},
}

func TestEvalStepCondition(t *testing.T) {
	stepCtx := &StepContext{
		BatchChange: BatchChangeAttributes{
			Name:        "test-batch-change",
			Description: "This batch change is just an experiment",
		},
		PreviousStep: execution.AfterStepResult{
			ChangedFiles: testChanges,
			Stdout:       "this is previous step's stdout",
			Stderr:       "this is previous step's stderr",
		},
		Steps: StepsContext{
			Changes: testChanges,
			Path:    "sub/directory/of/repo",
		},
		Outputs: map[string]any{},
		// Step is not set when evalStepCondition is called
		Repository: *testRepo1,
	}

	tests := []struct {
		run  string
		want bool
	}{
		{run: `true`, want: true},
		{run: `  true    `, want: true},
		{run: `TRUE`, want: false},
		{run: `false`, want: false},
		{run: `FALSE`, want: false},
		{run: `${{ eq repository.name "github.com/sourcegraph/src-cli" }}`, want: true},
		{run: `${{ eq steps.path "sub/directory/of/repo" }}`, want: true},
		{run: `${{ matches repository.name "github.com/sourcegraph/*" }}`, want: true},
	}

	for _, tc := range tests {
		got, err := EvalStepCondition(tc.run, stepCtx)
		if err != nil {
			t.Fatal(err)
		}

		if got != tc.want {
			t.Fatalf("wrong value. want=%t, got=%t", tc.want, got)
		}
	}
}

const rawYaml = `dist: release
env:
  - GO111MODULE=on
  - CGO_ENABLED=0
before:
  hooks:
    - go mod download
    - go mod tidy
    - go generate ./schema
`

func TestRenderStepTemplate(t *testing.T) {
	// To avoid bugs due to differences between test setup and actual code, we
	// do the actual parsing of YAML here to get an interface{} which we'll put
	// in the StepContext.
	var parsedYaml any
	if err := yaml.Unmarshal([]byte(rawYaml), &parsedYaml); err != nil {
		t.Fatalf("failed to parse YAML: %s", err)
	}

	stepCtx := &StepContext{
		BatchChange: BatchChangeAttributes{
			Name:        "test-batch-change",
			Description: "This batch change is just an experiment",
		},
		PreviousStep: execution.AfterStepResult{
			ChangedFiles: testChanges,
			Stdout:       "this is previous step's stdout",
			Stderr:       "this is previous step's stderr",
		},
		Outputs: map[string]any{
			"lastLine": "lastLine is this",
			"project":  parsedYaml,
		},
		Step: execution.AfterStepResult{
			ChangedFiles: testChanges,
			Stdout:       "this is current step's stdout",
			Stderr:       "this is current step's stderr",
		},
		Steps:      StepsContext{Changes: testChanges, Path: "sub/directory/of/repo"},
		Repository: *testRepo1,
	}

	tests := []struct {
		name    string
		stepCtx *StepContext
		run     string
		want    string
	}{
		{
			name:    "lower-case aliases",
			stepCtx: stepCtx,
			run: `${{ repository.search_result_paths }}
${{ repository.name }}
${{ batch_change.name }}
${{ batch_change.description }}
${{ previous_step.modified_files }}
${{ previous_step.added_files }}
${{ previous_step.deleted_files }}
${{ previous_step.renamed_files }}
${{ previous_step.stdout }}
${{ previous_step.stderr}}
${{ outputs.lastLine }}
${{ index outputs.project.env 1 }}
${{ step.modified_files }}
${{ step.added_files }}
${{ step.deleted_files }}
${{ step.renamed_files }}
${{ step.stdout}}
${{ step.stderr}}
${{ steps.modified_files }}
${{ steps.added_files }}
${{ steps.deleted_files }}
${{ steps.renamed_files }}
${{ steps.path }}
`,
			want: `README.md main.go
github.com/sourcegraph/src-cli
test-batch-change
This batch change is just an experiment
[go.mod]
[main.go.swp]
[.DS_Store]
[new-filename.txt]
this is previous step's stdout
this is previous step's stderr
lastLine is this
CGO_ENABLED=0
[go.mod]
[main.go.swp]
[.DS_Store]
[new-filename.txt]
this is current step's stdout
this is current step's stderr
[go.mod]
[main.go.swp]
[.DS_Store]
[new-filename.txt]
sub/directory/of/repo
`,
		},
		{
			name:    "empty context",
			stepCtx: &StepContext{},
			run: `${{ repository.search_result_paths }}
${{ repository.name }}
${{ previous_step.modified_files }}
${{ previous_step.added_files }}
${{ previous_step.deleted_files }}
${{ previous_step.renamed_files }}
${{ previous_step.stdout }}
${{ previous_step.stderr}}
${{ step.modified_files }}
${{ step.added_files }}
${{ step.deleted_files }}
${{ step.renamed_files }}
${{ step.stdout}}
${{ step.stderr}}
${{ steps.modified_files }}
${{ steps.added_files }}
${{ steps.deleted_files }}
${{ steps.renamed_files }}
${{ steps.path }}
`,
			want: `

[]
[]
[]
[]


[]
[]
[]
[]


[]
[]
[]
[]

`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer

			err := RenderStepTemplate("testing", tc.run, &out, tc.stepCtx)
			if err != nil {
				t.Fatal(err)
			}

			if out.String() != tc.want {
				t.Fatalf("wrong output:\n%s", cmp.Diff(tc.want, out.String()))
			}
		})
	}
}

func TestRenderStepMap(t *testing.T) {
	stepCtx := &StepContext{
		PreviousStep: execution.AfterStepResult{
			ChangedFiles: testChanges,
			Stdout:       "this is previous step's stdout",
			Stderr:       "this is previous step's stderr",
		},
		Outputs:    map[string]any{},
		Repository: *testRepo1,
	}

	input := map[string]string{
		"/tmp/my-file.txt":        `${{ previous_step.modified_files }}`,
		"/tmp/my-other-file.txt":  `${{ previous_step.added_files }}`,
		"/tmp/my-other-file2.txt": `${{ previous_step.deleted_files }}`,
	}

	have, err := RenderStepMap(input, stepCtx)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	want := map[string]string{
		"/tmp/my-file.txt":        "[go.mod]",
		"/tmp/my-other-file.txt":  "[main.go.swp]",
		"/tmp/my-other-file2.txt": "[.DS_Store]",
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("wrong output:\n%s", diff)
	}
}

func TestRenderChangesetTemplateField(t *testing.T) {
	// To avoid bugs due to differences between test setup and actual code, we
	// do the actual parsing of YAML here to get an interface{} which we'll put
	// in the StepContext.
	var parsedYaml any
	if err := yaml.Unmarshal([]byte(rawYaml), &parsedYaml); err != nil {
		t.Fatalf("failed to parse YAML: %s", err)
	}

	tmplCtx := &ChangesetTemplateContext{
		BatchChangeAttributes: BatchChangeAttributes{
			Name:        "test-batch-change",
			Description: "This batch change is just an experiment",
		},
		Outputs: map[string]any{
			"lastLine": "lastLine is this",
			"project":  parsedYaml,
		},
		Repository: *testRepo1,
		Steps: StepsContext{
			Changes: git.Changes{
				Modified: []string{"modified-file.txt"},
				Added:    []string{"added-file.txt"},
				Deleted:  []string{"deleted-file.txt"},
				Renamed:  []string{"renamed-file.txt"},
			},
			Path: "infrastructure/sub-project",
		},
	}

	tests := []struct {
		name    string
		tmplCtx *ChangesetTemplateContext
		tmpl    string
		want    string
	}{
		{
			name:    "lower-case aliases",
			tmplCtx: tmplCtx,
			tmpl: `${{ repository.search_result_paths }}
${{ repository.name }}
${{ batch_change.name }}
${{ batch_change.description }}
${{ outputs.lastLine }}
${{ index outputs.project.env 1 }}
${{ steps.modified_files }}
${{ steps.added_files }}
${{ steps.deleted_files }}
${{ steps.renamed_files }}
${{ steps.path }}
${{ batch_change_link }}
`,
			want: `README.md main.go
github.com/sourcegraph/src-cli
test-batch-change
This batch change is just an experiment
lastLine is this
CGO_ENABLED=0
[modified-file.txt]
[added-file.txt]
[deleted-file.txt]
[renamed-file.txt]
infrastructure/sub-project
${{ batch_change_link }}`,
		},
		{
			name:    "empty context",
			tmplCtx: &ChangesetTemplateContext{},
			tmpl: `${{ repository.search_result_paths }}
${{ repository.name }}
${{ steps.modified_files }}
${{ steps.added_files }}
${{ steps.deleted_files }}
${{ steps.renamed_files }}
${{ batch_change_link }}
`,
			want: `[]
[]
[]
[]
${{ batch_change_link }}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := RenderChangesetTemplateField("testing", tc.tmpl, tc.tmplCtx)
			if err != nil {
				t.Fatal(err)
			}

			if out != tc.want {
				t.Fatalf("wrong output:\n%s", cmp.Diff(tc.want, out))
			}
		})
	}
}

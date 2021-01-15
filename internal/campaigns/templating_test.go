package campaigns

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
	"gopkg.in/yaml.v3"
)

func TestParseGitStatus(t *testing.T) {
	const input = `M  README.md
M  another_file.go
A  new_file.txt
A  barfoo/new_file.txt
D  to_be_deleted.txt
R  README.md -> README.markdown
`
	parsed, err := parseGitStatus([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	want := StepChanges{
		Modified: []string{"README.md", "another_file.go"},
		Added:    []string{"new_file.txt", "barfoo/new_file.txt"},
		Deleted:  []string{"to_be_deleted.txt"},
		Renamed:  []string{"README.markdown"},
	}

	if !cmp.Equal(want, parsed) {
		t.Fatalf("wrong output:\n%s", cmp.Diff(want, parsed))
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
	var parsedYaml interface{}
	if err := yaml.Unmarshal([]byte(rawYaml), &parsedYaml); err != nil {
		t.Fatalf("failed to parse YAML: %s", err)
	}

	stepCtx := &StepContext{
		PreviousStep: StepResult{
			files: &StepChanges{
				Modified: []string{"go.mod"},
				Added:    []string{"main.go.swp"},
				Deleted:  []string{".DS_Store"},
				Renamed:  []string{"new-filename.txt"},
			},
			Stdout: bytes.NewBufferString("this is previous step's stdout"),
			Stderr: bytes.NewBufferString("this is previous step's stderr"),
		},
		Outputs: map[string]interface{}{
			"lastLine": "lastLine is this",
			"project":  parsedYaml,
		},
		Step: StepResult{
			files: &StepChanges{
				Modified: []string{"step-go.mod"},
				Added:    []string{"step-main.go.swp"},
				Deleted:  []string{"step-.DS_Store"},
				Renamed:  []string{"step-new-filename.txt"},
			},
			Stdout: bytes.NewBufferString("this is current step's stdout"),
			Stderr: bytes.NewBufferString("this is current step's stderr"),
		},
		Repository: graphql.Repository{
			Name: "github.com/sourcegraph/src-cli",
			FileMatches: map[string]bool{
				"README.md": true,
				"main.go":   true,
			},
		},
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
`,
			want: `README.md main.go
github.com/sourcegraph/src-cli
[go.mod]
[main.go.swp]
[.DS_Store]
[new-filename.txt]
this is previous step's stdout
this is previous step's stderr
lastLine is this
CGO_ENABLED=0
[step-go.mod]
[step-main.go.swp]
[step-.DS_Store]
[step-new-filename.txt]
this is current step's stdout
this is current step's stderr
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
${{ outputs.lastLine }}
${{ outputs.project }}
${{ step.modified_files }}
${{ step.added_files }}
${{ step.deleted_files }}
${{ step.renamed_files }}
${{ step.stdout}}
${{ step.stderr}}
`,
			want: `

[]
[]
[]
[]


<no value>
<no value>
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

			err := renderStepTemplate("testing", tc.run, &out, tc.stepCtx)
			if err != nil {
				t.Fatal(err)
			}

			if out.String() != tc.want {
				t.Fatalf("wrong output:\n%s", cmp.Diff(tc.want, out.String()))
			}
		})
	}
}

func TestRenderChangesetTemplateField(t *testing.T) {
	// To avoid bugs due to differences between test setup and actual code, we
	// do the actual parsing of YAML here to get an interface{} which we'll put
	// in the StepContext.
	var parsedYaml interface{}
	if err := yaml.Unmarshal([]byte(rawYaml), &parsedYaml); err != nil {
		t.Fatalf("failed to parse YAML: %s", err)
	}

	tmplCtx := &ChangesetTemplateContext{
		Outputs: map[string]interface{}{
			"lastLine": "lastLine is this",
			"project":  parsedYaml,
		},
		Repository: graphql.Repository{
			Name: "github.com/sourcegraph/src-cli",
			FileMatches: map[string]bool{
				"README.md": true,
				"main.go":   true,
			},
		},
		Steps: &StepChanges{
			Modified: []string{"modified-file.txt"},
			Added:    []string{"added-file.txt"},
			Deleted:  []string{"deleted-file.txt"},
			Renamed:  []string{"renamed-file.txt"},
		},
	}

	tests := []struct {
		name    string
		tmplCtx *ChangesetTemplateContext
		run     string
		want    string
	}{
		{
			name:    "lower-case aliases",
			tmplCtx: tmplCtx,
			run: `${{ repository.search_result_paths }}
${{ repository.name }}
${{ outputs.lastLine }}
${{ index outputs.project.env 1 }}
${{ steps.modified_files }}
${{ steps.added_files }}
${{ steps.deleted_files }}
${{ steps.renamed_files }}
`,
			want: `README.md main.go
github.com/sourcegraph/src-cli
lastLine is this
CGO_ENABLED=0
[modified-file.txt]
[added-file.txt]
[deleted-file.txt]
[renamed-file.txt]
`,
		},
		{
			name:    "empty context",
			tmplCtx: &ChangesetTemplateContext{},
			run: `${{ repository.search_result_paths }}
${{ repository.name }}
${{ outputs.lastLine }}
${{ outputs.project }}
${{ steps.modified_files }}
${{ steps.added_files }}
${{ steps.deleted_files }}
${{ steps.renamed_files }}
`,
			want: `

<no value>
<no value>
[]
[]
[]
[]
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := renderChangesetTemplateField("testing", tc.run, tc.tmplCtx)
			if err != nil {
				t.Fatal(err)
			}

			if out != tc.want {
				t.Fatalf("wrong output:\n%s", cmp.Diff(tc.want, out))
			}
		})
	}
}

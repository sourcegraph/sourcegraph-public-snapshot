package campaigns

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
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

func TestParsingAndRenderingTemplates(t *testing.T) {
	stepCtx := &StepContext{
		PreviousStep: StepResult{
			files: &StepChanges{
				Modified: []string{"go.mod"},
				Added:    []string{"main.go.swp"},
				Deleted:  []string{".DS_Store"},
				Renamed:  []string{"new-filename.txt"},
			},
			Stdout: bytes.NewBufferString("this is stdout"),
			Stderr: bytes.NewBufferString("this is stderr"),
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
			name:    "previous step file changes",
			stepCtx: stepCtx,
			run: `${{ .PreviousStep.ModifiedFiles }}
${{ .PreviousStep.AddedFiles }}
${{ .PreviousStep.DeletedFiles }}
${{ .PreviousStep.RenamedFiles }}
`,
			want: `[go.mod]
[main.go.swp]
[.DS_Store]
[new-filename.txt]
`,
		},
		{
			name:    "previous step output",
			stepCtx: stepCtx,
			run:     `${{ .PreviousStep.Stdout }} ${{ .PreviousStep.Stderr }}`,
			want:    `this is stdout this is stderr`,
		},
		{
			name:    "repository name",
			stepCtx: stepCtx,
			run:     `${{ .Repository.Name }}`,
			want:    `github.com/sourcegraph/src-cli`,
		},
		{
			name:    "search result paths",
			stepCtx: stepCtx,
			run:     `${{ .Repository.SearchResultPaths }}`,
			want:    `README.md main.go`,
		},
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
		`,
			want: `README.md main.go
		github.com/sourcegraph/src-cli
		[go.mod]
		[main.go.swp]
		[.DS_Store]
		[new-filename.txt]
		this is stdout
		this is stderr
		`,
		},
		{
			name:    "empty context",
			stepCtx: &StepContext{},
			run: `${{ .Repository.SearchResultPaths }}
${{ .Repository.Name }}
${{ previous_step.modified_files }}
${{ previous_step.added_files }}
${{ previous_step.deleted_files }}
${{ previous_step.renamed_files }}
${{ previous_step.stdout }}
${{ previous_step.stderr}}
`,
			want: `

[]
[]
[]
[]


`,
		},
		{
			name:    "empty context and aliases",
			stepCtx: &StepContext{},
			run: `${{ repository.search_result_paths }}
${{ repository.name }}
${{ previous_step.modified_files }}
${{ previous_step.added_files }}
${{ previous_step.deleted_files }}
${{ previous_step.renamed_files }}
${{ previous_step.stdout }}
${{ previous_step.stderr}}
`,
			want: `

[]
[]
[]
[]


`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := parseAsTemplate("testing", tc.run, tc.stepCtx)
			if err != nil {
				t.Fatal(err)
			}

			var out bytes.Buffer
			if err := parsed.Execute(&out, tc.stepCtx); err != nil {
				t.Fatalf("executing template failed: %s", err)
			}

			if out.String() != tc.want {
				t.Fatalf("wrong output:\n%s", cmp.Diff(tc.want, out.String()))
			}
		})
	}
}

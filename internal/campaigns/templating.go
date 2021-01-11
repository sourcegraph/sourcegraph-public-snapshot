package campaigns

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

func renderTemplate(name, tmpl string, out io.Writer, stepCtx *StepContext) error {
	t, err := parseAsTemplate(name, tmpl, stepCtx)
	if err != nil {
		return errors.Wrap(err, "parsing step run")
	}

	return t.Execute(out, stepCtx)
}

func parseAsTemplate(name, input string, stepCtx *StepContext) (*template.Template, error) {
	return template.New(name).Delims("${{", "}}").Funcs(stepCtx.ToFuncMap()).Parse(input)
}

func renderMap(m map[string]string, stepCtx *StepContext) (map[string]string, error) {
	rendered := make(map[string]string, len(m))

	for k, v := range rendered {
		var out bytes.Buffer

		if err := renderTemplate(k, v, &out, stepCtx); err != nil {
			return rendered, err
		}

		rendered[k] = out.String()
	}

	return rendered, nil
}

// StepContext represents the contextual information available when executing a
// step that's defined in a campaign spec.
type StepContext struct {
	PreviousStep StepResult
	Repository   graphql.Repository
}

// ToFuncMap returns a template.FuncMap to access fields on the StepContext in a
// text/template.
func (stepCtx *StepContext) ToFuncMap() template.FuncMap {
	return template.FuncMap{
		"join": func(list []string, sep string) string {
			return strings.Join(list, sep)
		},
		"split": func(s string, sep string) []string {
			return strings.Split(s, sep)
		},
		"previous_step": func() map[string]interface{} {
			result := map[string]interface{}{
				"modified_files": stepCtx.PreviousStep.ModifiedFiles(),
				"added_files":    stepCtx.PreviousStep.AddedFiles(),
				"deleted_files":  stepCtx.PreviousStep.DeletedFiles(),
				"renamed_files":  stepCtx.PreviousStep.RenamedFiles(),
			}

			if stepCtx.PreviousStep.Stdout != nil {
				result["stdout"] = stepCtx.PreviousStep.Stdout.String()
			} else {
				result["stdout"] = ""
			}

			if stepCtx.PreviousStep.Stderr != nil {
				result["stderr"] = stepCtx.PreviousStep.Stderr.String()
			} else {
				result["stderr"] = ""
			}

			return result
		},
		"repository": func() map[string]interface{} {
			return map[string]interface{}{
				"search_result_paths": stepCtx.Repository.SearchResultPaths(),
				"name":                stepCtx.Repository.Name,
			}
		},
	}
}

// StepResult represents the result of a previously executed step.
type StepResult struct {
	// files are the changes made to files by the step.
	files *StepChanges

	// Stdout is the output produced by the step on standard out.
	Stdout *bytes.Buffer
	// Stderr is the output produced by the step on standard error.
	Stderr *bytes.Buffer
}

// StepChanges are the changes made to files by a previous step in a repository.
type StepChanges struct {
	Modified []string
	Added    []string
	Deleted  []string
	Renamed  []string
}

// ModifiedFiles returns the files modified by a step.
func (r StepResult) ModifiedFiles() []string {
	if r.files != nil {
		return r.files.Modified
	}
	return []string{}
}

// AddedFiles returns the files added by a step.
func (r StepResult) AddedFiles() []string {
	if r.files != nil {
		return r.files.Added
	}
	return []string{}
}

// DeletedFiles returns the files deleted by a step.
func (r StepResult) DeletedFiles() []string {
	if r.files != nil {
		return r.files.Deleted
	}
	return []string{}
}

// RenamedFiles returns the new name of files that have been renamed by a step.
func (r StepResult) RenamedFiles() []string {
	if r.files != nil {
		return r.files.Renamed
	}
	return []string{}
}

func parseGitStatus(out []byte) (StepChanges, error) {
	result := StepChanges{}

	stripped := strings.TrimSpace(string(out))
	if len(stripped) == 0 {
		return result, nil
	}

	for _, line := range strings.Split(stripped, "\n") {
		if len(line) < 4 {
			return result, fmt.Errorf("git status line has unrecognized format: %q", line)
		}

		file := line[3:]

		switch line[0] {
		case 'M':
			result.Modified = append(result.Modified, file)
		case 'A':
			result.Added = append(result.Added, file)
		case 'D':
			result.Deleted = append(result.Deleted, file)
		case 'R':
			files := strings.Split(file, " -> ")
			newFile := files[len(files)-1]
			result.Renamed = append(result.Renamed, newFile)
		}
	}

	return result, nil
}

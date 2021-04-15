package executor

import (
	"bytes"
	"io"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/batches/git"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

func renderStepTemplate(name, tmpl string, out io.Writer, stepCtx *StepContext) error {
	t, err := parseAsTemplate(name, tmpl, stepCtx)
	if err != nil {
		return errors.Wrap(err, "parsing step run")
	}

	return t.Execute(out, stepCtx)
}

func parseAsTemplate(name, input string, stepCtx *StepContext) (*template.Template, error) {
	return template.New(name).Delims("${{", "}}").Funcs(stepCtx.ToFuncMap()).Parse(input)
}

func renderStepMap(m map[string]string, stepCtx *StepContext) (map[string]string, error) {
	rendered := make(map[string]string, len(m))

	for k, v := range m {
		var out bytes.Buffer

		if err := renderStepTemplate(k, v, &out, stepCtx); err != nil {
			return rendered, err
		}

		rendered[k] = out.String()
	}

	return rendered, nil
}

// TODO(mrnugget): This is bad and should be (a) removed or (b) moved to batches package
type BatchChangeAttributes struct {
	Name        string
	Description string
}

// StepContext represents the contextual information available when rendering a
// step's fields, such as "run" or "outputs", as templates.
type StepContext struct {
	// BatchChange are the attributes in the BatchSpec that are set on the BatchChange.
	BatchChange BatchChangeAttributes
	// Outputs are the outputs set by the current and all previous steps.
	Outputs map[string]interface{}
	// Step is the result of the current step. Empty when evaluating the "run" field
	// but filled when evaluating the "outputs" field.
	Step StepResult
	// PreviousStep is the result of the previous step. Empty when there is no
	// previous step.
	PreviousStep StepResult
	// Repository is the Sourcegraph repository in which the steps are executed.
	Repository graphql.Repository
}

// ToFuncMap returns a template.FuncMap to access fields on the StepContext in a
// text/template.
func (stepCtx *StepContext) ToFuncMap() template.FuncMap {
	newStepResult := func(res *StepResult) map[string]interface{} {
		m := map[string]interface{}{
			"modified_files": "",
			"added_files":    "",
			"deleted_files":  "",
			"renamed_files":  "",
			"stdout":         "",
			"stderr":         "",
		}
		if res == nil {
			return m
		}

		m["modified_files"] = res.ModifiedFiles()
		m["added_files"] = res.AddedFiles()
		m["deleted_files"] = res.DeletedFiles()
		m["renamed_files"] = res.RenamedFiles()

		if res.Stdout != nil {
			m["stdout"] = res.Stdout.String()
		}

		if res.Stderr != nil {
			m["stderr"] = res.Stderr.String()
		}

		return m
	}

	return template.FuncMap{
		"join":    strings.Join,
		"split":   strings.Split,
		"replace": strings.ReplaceAll,
		"join_if": func(sep string, elems ...string) string {
			var nonBlank []string
			for _, e := range elems {
				if e != "" {
					nonBlank = append(nonBlank, e)
				}
			}
			return strings.Join(nonBlank, sep)
		},
		"previous_step": func() map[string]interface{} {
			return newStepResult(&stepCtx.PreviousStep)
		},
		"step": func() map[string]interface{} {
			return newStepResult(&stepCtx.Step)
		},
		"outputs": func() map[string]interface{} {
			return stepCtx.Outputs
		},
		"repository": func() map[string]interface{} {
			return map[string]interface{}{
				"search_result_paths": stepCtx.Repository.SearchResultPaths(),
				"name":                stepCtx.Repository.Name,
			}
		},
		"batch_change": func() map[string]interface{} {
			return map[string]interface{}{
				"name":        stepCtx.BatchChange.Name,
				"description": stepCtx.BatchChange.Description,
			}
		},
	}
}

// StepResult represents the result of a previously executed step.
type StepResult struct {
	// files are the changes made to files by the step.
	files *git.Changes

	// Stdout is the output produced by the step on standard out.
	Stdout *bytes.Buffer
	// Stderr is the output produced by the step on standard error.
	Stderr *bytes.Buffer
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

type StepsContext struct {
	// Changes that have been made by executing all steps.
	Changes *git.Changes
	// Path is the relative-to-root directory in which the steps have been
	// executed. Default is "". No leading "/".
	Path string
}

// ChangesetTemplateContext represents the contextual information available
// when rendering a field of the ChangesetTemplate as a template.
type ChangesetTemplateContext struct {
	// BatchChangeAttributes are the attributes of the BatchChange that will be
	// created/updated.
	BatchChangeAttributes BatchChangeAttributes

	// Steps are the changes made by all steps that were executed.
	Steps StepsContext

	// Outputs are the outputs defined and initialized by the steps.
	Outputs map[string]interface{}

	// Repository is the repository in which the steps were executed.
	Repository graphql.Repository
}

// ToFuncMap returns a template.FuncMap to access fields on the StepContext in a
// text/template.
func (tmplCtx *ChangesetTemplateContext) ToFuncMap() template.FuncMap {
	return template.FuncMap{
		"join":    strings.Join,
		"split":   strings.Split,
		"replace": strings.ReplaceAll,
		"join_if": func(sep string, elems ...string) string {
			var nonBlank []string
			for _, e := range elems {
				if e != "" {
					nonBlank = append(nonBlank, e)
				}
			}
			return strings.Join(nonBlank, sep)
		},
		"repository": func() map[string]interface{} {
			return map[string]interface{}{
				"search_result_paths": tmplCtx.Repository.SearchResultPaths(),
				"name":                tmplCtx.Repository.Name,
			}
		},
		"batch_change": func() map[string]interface{} {
			return map[string]interface{}{
				"name":        tmplCtx.BatchChangeAttributes.Name,
				"description": tmplCtx.BatchChangeAttributes.Description,
			}
		},
		"outputs": func() map[string]interface{} {
			return tmplCtx.Outputs
		},
		"steps": func() map[string]interface{} {
			// Wrap the *StepChanges in a StepResult so we can use nil-safe
			// methods.
			res := StepResult{files: tmplCtx.Steps.Changes}

			return map[string]interface{}{
				"modified_files": res.ModifiedFiles(),
				"added_files":    res.AddedFiles(),
				"deleted_files":  res.DeletedFiles(),
				"renamed_files":  res.RenamedFiles(),
				"path":           tmplCtx.Steps.Path,
			}
		},
	}
}

func renderChangesetTemplateField(name, tmpl string, tmplCtx *ChangesetTemplateContext) (string, error) {
	var out bytes.Buffer

	t, err := template.New(name).Delims("${{", "}}").Funcs(tmplCtx.ToFuncMap()).Parse(tmpl)
	if err != nil {
		return "", err
	}

	if err := t.Execute(&out, tmplCtx); err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}

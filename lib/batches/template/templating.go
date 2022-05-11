package template

import (
	"bytes"
	"io"
	"sort"
	"strings"
	"text/template"

	"github.com/gobwas/glob"

	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const startDelim = "${{"
const endDelim = "}}"

var builtins = template.FuncMap{
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
	"matches": func(in, pattern string) (bool, error) {
		g, err := glob.Compile(pattern)
		if err != nil {
			return false, err
		}
		return g.Match(in), nil
	},
}

func isTrueOutput(output interface{ String() string }) bool {
	return strings.TrimSpace(output.String()) == "true"
}

func EvalStepCondition(condition string, stepCtx *StepContext) (bool, error) {
	if condition == "" {
		return true, nil
	}

	var out bytes.Buffer
	if err := RenderStepTemplate("step-condition", condition, &out, stepCtx); err != nil {
		return false, errors.Wrap(err, "parsing step if")
	}

	return isTrueOutput(&out), nil
}

func RenderStepTemplate(name, tmpl string, out io.Writer, stepCtx *StepContext) error {
	t, err := template.New(name).Delims(startDelim, endDelim).Funcs(builtins).Funcs(stepCtx.ToFuncMap()).Parse(tmpl)
	if err != nil {
		return errors.Wrap(err, "parsing step run")
	}

	return t.Execute(out, stepCtx)
}

func RenderStepMap(m map[string]string, stepCtx *StepContext) (map[string]string, error) {
	rendered := make(map[string]string, len(m))

	for k, v := range m {
		var out bytes.Buffer

		if err := RenderStepTemplate(k, v, &out, stepCtx); err != nil {
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

type Repository struct {
	Name        string
	Branch      string
	FileMatches []string
}

func (r Repository) SearchResultPaths() (list fileMatchPathList) {
	sort.Strings(r.FileMatches)
	return fileMatchPathList(r.FileMatches)
}

type fileMatchPathList []string

func (f fileMatchPathList) String() string { return strings.Join(f, " ") }

// StepContext represents the contextual information available when rendering a
// step's fields, such as "run" or "outputs", as templates.
type StepContext struct {
	// BatchChange are the attributes in the BatchSpec that are set on the BatchChange.
	BatchChange BatchChangeAttributes
	// Outputs are the outputs set by the current and all previous steps.
	Outputs map[string]any
	// Step is the result of the current step. Empty when evaluating the "run" field
	// but filled when evaluating the "outputs" field.
	Step execution.StepResult
	// Steps contains the path in which the steps are being executed and the
	// changes made by all steps that were executed up until the current step.
	Steps StepsContext
	// PreviousStep is the result of the previous step. Empty when there is no
	// previous step.
	PreviousStep execution.StepResult
	// Repository is the Sourcegraph repository in which the steps are executed.
	Repository Repository
}

// ToFuncMap returns a template.FuncMap to access fields on the StepContext in a
// text/template.
func (stepCtx *StepContext) ToFuncMap() template.FuncMap {
	newStepResult := func(res *execution.StepResult) map[string]any {
		m := map[string]any{
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
		"previous_step": func() map[string]any {
			return newStepResult(&stepCtx.PreviousStep)
		},
		"step": func() map[string]any {
			return newStepResult(&stepCtx.Step)
		},
		"steps": func() map[string]any {
			res := newStepResult(&execution.StepResult{Files: stepCtx.Steps.Changes})
			res["path"] = stepCtx.Steps.Path
			return res
		},
		"outputs": func() map[string]any {
			return stepCtx.Outputs
		},
		"repository": func() map[string]any {
			return map[string]any{
				"search_result_paths": stepCtx.Repository.SearchResultPaths(),
				"name":                stepCtx.Repository.Name,
				"branch":              stepCtx.Repository.Branch,
			}
		},
		"batch_change": func() map[string]any {
			return map[string]any{
				"name":        stepCtx.BatchChange.Name,
				"description": stepCtx.BatchChange.Description,
			}
		},
	}
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
	Outputs map[string]any

	// Repository is the repository in which the steps were executed.
	Repository Repository
}

// ToFuncMap returns a template.FuncMap to access fields on the StepContext in a
// text/template.
func (tmplCtx *ChangesetTemplateContext) ToFuncMap() template.FuncMap {
	return template.FuncMap{
		"repository": func() map[string]any {
			return map[string]any{
				"search_result_paths": tmplCtx.Repository.SearchResultPaths(),
				"name":                tmplCtx.Repository.Name,
				"branch":              tmplCtx.Repository.Branch,
			}
		},
		"batch_change": func() map[string]any {
			return map[string]any{
				"name":        tmplCtx.BatchChangeAttributes.Name,
				"description": tmplCtx.BatchChangeAttributes.Description,
			}
		},
		"outputs": func() map[string]any {
			return tmplCtx.Outputs
		},
		"steps": func() map[string]any {
			// Wrap the *StepChanges in a execution.StepResult so we can use nil-safe
			// methods.
			res := execution.StepResult{Files: tmplCtx.Steps.Changes}

			return map[string]any{
				"modified_files": res.ModifiedFiles(),
				"added_files":    res.AddedFiles(),
				"deleted_files":  res.DeletedFiles(),
				"renamed_files":  res.RenamedFiles(),
				"path":           tmplCtx.Steps.Path,
			}
		},
	}
}

func RenderChangesetTemplateField(name, tmpl string, tmplCtx *ChangesetTemplateContext) (string, error) {
	var out bytes.Buffer

	t, err := template.New(name).Delims(startDelim, endDelim).Funcs(builtins).Funcs(tmplCtx.ToFuncMap()).Parse(tmpl)
	if err != nil {
		return "", err
	}

	if err := t.Execute(&out, tmplCtx); err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
}

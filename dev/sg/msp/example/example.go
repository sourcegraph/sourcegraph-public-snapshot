package example

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const newProjectIDFuncKey = "newProjectID"

var templateFuncs = template.FuncMap{
	newProjectIDFuncKey: spec.NewProjectID,
}

type Template struct {
	// ID is spec.service.id - required.
	ID string
	// Name is a human-readable name, optional.
	Name string
	// Onwer is the first value in spec.service.owners - required.
	Owner string
	// Dev indicates if this template should render a dev environment.
	Dev bool
	// ProjectIDSuffixLength specifies the length of the generated project ID's
	// random suffix.
	ProjectIDSuffixLength int
}

func (t *Template) setDefaults() {
	if t.Name == "" {
		caser := cases.Title(language.English)
		t.Name = caser.String(strings.ReplaceAll(strings.ReplaceAll(t.ID, "_", " "), "-", " "))
	}
}

var (
	//go:embed service.template.yaml
	serviceTemplateYAML string
	serviceTemplate     = func() *template.Template {
		return template.Must(
			template.New("service").
				Funcs(templateFuncs).
				Parse(serviceTemplateYAML),
		)
	}
)

// NewService provides a simple MSP service specification.
func NewService(t Template) ([]byte, error) {
	t.setDefaults()

	var b bytes.Buffer
	if err := serviceTemplate().Execute(&b, t); err != nil {
		return nil, errors.Wrap(err, "executing template")
	}
	return b.Bytes(), nil
}

var (
	//go:embed job.template.yaml
	jobTemplateYAML string
	jobTemplate     = func() *template.Template {
		return template.Must(
			template.New("job").
				Funcs(templateFuncs).
				Parse(jobTemplateYAML),
		)
	}
)

// NewJob provides a simple MSP job specification.
func NewJob(t Template) ([]byte, error) {
	t.setDefaults()

	var b bytes.Buffer
	if err := jobTemplate().Execute(&b, t); err != nil {
		return nil, errors.Wrap(err, "executing template")
	}
	return b.Bytes(), nil
}

var (
	//go:embed environment.template.yaml
	environmentTemplateYAML string
	environmentTemplate     = func() *template.Template {
		return template.Must(
			template.New("environment").
				Funcs(templateFuncs).
				Parse(environmentTemplateYAML),
		)
	}
)

type EnvironmentTemplate struct {
	ServiceID     string
	EnvironmentID string
	// ProjectIDSuffixLength is the length of the random suffix appended to
	// the generated project ID.
	ProjectIDSuffixLength int
}

func NewEnvironment(t EnvironmentTemplate) (*yaml.Node, error) {
	var b bytes.Buffer
	if err := environmentTemplate().Execute(&b, t); err != nil {
		return nil, errors.Wrap(err, "executing template")
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(b.Bytes(), &doc); err != nil {
		return nil, errors.Wrap(err, "unmarshal template as YAML")
	}

	root := doc.Content[0]
	return root, nil
}

package example

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Template struct {
	// ID is spec.service.id - required.
	ID string
	// Name is a human-readable name, optional.
	Name string
	// Onwer is the first value in spec.service.owners - required.
	Owner string
	// Dev indicates if this template should render a dev environment.
	Dev bool
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
	serviceTemplate     = template.Must(template.New("service").Parse(serviceTemplateYAML))
)

// NewService provides a simple MSP service specification.
func NewService(t Template) ([]byte, error) {
	t.setDefaults()

	var b bytes.Buffer
	if err := serviceTemplate.Execute(&b, t); err != nil {
		return nil, errors.Wrap(err, "executing template")
	}
	return b.Bytes(), nil
}

var (
	//go:embed job.template.yaml
	jobTemplateYAML string
	jobTemplate     = template.Must(template.New("job").Parse(jobTemplateYAML))
)

// NewJob provides a simple MSP job specification.
func NewJob(t Template) ([]byte, error) {
	t.setDefaults()

	var b bytes.Buffer
	if err := jobTemplate.Execute(&b, t); err != nil {
		return nil, errors.Wrap(err, "executing template")
	}
	return b.Bytes(), nil
}

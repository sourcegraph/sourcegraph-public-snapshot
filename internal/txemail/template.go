package txemail

import (
	"bytes"
	htmltemplate "html/template"
	"io"
	"strings"
	texttemplate "text/template"

	"github.com/jordan-wright/email"
	"github.com/k3a/html2text"

	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

// MustParseTemplate calls ParseTemplate and panics if an error is returned.
// It is intended to be called in a package init func.
func MustParseTemplate(input txtypes.Templates) txtypes.ParsedTemplates {
	pt, err := ParseTemplate(input)
	if err != nil {
		panic("MustParseTemplate: " + err.Error())
	}
	return *pt
}

// MustValidate panics if the templates are unparsable, otherwise it returns
// them unmodified.
func MustValidate(input txtypes.Templates) txtypes.Templates {
	MustParseTemplate(input)
	return input
}

// ParseTemplate is a helper func for parsing the text/template and html/template
// templates together.
func ParseTemplate(input txtypes.Templates) (*txtypes.ParsedTemplates, error) {
	if input.Text == "" {
		input.Text = html2text.HTML2Text(input.HTML)
	}

	st, err := texttemplate.New("").Parse(strings.TrimSpace(input.Subject))
	if err != nil {
		return nil, err
	}

	tt, err := texttemplate.New("").Parse(strings.TrimSpace(input.Text))
	if err != nil {
		return nil, err
	}

	ht, err := htmltemplate.New("").Parse(strings.TrimSpace(input.HTML))
	if err != nil {
		return nil, err
	}

	return &txtypes.ParsedTemplates{Subj: st, Text: tt, Html: ht}, nil
}

func renderTemplate(t *txtypes.ParsedTemplates, data any, m *email.Email) error {
	render := func(tmpl interface {
		Execute(io.Writer, any) error
	}) ([]byte, error) {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	subject, err := render(t.Subj)
	if err != nil {
		return err
	}
	m.Subject = string(subject)

	m.Text, err = render(t.Text)
	if err != nil {
		return err
	}

	m.HTML, err = render(t.Html)
	return err
}

package txemail

import (
	"bytes"
	htmltemplate "html/template"
	"io"
	"strings"
	texttemplate "text/template"

	gophermail "gopkg.in/jpoehls/gophermail.v0"
)

// Templates contains the text and HTML templates for an email.
type Templates struct {
	Subject string // text/template subject template
	Text    string // text/template text body template
	HTML    string //  html/template HTML body template
}

// MustParseTemplate calls ParseTemplate and panics if an error is returned.
// It is intended to be called in a package init func.
func MustParseTemplate(input Templates) ParsedTemplates {
	pt, err := ParseTemplate(input)
	if err != nil {
		panic("MustParseTemplate: " + err.Error())
	}
	return *pt
}

// ParseTemplate is a helper func for parsing the text/template and html/template
// templates together. In the future it will also provide common template funcs
// and a common footer.
func ParseTemplate(input Templates) (*ParsedTemplates, error) {
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

	return &ParsedTemplates{subj: st, text: tt, html: ht}, nil
}

// ParsedTemplates contains parsed text and HTML email templates.
type ParsedTemplates struct {
	subj *texttemplate.Template
	text *texttemplate.Template
	html *htmltemplate.Template
}

func (t ParsedTemplates) render(data interface{}, m *gophermail.Message) error {
	render := func(tmpl interface {
		Execute(io.Writer, interface{}) error
	}) (string, error) {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	var err error
	m.Subject, err = render(t.subj)
	if err != nil {
		return err
	}
	m.Body, err = render(t.text)
	if err != nil {
		return err
	}
	m.HTMLBody, err = render(t.html)
	if err != nil {
		return err
	}
	return nil
}

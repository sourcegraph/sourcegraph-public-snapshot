package txemail

import (
	"bytes"
	"html"
	htmltemplate "html/template"
	"io"
	"strings"
	texttemplate "text/template"

	"github.com/microcosm-cc/bluemonday"
	gfm "github.com/shurcooL/github_flavored_markdown"
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

// MustValidate panics if the templates are unparsable, otherwise it returns
// them unmodified.
func MustValidate(input Templates) Templates {
	MustParseTemplate(input)
	return input
}

// ParseTemplate is a helper func for parsing the text/template and html/template
// templates together. In the future it will also provide common template funcs
// and a common footer.
func ParseTemplate(input Templates) (*ParsedTemplates, error) {
	st, err := texttemplate.New("").Funcs(textFuncMap).Parse(strings.TrimSpace(input.Subject))
	if err != nil {
		return nil, err
	}

	tt, err := texttemplate.New("").Funcs(textFuncMap).Parse(strings.TrimSpace(input.Text))
	if err != nil {
		return nil, err
	}

	ht, err := htmltemplate.New("").Funcs(htmlFuncMap).Parse(strings.TrimSpace(input.HTML))
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

var (
	textFuncMap = map[string]interface{}{
		// Removes HTML tags (which are valid Markdown) from the source, for display in a text-only
		// setting.
		"markdownToText": func(markdownSource string) string {
			p := bluemonday.StrictPolicy()
			return html.UnescapeString(p.Sanitize(markdownSource))
		},
	}

	htmlFuncMap = map[string]interface{}{
		// Renders Markdown for display in an HTML email.
		"markdownToSafeHTML": func(markdownSource string) htmltemplate.HTML {
			unsafeHTML := gfm.Markdown([]byte(markdownSource))

			// The recommended policy at https://github.com/russross/blackfriday#extensions
			p := bluemonday.UGCPolicy()
			return htmltemplate.HTML(p.SanitizeBytes(unsafeHTML))
		},
	}
)

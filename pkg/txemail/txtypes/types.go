package txtypes

import (
	"bytes"
	htmltemplate "html/template"
	"io"
	texttemplate "text/template"

	gophermail "gopkg.in/jpoehls/gophermail.v0"
)

// Message describes an email message to be sent.
type Message struct {
	FromName   string   // email "From" address proper name
	To         []string // email "To" recipients
	ReplyTo    *string  // optional "ReplyTo" address
	MessageID  *string  // optional "Message-ID" header
	References []string // optional "References" header list

	Template Templates   // unparsed subject/body templates
	Data     interface{} // template data
}

// Templates contains the text and HTML templates for an email.
type Templates struct {
	Subject string // text/template subject template
	Text    string // text/template text body template
	HTML    string //  html/template HTML body template
}

// ParsedTemplates contains parsed text and HTML email templates.
type ParsedTemplates struct {
	Subj *texttemplate.Template
	Text *texttemplate.Template
	Html *htmltemplate.Template
}

func (t ParsedTemplates) Render(data interface{}, m *gophermail.Message) error {
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
	m.Subject, err = render(t.Subj)
	if err != nil {
		return err
	}
	m.Body, err = render(t.Text)
	if err != nil {
		return err
	}
	m.HTMLBody, err = render(t.Html)
	if err != nil {
		return err
	}
	return nil
}

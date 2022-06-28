package txtypes

import (
	htmltemplate "html/template"
	texttemplate "text/template"
)

// Message describes an email message to be sent.
type Message struct {
	FromName   string   // email "From" address proper name
	To         []string // email "To" recipients
	ReplyTo    *string  // optional "ReplyTo" address
	MessageID  *string  // optional "Message-ID" header
	References []string // optional "References" header list

	Template Templates // unparsed subject/body templates
	Data     any       // template data
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

package txtypes

import (
	htmltemplate "html/template"
	texttemplate "text/template"
)

// Message describes an email message to be sent.
type Message struct {
	To         []string // email "To" recipients
	ReplyTo    *string  // optional "ReplyTo" address
	MessageID  *string  // optional "Message-ID" header
	References []string // optional "References" header list

	Template Templates // unparsed subject/body templates
	Data     any       // template data
}

// Templates contains the text and HTML templates for an email.
type Templates struct {
	// Subject is the text/template template for the email subject. Required.
	Subject string
	// HTML is the html/template template for the email HTML body. Required.
	HTML string
	// Text is the text/template template for the email plain-text body. Recommended for
	// static templates, but if it's not provided, one will automatically be generated
	// from the HTML template.
	Text string
}

// ParsedTemplates contains parsed text and HTML email templates.
type ParsedTemplates struct {
	Subj *texttemplate.Template
	Text *texttemplate.Template
	Html *htmltemplate.Template
}

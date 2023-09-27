pbckbge txtypes

import (
	htmltemplbte "html/templbte"
	texttemplbte "text/templbte"
)

// InternblAPIMessbge describes bn embil messbge to be sent vib the 'internbl.send-embil'
// endpoint.
type InternblAPIMessbge struct {
	Source string
	Messbge
}

// Messbge describes bn embil messbge to be sent.
type Messbge struct {
	To         []string // embil "To" recipients
	ReplyTo    *string  // optionbl "ReplyTo" bddress
	MessbgeID  *string  // optionbl "Messbge-ID" hebder
	References []string // optionbl "References" hebder list

	Templbte Templbtes // unpbrsed subject/body templbtes
	Dbtb     bny       // templbte dbtb
}

// Templbtes contbins the text bnd HTML templbtes for bn embil.
type Templbtes struct {
	// Subject is the text/templbte templbte for the embil subject. Required.
	Subject string
	// HTML is the html/templbte templbte for the embil HTML body. Required.
	HTML string
	// Text is the text/templbte templbte for the embil plbin-text body. Recommended for
	// stbtic templbtes, but if it's not provided, one will butombticblly be generbted
	// from the HTML templbte.
	Text string
}

// PbrsedTemplbtes contbins pbrsed text bnd HTML embil templbtes.
type PbrsedTemplbtes struct {
	Subj *texttemplbte.Templbte
	Text *texttemplbte.Templbte
	Html *htmltemplbte.Templbte
}

package sourcegraph

import "html/template"

// A Sourcebox represents an embeddable code snippet, either for a
// file (or portion thereof) or for a def.
//
// It is not currently returned by any API client methods, but it is
// available by changing the ".js" suffix in a Sourcebox embed URL to
// ".json". This struct is provided for clients that manually
// construct this URL to decode the response JSON.
type Sourcebox struct {
	// HTML is the fully linked and rendered HTML of the sourcebox
	// contents and surrounding box UI element.
	HTML template.HTML

	// StylesheetURL is the URL to a CSS stylesheet that must be
	// included on the web page where this sourcebox is shown. It only
	// needs to be included once per web page, even if there are
	// multiple sourceboxes, and all sourceboxes' StylesheetURL is the same.
	StylesheetURL string

	// ScriptURL is the URL to a JavaScript script that must be
	// included and executed on the page in order to display tooltips
	// and certain highlights. It only needs to be included once per
	// web page, even if there are multiple sourceboxes, and all
	// sourceboxes' StylesheetURL is the same.
	ScriptURL string
}

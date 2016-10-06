package htmlutil

import "github.com/microcosm-cc/bluemonday"

// HTML is a type which marshals into {__html: "html code here"} to designate
// that this value is sanitized HTML code, see
// https://facebook.github.io/react/tips/dangerously-set-inner-html.html
type HTML struct {
	HTML string `json:"__html"`
}

var sanitizePolicy = bluemonday.UGCPolicy().AllowAttrs("class").Matching(bluemonday.SpaceSeparatedTokens).Globally()

// EmptyHTML returns the type *HTML with an empty HTML string.
func EmptyHTML() *HTML {
	return &HTML{}
}

// Sanitize makes sure that the given HTML code is safe against XSS attacks
// and returns it wrapped in the *HTML type to designate that this is
// safe HTML. *HTML should always be created via this package.
func Sanitize(html string) *HTML {
	return &HTML{
		HTML: sanitizePolicy.Sanitize(html),
	}
}

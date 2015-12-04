package app

import (
	"html/template"
	"net/http"
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

func sanitizeHTML(html string) template.HTML {
	return template.HTML(bluemonday.UGCPolicy().Sanitize(html))
}

var formattingClasses = regexp.MustCompile("^defn\\-popover( match| highlight| highlight-primary| ref| def| pln| str| kwd| com| typ| lit| pun| opn| clo| tag| atn| atv| dec| var| fun)*$")

func sanitizeFormattedCode(html string) template.HTML {
	p := bluemonday.NewPolicy()
	p.RequireParseableURLs(true)
	p.AllowRelativeURLs(true)
	p.AllowURLSchemes("http", "https")
	p.AllowElements("a", "span")
	p.AllowAttrs("href").OnElements("a")
	p.AllowAttrs("class").Matching(formattingClasses).OnElements("a", "span")
	return template.HTML(p.Sanitize(html))
}

func copyRequest(r *http.Request) *http.Request {
	rCopy := *r
	urlCopy := *r.URL
	rCopy.URL = &urlCopy
	return &rCopy
}

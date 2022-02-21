package main

import (
	"strings"
	"text/template"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var searchResultsAlertTemplate *template.Template

func init() {
	var err error

	if searchResultsAlertTemplate, err = parseTemplate(searchResultsAlertTemplateContent); err != nil {
		// This shouldn't fail, since we control the template content via a
		// constant below.
		panic(err)
	}
}

// ProposedQuery is a suggested query to run when we emit an alert.
type ProposedQuery struct {
	Description string
	Query       string
}

// searchResultsAlert is a type that can be used to unmarshal values returned by
// the searchResultsAlertFragment GraphQL fragment below.
type searchResultsAlert struct {
	Title           string
	Description     string
	ProposedQueries []ProposedQuery
}

// Render renders an alert to a string ready to be output to a console,
// respecting the colour configuration in use by the current process. If the
// alert is empty, then an empty string will be returned.
func (alert *searchResultsAlert) Render() (string, error) {
	b := &strings.Builder{}
	if err := searchResultsAlertTemplate.Execute(b, alert); err != nil {
		return "", errors.Wrap(err, "rendering alert template")
	}
	return b.String(), nil
}

// searchResultsAlertFragment provides a GraphQL fragment that can be used to
// hydrate a searchResultsAlert instance.
const searchResultsAlertFragment = `
	fragment SearchResultsAlertFields on SearchResults {
		alert {
			title
			description
			proposedQueries {
				description
				query
			}
		}
	}
`

const searchResultsAlertTemplateContent = `
	{{- if gt (len .Title) 0 -}}
		{{- color "search-alert-title"}}❗{{.Title}}{{color "nc"}}{{"\n"}}
	{{- end -}}

	{{- if gt (len .Description) 0 -}}
		{{- color "search-alert-description"}}  {{.Description}}{{color "nc"}}{{"\n"}}
	{{- end -}}

	{{- if gt (len .ProposedQueries) 0 -}}
		{{- color "search-alert-proposed-title"}}  Did you mean:{{color "nc" -}}
		{{- "\n" -}}
		{{- range .ProposedQueries -}}
			{{- color "search-alert-proposed-query"}}  {{.Query}}{{color "nc" -}}
			{{- " - " -}}
			{{- color "search-alert-proposed-description"}}{{.Description}}{{color "nc" -}}
			{{- "\n" -}}
		{{- end -}}
	{{- end -}}
`

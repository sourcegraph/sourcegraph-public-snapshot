package background

import (
	_ "embed"

	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

var (
	//go:embed template.html.tmpl
	htmlTemplate string

	//go:embed template.txt.tmpl
	textTemplate string
)

var newSearchResultsEmailTemplates = txemail.MustValidate(txtypes.Templates{
	Subject: `{{ if .IsTest }}Test: {{ end }}[{{.Priority}} event] {{.Description}}`,
	Text:    textTemplate,
	HTML:    htmlTemplate,
})

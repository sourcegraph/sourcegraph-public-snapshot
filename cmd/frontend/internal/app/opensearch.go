package app

import (
	"bytes"
	"html/template"
	"net/http"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

var openSearchDescription = template.Must(template.New("").Parse(`
<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/" xmlns:moz="http://www.mozilla.org/2006/browser/search/">
  <ShortName>{{.SiteName}}</ShortName>
  <Description>Search {{.SiteName}}</Description>
  <InputEncoding>UTF-8</InputEncoding>
  <Image width="16" height="16" type="image/png">{{.BaseURL}}/.assets/img/favicon.png</Image>
  <Url type="text/html" method="GET" template="{{.SearchURL}}" />
  <SearchForm>{{.BaseURL}}/search</SearchForm>
</OpenSearchDescription>
`))

func openSearch(w http.ResponseWriter, r *http.Request) {
	type vars struct {
		SiteName  string
		BaseURL   string
		SearchURL string
	}
	data := vars{
		BaseURL: globals.AppURL.String(),
	}
	if globals.AppURL.String() == "https://sourcegraph.com" {
		data.SiteName = "Sourcegraph"
	} else {
		data.SiteName = "Sourcegraph (" + globals.AppURL.Host + ")"
	}
	openSearchConfiguration := conf.Get().OpenSearch
	if openSearchConfiguration != nil {
		data.SearchURL = openSearchConfiguration.SearchUrl
	} else {
		data.SearchURL = data.BaseURL + "/search?q={searchTerms}"
	}

	var buf bytes.Buffer
	if err := openSearchDescription.Execute(&buf, data); err != nil {
		log15.Error("Failed to execute OpenSearch template", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	_, _ = buf.WriteTo(w)
}

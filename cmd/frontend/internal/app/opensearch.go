package app

import (
	"bytes"
	"html/template"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

var openSearchDescription = template.Must(template.New("").Parse(`
<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/" xmlns:moz="http://www.mozilla.org/2006/browser/search/">
  <ShortName>{{.SiteName}}</ShortName>
  <Description>Search {{.SiteName}}</Description>
  <InputEncoding>UTF-8</InputEncoding>
  <Image width="16" height="16" type="image/png">{{.BaseURL}}/.assets/img/favicon.png</Image>
  <Url type="text/html" method="GET" template="{{.BaseURL}}/-/search?q={searchTerms}&amp;profile=last" />
  <SearchForm>{{.BaseURL}}/-/search</SearchForm>
</OpenSearchDescription>
`))

func openSearch(w http.ResponseWriter, r *http.Request) {
	// TODO: Omnisearch disabled currently. Re-enable in the future.
	// See https://github.com/sourcegraph/sourcegraph/issues/6798
	w.WriteHeader(http.StatusNotFound)
	return

	type vars struct {
		SiteName string
		BaseURL  string
	}
	data := vars{
		BaseURL: conf.AppURL.String(),
	}
	if conf.AppURL.String() == "https://sourcegraph.com" {
		data.SiteName = "Sourcegraph"
	} else {
		data.SiteName = "Sourcegraph (" + conf.AppURL.Host + ")"
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

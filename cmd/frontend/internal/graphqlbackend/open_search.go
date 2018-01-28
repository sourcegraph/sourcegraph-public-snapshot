package graphqlbackend

import (
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

type openSearchSettings struct {
	searchURL string
}

func (s openSearchSettings) SearchURL() string { return s.searchURL }

func (r *schemaResolver) OpenSearchSettings() *openSearchSettings {
	var openSearch = conf.Get().OpenSearch
	searchURL := globals.AppURL.String() + "/search?q={searchTerms}"
	if openSearch != nil {
		searchURL = conf.Get().OpenSearch.SearchUrl
	}

	return &openSearchSettings{
		searchURL: searchURL,
	}
}

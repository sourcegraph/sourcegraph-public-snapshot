package graphqlbackend

import (
	"encoding/json"
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

type searchScope struct {
	JName  string `json:"name"`
	JValue string `json:"value"`
}

func (s searchScope) Name() string  { return s.JName }
func (s searchScope) Value() string { return s.JValue }

var (
	searchScopes     = env.Get("SEARCH_SCOPES", defaultSearchScopesJSON(), `JSON array of predefined search scopes - [{"name":"...","value":"..."}, ...]`)
	searchScopesList []*searchScope
)

func defaultSearchScopesJSON() string {
	var prependScope string
	var defaultSearchScopes []*searchScope
	useActiveInactiveRepos := inactiveRepos != ""
	if useActiveInactiveRepos {
		prependScope = "repogroup:active "
		defaultSearchScopes = append(defaultSearchScopes,
			&searchScope{JName: "All active repos", JValue: "repogroup:active"},
		)
	} else if envvar.DeploymentOnPrem() {
		defaultSearchScopes = append(defaultSearchScopes, &searchScope{JName: "All", JValue: ""})
	} else {
		prependScope = "repogroup:sample "
		defaultSearchScopes = append(defaultSearchScopes, &searchScope{JName: "Sample repos", JValue: "repogroup:sample"})
	}

	scopes := []*searchScope{
		{JName: "Test code", JValue: "file:(test|spec)"},
		{JName: "Non-test files", JValue: "-file:(test|spec)"},
		{JName: "JSON files", JValue: `file:\.json$`},
		{JName: "Text documents", JValue: `file:\.(txt|md)$`},
		{JName: "Non-vendor code", JValue: "-file:vendor/ -file:node_modules/"},
		{JName: "Vendored code", JValue: "file:(vendor|node_modules)/"},
	}
	if prependScope != "" {
		for _, scope := range scopes {
			scope.JValue = prependScope + scope.JValue
		}
	}
	defaultSearchScopes = append(defaultSearchScopes, scopes...)

	if useActiveInactiveRepos {
		defaultSearchScopes = append(defaultSearchScopes, &searchScope{JName: "Inactive repos", JValue: "repogroup:inactive"})
	}

	b, _ := json.Marshal(defaultSearchScopes)
	return string(b)
}

func init() {
	if err := json.Unmarshal([]byte(searchScopes), &searchScopesList); err != nil {
		log.Fatal("SEARCH_SCOPES:", err)
	}
}

func (r *rootResolver) SearchScopes2() []*searchScope {
	return searchScopesList
}

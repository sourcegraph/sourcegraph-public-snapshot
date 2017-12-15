package graphqlbackend

import (
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

type searchScope struct {
	JName  string `json:"name"`
	JValue string `json:"value"`
}

func (s searchScope) Name() string  { return s.JName }
func (s searchScope) Value() string { return s.JValue }

var (
	searchScopesList []*searchScope
)

func init() {
	if conf.Get().SearchScopes == nil {
		searchScopesList = defaultSearchScopes()
	} else {
		searchScopesList = make([]*searchScope, len(conf.Get().SearchScopes))
		for i, s := range conf.Get().SearchScopes {
			searchScopesList[i] = &searchScope{JName: s.Name, JValue: s.Value}
		}
	}
}

func defaultSearchScopes() []*searchScope {
	var prependScope string
	var defaultSearchScopes []*searchScope
	useActiveInactiveRepos := envvar.DeploymentOnPrem() && inactiveRepos != ""
	if useActiveInactiveRepos {
		prependScope = "repogroup:active "
		defaultSearchScopes = append(defaultSearchScopes,
			&searchScope{JName: "All active repos", JValue: "repogroup:active"},
		)
	} else {
		if !envvar.DeploymentOnPrem() {
			prependScope = "repogroup:sample "
			defaultSearchScopes = append(defaultSearchScopes, &searchScope{JName: "Sample repositories", JValue: "repogroup:sample"})
		}

		defaultSearchScopes = append(defaultSearchScopes, &searchScope{JName: "All repositories", JValue: ""})
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

	return defaultSearchScopes
}

func (r *schemaResolver) SearchScopes2() []*searchScope {
	return searchScopesList
}

package graphqlbackend

import (
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
	searchScopesList = make([]*searchScope, len(conf.Get().SearchScopes))
	for i, s := range conf.Get().SearchScopes {
		searchScopesList[i] = &searchScope{JName: s.Name, JValue: s.Value}
	}
}

func (r *schemaResolver) SearchScopes() []*searchScope {
	return searchScopesList
}

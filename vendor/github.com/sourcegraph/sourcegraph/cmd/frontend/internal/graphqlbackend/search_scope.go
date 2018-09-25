package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

type searchScope struct {
	id          *string
	name        string
	value       string
	description *string
}

func (s searchScope) ID() *string          { return s.id }
func (s searchScope) Name() string         { return s.name }
func (s searchScope) Value() string        { return s.value }
func (s searchScope) Description() *string { return s.description }

var (
	searchScopesList []*searchScope
)

func init() {
	searchScopesList = make([]*searchScope, len(conf.GetTODO().SearchScopes))
	for i, s := range conf.GetTODO().SearchScopes {
		if s.Id != "" {
			searchScopesList[i].id = &s.Id
			searchScopesList[i].description = &s.Description
		}
		searchScopesList[i] = &searchScope{name: s.Name, value: s.Value}
	}
}

func (r *schemaResolver) SearchScopes() []*searchScope {
	return searchScopesList
}

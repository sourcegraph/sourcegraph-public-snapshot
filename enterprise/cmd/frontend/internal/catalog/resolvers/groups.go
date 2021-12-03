package resolvers

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
)

func (r *rootResolver) Groups() []gql.GroupResolver {
	_, groups, _ := catalog.Data()

	var groupResolvers []gql.GroupResolver
	for _, group := range groups {
		groupResolvers = append(groupResolvers, &groupResolver{group: group, db: r.db})
	}
	return groupResolvers
}

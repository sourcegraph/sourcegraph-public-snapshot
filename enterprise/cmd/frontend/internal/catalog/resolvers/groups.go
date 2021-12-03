package resolvers

import (
	"context"

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

func (r *rootResolver) Group(ctx context.Context, args *gql.GroupArgs) (gql.GroupResolver, error) {
	_, groups, _ := catalog.Data()
	for _, group := range groups {
		if group.Name == args.Name {
			return &groupResolver{group: group, db: r.db}, nil
		}
	}
	return nil, nil
}

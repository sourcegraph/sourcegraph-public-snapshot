package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func allGroups(db database.DB) []*groupResolver {
	groups := catalog.Groups()

	var groupResolvers []*groupResolver
	for _, group := range groups {
		groupResolvers = append(groupResolvers, &groupResolver{group: group, db: db})
	}
	return groupResolvers
}

func groupByID(db database.DB, id graphql.ID) *groupResolver {
	groups := catalog.Groups()
	for _, g := range groups {
		gr := &groupResolver{group: g, db: db}
		if gr.ID() == id {
			return gr
		}
	}
	return nil
}

func (r *rootResolver) Groups() []gql.GroupResolver {
	var groupResolvers []gql.GroupResolver
	for _, group := range allGroups(r.db) {
		groupResolvers = append(groupResolvers, group)
	}
	return groupResolvers
}

func groupByName(db database.DB, name string) *groupResolver {
	groups := catalog.Groups()
	for _, group := range groups {
		if group.Name == name {
			return &groupResolver{group: group, db: db}
		}
	}
	return nil
}

func (r *rootResolver) Group(ctx context.Context, args *gql.GroupArgs) (gql.GroupResolver, error) {
	return groupByName(r.db, args.Name), nil
}

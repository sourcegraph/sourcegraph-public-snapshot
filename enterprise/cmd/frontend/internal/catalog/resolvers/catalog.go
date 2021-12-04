package resolvers

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type catalogResolver struct {
	db database.DB
}

func (r *catalogResolver) Entities(ctx context.Context, args *gql.CatalogEntitiesArgs) (gql.CatalogEntityConnectionResolver, error) {
	components := dummyData(r.db)

	var query string
	if args.Query != nil {
		query = *args.Query
	}
	match := getQueryMatcher(query)

	var keep []gql.CatalogEntity
	for _, c := range components {
		if match(c) {
			keep = append(keep, c)
		}
	}

	return &catalogEntityConnectionResolver{entities: wrapInCatalogEntityInterfaceType(keep)}, nil
}

func parseQuery(q string) (literal string, groupID graphql.ID) {
	parts := strings.Fields(q)
	for _, part := range parts {
		const groupPrefix = "group:"
		if strings.HasPrefix(part, groupPrefix) {
			groupID = graphql.ID(strings.TrimPrefix(part, groupPrefix))
		} else {
			literal += part
		}
	}
	return
}

func getQueryMatcher(q string) func(*catalogComponentResolver) bool {
	literal, groupID := parseQuery(q)
	group := groupByID(groupID)
	isComponentInGroup := func(c *catalogComponentResolver) bool {
		if c.component.OwnedBy == group.group.Name {
			return true
		}
		for _, dg := range group.DescendentGroups() {
			if c.component.OwnedBy == dg.Name() {
				return true
			}
		}
		return false
	}

	return func(c *catalogComponentResolver) bool {
		return strings.Contains(c.component.Name, literal) && (group == nil || isComponentInGroup(c))
	}
}

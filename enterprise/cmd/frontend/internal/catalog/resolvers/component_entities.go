package resolvers

import (
	"context"
	"fmt"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *catalogComponentResolver) RelatedEntities(ctx context.Context, args *gql.CatalogEntityRelatedEntitiesArgs) (gql.CatalogEntityRelatedEntityConnectionResolver, error) {
	q := parseQuery(r.db, fmt.Sprintf("%s relatedToEntity:%s", strOrEmpty(args.Query), r.ID()))
	graph := makeGraphData(r.db, q, false)
	var edges []gql.CatalogEntityRelatedEntityEdgeResolver
	for _, edge := range graph.edges {
		edges = append(edges, &catalogEntityRelatedEntityEdgeResolver{
			type_: edge.Type(),
			node:  edge.InNode(),
		})
	}
	return &catalogEntityRelatedEntityConnectionResolver{edges: edges}, nil
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

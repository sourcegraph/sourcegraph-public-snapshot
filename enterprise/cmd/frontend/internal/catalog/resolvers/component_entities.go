package resolvers

import (
	"context"
	"fmt"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *catalogComponentResolver) RelatedEntities(ctx context.Context) (gql.CatalogEntityRelatedEntityConnectionResolver, error) {
	graph := makeGraphData(r.db, parseQuery(r.db, fmt.Sprintf("relatedToEntity:%s", r.ID())))
	var edges []gql.CatalogEntityRelatedEntityEdgeResolver
	for _, edge := range graph.edges {
		edges = append(edges, &catalogEntityRelatedEntityEdgeResolver{
			type_: edge.Type(),
			node:  edge.InNode(),
		})
	}
	return &catalogEntityRelatedEntityConnectionResolver{edges: edges}, nil
}

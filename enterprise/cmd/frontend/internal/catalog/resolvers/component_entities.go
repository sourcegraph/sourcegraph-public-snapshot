package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *catalogComponentResolver) RelatedEntities(ctx context.Context) (gql.CatalogEntityRelatedEntityConnectionResolver, error) {
	graph := makeGraphData(r.db, nil)
	var edges []gql.CatalogEntityRelatedEntityEdgeResolver
	for _, edge := range graph.edges {
		if edge.OutNode().ID() == r.ID() {
			edges = append(edges, &catalogEntityRelatedEntityEdgeResolver{
				type_: edge.Type(),
				node:  edge.InNode(),
			})
		}
	}
	return &catalogEntityRelatedEntityConnectionResolver{edges: edges}, nil
}

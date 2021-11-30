package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *catalogComponentResolver) RelatedEntities(ctx context.Context) (gql.CatalogEntityRelatedEntityConnectionResolver, error) {
	graph := dummyGraph(r.db, r.ID())
	var edges []gql.CatalogEntityRelatedEntityEdgeResolver
	for _, edge := range graph.edges {
		if edge.InNode().ID() == r.ID() {
			edges = append(edges, &catalogEntityRelatedEntityEdgeResolver{
				node:  &gql.CatalogEntityResolver{edge.OutNode()},
				type_: "DEPENDS_ON",
			})
		} else if edge.OutNode().ID() == r.ID() {
			edges = append(edges, &catalogEntityRelatedEntityEdgeResolver{
				node:  &gql.CatalogEntityResolver{edge.InNode()},
				type_: "DEPENDENCY_OF",
			})
		}
	}
	return &catalogEntityRelatedEntityConnectionResolver{edges: edges}, nil
}

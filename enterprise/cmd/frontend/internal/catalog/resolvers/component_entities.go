package resolvers

import (
	"context"
	"fmt"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *componentResolver) RelatedEntities(ctx context.Context, args *gql.ComponentRelatedEntitiesArgs) (gql.ComponentRelatedEntityConnectionResolver, error) {
	q := parseQuery(r.db, fmt.Sprintf("%s relatedToEntity:%s", strOrEmpty(args.Query), r.ID()))
	graph := makeGraphData(r.db, q, false)
	var edges []gql.ComponentRelatedEntityEdgeResolver
	for _, edge := range graph.edges {
		edges = append(edges, &componentRelatedEntityEdgeResolver{
			type_: edge.Type(),
			node:  edge.InNode(),
		})
	}
	return &componentRelatedEntityConnectionResolver{edges: edges}, nil
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

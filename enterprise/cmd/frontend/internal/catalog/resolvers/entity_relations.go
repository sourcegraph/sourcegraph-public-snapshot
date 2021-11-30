package resolvers

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type catalogEntityRelatedEntityConnectionResolver struct {
	edges []gql.CatalogEntityRelatedEntityEdgeResolver
}

func (r *catalogEntityRelatedEntityConnectionResolver) Edges() []gql.CatalogEntityRelatedEntityEdgeResolver {
	return r.edges
}

type catalogEntityRelatedEntityEdgeResolver struct {
	node  *gql.CatalogEntityResolver
	type_ gql.CatalogEntityRelationType
}

func (r *catalogEntityRelatedEntityEdgeResolver) Node() *gql.CatalogEntityResolver {
	return r.node
}

func (r *catalogEntityRelatedEntityEdgeResolver) Type() gql.CatalogEntityRelationType {
	return r.type_
}

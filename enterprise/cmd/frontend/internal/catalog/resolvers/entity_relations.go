package resolvers

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type componentRelatedEntityConnectionResolver struct {
	edges []gql.ComponentRelatedEntityEdgeResolver
}

func (r *componentRelatedEntityConnectionResolver) Edges() []gql.ComponentRelatedEntityEdgeResolver {
	return r.edges
}

type componentRelatedEntityEdgeResolver struct {
	node  gql.ComponentResolver
	type_ gql.ComponentRelationType
}

func (r *componentRelatedEntityEdgeResolver) Node() gql.ComponentResolver {
	return r.node
}

func (r *componentRelatedEntityEdgeResolver) Type() gql.ComponentRelationType {
	return r.type_
}

package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
)

func (r *catalogResolver) Graph(ctx context.Context) (gql.CatalogGraphResolver, error) {
	var graph catalogGraphResolver

	components := catalog.Components()
	var entities []gql.CatalogEntity
	for _, c := range components {
		entities = append(entities, &catalogComponentResolver{component: c, db: r.db})
	}
	graph.nodes = wrapInCatalogEntityInterfaceType(entities)

	seeds := []int{1, 5, 7, 8, 11}
	for _, x := range seeds {
		edge := catalogEntityRelationEdgeResolver{
			outNode: graph.nodes[x%len(graph.nodes)],
			outType: "DEPENDS_ON",
			inNode:  graph.nodes[(x*2)%len(graph.nodes)],
			inType:  "DEPENDENCY_OF",
		}

		graph.edges = append(graph.edges, &edge)
	}

	return &graph, nil
}

type catalogGraphResolver struct {
	nodes []*gql.CatalogEntityResolver
	edges []gql.CatalogEntityRelationEdgeResolver
}

func (r *catalogGraphResolver) Nodes() []*gql.CatalogEntityResolver            { return r.nodes }
func (r *catalogGraphResolver) Edges() []gql.CatalogEntityRelationEdgeResolver { return r.edges }

type catalogEntityRelationEdgeResolver struct {
	outNode *gql.CatalogEntityResolver
	outType gql.CatalogEntityRelationType

	inNode *gql.CatalogEntityResolver
	inType gql.CatalogEntityRelationType
}

func (r *catalogEntityRelationEdgeResolver) OutNode() *gql.CatalogEntityResolver    { return r.outNode }
func (r *catalogEntityRelationEdgeResolver) OutType() gql.CatalogEntityRelationType { return r.outType }
func (r *catalogEntityRelationEdgeResolver) InNode() *gql.CatalogEntityResolver     { return r.inNode }
func (r *catalogEntityRelationEdgeResolver) InType() gql.CatalogEntityRelationType  { return r.inType }

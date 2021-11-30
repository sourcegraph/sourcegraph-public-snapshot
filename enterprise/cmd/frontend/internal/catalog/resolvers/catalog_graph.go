package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func dummyGraph(db database.DB, filterID graphql.ID) *catalogGraphResolver {
	var graph catalogGraphResolver

	components := catalog.Components()
	var entities []gql.CatalogEntity
	for _, c := range components {
		entities = append(entities, &catalogComponentResolver{component: c, db: db})
	}
	graph.nodes = wrapInCatalogEntityInterfaceType(entities)

	seeds := []int{1, 2, 3, 5, 7, 8, 11, 1, 12, 5, 7, 3, 2, 11, 5, 4, 8, 4, 15, 3, 5}
	seeds = append(seeds, seeds...)
	y := 17
	seen := map[graphql.ID]struct{}{}
	i := 0
	for _, x := range seeds {
		i++
		outNode := graph.nodes[(x+i)%len(graph.nodes)]
		inNode := graph.nodes[y%len(graph.nodes)]
		y += x
		if outNode == inNode {
			continue
		}
		if filterID != "" && outNode.ID() != filterID && inNode.ID() != filterID {
			continue
		}

		key := outNode.ID() + inNode.ID()
		if _, seen := seen[key]; seen {
			continue
		}
		seen[key] = struct{}{}

		edge := catalogEntityRelationEdgeResolver{
			outNode: outNode,
			outType: "DEPENDS_ON",
			inNode:  inNode,
			inType:  "DEPENDENCY_OF",
		}

		graph.edges = append(graph.edges, &edge)
	}

	return &graph
}

func (r *catalogResolver) Graph(ctx context.Context) (gql.CatalogGraphResolver, error) {
	return dummyGraph(r.db, ""), nil
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

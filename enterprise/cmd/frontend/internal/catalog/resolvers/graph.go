package resolvers

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/catalog"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func makeGraphData(db database.DB, q *queryMatcher, allEdges bool) *catalogGraphResolver {
	var graph catalogGraphResolver

	components := catalog.Components()
	edges := catalog.Edges()
	for _, c := range components {
		cr := &componentResolver{component: c, db: db}
		if q != nil && q.matchNode(cr) {
			graph.nodes = append(graph.nodes, cr)
		}
	}

	findNodeByName := func(name string) gql.ComponentResolver {
		for _, node := range graph.nodes {
			if node.Name() == name {
				return node
			}
		}
		return nil
	}

	// edgeMatches := map[*gql.ComponentResolver]struct{}{}
	for _, e := range edges {
		outNode := findNodeByName(e.Out)
		inNode := findNodeByName(e.In)
		if outNode == nil || inNode == nil {
			continue
		}
		edge := &componentRelationEdgeResolver{
			type_:   gql.ComponentRelationType(e.Type),
			outNode: outNode,
			inNode:  inNode,
		}
		if allEdges || q.matchEdge(edge) {
			graph.edges = append(graph.edges, edge)
		}
		// edgeMatches[inNode] = struct{}{}
		// edgeMatches[outNode] = struct{}{}
	}

	// keepNodes := graph.nodes[:0]
	// for _, node := range graph.nodes {
	// 	if _, edgeMatches := edgeMatches[node]; edgeMatches {
	// 		keepNodes = append(keepNodes, node)
	// 	}
	// }
	// graph.nodes = keepNodes

	return &graph
}

func (r *rootResolver) Graph(ctx context.Context, args *gql.CatalogGraphArgs) (gql.CatalogGraphResolver, error) {
	// TODO(sqs): support literal query search
	var query string
	if args.Query != nil {
		query = *args.Query
	}

	return makeGraphData(r.db, parseQuery(r.db, query), true), nil
}

type catalogGraphResolver struct {
	nodes []gql.ComponentResolver
	edges []gql.ComponentRelationEdgeResolver
}

func (r *catalogGraphResolver) Nodes() []gql.ComponentResolver             { return r.nodes }
func (r *catalogGraphResolver) Edges() []gql.ComponentRelationEdgeResolver { return r.edges }

type componentRelationEdgeResolver struct {
	type_   gql.ComponentRelationType
	outNode gql.ComponentResolver
	inNode  gql.ComponentResolver
}

func (r *componentRelationEdgeResolver) Type() gql.ComponentRelationType { return r.type_ }
func (r *componentRelationEdgeResolver) OutNode() gql.ComponentResolver  { return r.outNode }
func (r *componentRelationEdgeResolver) InNode() gql.ComponentResolver   { return r.inNode }

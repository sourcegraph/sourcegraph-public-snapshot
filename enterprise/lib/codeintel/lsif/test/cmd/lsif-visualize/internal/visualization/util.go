package visualization

import (
	reader "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/protocol/reader"
	reader2 "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/test/internal/reader"
)

//
// TODO - move these functions into shared internal

// forEachInV calls the given function on each sink vertex adjacent to the given
// edge. If any invocation returns false, iteration of the adjacent vertices will
// not complete and false will be returned immediately.
func forEachInV(edge reader.Edge, f func(inV int) bool) bool {
	if edge.InV != 0 {
		if !f(edge.InV) {
			return false
		}
	}
	for _, inV := range edge.InVs {
		if !f(inV) {
			return false
		}
	}
	return true
}

// buildForwardGraph returns a map from OutV to InV/InVs properties across all edges of the graph.
func buildForwardGraph(stasher *reader2.Stasher) map[int][]int {
	edges := map[int][]int{}
	_ = stasher.Edges(func(lineContext reader2.LineContext, edge reader.Edge) bool {
		// Note: skip contains relationships because it ruins the visualizer
		// We need to replace this with a smarter graph output that won't go up/down
		// contains relationships: if we have a range, we have ALL ranges in that
		// document due to this relationship.
		// if lineContext.Element.Label == "contains" {
		// 	return true
		// }

		return forEachInV(edge, func(inV int) bool {
			edges[edge.OutV] = append(edges[edge.OutV], inV)
			return true
		})
	})

	return edges
}

func invertEdges(m map[int][]int) map[int][]int {
	inverted := map[int][]int{}
	for k, vs := range m {
		for _, v := range vs {
			inverted[v] = append(inverted[v], k)
		}
	}

	return inverted
}

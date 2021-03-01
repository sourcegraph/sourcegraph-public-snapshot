package validation

import (
	"sort"

	reader "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/protocol/reader"
	reader2 "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/test/internal/reader"
)

var reachabilityIgnoreList = []string{"metaData", "project", "document", "$event"}

// ensureReachability ensures that every vertex (except for metadata, project, document, and $events)
// is reachable by tracing the forward edges starting at the set of range vertices and the document
// that contains them.
func ensureReachability(ctx *ValidationContext) bool {
	visited := traverseGraph(ctx)

	return ctx.Stasher.Vertices(func(lineContext reader2.LineContext) bool {
		for _, label := range reachabilityIgnoreList {
			if lineContext.Element.Label == label {
				return true
			}
		}

		if _, ok := visited[lineContext.Element.ID]; !ok {
			ctx.AddError("vertex %d unreachable from any range", lineContext.Element.ID).AddContext(lineContext)
			return false
		}

		return true
	})
}

// traverseGraph returns a set of vertex identifiers which are reachable by tracing the forward edges
// of the graph starting from the set of contains edges between documents and ranges.
func traverseGraph(ctx *ValidationContext) map[int]struct{} {
	var frontier []int
	_ = ctx.Stasher.Edges(func(lineContext reader2.LineContext, edge reader.Edge) bool {
		if lineContext.Element.Label == "contains" {
			if outContext, ok := ctx.Stasher.Vertex(edge.OutV); ok && outContext.Element.Label == "document" {
				frontier = append(append(frontier, edge.OutV), eachInV(edge)...)
			}
		}

		return true
	})

	edges := buildForwardGraph(ctx)
	visited := map[int]struct{}{}

	for len(frontier) > 0 {
		var top int
		top, frontier = frontier[0], frontier[1:]
		if _, ok := visited[top]; ok {
			continue
		}

		visited[top] = struct{}{}
		frontier = append(frontier, edges[top]...)
	}

	return visited
}

// buildForwardGraph returns a map from OutV to InV/InVs properties across all edges of the graph.
func buildForwardGraph(ctx *ValidationContext) map[int][]int {
	edges := map[int][]int{}
	_ = ctx.Stasher.Edges(func(lineContext reader2.LineContext, edge reader.Edge) bool {
		return forEachInV(edge, func(inV int) bool {
			edges[edge.OutV] = append(edges[edge.OutV], inV)
			return true
		})
	})

	return edges
}

// ensureRangeOwnership ensures that every range vertex is adjacent to a contains
// edge to some document.
func ensureRangeOwnership(ctx *ValidationContext) bool {
	ownershipMap := ctx.OwnershipMap()
	if ownershipMap == nil {
		return false
	}

	return ctx.Stasher.Vertices(func(lineContext reader2.LineContext) bool {
		if lineContext.Element.Label == "range" {
			if _, ok := ownershipMap[lineContext.Element.ID]; !ok {
				ctx.AddError("range %d not owned by any document", lineContext.Element.ID).AddContext(lineContext)
				return false
			}
		}

		return true
	})
}

// ensureDisjointRanges ensures that the set of ranges within a single document are either
// properly nested or completely disjoint.
func ensureDisjointRanges(ctx *ValidationContext) bool {
	ownershipMap := ctx.OwnershipMap()
	if ownershipMap == nil {
		return false
	}

	valid := true
	for documentID, rangeIDs := range invertOwnershipMap(ownershipMap) {
		ranges := make([]reader2.LineContext, 0, len(rangeIDs))
		for _, rangeID := range rangeIDs {
			if r, ok := ctx.Stasher.Vertex(rangeID); ok {
				ranges = append(ranges, r)
			}
		}

		if !ensureDisjoint(ctx, documentID, ranges) {
			valid = false
		}
	}

	return valid
}

// ensureDisjoint marks an error for each pair from the set of ranges which overlap but are not properly
// nested within one `another.
func ensureDisjoint(ctx *ValidationContext, documentID int, ranges []reader2.LineContext) bool {
	sort.Slice(ranges, func(i, j int) bool {
		r1 := ranges[i].Element.Payload.(reader.Range)
		r2 := ranges[j].Element.Payload.(reader.Range)

		// Sort by starting offset (if on the same line, break ties by start character)
		return r1.Start.Line < r2.Start.Line || (r1.Start.Line == r2.Start.Line && r1.Start.Character < r2.Start.Character)
	})

	for i := 1; i < len(ranges); i++ {
		lineContext1 := ranges[i-1]
		lineContext2 := ranges[i]
		r1 := lineContext1.Element.Payload.(reader.Range)
		r2 := lineContext2.Element.Payload.(reader.Range)

		// r1 ends after r2, so r1 properly encloses r2
		if r1.End.Line > r2.End.Line || (r1.End.Line == r2.End.Line && r1.End.Character >= r2.End.Character) {
			continue
		}

		// r1 ends before r2 starts so they are disjoint
		if r1.End.Line < r2.Start.Line || (r1.End.Line == r2.Start.Line && r1.End.Character < r2.Start.Character) {
			continue
		}

		ctx.AddError("ranges overlap in document %d", documentID).AddContext(lineContext1, lineContext2)
		return false
	}

	return true
}

// ensureItemContains ensures that the inVs of every item edge refer to range that belong
// to the document specified by the item edge's document property.
func ensureItemContains(ctx *ValidationContext) bool {
	ownershipMap := ctx.OwnershipMap()
	if ownershipMap == nil {
		return false
	}

	return ctx.Stasher.Edges(func(lineContext reader2.LineContext, edge reader.Edge) bool {
		if lineContext.Element.Label == "item" {
			return forEachInV(edge, func(inV int) bool {
				if ownershipMap[inV].DocumentID != edge.Document {
					ctx.AddError("vertex should be %d owned by document %d", inV, edge.Document).AddContext(lineContext, ownershipMap[inV].LineContext)
					return false
				}

				return true
			})
		}

		return true
	})
}

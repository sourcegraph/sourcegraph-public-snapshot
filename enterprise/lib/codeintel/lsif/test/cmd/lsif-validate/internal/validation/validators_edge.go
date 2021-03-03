package validation

import (
	"strings"

	reader "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/protocol/reader"
	reader2 "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/test/internal/reader"
)

// validateContainsEdge ensures that a range edge attaches a document to a set of ranges.
func validateContainsEdge(ctx *ValidationContext, lineContext reader2.LineContext) bool {
	return validateEdge(ctx, lineContext, nil, func(ctx *ValidationContext, edgeContext, outContext, inContext reader2.LineContext) bool {
		if outContext.Element.Label != "document" {
			// Skip validation of document/project edges
			return true
		}

		return validateLabels(ctx, edgeContext, inContext, []string{"range"})
	})
}

// validateItemEdge ensures that an item edge attaches definition/reference results to ranges
// (or in the case of reference results, possibly another reference result).
func validateItemEdge(ctx *ValidationContext, lineContext reader2.LineContext) bool {
	return validateEdge(ctx, lineContext, nil, func(ctx *ValidationContext, edgeContext, outContext, inContext reader2.LineContext) bool {
		if outContext.Element.Label == "referenceResult" {
			return validateLabels(ctx, edgeContext, inContext, []string{"range", "referenceResult"})
		}

		return validateLabels(ctx, edgeContext, inContext, []string{"range"})
	})
}

// makeGenericEdgeValidator returns an ElementValidator that ensures the edge's outV property
// refers to a vertex with one of the given out labels, and the edge's inV/inVs properties refers
// to vertices with with one of the given in labels.
func makeGenericEdgeValidator(outLabels, inLabels []string) ElementValidator {
	outValidator := func(ctx *ValidationContext, edgeContext, outContext reader2.LineContext) bool {
		return validateLabels(ctx, edgeContext, outContext, outLabels)
	}

	inValidator := func(ctx *ValidationContext, edgeContext, outContext, inContext reader2.LineContext) bool {
		return validateLabels(ctx, edgeContext, inContext, inLabels)
	}

	return func(ctx *ValidationContext, lineContext reader2.LineContext) bool {
		return validateEdge(ctx, lineContext, outValidator, inValidator)
	}
}

// OutValidator is the type of function that is invoked to validate the source vertex of an edge.
type OutValidator func(ctx *ValidationContext, edgeContext, outContext reader2.LineContext) bool

// InValidator is the type of function that is invoked to validate the sink vertex of an edge.
type InValidator func(ctx *ValidationContext, edgeContext, outContext, inContext reader2.LineContext) bool

// validateEdge validates the source and sink vertices of the given edge by invoking the given out and
// in validators. This also ensures that there is at least one sink vertex attached to each edge, and
// if a document property is present that it refers to a known document vertex.
func validateEdge(ctx *ValidationContext, lineContext reader2.LineContext, outValidator OutValidator, inValidator InValidator) bool {
	edge, ok := lineContext.Element.Payload.(reader.Edge)
	if !ok {
		ctx.AddError("illegal payload").AddContext(lineContext)
		return false
	}

	outContext, ok := validateOutV(ctx, lineContext, edge, outValidator)
	if !ok {
		return false
	}

	if !validateInVs(ctx, lineContext, outContext, edge, inValidator) {
		return false
	}

	if !validateEdgeDocument(ctx, lineContext, edge) {
		return false
	}

	return true
}

// validateOutV validates the OutV property of the given edge.
func validateOutV(ctx *ValidationContext, lineContext reader2.LineContext, edge reader.Edge, outValidator OutValidator) (reader2.LineContext, bool) {
	outContext, ok := ctx.Stasher.Vertex(edge.OutV)
	if !ok {
		ctx.AddError("no such vertex %d", edge.OutV).AddContext(lineContext)
		return reader2.LineContext{}, false
	}

	return outContext, outValidator == nil || outValidator(ctx, lineContext, outContext)
}

// validateInVs validates the InV/InVs properties of the given edge.
func validateInVs(ctx *ValidationContext, lineContext, outContext reader2.LineContext, edge reader.Edge, inValidator InValidator) bool {
	if !forEachInV(edge, func(inV int) bool {
		inContext, ok := ctx.Stasher.Vertex(inV)
		if !ok {
			ctx.AddError("no such vertex %d", inV).AddContext(lineContext)
			return false
		}

		return inValidator == nil || inValidator(ctx, lineContext, outContext, inContext)
	}) {
		return false
	}

	if edge.InV == 0 && len(edge.InVs) == 0 {
		ctx.AddError("no InVs are specified").AddContext(lineContext)
		return false
	}

	return true
}

// validateEdgeDocument validates the document property of the given edge.
func validateEdgeDocument(ctx *ValidationContext, lineContext reader2.LineContext, edge reader.Edge) bool {
	if edge.Document == 0 {
		return true
	}

	documentContext, ok := ctx.Stasher.Vertex(edge.Document)
	if !ok {
		ctx.AddError("no such vertex %d", edge.Document).AddContext(lineContext)
		return false
	}
	if !validateLabels(ctx, lineContext, documentContext, []string{"document"}) {
		return false
	}

	return true
}

// validateLabels marks an error and returns false if the given adjacentLineContext does not have one of the given
// labels. The error will contain the given lineContext, which is meant to represent the edge that dictates the
// relationship between its adjacent vertices.
func validateLabels(ctx *ValidationContext, lineContext, adjacentLineContext reader2.LineContext, labels []string) bool {
	for _, label := range labels {
		if adjacentLineContext.Element.Label == label {
			return true
		}
	}

	adjacentID := adjacentLineContext.Element.ID
	types := strings.Join(labels, ", ")
	ctx.AddError("expected vertex %d to be of type %s", adjacentID, types).AddContext(adjacentLineContext, lineContext)
	return false
}

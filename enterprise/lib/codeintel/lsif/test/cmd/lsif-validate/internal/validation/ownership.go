package validation

import (
	reader "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/protocol/reader"
	reader2 "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/test/internal/reader"
)

// OwnershipContext bundles an document identifier and a contains edge that refers to that
// document via its OutV property.
type OwnershipContext struct {
	DocumentID  int
	LineContext reader2.LineContext
}

// ownershipMap uses the given context's Stasher to create a mapping from range identifiers
// to an OwnershipContext value, which bundles a document identifier as well as the parsed
// edge element that ties them together.
func ownershipMap(ctx *ValidationContext) map[int]OwnershipContext {
	ownershipMap := map[int]OwnershipContext{}

	if !ctx.Stasher.Edges(func(lineContext reader2.LineContext, edge reader.Edge) bool {
		if lineContext.Element.Label != "contains" {
			return true
		}
		edge, ok := lineContext.Element.Payload.(reader.Edge)
		if !ok {
			return true
		}
		if outContext, ok := ctx.Stasher.Vertex(edge.OutV); !ok || outContext.Element.Label != "document" {
			return true
		}

		return forEachInV(edge, func(inV int) bool {
			if other, ok := ownershipMap[inV]; ok {
				ctx.AddError("range %d already claimed by document %d", inV, other.DocumentID).AddContext(lineContext, other.LineContext)
				return false
			}

			ownershipMap[inV] = OwnershipContext{DocumentID: edge.OutV, LineContext: lineContext}
			return true
		})
	}) {
		return nil
	}

	return ownershipMap
}

// invertOwnershipMap converts the given ownership map to return a map from document
// identifiers to the set of range identifiers that document contains.
func invertOwnershipMap(m map[int]OwnershipContext) map[int][]int {
	inverted := map[int][]int{}
	for rangeID, ownershipContext := range m {
		inverted[ownershipContext.DocumentID] = append(inverted[ownershipContext.DocumentID], rangeID)
	}

	return inverted
}

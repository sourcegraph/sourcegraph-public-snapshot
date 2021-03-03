package validation

import reader2 "github.com/sourcegraph/sourcegraph/enterprise/lib/codeintel/lsif/test/internal/reader"

// ElementValidator validates specific properties of a single vertex or edge element.
type ElementValidator func(ctx *ValidationContext, lineContext reader2.LineContext) bool

// vertexValidators is a map from vertex labels to that vertex type's validator.
var vertexValidators = map[string]ElementValidator{
	"metaData": validateMetaDataVertex,
	"document": validateDocumentVertex,
	"range":    validateRangeVertex,
}

// edgeValidators is a map from edge labels to that edge type's validator.
var edgeValidators = map[string]ElementValidator{
	"contains":                validateContainsEdge,
	"item":                    validateItemEdge,
	"next":                    makeGenericEdgeValidator([]string{"range", "resultSet"}, []string{"resultSet"}),
	"textDocument/definition": makeGenericEdgeValidator([]string{"range", "resultSet"}, []string{"definitionResult"}),
	"textDocument/references": makeGenericEdgeValidator([]string{"range", "resultSet"}, []string{"referenceResult"}),
	"textDocument/hover":      makeGenericEdgeValidator([]string{"range", "resultSet"}, []string{"hoverResult"}),
	"moniker":                 makeGenericEdgeValidator([]string{"range", "resultSet"}, []string{"moniker"}),
	"nextMoniker":             makeGenericEdgeValidator([]string{"moniker"}, []string{"moniker"}),
	"packageInformation":      makeGenericEdgeValidator([]string{"moniker"}, []string{"packageInformation"}),
}

// RelationshipValidator validates a specific property across all vertex and edges
// registered to the given context's stasher.
type RelationshipValidator func(ctx *ValidationContext) bool

// relationshipValidators is the set of validators that operate across the entire LSIF graph.
var relationshipValidators = []RelationshipValidator{
	ensureReachability,
	ensureRangeOwnership,
	ensureDisjointRanges,
	ensureItemContains,
}

package types

import "github.com/sourcegraph/scip/bindings/go/scip"

// FlattenDocuments merges elements of the given slice with the same relative path. This allows us to make
// the assumption post-canonicalization that each index has one representation of a given document path in
// the database. This function returns a new slice.
func FlattenDocuments(documents []*scip.Document) []*scip.Document {
	documentMap := make(map[string]*scip.Document, len(documents))
	for _, document := range documents {
		existing, ok := documentMap[document.RelativePath]
		if !ok {
			documentMap[document.RelativePath] = document
			continue
		}
		if existing.Language != document.Language {
			_ = 0 // TODO - warn?
		}

		existing.Symbols = append(existing.Symbols, document.Symbols...)
		existing.Occurrences = append(existing.Occurrences, document.Occurrences...)
	}

	flattened := make([]*scip.Document, 0, len(documentMap))
	for _, document := range documentMap {
		flattened = append(flattened, document)
	}

	return flattened
}

// FlattenSymbol merges elements of the given slice with the same symbol name. This allows us to make the
// assumption post-canonicalization that each index and document refer to one symbol metadata object uniquely.
// This function returns a new slice.
func FlattenSymbols(symbols []*scip.SymbolInformation) []*scip.SymbolInformation {
	symbolMap := make(map[string]*scip.SymbolInformation, len(symbols))
	for _, symbol := range symbols {
		existing, ok := symbolMap[symbol.Symbol]
		if !ok {
			symbolMap[symbol.Symbol] = symbol
			continue
		}

		existing.Documentation = combineDocumentation(existing.Documentation, symbol.Documentation)
		existing.Relationships = append(existing.Relationships, symbol.Relationships...)
	}

	flattened := make([]*scip.SymbolInformation, 0, len(symbolMap))
	for _, symbol := range symbolMap {
		flattened = append(flattened, symbol)
	}

	return flattened
}

// FlattenOccurrences merges elements of the given slice with equivalent bounds. This function returns a new slice.
func FlattenOccurrences(occurrences []*scip.Occurrence) []*scip.Occurrence {
	if len(occurrences) == 0 {
		return occurrences
	}

	_ = SortOccurrences(occurrences)
	flattened := make([]*scip.Occurrence, 0, len(occurrences))
	flattened = append(flattened, occurrences[0])

	for _, occurrence := range occurrences[1:] {
		top := flattened[len(flattened)-1]

		if !rawRangesEqual(top.Range, occurrence.Range) {
			flattened = append(flattened, occurrence)
			continue
		}
		if top.Symbol != occurrence.Symbol {
			flattened = append(flattened, occurrence)
			continue
		}

		if top.SyntaxKind == scip.SyntaxKind_UnspecifiedSyntaxKind {
			// Take first valid syntax kind
			top.SyntaxKind = occurrence.SyntaxKind
		}

		// Combine all other fields
		top.SymbolRoles |= occurrence.SymbolRoles
		top.OverrideDocumentation = append(top.OverrideDocumentation, occurrence.OverrideDocumentation...)
		top.Diagnostics = append(top.Diagnostics, occurrence.Diagnostics...)
	}

	return flattened
}

// FlattenRelationship merges elements of the given slice with equivalent symbol names. This function returns a new
// slice.
func FlattenRelationship(relationships []*scip.Relationship) []*scip.Relationship {
	relationshipMap := make(map[string][]*scip.Relationship, len(relationships))
	for _, relationship := range relationships {
		relationshipMap[relationship.Symbol] = append(relationshipMap[relationship.Symbol], relationship)
	}

	flattened := make([]*scip.Relationship, 0, len(relationshipMap))
	for _, relationships := range relationshipMap {
		combined := relationships[0]
		for _, relationship := range relationships[1:] {
			combined.IsReference = combined.IsReference || relationship.IsReference
			combined.IsImplementation = combined.IsImplementation || relationship.IsImplementation
			combined.IsTypeDefinition = combined.IsTypeDefinition || relationship.IsTypeDefinition
			combined.IsDefinition = combined.IsDefinition || relationship.IsDefinition
		}

		flattened = append(flattened, combined)
	}

	return flattened
}

// combineDocumentation merges documentation components from two separate symbol information objects.
func combineDocumentation(existing, additional []string) []string {
	filtered := make([]string, 0, len(additional))
	for _, s := range additional {
		if !stringSliceContains(existing, s) {
			filtered = append(filtered, s)
		}
	}

	return append(existing, filtered...)
}

func stringSliceContains(slice []string, target string) bool {
	for _, candidate := range slice {
		if target == candidate {
			return true
		}
	}

	return false
}

package types

import "github.com/sourcegraph/scip/bindings/go/scip"

// TODO - document
func FlattenDocuments(documents []*scip.Document) []*scip.Document {
	documentMap := make(map[string]*scip.Document, len(documents))
	for _, document := range documents {
		existing, ok := documentMap[document.RelativePath]
		if !ok {
			documentMap[document.RelativePath] = document
			continue
		}
		if existing.Language != document.Language {
			// TODO - warn?
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

// TODO - document
func FlattenSymbols(symbols []*scip.SymbolInformation) []*scip.SymbolInformation {
	symbolMap := make(map[string]*scip.SymbolInformation, len(symbols))
	for _, symbol := range symbols {
		existing, ok := symbolMap[symbol.Symbol]
		if !ok {
			symbolMap[symbol.Symbol] = symbol
			continue
		}

		existing.Documentation = append(existing.Documentation, symbol.Documentation...)
		existing.Relationships = append(existing.Relationships, symbol.Relationships...)
	}

	flattened := make([]*scip.SymbolInformation, 0, len(symbolMap))
	for _, symbol := range symbolMap {
		flattened = append(flattened, symbol)
	}

	return flattened
}

// TODO - document
func FlattenOccurrences(occurrences []*scip.Occurrence) []*scip.Occurrence {
	_ = SortOccurrences(occurrences)

	flattened := make([]*scip.Occurrence, 0, len(occurrences))
	for _, occurrence := range occurrences {
		if len(flattened) == 0 || flattened[len(flattened)-1].Symbol != occurrence.Symbol {
			flattened = append(flattened, occurrence)
			continue
		}
		existing := flattened[len(flattened)-1]
		if existing.SyntaxKind != occurrence.SyntaxKind {
			// TODO - warn?
		}

		existing.SymbolRoles |= occurrence.SymbolRoles
		existing.OverrideDocumentation = append(existing.OverrideDocumentation, occurrence.OverrideDocumentation...)
		existing.Diagnostics = append(existing.Diagnostics, occurrence.Diagnostics...)
	}

	return flattened
}

// TODO - document
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

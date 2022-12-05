package types

import "github.com/sourcegraph/scip/bindings/go/scip"

// CanonicalizeDocument deterministically re-orders the fields of the given document.
func CanonicalizeDocument(document *scip.Document) *scip.Document {
	document.Occurrences = CanonicalizeOccurrences(document.Occurrences)
	document.Symbols = CanonicalizeSymbols(document.Symbols)
	return document
}

// CanonicalizeOccurrences deterministically re-orders the fields of the given occurrence slice.
func CanonicalizeOccurrences(occurrences []*scip.Occurrence) []*scip.Occurrence {
	canonicalized := make([]*scip.Occurrence, 0, len(occurrences))
	for _, occurrence := range FlattenOccurrences(occurrences) {
		canonicalized = append(canonicalized, CanonicalizeOccurrence(occurrence))
	}

	return SortOccurrences(canonicalized)
}

// CanonicalizeOccurrence deterministically re-orders the fields of the given occurrence.
func CanonicalizeOccurrence(occurrence *scip.Occurrence) *scip.Occurrence {
	// Express ranges as three-components if possible
	occurrence.Range = scip.NewRange(occurrence.Range).SCIPRange()
	occurrence.Diagnostics = CanonicalizeDiagnostics(occurrence.Diagnostics)
	return occurrence
}

// CanonicalizeDiagnostics deterministically re-orders the fields of the given diagnostic slice.
func CanonicalizeDiagnostics(diagnostics []*scip.Diagnostic) []*scip.Diagnostic {
	canonicalized := make([]*scip.Diagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		canonicalized = append(canonicalized, CanonicalizeDiagnostic(diagnostic))
	}

	return SortDiagnostics(canonicalized)
}

// CanonicalizeDiagnostic deterministically re-orders the fields of the given diagnostic.
func CanonicalizeDiagnostic(diagnostic *scip.Diagnostic) *scip.Diagnostic {
	diagnostic.Tags = SortDiagnosticTags(diagnostic.Tags)
	return diagnostic
}

// CanonicalizeSymbols deterministically re-orders the fields of the given symbols slice.
func CanonicalizeSymbols(symbols []*scip.SymbolInformation) []*scip.SymbolInformation {
	canonicalized := make([]*scip.SymbolInformation, 0, len(symbols))
	for _, symbol := range FlattenSymbols(symbols) {
		canonicalized = append(canonicalized, CanonicalizeSymbol(symbol))
	}

	return SortSymbols(canonicalized)
}

// CanonicalizeSymbol deterministically re-orders the fields of the given symbol.
func CanonicalizeSymbol(symbol *scip.SymbolInformation) *scip.SymbolInformation {
	symbol.Relationships = CanonicalizeRelationships(symbol.Relationships)
	return symbol
}

// CanonicalizeRelationships deterministically re-orders the fields of the given relationship slice.
func CanonicalizeRelationships(relationships []*scip.Relationship) []*scip.Relationship {
	return SortRelationships(FlattenRelationship(relationships))
}

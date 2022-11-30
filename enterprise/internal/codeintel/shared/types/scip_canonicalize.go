package types

import (
	"sort"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

// CanonicalizeDocument deterministically re-orders the fields of the given document.
// The input is modified in-place but returned for convenience.
func CanonicalizeDocument(document *scip.Document) *scip.Document {
	_ = CanonicalizeOccurrences(document.Occurrences)
	_ = CanonicalizeSymbols(document.Symbols)

	return document
}

// CanonicalizeOccurrences deterministically re-orders the fields of the given occurrence slice.
// The input is modified in-place but returned for convenience.
func CanonicalizeOccurrences(occurrences []*scip.Occurrence) []*scip.Occurrence {
	for _, occurrence := range occurrences {
		_ = CanonicalizeOccurrence(occurrence)
	}

	return SortOccurrences(occurrences)
}

// CanonicalizeOccurrence deterministically re-orders the fields of the given occurrence.
// The input is modified in-place but returned for convenience.
func CanonicalizeOccurrence(occurrence *scip.Occurrence) *scip.Occurrence {
	// Express ranges as three-components if possible
	occurrence.Range = scip.NewRange(occurrence.Range).SCIPRange()
	_ = CanonicalizeDiagnostics(occurrence.Diagnostics)
	return occurrence
}

// CanonicalizeDiagnostics deterministically re-orders the fields of the given diagnostic slice.
// The input is modified in-place but returned for convenience.
func CanonicalizeDiagnostics(diagnostics []*scip.Diagnostic) []*scip.Diagnostic {
	for _, diagnostic := range diagnostics {
		_ = CanonicalizeDiagnostic(diagnostic)
	}

	return SortDiagnostics(diagnostics)
}

// CanonicalizeDiagnostic deterministically re-orders the fields of the given diagnostic.
// The input is modified in-place but returned for convenience.
func CanonicalizeDiagnostic(diagnostic *scip.Diagnostic) *scip.Diagnostic {
	sort.Slice(diagnostic.Tags, func(i, j int) bool {
		return diagnostic.Tags[i] < diagnostic.Tags[j]
	})

	return diagnostic
}

// CanonicalizeSymbols deterministically re-orders the fields of the given symbols slice.
// The input is modified in-place but returned for convenience.
func CanonicalizeSymbols(symbols []*scip.SymbolInformation) []*scip.SymbolInformation {
	for _, symbol := range symbols {
		_ = CanonicalizeSymbol(symbol)
	}

	return SortSymbols(symbols)
}

// CanonicalizeSymbol deterministically re-orders the fields of the given symbol.
// The input is modified in-place but returned for convenience.
func CanonicalizeSymbol(symbol *scip.SymbolInformation) *scip.SymbolInformation {
	sort.Slice(symbol.Relationships, func(i, j int) bool {
		return symbol.Relationships[i].Symbol < symbol.Relationships[j].Symbol
	})

	return symbol
}

package types

import (
	"unicode/utf8"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

// SanitizeDocument ensures that all strings in the given document are valid UTF-8.
// This is a requirement for successful protobuf encoding.
func SanitizeDocument(document *scip.Document) *scip.Document {
	document.Language = sanitizeString(document.Language)
	document.RelativePath = sanitizeString(document.RelativePath)
	document.Occurrences = SanitizeOccurrences(document.Occurrences)
	document.Symbols = SanitizeSymbols(document.Symbols)
	return document
}

// SanitizeOccurrences ensures that all strings in the given occurrence slice are valid UTF-8.
// The input slice is modified in-place but returned for convenience.
// This is a requirement for successful protobuf encoding.
func SanitizeOccurrences(occurrences []*scip.Occurrence) []*scip.Occurrence {
	for i, occurrence := range occurrences {
		occurrences[i] = SanitizeOccurrence(occurrence)
	}

	return occurrences
}

// SanitizeOccurrence ensures that all strings in the given occurrence are valid UTF-8.
// This is a requirement for successful protobuf encoding.
func SanitizeOccurrence(occurrence *scip.Occurrence) *scip.Occurrence {
	occurrence.Symbol = sanitizeString(occurrence.Symbol)
	occurrence.OverrideDocumentation = sanitizeStringSlice(occurrence.OverrideDocumentation)
	occurrence.Diagnostics = SanitizeDiagnostics(occurrence.Diagnostics)
	return occurrence
}

// SanitizeDiagnostics ensures that all strings in the given diagnostic slice are valid UTF-8.
// The input slice is modified in-place but returned for convenience.
// This is a requirement for successful protobuf encoding.
func SanitizeDiagnostics(diagnostics []*scip.Diagnostic) []*scip.Diagnostic {
	for i, diagnostic := range diagnostics {
		diagnostics[i] = SanitizeDiagnostic(diagnostic)
	}

	return diagnostics
}

// SanitizeDiagnostic ensures that all strings in the given diagnostic are valid UTF-8.
// This is a requirement for successful protobuf encoding.
func SanitizeDiagnostic(diagnostic *scip.Diagnostic) *scip.Diagnostic {
	diagnostic.Code = sanitizeString(diagnostic.Code)
	diagnostic.Message = sanitizeString(diagnostic.Message)
	diagnostic.Source = sanitizeString(diagnostic.Source)
	return diagnostic
}

// SanitizeSymbols ensures that all strings in the given symbols slice are valid UTF-8.
// The input slice is modified in-place but returned for convenience.
// This is a requirement for successful protobuf encoding.
func SanitizeSymbols(symbols []*scip.SymbolInformation) []*scip.SymbolInformation {
	for i, symbol := range symbols {
		symbols[i] = SanitizeSymbol(symbol)
	}

	return symbols
}

// SanitizeSymbol ensures that all strings in the given symbol are valid UTF-8.
// This is a requirement for successful protobuf encoding.
func SanitizeSymbol(symbol *scip.SymbolInformation) *scip.SymbolInformation {
	symbol.Symbol = sanitizeString(symbol.Symbol)
	symbol.Documentation = sanitizeStringSlice(symbol.Documentation)

	for _, relationship := range symbol.Relationships {
		relationship.Symbol = sanitizeString(relationship.Symbol)
	}

	return symbol
}

// sanitizeStringSlice ensures the strings in the given slice are all valid UTF-8.
// The input slice is modified in-place but returned for convenience.
// This is a requirement for successful protobuf encoding.
func sanitizeStringSlice(ss []string) []string {
	for i, s := range ss {
		ss[i] = sanitizeString(s)
	}

	return ss
}

// sanitizeString coerces a string into valid UTF-8 (if it's not already).
func sanitizeString(s string) string {
	if utf8.ValidString(s) {
		return s
	}

	// This seems redundant, but it isn't: it magically makes the string valid UTF-8
	return string([]rune(s))
}

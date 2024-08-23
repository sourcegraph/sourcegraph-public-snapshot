package scip

import (
	"sort"

	"golang.org/x/exp/slices"
)

// FindSymbol returns the symbol with the given name in the given document. If there is no symbol by
// that name, this function returns nil. Prefer using FindSymbolBinarySearch over this function.
func FindSymbol(document *Document, symbolName string) *SymbolInformation {
	for _, symbol := range document.Symbols {
		if symbol.Symbol == symbolName {
			return symbol
		}
	}

	return nil
}

// FindSymbolBinarySearch attempts to find the SymbolInformation in the given document.
//
// Pre-condition: The symbols array must be sorted in ascending order based on the symbol name,
// and SymbolInformation values must be merged. This guarantee is upheld by CanonicalizeDocument.
func FindSymbolBinarySearch(canonicalizedDocument *Document, symbolName string) *SymbolInformation {
	i, found := slices.BinarySearchFunc(canonicalizedDocument.Symbols, symbolName, func(sym *SymbolInformation, lookup string) int {
		if sym.Symbol < lookup {
			return -1
		} else if sym.Symbol == lookup {
			return 0
		}
		return 1
	})
	if found {
		return canonicalizedDocument.Symbols[i]
	}
	return nil
}

// SortDocuments sorts the given documents slice (in-place) and returns it (for convenience). Documents
// are sorted in ascending order of their relative path.
func SortDocuments(documents []*Document) []*Document {
	sort.Slice(documents, func(i, j int) bool {
		return documents[i].RelativePath < documents[j].RelativePath
	})

	return documents
}

// FindOccurrences filters the given slice of occurrences and returns those that contain the position
// represented by line and character. The order of the output slice is "inside-out", so that earlier
// occurrences are properly enclosed by later occurrences.
func FindOccurrences(occurrences []*Occurrence, targetLine, targetCharacter int32) []*Occurrence {
	var filtered []*Occurrence
	pos := Position{targetLine, targetCharacter}
	for _, occurrence := range occurrences {
		if NewRangeUnchecked(occurrence.Range).Contains(pos) {
			filtered = append(filtered, occurrence)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		// Ordered so that the least precise (largest) range comes last
		return NewRangeUnchecked(filtered[i].Range).CompareStrict(NewRangeUnchecked(filtered[j].Range)) > 0
	})

	return filtered
}

// SortOccurrences sorts the given occurrence slice (in-place) and returns it (for convenience).
// Occurrences sorted in ascending order of their range's starting position, where enclosing ranges
// come before the enclosed. If there are multiple occurrences with the exact same range, then the
// occurrences are sorted by symbol name.
func SortOccurrences(occurrences []*Occurrence) []*Occurrence {
	sort.Slice(occurrences, func(i, j int) bool {
		r1 := NewRangeUnchecked(occurrences[i].Range)
		r2 := NewRangeUnchecked(occurrences[j].Range)
		if ret := r1.CompareStrict(r2); ret != 0 {
			return ret < 0
		}
		return occurrences[i].Symbol < occurrences[j].Symbol
	})

	return occurrences
}

// rawRangesEqual compares the given SCIP-encoded raw ranges for equality.
func rawRangesEqual(a, b []int32) bool {
	if len(a) == len(b) {
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}

		return true
	}

	ra := NewRangeUnchecked(a)
	rb := NewRangeUnchecked(b)

	return ra.Start.Line == rb.Start.Line && ra.Start.Character == rb.Start.Character && ra.End.Line == rb.End.Line && ra.End.Character == rb.End.Character
}

// SortRanges sorts the given range slice (in-place) and returns it (for convenience). Ranges are
// sorted in ascending order of starting position, where enclosing ranges come before the enclosed.
func SortRanges(ranges []Range) []Range {
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].LessStrict(ranges[j])
	})
	return ranges
}

// SortSymbols sorts the given symbols slice (in-place) and returns it (for convenience).
// Symbol information objects are sorted in ascending order by name.
func SortSymbols(symbols []*SymbolInformation) []*SymbolInformation {
	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].Symbol < symbols[j].Symbol
	})

	return symbols
}

// SortDiagnostics sorts the given diagnostics slice (in-place) and returns it (for convenience).
// Diagnostics are sorted first by severity (more severe earlier in the slice) and then by the
// diagnostic message.
func SortDiagnostics(diagnostics []*Diagnostic) []*Diagnostic {
	sort.Slice(diagnostics, func(i, j int) bool {
		if diagnostics[i].Severity < diagnostics[j].Severity {
			return true
		} else if diagnostics[i].Severity == diagnostics[j].Severity {
			return diagnostics[i].Message < diagnostics[j].Message
		}

		return false
	})

	return diagnostics
}

// SortDiagnosticTags sorts the given diagnostic tags slice (in-place) and returns it (for convenience).
func SortDiagnosticTags(tags []DiagnosticTag) []DiagnosticTag {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i] < tags[j]
	})

	return tags
}

// SortRelationships sorts the given symbol relationships slice (in-place) and returns it (for convenience).
func SortRelationships(relationships []*Relationship) []*Relationship {
	sort.Slice(relationships, func(i, j int) bool {
		return relationships[i].Symbol < relationships[j].Symbol
	})

	return relationships
}

// comparePositionToRange compares the given target position represented by line and character
// against the four-element range bound. This method returns
//
// - -1 if the position occurs before the range,
// - +1 if the position occurs after the range, and
// - +0 if the position is enclosed by the range.
//
// Note that ranges are half-closed intervals, so a match on the leading end of the range will
// be considered enclosed, but a match on the trailing edge will not.
func comparePositionToRange(
	startLine int32,
	startCharacter int32,
	endLine int32,
	endCharacter int32,
	targetLine int32,
	targetCharacter int32,
) int {
	// line before range
	if targetLine < startLine {
		return 1
	}

	// line after range
	if targetLine > endLine {
		return -1
	}

	// on first line, character before start of range
	if targetLine == startLine && targetCharacter < startCharacter {
		return 1
	}

	// on last line; character after end of range
	if targetLine == endLine && targetCharacter >= endCharacter {
		return -1
	}

	// enclosed by range
	return 0
}

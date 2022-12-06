package types

import (
	"sort"

	"github.com/sourcegraph/scip/bindings/go/scip"
)

// FindSymbol returns the symbol with the given name in the given document. If there is no symbol by
// that name, this function returns nil.
func FindSymbol(document *scip.Document, symbolName string) *scip.SymbolInformation {
	for _, symbol := range document.Symbols {
		if symbol.Symbol == symbolName {
			return symbol
		}
	}

	return nil
}

// SortDocuments sorts the given documents slice (in-place) and returns it (for convenience). Documents
// are sorted in ascending order of their relative path.
func SortDocuments(documents []*scip.Document) []*scip.Document {
	sort.Slice(documents, func(i, j int) bool {
		return documents[i].RelativePath < documents[j].RelativePath
	})

	return documents
}

// FindOccurrences filters the given slice of occurrences and returns those that contain the position
// represented by line and character. The order of the output slice is "inside-out", so that earlier
// occurrences are properly enclosed by later occurrences.
func FindOccurrences(occurrences []*scip.Occurrence, targetLine, targetCharacter int32) []*scip.Occurrence {
	var filtered []*scip.Occurrence
	for _, occurrence := range occurrences {
		if compareRanges(occurrence.Range, targetLine, targetCharacter) == 0 {
			filtered = append(filtered, occurrence)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		// Ordered so that the least precise (largest) range comes first
		return compareRanges(filtered[i].Range, filtered[j].Range...) > 0
	})

	return filtered
}

// SortOccurrences sorts the given occurrence slice (in-place) and returns it (for convenience).
// Occurrences sorted in ascending order of their range's starting position, where enclosing ranges
// come before the enclosed. If there are multiple occurrences with the exact same range, then the
// occurrences are sorted by symbol name.
func SortOccurrences(occurrences []*scip.Occurrence) []*scip.Occurrence {
	sort.Slice(occurrences, func(i, j int) bool {
		if rawRangesEqual(occurrences[i].Range, occurrences[j].Range) {
			return occurrences[i].Symbol < occurrences[j].Symbol
		}

		return compareRanges(occurrences[i].Range, occurrences[j].Range...) <= 0
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

	ra := scip.NewRange(a)
	rb := scip.NewRange(b)

	return ra.Start.Line == rb.Start.Line && ra.Start.Character == rb.Start.Character && ra.End.Line == rb.End.Line && ra.End.Character == rb.End.Character
}

// SortRanges sorts the given range slice (in-place) and returns it (for convenience). Ranges are
// sorted in ascending order of starting position, where enclosing ranges come before the enclosed.
func SortRanges(ranges []*scip.Range) []*scip.Range {
	sort.Slice(ranges, func(i, j int) bool {
		return comparePositionToRange(
			ranges[i].Start.Line,
			ranges[i].Start.Character,
			ranges[i].End.Line,
			ranges[i].End.Character,
			ranges[j].Start.Line,
			ranges[j].Start.Character,
		) <= 0
	})

	return ranges
}

// SortSymbols sorts the given symbols slice (in-place) and returns it (for convenience).
// Symbol information objects are sorted in ascending order by name.
func SortSymbols(symbols []*scip.SymbolInformation) []*scip.SymbolInformation {
	sort.Slice(symbols, func(i, j int) bool {
		return symbols[i].Symbol < symbols[j].Symbol
	})

	return symbols
}

// SortDiagnostics sorts the given diagnostics slice (in-place) and returns it (for convenience).
// Diagnostics are sorted first by severity (more severe earlier in the slice) and then by the
// diagnostic message.
func SortDiagnostics(diagnostics []*scip.Diagnostic) []*scip.Diagnostic {
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
func SortDiagnosticTags(tags []scip.DiagnosticTag) []scip.DiagnosticTag {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i] < tags[j]
	})

	return tags
}

// SortRelationships sorts the given symbol relationships slice (in-place) and returns it (for convenience).
func SortRelationships(relationships []*scip.Relationship) []*scip.Relationship {
	sort.Slice(relationships, func(i, j int) bool {
		return relationships[i].Symbol < relationships[j].Symbol
	})

	return relationships
}

// compareRanges compares the order of the leading edge of the two ranges. This method returns
//
// - -1 if the leading edge of r2 occurs before r1,
// - +1 if the leading edge of r2 occurs after r1, and
// - +0 if the leading edge of r2 is enclosed by r1.
//
// Note that ranges are half-closed intervals, so a match on the leading end of the range will
// be considered enclosed, but a match on the trailing edge will not.
func compareRanges(r1 []int32, r2 ...int32) int {
	startLine, startCharacter, endLine, endCharacter := unpackRange(r1)

	return comparePositionToRange(
		startLine,
		startCharacter,
		endLine,
		endCharacter,
		r2[0],
		r2[1],
	)
}

// unpackRange unpacks the raw SCIP range into a four-element range bound. This function
// duplicates some of the logic in the SCIP repository, but we're dealing heavily with raw
// encoded proto messages in the database layer here as well, and we'd like to avoid boxing
// into a scip.Range unnecessarily.
func unpackRange(r []int32) (int32, int32, int32, int32) {
	if len(r) == 3 {
		return r[0], r[1], r[0], r[2]
	}

	return r[0], r[1], r[2], r[3]
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

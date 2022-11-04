package keyword

import (
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

// annotatedMatchesToChunkMatches converts a string containing matches annotated with annotationCharacter into chunk matches.
// Each match should be surrounded by annotationCharacter (e.g., if annotationCharacter="_" -> `func _match1__match2_()`).
// It does not support matches across new lines.
func annotatedMatchesToChunkMatches(annotatedMatches string, annotationCharacter string) result.ChunkMatches {
	lines := strings.Split(annotatedMatches, "\n")
	matchesPerLine := [][]string{}
	for i := 0; i < len(lines); i++ {
		matchesPerLine = append(matchesPerLine, strings.Split(lines[i], annotationCharacter))
	}

	chunkMatches := result.ChunkMatches{}
	offset := 0
	for lineIdx, line := range matchesPerLine {
		lineOffset := offset
		ranges := result.Ranges{}
		for i := 0; i < len(line); i++ {
			if i%2 == 1 {
				// Match
				ranges = append(ranges, result.Range{
					Start: result.Location{Offset: lineOffset, Line: lineIdx, Column: lineOffset},
					End:   result.Location{Offset: lineOffset + len(line[i]), Line: lineIdx, Column: lineOffset + len(line[i])},
				})
			}
			lineOffset += len(line[i])
		}

		lineContent := strings.Join(line, "")
		if len(ranges) > 0 {
			chunkMatches = append(chunkMatches, result.ChunkMatch{
				ContentStart: result.Location{Offset: offset, Line: lineIdx, Column: 0},
				Content:      lineContent,
				Ranges:       ranges,
			})
		}

		offset += len(lineContent) + 1 // including new line
	}

	return chunkMatches
}

func TestGroupChunkMatches(t *testing.T) {
	const annotatedMatches = `
func _abc_(_def_ int) {
	// _abc_ _def_ _123_
}

const xyz = 1 + 1
const _abc_ = _def_

// xyz
func small_abc__abc_() {

	// _abc_ _def_

	// _123_ _def_ _abc_
}
`
	chunkMatches := annotatedMatchesToChunkMatches(annotatedMatches, "_")
	fileMatch := result.FileMatch{File: result.File{}, Symbols: []*result.SymbolMatch{}, LimitHit: false, ChunkMatches: chunkMatches}
	groups := groupChunkMatches(&fileMatch, 0.0, chunkMatches, 3)

	groupRelevancy := []bool{true, false, true}
	if len(groups) != len(groupRelevancy) {
		t.Fatalf("expected %d groups, got %d", len(groupRelevancy), len(groups))
	}

	for idx, group := range groups {
		isRelevant := group.IsRelevant()
		if isRelevant != groupRelevancy[idx] {
			t.Fatalf("expected group %d relevancy to be %v, got %v", idx, groupRelevancy[idx], isRelevant)
		}
	}
}

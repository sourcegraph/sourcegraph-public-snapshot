package search

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/byteutils"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

const (
	matchContextLines = 1
	maxLinesPerHunk   = 5
	maxHunksPerFile   = 3
	maxFiles          = 5
)

var escaper = strings.NewReplacer(" ", `\ `)

func escapeUTF8AndSpaces(s string) string {
	return escaper.Replace(strings.ToValidUTF8(s, "�"))
}

func FormatDiff(rawDiff []*diff.FileDiff, highlights map[int]MatchedFileDiff) (string, result.Ranges) {
	var buf strings.Builder
	var loc result.Location
	var ranges result.Ranges

	fileCount := 0
	for fileIdx, fileDiff := range rawDiff {
		fdh, ok := highlights[fileIdx]
		if !ok && len(highlights) > 0 {
			continue
		}
		if fileCount >= maxFiles {
			break
		}
		fileCount++

		ranges = append(ranges, fdh.OldFile.Add(loc)...)
		// NOTE(@camdencheek): this does not correctly update the highlight ranges of files with spaces in the name.
		// Doing so would require a smarter replacer.
		escaped := escapeUTF8AndSpaces(fileDiff.OrigName)
		buf.WriteString(escaped)
		buf.WriteByte(' ')
		loc.Offset = buf.Len()
		loc.Column = len(escaped) + len(" ")

		ranges = append(ranges, fdh.NewFile.Add(loc)...)
		buf.WriteString(escapeUTF8AndSpaces(fileDiff.NewName))
		buf.WriteByte('\n')
		loc.Offset = buf.Len()
		loc.Line++
		loc.Column = 0

		filteredHunks, filteredHighlights := splitHunkMatches(fileDiff.Hunks, fdh.MatchedHunks, matchContextLines, maxLinesPerHunk)

		hunkCount := 0
		for hunkIdx, hunk := range filteredHunks {
			hmh, ok := filteredHighlights[hunkIdx]
			if !ok && len(filteredHighlights) > 0 {
				continue
			}
			if hunkCount >= maxHunksPerFile {
				break
			}
			hunkCount++

			fmt.Fprintf(&buf,
				"@@ -%d,%d +%d,%d @@ %s\n",
				hunk.OrigStartLine,
				hunk.OrigLines,
				hunk.NewStartLine,
				hunk.NewLines,
				hunk.Section,
			)
			loc.Offset = buf.Len()
			loc.Line++
			loc.Column = 0

			lr := byteutils.NewLineReader(hunk.Body)
			lineIdx := -1
			for lr.Scan() {
				line := lr.Line()
				lineIdx++

				if len(line) == 0 {
					continue
				}

				prefix, lineWithoutPrefix := line[0], line[1:]
				buf.WriteByte(prefix)
				loc.Offset = buf.Len()
				loc.Column = 1

				if lineHighlights, ok := hmh.MatchedLines[lineIdx]; ok {
					ranges = append(ranges, lineHighlights.Add(loc)...)
				}

				buf.Write(bytes.ToValidUTF8(lineWithoutPrefix, []byte("�")))
				buf.WriteByte('\n')
				loc.Offset = buf.Len()
				loc.Line++
				loc.Column = 0
			}
		}
	}

	return buf.String(), ranges
}

// splitHunkMatches returns a list of hunks that are a subset of the input hunks,
// filtered down to only hunks that matched, determined by whether the hunk has highlights.
// and non-matching changed lines are eliminated, and the hunk header (start/end
// lines) are adjusted accordingly. The provided highlights are adjusted accordingly.
func splitHunkMatches(hunks []*diff.Hunk, hunkHighlights map[int]MatchedHunk, matchContextLines, maxLinesPerHunk int) (results []*diff.Hunk, newHighlights map[int]MatchedHunk) {
	addExtraHunkMatchesSection := func(hunk *diff.Hunk, extraHunkMatches int) {
		if extraHunkMatches > 0 {
			if hunk.Section != "" {
				hunk.Section += " "
			}
			hunk.Section += fmt.Sprintf("... +%d", extraHunkMatches)
		}
	}

	newHighlights = make(map[int]MatchedHunk, len(hunkHighlights))

	for i, hunk := range hunks {
		var cur *diff.Hunk
		var curLines [][]byte
		origHighlights := hunkHighlights[i].MatchedLines
		curHighlights := make(map[int]result.Ranges, len(hunkHighlights))

		lines := bytes.SplitAfter(hunk.Body, []byte("\n"))
		lineInfo := computeDiffHunkInfo(lines, hunkHighlights[i].MatchedLines, matchContextLines)

		extraHunkMatches := 0
		var origLineOffset, newLineOffset int32
		for i, line := range lines {
			if len(line) == 0 {
				break
			}
			lineInfo := lineInfo[i]
			if lineInfo.matching || lineInfo.context {
				if maxLinesPerHunk > 0 && len(curLines) == matchContextLines*2+maxLinesPerHunk {
					if lineInfo.matching {
						extraHunkMatches++
					}
					continue
				}

				if cur == nil {
					cur = &diff.Hunk{
						OrigStartLine: hunk.OrigStartLine + origLineOffset,
						NewStartLine:  hunk.NewStartLine + newLineOffset,
						Section:       hunk.Section,
					}
				}

				if !lineInfo.added {
					cur.OrigLines++
				}
				if !lineInfo.removed {
					cur.NewLines++
				}
				if len(origHighlights[i]) > 0 {
					curHighlights[len(curLines)] = origHighlights[i]
				}
				curLines = append(curLines, line)
			} else if cur != nil {
				addExtraHunkMatchesSection(cur, extraHunkMatches)
				cur.Body = bytes.Join(curLines, nil)
				if len(curHighlights) > 0 {
					newHighlights[len(results)] = MatchedHunk{MatchedLines: curHighlights}
					curHighlights = make(map[int]result.Ranges)
				}
				results = append(results, cur)
				cur = nil
				curLines = nil
				extraHunkMatches = 0
			}

			if !lineInfo.added {
				origLineOffset++
			}
			if !lineInfo.removed {
				newLineOffset++
			}
		}

		if cur != nil {
			addExtraHunkMatchesSection(cur, extraHunkMatches)
			cur.Body = bytes.Join(curLines, nil)
			if len(curHighlights) > 0 {
				newHighlights[len(results)] = MatchedHunk{MatchedLines: curHighlights}
			}
			results = append(results, cur)
		}
	}

	return results, newHighlights
}

type diffHunkLineInfo struct {
	added    bool // line starts with '+'
	removed  bool // line starts with '-'
	matching bool // line matches query (only computed for changed lines)
	context  bool // include because it's context for a matching changed line
}

func (info diffHunkLineInfo) changed() bool { return info.added || info.removed }

func computeDiffHunkInfo(lines [][]byte, lineHighlights map[int]result.Ranges, matchContextLines int) []diffHunkLineInfo {
	// Return context line numbers for a given line number.
	contextLines := func(line int) (start, end int) {
		start = line - matchContextLines
		if start < 0 {
			start = 0
		}
		end = line + matchContextLines
		if end > len(lines)-1 {
			end = len(lines) - 1
		}
		return
	}

	lineInfo := make([]diffHunkLineInfo, len(lines))
	for i, line := range lines {
		lineInfo[i].added, lineInfo[i].removed = diffHunkLineStatus(line)
		if lineInfo[i].changed() {
			_, ok := lineHighlights[i]
			lineInfo[i].matching = ok || len(lineHighlights) == 0
			if lineInfo[i].matching {
				// Mark context lines before/after matching lines.
				start, end := contextLines(i)
				for j := start; j <= end; j++ {
					if i == j {
						continue
					}
					lineInfo[j].context = true
				}
			}
		}
	}
	return lineInfo
}

func diffHunkLineStatus(line []byte) (added, removed bool) {
	if len(line) >= 1 {
		added = line[0] == '+'
		removed = line[0] == '-'
	}
	return
}

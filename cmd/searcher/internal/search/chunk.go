package search

import (
	"bytes"
	"sort"
	"unicode/utf8"

	"github.com/sourcegraph/sourcegraph/internal/searcher/protocol"
)

// chunkRanges groups a set of ranges into chunks of adjacent ranges.
//
// `interChunkLines` is the minimum number of lines allowed between chunks. If
// two chunks would have fewer than `interChunkLines` lines between them, they
// are instead merged into a single chunk. For example, calling `chunkRanges`
// with `interChunkLines == 0` means ranges on two adjacent lines would be
// returned as two separate chunks.
//
// This function guarantees that the chunks returned are ordered by line number,
// have no overlapping lines, and the line ranges covered are spaced apart by
// a minimum of `interChunkLines`. More precisely, for any return value `rangeChunks`:
// rangeChunks[i].cover.End.Line + interChunkLines < rangeChunks[i+1].cover.Start.Line
func chunkRanges(ranges []protocol.Range, interChunkLines int32) []rangeChunk {
	// Sort by range start
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Start.Offset < ranges[j].Start.Offset
	})

	// guestimate size to minimize allocations. This assumes ~2 matches per
	// chunk. Additionally, since allocations are doubled on realloc, this
	// should only realloc once for small ranges.
	chunks := make([]rangeChunk, 0, len(ranges)/2)
	for i, rr := range ranges {
		if i == 0 {
			// First iteration, there are no chunks, so create a new one
			chunks = append(chunks, rangeChunk{
				cover:  rr,
				ranges: ranges[:1],
			})
			continue
		}

		lastChunk := &chunks[len(chunks)-1] // pointer for mutability
		if lastChunk.cover.End.Line+interChunkLines >= rr.Start.Line {
			// The current range overlaps with the current chunk, so merge them
			lastChunk.ranges = ranges[i-len(lastChunk.ranges) : i+1]

			// Expand the chunk coverRange if needed
			if rr.End.Offset > lastChunk.cover.End.Offset {
				lastChunk.cover.End = rr.End
			}
		} else {
			// No overlap, so create a new chunk
			chunks = append(chunks, rangeChunk{
				cover:  rr,
				ranges: ranges[i : i+1],
			})
		}
	}
	return chunks
}

func chunksToMatches(buf []byte, chunks []rangeChunk, contextLines int32) []protocol.ChunkMatch {
	chunkMatches := make([]protocol.ChunkMatch, 0, len(chunks))
	for _, chunk := range chunks {
		extendedRange := extendRangeToLines(chunk.cover, buf)
		rangeWithContext := addContextLines(extendedRange, buf, contextLines)
		chunkMatches = append(chunkMatches, protocol.ChunkMatch{
			// NOTE: we must copy the content here because the reference
			// must not outlive the backing mmap, which may be cleaned
			// up before the match is serialized for the network.
			Content:      string(bytes.ToValidUTF8(buf[rangeWithContext.Start.Offset:rangeWithContext.End.Offset], []byte("�"))),
			ContentStart: rangeWithContext.Start,
			Ranges:       chunk.ranges,
		})
	}
	return chunkMatches
}

// extendRangeWithContext adds contextLines worth of context to the range.
func extendRangeToLines(inputRange protocol.Range, buf []byte) protocol.Range {
	firstLineStart := lineStart(buf, inputRange.Start.Offset)
	lastLineStart := lineStart(buf, inputRange.End.Offset)
	lastLineEnd := lineEnd(buf,
		// We want the end of the line containing the last byte of the
		// match, not the first byte after the match. In the case of a
		// zero-width match between lines, prefer the line after rather
		// than the line before (like we do for lineStart).
		max(inputRange.End.Offset, max(inputRange.End.Offset, 1)-1 /* prevent underflow */),
	)

	return protocol.Range{
		Start: protocol.Location{
			Offset: firstLineStart,
			Line:   inputRange.Start.Line,
			Column: 0,
		},
		End: protocol.Location{
			Offset: lastLineEnd,
			Line:   inputRange.End.Line,
			Column: int32(utf8.RuneCount(buf[lastLineStart:lastLineEnd])),
		},
	}
}

func addContextLines(inputRange protocol.Range, buf []byte, contextLines int32) protocol.Range {
	if contextLines == 0 {
		return inputRange
	}
	firstLineStart := inputRange.Start.Offset
	lastLineEnd := inputRange.End.Offset

	precedingLinesAdded := 0
	succeedingLinesAdded := 0

	for i := int32(0); i < contextLines; i++ {
		if firstLineStart > 0 {
			firstLineStart = lineStart(buf, firstLineStart-1)
			precedingLinesAdded += 1
		}

		if int(lastLineEnd) < len(buf) {
			lastLineEnd = lineEnd(buf, lastLineEnd)
			succeedingLinesAdded += 1
		}
	}

	lastLineStart := lineStart(buf, lastLineEnd)

	return protocol.Range{
		Start: protocol.Location{
			Offset: firstLineStart,
			Line:   inputRange.Start.Line - int32(precedingLinesAdded),
			Column: 0,
		},
		End: protocol.Location{
			Offset: lastLineEnd,
			Line:   inputRange.End.Line + int32(succeedingLinesAdded),
			Column: int32(utf8.RuneCount(buf[lastLineStart:lastLineEnd])),
		},
	}
}

func lineStart(buf []byte, offset int32) int32 {
	start := int32(0)
	if loc := bytes.LastIndexByte(buf[:offset], '\n'); loc >= 0 {
		start = int32(loc) + 1
	}
	return start
}

func lineEnd(buf []byte, offset int32) int32 {
	end := int32(len(buf))
	if loc := bytes.IndexByte(buf[offset:], '\n'); loc >= 0 {
		end = int32(loc) + offset + 1
	}
	return end
}

// columnHelper is a helper struct which caches the number of runes last
// counted. If we naively use utf8.RuneCount for each match on a line, this
// leads to an O(nm) algorithm where m is the number of matches and n is the
// length of the line. Since the matches are sorted by increasing offset, we
// can avoid searching through the part of the line already processed, which
// makes this operation O(n) instead.
type columnHelper struct {
	data []byte

	// 0 values for all these are valid values
	lastLineOffset int
	lastOffset     int
	lastRuneCount  int
}

// get returns the column for the match. 'lineOffset' is the byte offset for the
// start of the line in the data buffer, and 'offset' is the byte offset of the
// rune in data.
func (c *columnHelper) get(lineOffset int, offset int) int {
	var runeCount int

	if lineOffset == c.lastLineOffset && offset >= c.lastOffset {
		// Can count from last calculation
		runeCount = c.lastRuneCount + utf8.RuneCount(c.data[c.lastOffset:offset])
	} else {
		// Need to count from the beginning of line
		runeCount = utf8.RuneCount(c.data[lineOffset:offset])
	}

	c.lastLineOffset = lineOffset
	c.lastOffset = offset
	c.lastRuneCount = runeCount

	return runeCount
}

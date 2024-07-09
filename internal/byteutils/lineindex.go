package byteutils

import (
	"bytes"
	"math"
	"strings"
)

// NewLineIndex creates a new LineIndex from some file content.
func NewLineIndex[T ~string | ~[]byte](content T) LineIndex {
	if len(content) > math.MaxUint32 {
		panic("content too large")
	}

	// PERF: count the newlines in advance to allocate the index slice exactly
	// Explicitly case on the type rather than casting because the generics
	// seem to break the optimization that allows the allocation to be elided.
	var newlineCount int
	switch v := any(content).(type) {
	case string:
		newlineCount = strings.Count(v, "\n")
	case []byte:
		newlineCount = bytes.Count(v, []byte("\n"))
	}

	index := make(LineIndex, 0, newlineCount+2)
	index = append(index, 0)
	offset := 0
	for {
		var loc int
		switch v := any(content).(type) {
		case string:
			loc = strings.IndexByte(v[offset:], '\n')
		case []byte:
			loc = bytes.IndexByte(v[offset:], '\n')
		}
		if loc == -1 {
			break
		}
		index = append(index, uint32(offset+loc+1))
		offset += loc + 1
	}
	index = append(index, uint32(len(content)))
	return index
}

// LineIndex contains the line boundaries of the indexed content.
// Its structure is:
// - A leading 0
// - A sorted list of every byte offset _after_ a newline byte
// - A trailing len(content)
//
// This means:
// - LineIndex[N] is the offset of the first byte of line N
// - LineIndex[N+1] is the offset of the first byte after line N
// - content[LineIndex[N]:LineIndex[N+1]] is the contents of line N
type LineIndex []uint32

// LineRange returns a range that can be used to slice the indexed content to obtain
// the line for the given number. The range is guaranteed to be a valid slice
// into the content if the content is unchanged. If the line number refers to a
// line that does not exist, a zero-length range will be returned pointing to
// the beginning (for underflow) or end (for overflow) of the file.
//
// lineNumber is 0-indexed, and the returned range includes the terminating
// newline (if it exists). Equivalent to Lines(lineNumber, lineNumber + 1).
func (l LineIndex) LineRange(lineNumber int) (int, int) {
	return l.LinesRange(lineNumber, lineNumber+1)
}

// LinesRange returns a range that can be used to slice the indexed content to
// obtain the lines for the given half-open range. The range is guaranteed to
// be a valid slice into the content if the content is unchanged. If the
// requested range of lines does not exist, it will be truncated to return the
// set of lines in that range that does exist.
//
// line numbers are 0-indexed, and the returned range includes the terminating
// newline (if it exists).
func (l LineIndex) LinesRange(startLine, endLine int) (int, int) {
	startLine = min(max(0, startLine), len(l)-1)
	endLine = min(max(startLine, endLine), len(l)-1)
	return int(l[startLine]), int(l[endLine])
}

// For the purpose of this package, a line is defined as:
// - zero or more non-newline bytes terminated by a newline byte
// - OR one more non-newline terminated by the end of the file.
//
// Equivalently, the regex `[^\n]*\n|[^\n]+$`
//
// Equivalently, a newline at the last byte of the file does not
// start an empty last line.
//
// Notably, this is at odds with the POSIX standard:
// https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap03.html#tag_03_206
func (l LineIndex) LineCount() int {
	lastLineEnd := l[len(l)-1]
	contentEnd := l[len(l)-2]
	if lastLineEnd == contentEnd {
		return len(l) - 2
	}
	return len(l) - 1
}

// NewlineCount is simply the number of newline bytes in the content
func (l LineIndex) NewlineCount() int {
	return len(l) - 2
}

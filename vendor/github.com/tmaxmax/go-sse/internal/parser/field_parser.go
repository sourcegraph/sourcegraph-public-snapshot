package parser

import (
	"errors"
	"strings"
)

// FieldParser extracts fields from a byte slice.
type FieldParser struct {
	err  error
	data string

	started bool

	keepComments bool
	removeBOM    bool
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func trimFirstSpace(c string) string {
	if c != "" && c[0] == ' ' {
		return c[1:]
	}
	return c
}

func (f *FieldParser) scanSegment(chunk string, out *Field) bool {
	colonPos, l := strings.IndexByte(chunk, ':'), len(chunk)
	if colonPos > maxFieldNameLength {
		return false
	}
	if colonPos == -1 {
		colonPos = l
	}

	name, ok := getFieldName(chunk[:colonPos])
	if ok {
		out.Name = name
		out.Value = trimFirstSpace(chunk[min(colonPos+1, l):])
		return true
	} else if chunk == "" {
		// scanSegment is called only with chunks which end with a newline in the input.
		// If chunk is empty, it means that this is a blank line which ends the event,
		// so an empty Field needs to be returned.
		out.Name = ""
		out.Value = ""
		return true
	} else if colonPos == 0 && f.keepComments {
		out.Name = FieldNameComment
		out.Value = trimFirstSpace(chunk[min(1, l):])
		return true
	}

	return false
}

// ErrUnexpectedEOF is returned when the input is completely parsed but no complete field was found at the end.
var ErrUnexpectedEOF = errors.New("go-sse: unexpected end of input")

// Next parses the next available field in the remaining buffer.
// It returns false if there are no more fields to parse.
func (f *FieldParser) Next(r *Field) bool {
	for f.data != "" {
		f.started = true

		chunk, rem, hasNewline := NextChunk(f.data)
		if !hasNewline {
			f.err = ErrUnexpectedEOF
			return false
		}

		f.data = rem

		if !f.scanSegment(chunk, r) {
			continue
		}

		return true
	}

	return false
}

// Reset changes the buffer from which fields are parsed.
func (f *FieldParser) Reset(data string) {
	f.data = data
	f.err = nil
	f.started = false
	f.doRemoveBOM()
}

// Err returns the last error encountered by the parser. It is either nil or ErrUnexpectedEOF.
func (f *FieldParser) Err() error {
	return f.err
}

// Started tells whether parsing has started (a call to Next which consumed input was made
// or the BOM was removed, if it existed). Started will be true if the FieldParser has advanced
// through the data.
func (f *FieldParser) Started() bool {
	return f.started
}

// KeepComments configures the FieldParser to parse/ignore comment fields.
// By default comment fields are ignored.
func (f *FieldParser) KeepComments(shouldKeep bool) {
	f.keepComments = shouldKeep
}

// RemoveBOM configures the FieldParser to try and remove the Unicode BOM
// when parsing the first field, if it exists.
// If, at the time this option is set, the input is untouched (no fields were parsed),
// it will also be attempted to remove the BOM.
func (f *FieldParser) RemoveBOM(shouldRemove bool) {
	f.removeBOM = shouldRemove
	f.doRemoveBOM()
}

func (f *FieldParser) doRemoveBOM() {
	const bom = "\xEF\xBB\xBF"
	if f.removeBOM && !f.started && strings.HasPrefix(f.data, bom) {
		f.data = f.data[len(bom):]
		f.started = true
	}
}

// NewFieldParser creates a parser that extracts fields from the given string.
func NewFieldParser(data string) *FieldParser {
	return &FieldParser{data: data}
}

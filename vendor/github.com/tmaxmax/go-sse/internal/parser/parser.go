package parser

import (
	"bufio"
	"io"
	"unsafe"
)

// splitFunc is a split function for a bufio.Scanner that splits a sequence of
// bytes into SSE events. Each event ends with two consecutive newline sequences,
// where a newline sequence is defined as either "\n", "\r", or "\r\n".
func splitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if len(data) == 0 {
		return 0, nil, nil
	}

	var start int
	for {
		index, endlineLen := NewlineIndex((*(*string)(unsafe.Pointer(&data)))[advance:])
		advance += index + endlineLen
		if index == 0 {
			// If it was a blank line, skip it.
			start += endlineLen
		}
		// We've reached the end of data or a second newline follows and the line isn't blank.
		// The latter means we have an event.
		if advance == len(data) || (isNewlineChar(data[advance]) && index > 0) {
			break
		}
	}

	if l := len(data); advance == l && !atEOF {
		// We have reached the end of the buffer but have not yet seen two consecutive
		// newline sequences, so we request more data.
		return 0, nil, nil
	} else if advance < l {
		// We have found a newline. Consume the end-of-line sequence.
		advance++
		// Consume one more character if end-of-line is "\r\n".
		if advance < l && data[advance-1] == '\r' && data[advance] == '\n' {
			advance++
		}
	}

	token = data[start:advance]

	return advance, token, nil
}

// Parser extracts fields from a reader. Reading is buffered using a bufio.Scanner.
// The Parser also removes the UTF-8 BOM if it exists.
type Parser struct {
	inputScanner *bufio.Scanner
	fieldScanner *FieldParser
}

// Next parses a single field from the reader. It returns false when there are no more fields to parse.
func (r *Parser) Next(f *Field) bool {
	if !r.fieldScanner.Next(f) {
		if !r.inputScanner.Scan() {
			// Do this to signal EOF, which bufio.Scanner suppresses.
			if r.inputScanner.Err() == nil {
				r.inputScanner = nil
			}
			return false
		}

		if r.fieldScanner.Started() {
			// If scanning was started, then an event was already processed at this point and the BOM was
			// already removed, if it existed. We don't need to remove it anymore, so disable the option.
			r.fieldScanner.RemoveBOM(false)
		}

		// The allocation made inside `Text` is not an issue and should even improve performance.
		// If the Field returned from `Next` wouldn't own its resources, then the caller would have
		// to allocate new memory and copy each field value. This way, not only the caller doesn't
		// have to worry about allocations and ownership, but also bigger and less frequent allocations
		// are made, compared to the previous usage – allocations are now made per event, not per field value.
		r.fieldScanner.Reset(r.inputScanner.Text())

		return r.fieldScanner.Next(f)
	}

	return true
}

// Err returns the last read error. At the end of input
// it will always be equal to io.EOF.
func (r *Parser) Err() error {
	if err := r.fieldScanner.Err(); err != nil {
		return err
	}
	if r.inputScanner == nil {
		// Recover the EOF suppressed by bufio.Scanner.
		// We need it inside the client, to know when to retry.
		return io.EOF
	}
	return r.inputScanner.Err()
}

// Buffer sets the buffer used to scan the input.
// For more information, see the documentation on bufio.Scanner.Buffer.
// Do not call this after parsing has started – the method will panic!
func (r *Parser) Buffer(buf []byte, max int) {
	r.inputScanner.Buffer(buf, max)
}

// New returns a Parser that extracts fields from a reader.
func New(r io.Reader) *Parser {
	sc := bufio.NewScanner(r)
	sc.Split(splitFunc)

	fsc := NewFieldParser("")
	fsc.RemoveBOM(true)

	return &Parser{inputScanner: sc, fieldScanner: fsc}
}

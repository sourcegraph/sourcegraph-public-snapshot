package byteutils

import "bytes"

// NewLineReader creates a new lineReader instance that reads lines from data.
// It is more memory effective than bytes.Split, because it does not require 24 bytes
// for each subslice it generates, and instead returns one subslice at a time.
// Benchmarks prove it is faster _and_ more memory efficient than bytes.Split, see
// the test file for details.
// Note: This behaves slightly differently to bytes.Split!
// For an empty input, it does NOT read a single line, like bytes.Split would.
// Also, it does NOT return a final empty line if the input is terminated with
// a final newline.
//
// data is the byte slice to read lines from.
//
// A lineReader can be used to iterate over lines in a byte slice.
//
// For example:
//
// data := []byte("hello\nworld\n")
// reader := bytes.NewLineReader(data)
//
//	for reader.Scan() {
//	    line := reader.Line()
//	    // Use line...
//	}
func NewLineReader(data []byte) lineReader {
	return lineReader{data: data}
}

// lineReader is a struct that can be used to iterate over lines in a byte slice.
type lineReader struct {
	data    []byte
	current []byte
}

// Scan advances the lineReader to the next line and returns true, or returns false if there are no more lines.
// The lineReader's current field will be updated to contain the next line.
// Scan must be called before calling Line.
func (r *lineReader) Scan() bool {
	// If we are at the end of the data, stop
	if len(r.data) == 0 {
		return false
	}
	// Find the next newline
	i := bytes.IndexByte(r.data, '\n')
	if i >= 0 {
		// Exclude the newline from the line
		r.current = r.data[:i]
		// Advance past the newline
		r.data = r.data[i+1:]
		return true
	}
	// Otherwise include the last byte
	r.current = r.data
	r.data = nil
	return true
}

// Line returns the current line.
// The line is valid until the next call to Scan.
// Scan must be called before calling Line.
func (r *lineReader) Line() []byte {
	return r.current
}

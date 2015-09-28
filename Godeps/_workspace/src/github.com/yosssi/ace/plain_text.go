package ace

import (
	"bytes"
	"io"
	"strings"
)

// plainText represents a plain text.
type plainText struct {
	elementBase
	insertBr bool
}

// WriteTo writes data to w.
func (e *plainText) WriteTo(w io.Writer) (int64, error) {
	var bf bytes.Buffer

	// Write the plain text.
	bf.WriteString(strings.Join(e.ln.tokens[1:], space))

	if len(e.ln.tokens) > 1 && e.insertBr {
		bf.WriteString(htmlBr)
	}

	// Write the children's HTML.
	if len(e.children) > 0 {
		bf.WriteString(lf)

		if i, err := e.writeChildren(&bf); err != nil {
			return i, err
		}
	}

	// Write the buffer.
	i, err := w.Write(bf.Bytes())

	return int64(i), err
}

// ContainPlainText returns true.
func (e *plainText) ContainPlainText() bool {
	return true
}

// InsertBr returns true if the br tag is inserted to the line.
func (e *plainText) InsertBr() bool {
	return e.insertBr
}

// newPlainText creates and returns a plain text.
func newPlainText(ln *line, rslt *result, src *source, parent element, opts *Options) *plainText {
	return &plainText{
		elementBase: newElementBase(ln, rslt, src, parent, opts),
		insertBr:    ln.tokens[0] == doublePipe,
	}
}

package ace

import "io"

// HTML
const (
	htmlBr = "<br>"
)

// plainTextInner represents a plain text inner.
type plainTextInner struct {
	elementBase
	insertBr bool
}

// WriteTo writes data to w.
func (e *plainTextInner) WriteTo(w io.Writer) (int64, error) {
	s := ""

	if (e.parent.Base().ln.indent+1)*2 <= len(e.ln.str) {
		s = e.ln.str[(e.parent.Base().ln.indent+1)*2:]
	}

	if e.insertBr && !e.lastChild {
		s += htmlBr
	}

	i, err := w.Write([]byte(s + lf))

	return int64(i), err
}

// CanHaveChildren returns false.
func (e *plainTextInner) CanHaveChildren() bool {
	return false
}

// newPlainTextInner creates and returns a plain text.
func newPlainTextInner(ln *line, rslt *result, src *source, parent element, insertBr bool, opts *Options) *plainTextInner {
	return &plainTextInner{
		elementBase: newElementBase(ln, rslt, src, parent, opts),
		insertBr:    insertBr,
	}
}

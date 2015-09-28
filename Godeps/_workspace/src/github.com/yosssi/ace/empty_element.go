package ace

import "io"

// emptyElement represents an empty element.
type emptyElement struct {
	elementBase
}

// Do nothing.
func (e *emptyElement) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

// CanHaveChildren returns false.
func (e *emptyElement) CanHaveChildren() bool {
	return false
}

// newEmpty creates and returns an empty element.
func newEmptyElement(ln *line, rslt *result, src *source, parent element, opts *Options) *emptyElement {
	return &emptyElement{
		elementBase: newElementBase(ln, rslt, src, parent, opts),
	}
}

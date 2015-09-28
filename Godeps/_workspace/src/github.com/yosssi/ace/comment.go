package ace

import "io"

// comment represents a comment.
type comment struct {
	elementBase
}

// Do nothing.
func (e *comment) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

// ContainPlainText returns true.
func (e *comment) ContainPlainText() bool {
	return true
}

// newComment creates and returns a comment.
func newComment(ln *line, rslt *result, src *source, parent element, opts *Options) *comment {
	return &comment{
		elementBase: newElementBase(ln, rslt, src, parent, opts),
	}
}

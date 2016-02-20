package ace

import (
	"bytes"
	"fmt"
	"io"
)

// helperMethodContent represents a helper method content.
type helperMethodContent struct {
	elementBase
	name string
}

// WriteTo writes data to w.
func (e *helperMethodContent) WriteTo(w io.Writer) (int64, error) {
	var bf bytes.Buffer

	inner := e.src.inner
	if inner == nil {
		return 0, fmt.Errorf("inner is not specified [file: %s][line: %d]", e.ln.fileName(), e.ln.no)
	}

	// Write a define action.
	bf.WriteString(fmt.Sprintf(actionDefine, e.opts.DelimLeft, inner.path+doubleColon+e.name, e.opts.DelimRight))

	// Write the children's HTML.
	if i, err := e.writeChildren(&bf); err != nil {
		return i, err
	}

	// Write an end action.
	bf.WriteString(fmt.Sprintf(actionEnd, e.opts.DelimLeft, e.opts.DelimRight))

	// Write the buffer.
	i, err := w.Write(bf.Bytes())

	return int64(i), err
}

// newHelperMethodContent creates and returns a helper method content.
func newHelperMethodContent(ln *line, rslt *result, src *source, parent element, opts *Options) (*helperMethodContent, error) {
	if len(ln.tokens) < 3 || ln.tokens[2] == "" {
		return nil, fmt.Errorf("no name is specified [file: %s][line: %d]", ln.fileName(), ln.no)
	}

	e := &helperMethodContent{
		elementBase: newElementBase(ln, rslt, src, parent, opts),
		name:        ln.tokens[2],
	}

	return e, nil
}

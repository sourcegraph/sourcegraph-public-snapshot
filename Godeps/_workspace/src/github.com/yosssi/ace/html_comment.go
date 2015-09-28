package ace

import (
	"bytes"
	"io"
	"strings"
)

// htmlComment represents an HTML comment.
type htmlComment struct {
	elementBase
}

// WriteTo writes data to w.
func (e *htmlComment) WriteTo(w io.Writer) (int64, error) {
	var bf bytes.Buffer

	// Write an open tag.
	bf.WriteString(e.opts.DelimLeft)
	bf.WriteString(preDefinedFuncNameHTML)
	bf.WriteString(space)
	bf.WriteString(doubleQuote)
	bf.WriteString(lt)
	bf.WriteString(exclamation)
	bf.WriteString(doubleQuote)
	bf.WriteString(e.opts.DelimRight)
	bf.WriteString(hyphen)
	bf.WriteString(hyphen)

	// Write the HTML comment
	if len(e.ln.tokens) > 1 {
		bf.WriteString(space)
		bf.WriteString(strings.Join(e.ln.tokens[1:], space))
	}

	// Write the children's HTML.
	if len(e.children) > 0 {
		bf.WriteString(lf)

		if i, err := e.writeChildren(&bf); err != nil {
			return i, err
		}
	} else {
		bf.WriteString(space)

	}

	// Write a close tag.
	bf.WriteString(hyphen)
	bf.WriteString(hyphen)
	bf.WriteString(gt)

	// Write the buffer.
	i, err := w.Write(bf.Bytes())

	return int64(i), err
}

// ContainPlainText returns true.
func (e *htmlComment) ContainPlainText() bool {
	return true
}

// newHTMLComment creates and returns an HTML comment.
func newHTMLComment(ln *line, rslt *result, src *source, parent element, opts *Options) *htmlComment {
	return &htmlComment{
		elementBase: newElementBase(ln, rslt, src, parent, opts),
	}
}

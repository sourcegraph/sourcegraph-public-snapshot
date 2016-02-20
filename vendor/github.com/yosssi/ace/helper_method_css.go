package ace

import (
	"bytes"
	"io"
)

// helperMethodCSS represents a helper method css.
type helperMethodCSS struct {
	elementBase
}

// WriteTo writes data to w.
func (e *helperMethodCSS) WriteTo(w io.Writer) (int64, error) {
	var bf bytes.Buffer

	// Write an open tag.
	bf.WriteString(lt)
	bf.WriteString(`style type="text/css"`)
	bf.WriteString(gt)

	bf.WriteString(lf)

	// Write the children's HTML.
	if i, err := e.writeChildren(&bf); err != nil {
		return i, err
	}

	// Write an open tag.
	bf.WriteString(lt)
	bf.WriteString(slash)
	bf.WriteString("style")
	bf.WriteString(gt)

	// Write the buffer.
	i, err := w.Write(bf.Bytes())

	return int64(i), err
}

// ContainPlainText returns true.
func (e *helperMethodCSS) ContainPlainText() bool {
	return true
}

// helperMethodCSS creates and returns a helper method css.
func newHelperMethodCSS(ln *line, rslt *result, src *source, parent element, opts *Options) *helperMethodCSS {
	return &helperMethodCSS{
		elementBase: newElementBase(ln, rslt, src, parent, opts),
	}
}

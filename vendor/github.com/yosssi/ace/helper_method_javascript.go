package ace

import (
	"bytes"
	"io"
)

// helperMethodJavascript represents a helper method javascript.
type helperMethodJavascript struct {
	elementBase
}

// WriteTo writes data to w.
func (e *helperMethodJavascript) WriteTo(w io.Writer) (int64, error) {
	var bf bytes.Buffer

	// Write an open tag.
	bf.WriteString(lt)
	bf.WriteString(`script type="text/javascript"`)
	bf.WriteString(gt)

	bf.WriteString(lf)

	// Write the children's HTML.
	if i, err := e.writeChildren(&bf); err != nil {
		return i, err
	}

	// Write an open tag.
	bf.WriteString(lt)
	bf.WriteString(slash)
	bf.WriteString("script")
	bf.WriteString(gt)

	// Write the buffer.
	i, err := w.Write(bf.Bytes())

	return int64(i), err
}

// ContainPlainText returns true.
func (e *helperMethodJavascript) ContainPlainText() bool {
	return true
}

// helperMethodJavascript creates and returns a helper method javascript.
func newHelperMethodJavascript(ln *line, rslt *result, src *source, parent element, opts *Options) *helperMethodJavascript {
	return &helperMethodJavascript{
		elementBase: newElementBase(ln, rslt, src, parent, opts),
	}
}

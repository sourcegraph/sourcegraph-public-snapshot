package ace

import (
	"bytes"
	"fmt"
	"io"
)

// helperMethodYield represents a helper method yield.
type helperMethodYield struct {
	elementBase
	templateName string
}

// WriteTo writes data to w.
func (e *helperMethodYield) WriteTo(w io.Writer) (int64, error) {
	var bf bytes.Buffer

	inner := e.src.inner
	if inner == nil {
		return 0, fmt.Errorf("inner is not specified [file: %s][line: %d]", e.ln.fileName(), e.ln.no)
	}

	var templateExists bool

	for _, innerE := range e.rslt.inner {
		ln := innerE.Base().ln
		if ln.isHelperMethodOf(helperMethodNameContent) && len(ln.tokens) > 2 && ln.tokens[2] == e.templateName {
			templateExists = true
			break
		}
	}

	if templateExists {
		bf.WriteString(fmt.Sprintf(actionTemplateWithPipeline, e.opts.DelimLeft, inner.path+doubleColon+e.templateName, dot, e.opts.DelimRight))
	} else {
		// Write the children's HTML.
		if i, err := e.writeChildren(&bf); err != nil {
			return i, err
		}
	}

	i, err := w.Write(bf.Bytes())

	return int64(i), err
}

// newHelperMethodYield creates and returns a helper method yield.
func newHelperMethodYield(ln *line, rslt *result, src *source, parent element, opts *Options) (*helperMethodYield, error) {
	if len(ln.tokens) < 3 {
		return nil, fmt.Errorf("no template name is specified [file: %s][line: %d]", ln.fileName(), ln.no)
	}

	e := &helperMethodYield{
		elementBase:  newElementBase(ln, rslt, src, parent, opts),
		templateName: ln.tokens[2],
	}

	return e, nil
}

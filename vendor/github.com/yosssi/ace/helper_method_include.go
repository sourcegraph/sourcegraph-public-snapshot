package ace

import (
	"fmt"
	"io"
	"strings"
)

// helperMethodInclude represents a helper method include.
type helperMethodInclude struct {
	elementBase
	templateName string
	pipeline     string
}

// WriteTo writes data to w.
func (e *helperMethodInclude) WriteTo(w io.Writer) (int64, error) {
	var s string

	if e.pipeline == "" {
		s = fmt.Sprintf(actionTemplate, e.opts.DelimLeft, e.templateName, e.opts.DelimRight)
	} else {
		s = fmt.Sprintf(actionTemplateWithPipeline, e.opts.DelimLeft, e.templateName, e.pipeline, e.opts.DelimRight)
	}

	i, err := w.Write([]byte(s))

	return int64(i), err
}

// newHelperMethodInclude creates and returns a helper method include.
func newHelperMethodInclude(ln *line, rslt *result, src *source, parent element, opts *Options) (*helperMethodInclude, error) {
	if len(ln.tokens) < 3 {
		return nil, fmt.Errorf("no template name is specified [file: %s][line: %d]", ln.fileName(), ln.no)
	}

	var pipeline string

	if len(ln.tokens) > 3 {
		pipeline = strings.Join(ln.tokens[3:], space)
	}

	e := &helperMethodInclude{
		elementBase:  newElementBase(ln, rslt, src, parent, opts),
		templateName: ln.tokens[2],
		pipeline:     pipeline,
	}

	return e, nil
}

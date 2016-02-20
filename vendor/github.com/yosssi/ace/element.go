package ace

import (
	"fmt"
	"io"
)

// Helper method names
const (
	helperMethodNameConditionalComment = "conditionalComment"
	helperMethodNameContent            = "content"
	helperMethodNameCSS                = "css"
	helperMethodNameDoctype            = "doctype"
	helperMethodNameYield              = "yield"
	helperMethodNameInclude            = "include"
	helperMethodNameJavascript         = "javascript"
)

// element is an interface for storing an element.
type element interface {
	io.WriterTo
	AppendChild(child element)
	ContainPlainText() bool
	Base() *elementBase
	CanHaveChildren() bool
	InsertBr() bool
	SetLastChild(lastChild bool)
}

// newElement creates and returns an element.
func newElement(ln *line, rslt *result, src *source, parent element, opts *Options) (element, error) {
	var e element
	var err error

	switch {
	case parent != nil && parent.ContainPlainText():
		e = newPlainTextInner(ln, rslt, src, parent, parent.InsertBr(), opts)
	case ln.isEmpty():
		e = newEmptyElement(ln, rslt, src, parent, opts)
	case ln.isComment():
		e = newComment(ln, rslt, src, parent, opts)
	case ln.isHTMLComment():
		e = newHTMLComment(ln, rslt, src, parent, opts)
	case ln.isHelperMethod():
		switch {
		case ln.isHelperMethodOf(helperMethodNameConditionalComment):
			e, err = newHelperMethodConditionalComment(ln, rslt, src, parent, opts)
		case ln.isHelperMethodOf(helperMethodNameContent):
			e, err = newHelperMethodContent(ln, rslt, src, parent, opts)
		case ln.isHelperMethodOf(helperMethodNameCSS):
			e = newHelperMethodCSS(ln, rslt, src, parent, opts)
		case ln.isHelperMethodOf(helperMethodNameDoctype):
			e, err = newHelperMethodDoctype(ln, rslt, src, parent, opts)
		case ln.isHelperMethodOf(helperMethodNameInclude):
			e, err = newHelperMethodInclude(ln, rslt, src, parent, opts)
		case ln.isHelperMethodOf(helperMethodNameJavascript):
			e = newHelperMethodJavascript(ln, rslt, src, parent, opts)
		case ln.isHelperMethodOf(helperMethodNameYield):
			e, err = newHelperMethodYield(ln, rslt, src, parent, opts)
		default:
			err = fmt.Errorf("the helper method name is invalid [file: %s][line: %d]", ln.fileName(), ln.no)
		}
	case ln.isPlainText():
		e = newPlainText(ln, rslt, src, parent, opts)
	case ln.isAction():
		e = newAction(ln, rslt, src, parent, opts)
	default:
		e, err = newHTMLTag(ln, rslt, src, parent, opts)
	}

	return e, err
}

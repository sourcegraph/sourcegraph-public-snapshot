package ace

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Comment types
const (
	commentTypeHidden   = "hidden"
	commentTypeRevealed = "revealed"
)

// helperMethodConditionalComment represents a helper method
// conditional comment.
type helperMethodConditionalComment struct {
	elementBase
	commentType string
	condition   string
}

// WriteTo writes data to w.
func (e *helperMethodConditionalComment) WriteTo(w io.Writer) (int64, error) {
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
	if e.commentType == commentTypeHidden {
		bf.WriteString(hyphen)
		bf.WriteString(hyphen)
	}
	bf.WriteString(bracketOpen)
	bf.WriteString("if ")
	bf.WriteString(e.condition)
	bf.WriteString(bracketClose)
	bf.WriteString(gt)

	bf.WriteString(lf)

	// Write the children's HTML.
	if i, err := e.writeChildren(&bf); err != nil {
		return i, err
	}

	// Write a close tag.
	bf.WriteString(e.opts.DelimLeft)
	bf.WriteString(preDefinedFuncNameHTML)
	bf.WriteString(space)
	bf.WriteString(doubleQuote)
	bf.WriteString(lt)
	bf.WriteString(exclamation)
	bf.WriteString(doubleQuote)
	bf.WriteString(e.opts.DelimRight)
	bf.WriteString(bracketOpen)
	bf.WriteString("endif")
	bf.WriteString(bracketClose)
	if e.commentType == commentTypeHidden {
		bf.WriteString(hyphen)
		bf.WriteString(hyphen)
	}
	bf.WriteString(gt)

	// Write the buffer.
	i, err := w.Write(bf.Bytes())

	return int64(i), err
}

// ContainPlainText returns true.
func (e *helperMethodConditionalComment) ContainPlainText() bool {
	return true
}

// newHelperMethodConditionalComment creates and returns an HTML comment.
func newHelperMethodConditionalComment(ln *line, rslt *result, src *source, parent element, opts *Options) (*helperMethodConditionalComment, error) {
	switch len(ln.tokens) {
	case 2:
		return nil, fmt.Errorf("no comment type is specified [file: %s][line: %d]", ln.fileName(), ln.no)
	case 3:
		return nil, fmt.Errorf("no condition is specified [file: %s][line: %d]", ln.fileName(), ln.no)
	}

	commentType := ln.tokens[2]

	if commentType != commentTypeHidden && commentType != commentTypeRevealed {
		return nil, fmt.Errorf("the comment type is invalid [file: %s][line: %d]", ln.fileName(), ln.no)
	}

	e := &helperMethodConditionalComment{
		elementBase: newElementBase(ln, rslt, src, parent, opts),
		commentType: commentType,
		condition:   strings.Join(ln.tokens[3:], space),
	}

	return e, nil
}

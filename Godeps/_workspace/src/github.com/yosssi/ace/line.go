package ace

import (
	"fmt"
	"strings"
)

const unicodeSpace = 32

const indentTop = 0

// line represents a line of codes.
type line struct {
	no     int
	str    string
	indent int
	tokens []string
	opts   *Options
	file   *File
}

// isEmpty returns true if the line is empty.
func (l *line) isEmpty() bool {
	return strings.TrimSpace(l.str) == ""
}

// isTopIndent returns true if the line's indent is the top level.
func (l *line) isTopIndent() bool {
	return l.indent == indentTop
}

// isHelperMethod returns true if the line is a helper method.
func (l *line) isHelperMethod() bool {
	return len(l.tokens) > 1 && l.tokens[0] == equal
}

// isHelperMethodOf returns true if the line is a specified helper method.
func (l *line) isHelperMethodOf(name string) bool {
	return l.isHelperMethod() && l.tokens[1] == name
}

// isPlainText returns true if the line is a plain text.
func (l *line) isPlainText() bool {
	return len(l.tokens) > 0 && (l.tokens[0] == pipe || l.tokens[0] == doublePipe)
}

// isComment returns true if the line is a comment.
func (l *line) isComment() bool {
	return len(l.tokens) > 0 && l.tokens[0] == slash
}

// isHTMLComment returns true if the line is an HTML comment.
func (l *line) isHTMLComment() bool {
	return len(l.tokens) > 0 && l.tokens[0] == slash+slash
}

// isAction returns true if the line is an action.
func (l *line) isAction() bool {
	str := strings.TrimSpace(l.str)
	return strings.HasPrefix(str, l.opts.DelimLeft) && strings.HasSuffix(str, l.opts.DelimRight)
}

// fileName returns the file name.
func (l *line) fileName() string {
	return l.file.path + dot + l.opts.Extension
}

// childOf returns true is the line is a child of the element.
func (l *line) childOf(parent element) (bool, error) {
	var ok bool
	var err error

	switch {
	case l.isEmpty():
		ok = true
	case parent.ContainPlainText():
		switch {
		case parent.Base().ln.indent < l.indent:
			ok = true
		}
	default:
		switch {
		case l.indent == parent.Base().ln.indent+1:
			ok = true
		case l.indent > parent.Base().ln.indent+1:
			err = fmt.Errorf("the indent is invalid [file: %s][line: %d]", l.fileName(), l.no)
		}
	}

	return ok, err
}

// newLine creates and returns a line.
func newLine(no int, str string, opts *Options, f *File) *line {
	return &line{
		no:     no,
		str:    str,
		indent: indent(str),
		tokens: strings.Split(strings.TrimLeft(str, space), space),
		opts:   opts,
		file:   f,
	}
}

// indent returns the line's indent.
func indent(str string) int {
	var i int

	for _, b := range str {
		if b != unicodeSpace {
			break
		}
		i++
	}

	return i / 2
}

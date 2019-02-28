package types

import (
	"regexp/syntax"
	"strings"
)

func escapeNonTerminalEOL(r *syntax.Regexp) {
	if r.Op == syntax.OpConcat && len(r.Sub) > 0 {
		for i, child := range r.Sub[:len(r.Sub)-1] {
			if child.Op == syntax.OpEndLine || child.Op == syntax.OpEndText {
				r.Sub[i] = &syntax.Regexp{
					Op:   syntax.OpLiteral,
					Rune: []rune{'$'},
				}
			}
		}
	}
	for _, child := range r.Sub {
		escapeNonTerminalEOL(child)
	}
}

var escapeErrorMessages = []string{
	"missing argument to repetition operator: `",
	"missing closing ",
}

func fixupCompileErrors(value string, err error) (string, error) {
	msg := err.Error()
	index := -1

	for _, errorMsg := range escapeErrorMessages {
		index = strings.Index(msg, errorMsg)
		if index > -1 {
			index = len(errorMsg) + index
			break
		}
	}

	if index == -1 {
		return value, err
	}

	runeToEscape := flipRune(rune(msg[index]))

	out := ""
	// Loop through and escape all runeToEscape
	for _, r := range value {
		if r == runeToEscape {
			out += string('\\') + string(r)
		} else {
			out += string(r)
		}
	}

	return out, nil
}

// flipRune maps opening block characters (e.g. ), ]) to their opening
// counterparts. If the rune provided is not one of those, this func returns
// the identity of the rune.
func flipRune(r rune) rune {
	switch r {
	case ')':
		return '('
	case ']':
		return '['
	default:
		return r
	}

}

func parseRegexp(value string) (string, error) {
	var r *syntax.Regexp
	// If we can't fix the query up, we want to return the original error from the user entered query.
	var originalErr error
	var err error

	r, originalErr = syntax.Parse(value, syntax.Perl)
	if originalErr != nil {
		var s string

		s, err = fixupCompileErrors(value, originalErr)
		if err != nil {
			return value, originalErr
		}

		r, err = syntax.Parse(s, syntax.Perl)
	}

	if err != nil {
		return value, originalErr
	}

	escapeNonTerminalEOL(r)

	return r.String(), nil
}

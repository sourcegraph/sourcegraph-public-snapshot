package comby

import (
	"strings"
	"unicode/utf8"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

var MatchHoleRegexp = lazyregexp.New(splitOnHolesPattern())

func splitOnHolesPattern() string {
	word := `\w+`
	whitespaceAndOptionalWord := `[ ]+(` + word + `)?`
	holeAnything := `:\[` + word + `\]`
	holeEllipses := `\.\.\.`
	holeAlphanum := `:\[\[` + word + `\]\]`
	holeWithPunctuation := `:\[` + word + `\.\]`
	holeWithNewline := `:\[` + word + `\\n\]`
	holeWhitespace := `:\[` + whitespaceAndOptionalWord + `\]`
	return strings.Join([]string{
		holeAnything,
		holeEllipses,
		holeAlphanum,
		holeWithPunctuation,
		holeWithNewline,
		holeWhitespace,
	}, "|")
}

var matchRegexpPattern = lazyregexp.New(`:\[(\w+)?~(.*)\]`)

type Term interface {
	term()
	String() string
}

type Literal string
type Hole string

func (Literal) term() {}
func (t Literal) String() string {
	return string(t)
}

func (Hole) term() {}
func (t Hole) String() string {
	return string(t)
}

// parseTemplate parses a comby pattern to a list of Terms where a Term is
// either a literal or hole metasyntax.
func parseTemplate(buf []byte) []Term {
	// Track context of whether we are inside an opening hole, e.g., after
	// ':['. Value is greater than 1 when inside.
	var open int
	// Track whether we are balanced inside a regular expression character
	// set like '[a]' inside an open hole, e.g., :[foo~[a]]. Value is greater
	// than 1 when inside.
	var inside int

	var start int
	var r rune
	var token []rune
	var result []Term

	next := func() rune {
		r, start := utf8.DecodeRune(buf)
		buf = buf[start:]
		return r
	}

	appendTerm := func(term Term) {
		result = append(result, term)
		// Reset token, but reuse the backing memory
		token = token[:0]
	}

	for len(buf) > 0 {
		r = next()
		switch r {
		case ':':
			if open > 0 {
				// ':' inside a hole, likely part of a regexp pattern.
				token = append(token, ':')
				continue
			}
			if len(buf[start:]) > 0 {
				// Look ahead and see if this is the start of a hole.
				if r, _ = utf8.DecodeRune(buf); r == '[' {
					// It is the start of a hole, consume the '['.
					_ = next()
					open++
					appendTerm(Literal(token))
					// Persist the literal token scanned up to this point.
					token = append(token, ':', '[')
					continue
				}
				// Something else, push the ':' we saw and continue.
				token = append(token, ':')
				continue
			}
			// Trailing ':'
			token = append(token, ':')
		case '\\':
			if len(buf[start:]) > 0 && open > 0 {
				// Assume this is an escape sequence for a regexp hole.
				r = next()
				token = append(token, '\\', r)
				continue
			}
			token = append(token, '\\')
		case '[':
			if open > 0 {
				// Assume this is a character set inside a regexp hole.
				inside++
				token = append(token, '[')
				continue
			}
			token = append(token, '[')
		case ']':
			if open > 0 && inside > 0 {
				// This ']' closes a regular expression inside a hole.
				inside--
				token = append(token, ']')
				continue
			}
			if open > 0 {
				// This ']' closes a hole.
				open--
				token = append(token, ']')
				appendTerm(Hole(token))
				continue
			}
			token = append(token, r)
		default:
			token = append(token, r)
		}
	}
	if len(token) > 0 {
		result = append(result, Literal(token))
	}
	return result
}

var onMatchWhitespace = lazyregexp.New(`[\s]+`)

// StructuralPatToRegexpQuery converts a comby pattern to an approximate regular
// expression query. It converts whitespace in the pattern so that content
// across newlines can be matched in the index. As an incomplete approximation,
// we use the regex pattern .*? to scan ahead. A shortcircuit option returns a
// regexp query that may find true matches faster, but may miss all possible
// matches.
//
// Example:
// "ParseInt(:[args]) if err != nil" -> "ParseInt(.*)\s+if\s+err!=\s+nil"
func StructuralPatToRegexpQuery(pattern string, shortcircuit bool) string {
	var pieces []string

	terms := parseTemplate([]byte(pattern))
	for _, term := range terms {
		if term.String() == "" {
			continue
		}
		switch v := term.(type) {
		case Literal:
			piece := regexp.QuoteMeta(v.String())
			piece = onMatchWhitespace.ReplaceAllLiteralString(piece, `[\s]+`)
			pieces = append(pieces, piece)
		case Hole:
			if matchRegexpPattern.MatchString(v.String()) {
				extractedRegexp := matchRegexpPattern.ReplaceAllString(v.String(), `$2`)
				pieces = append(pieces, extractedRegexp)
			}
		default:
			panic("Unreachable")
		}
	}

	if len(pieces) == 0 {
		// Match anything.
		return "(?:.|\\s)*?"
	}

	if shortcircuit {
		// As a shortcircuit, do not match across newlines of structural search pieces.
		return "(?:" + strings.Join(pieces, ").*?(?:") + ")"
	}
	return "(?:" + strings.Join(pieces, ")(?:.|\\s)*?(?:") + ")"
}

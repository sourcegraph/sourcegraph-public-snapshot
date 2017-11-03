package search2

import (
	"fmt"
	"strconv"
	"unicode"
)

// Parse parses a query into tokens.
//
// query ::= ( token )*
// token ::= [ field ":" ] ( term )
//
// If a double-quote character is used inside a term, it must appear
// as the sequence \" (backslash followed by double quote).
func Parse(query string) (tokens Tokens, err error) {
	type parseState int
	const (
		outer parseState = iota
		fieldOrTerm
		term
		quotedTerm
		quotedTermEscape
	)
	state := outer
	var minus bool
	var field Field
	start := -1
	for i, r := range query {
		switch state {
		case outer:
			switch {
			case unicode.IsSpace(r):

			case r == '"':
				state = quotedTerm
				start = i

			case r == '-':
				state = fieldOrTerm
				minus = true
				start = i

			default:
				state = fieldOrTerm
				start = i
			}

		case fieldOrTerm, term:
			switch {
			case ((!minus && start == i) || (minus && start+1 == i) || (state == term && start == i)) && r == '"':
				state = quotedTerm
				if minus && field == "" {
					field = "-"
				}
				start = i

			case unicode.IsSpace(r):
				if minus && field == "" {
					field = "-"
					start++
				}
				tokens = append(tokens, Token{Field: field, Value: Value{Value: query[start:i]}}) // only possible fields are "" or "-"
				state = outer
				field = ""
				minus = false
				start = -1

			case state == fieldOrTerm && r == ':':
				field = Field(query[start:i])
				state = term
				start = i + 1
			}

		case quotedTerm:
			switch r {
			case '"':
				var value string
				value, err = strconv.Unquote(query[start : i+1])
				if err != nil {
					err = &ParseError{Character: i, Message: err.Error()}
					return
				}
				tokens = append(tokens, Token{Field: field, Value: Value{Value: value, Quoted: true}})
				state = outer
				field = ""
				minus = false
				start = -1

			case '\\':
				state = quotedTermEscape
			}

		case quotedTermEscape:
			state = quotedTerm
		}
	}

	if start >= 0 {
		end := len(query)
		if state == quotedTermEscape {
			state = quotedTerm
			end--
		}
		if minus && field == "" {
			field = "-"
			start++
		}
		value := query[start:end]
		if state == quotedTerm {
			value, err = strconv.Unquote(value + `"`)
			if err != nil {
				err = &ParseError{Character: len(query), Message: err.Error()}
				return
			}
		}
		tokens = append(tokens, Token{Field: field, Value: Value{Value: value, Quoted: state == quotedTerm}})
	}

	if tokens == nil {
		tokens = Tokens{}
	}

	return
}

// ParseError occurs when Parse is called with an invalid query string.
type ParseError struct {
	Character int
	Message   string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("query parse error: %s (at character %d)", e.Message, e.Character)
}

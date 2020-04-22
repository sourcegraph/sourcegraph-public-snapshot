package query

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

type ExpectedOperand struct {
	Msg string
}

func (e *ExpectedOperand) Error() string {
	return e.Msg
}

/*
Parser implements a parser for the following grammar:

OrTerm     → AndTerm { OR AndTerm }
AndTerm    → Term { AND Term }
Term       → (OrTerm) | Parameters
Parameters → Parameter { " " Parameter }
*/

type Node interface {
	String() string
	node()
}

// All terms that implement Node.
func (Parameter) node() {}
func (Operator) node()  {}

// Parameter is a leaf node of expressions.
type Parameter struct {
	Field   string `json:"field"`   // The repo part in repo:sourcegraph.
	Value   string `json:"value"`   // The sourcegraph part in repo:sourcegraph.
	Negated bool   `json:"negated"` // True if the - prefix exists, as in -repo:sourcegraph.
	Quoted  bool   `json:"quoted"`  // True if the parsed value was quoted.
}

type operatorKind int

const (
	Or operatorKind = iota
	And
	Concat
)

// Operator is a nonterminal node of kind Kind with child nodes Operands.
type Operator struct {
	Kind     operatorKind
	Operands []Node
}

func (node Parameter) String() string {
	var v string
	switch {
	case node.Field == "":
		v = node.Value
	case node.Negated:
		v = fmt.Sprintf("-%s:%s", node.Field, node.Value)
	default:
		v = fmt.Sprintf("%s:%s", node.Field, node.Value)
	}
	return strconv.Quote(v)
}

func (node Operator) String() string {
	var result []string
	for _, child := range node.Operands {
		result = append(result, child.String())
	}
	var kind string
	switch node.Kind {
	case Or:
		kind = "or"
	case And:
		kind = "and"
	case Concat:
		kind = "concat"
	}

	return fmt.Sprintf("(%s %s)", kind, strings.Join(result, " "))
}

type keyword string

// Reserved keyword syntax.
const (
	AND    keyword = "and"
	OR     keyword = "or"
	LPAREN keyword = "("
	RPAREN keyword = ")"
	SQUOTE keyword = "'"
	DQUOTE keyword = "\""
	SLASH  keyword = "/"
)

func isSpace(buf []byte) bool {
	r, _ := utf8.DecodeRune(buf)
	return unicode.IsSpace(r)
}

// skipSpace returns the number of whitespace bytes skipped from the beginning of a buffer buf.
func skipSpace(buf []byte) int {
	count := 0
	for len(buf) > 0 {
		r, advance := utf8.DecodeRune(buf)
		if !unicode.IsSpace(r) {
			break
		}
		count += advance
		buf = buf[advance:]
	}
	return count
}

type heuristic struct {
	parensAsPatterns    bool // if true, parses parens as patterns rather than expression groups.
	allowDanglingParens bool // if true, disables parsing parentheses as expression groups.
}

type parser struct {
	buf          []byte
	pos          int
	balanced     int
	heuristic    heuristic
	unambiguated bool // if true, this signal implies that at least one expression was unambiguated by explicit parentheses.

}

func (p *parser) done() bool {
	return p.pos >= len(p.buf)
}

func (p *parser) next() rune {
	if p.done() {
		panic("eof")
	}
	r, advance := utf8.DecodeRune(p.buf[p.pos:])
	p.pos += advance
	return r
}

// peek looks ahead n runes in the input and returns a string if it succeeds, or
// an error if the length exceeds what's available in the buffer.
func (p *parser) peek(n int) (string, error) {
	start := p.pos
	defer func() {
		p.pos = start // backtrack
	}()

	var result []rune
	for i := 0; i < n; i++ {
		if p.done() {
			return "", io.ErrShortBuffer
		}
		next := p.next()
		result = append(result, next)
	}
	return string(result), nil
}

// match returns whether it succeeded matching a keyword at the current
// position. It does not advance the position.
func (p *parser) match(keyword keyword) bool {
	v, err := p.peek(len(string(keyword)))
	if err != nil {
		return false
	}
	return strings.EqualFold(v, string(keyword))
}

// expect returns the result of match, and advances the position if it succeeds.
func (p *parser) expect(keyword keyword) bool {
	if !p.match(keyword) {
		return false
	}
	p.pos += len(string(keyword))
	return true
}

// matchKeyword is like match but expects the keyword to be preceded and followed by whitespace.
func (p *parser) matchKeyword(keyword keyword) bool {
	if p.pos == 0 {
		return false
	}
	if !isSpace(p.buf[p.pos-1 : p.pos]) {
		return false
	}
	v, err := p.peek(len(string(keyword)))
	if err != nil {
		return false
	}
	after := p.pos + len(string(keyword))
	if after+1 > len(p.buf) || !isSpace(p.buf[after:after+1]) {
		return false
	}
	return strings.ToLower(v) == string(keyword)
}

// skipSpaces advances the input and places the parser position at the next
// non-space value.
func (p *parser) skipSpaces() error {
	if p.pos > len(p.buf) {
		return io.ErrShortBuffer
	}

	p.pos += skipSpace(p.buf[p.pos:])
	if p.pos > len(p.buf) {
		return io.ErrShortBuffer
	}
	return nil
}

// ScanDelimited takes a delimited (e.g., quoted) value for some arbitrary
// delimiter, returning the undelimited value, and the end position of the
// original delimited value (i.e., including quotes). `\` is treated as an
// escape character for the delimiter and traditional string escape sequences.
// The input buffer must start with the chosen delimiter.
func ScanDelimited(buf []byte, delimiter rune) (string, int, error) {
	var count, advance int
	var r rune
	var result []rune

	next := func() rune {
		r, advance := utf8.DecodeRune(buf)
		count += advance
		buf = buf[advance:]
		return r
	}

	r = next()
	if r != delimiter {
		panic(fmt.Sprintf("ScanDelimited expects the input buffer to start with delimiter %s, but it starts with %s.", string(delimiter), string(r)))
	}

loop:
	for len(buf) > 0 {
		r = next()
		switch {
		case r == delimiter:
			break loop
		case r == '\\':
			// Handle escape sequence.
			if len(buf[advance:]) > 0 {
				r = next()
				switch r {
				case 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', delimiter:
					result = append(result, r)
				default:
					return "", count, errors.New("unrecognized escape sequence")
				}
				if len(buf) <= 0 {
					return "", count, errors.New("unterminated literal: expected " + string(delimiter))
				}
			} else {
				return "", count, errors.New("unterminated escape sequence")
			}
		default:
			result = append(result, r)
		}
	}

	if r != delimiter {
		return "", count, errors.New("unterminated literal: expected " + string(delimiter))
	}
	return string(result), count, nil
}

// ScanField scans an optional '-' at the beginning of a string, and then scans
// one or more alphabetic characters until it encounters a ':', in which case it
// returns the value before the colon and its length. In all other cases it
// returns the empty string and zero length.
func ScanField(buf []byte) (string, int) {
	var count int
	var r rune
	var result []rune
	allowed := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	next := func() rune {
		r, advance := utf8.DecodeRune(buf)
		count += advance
		buf = buf[advance:]
		return r
	}

	r = next()
	if r != '-' && !strings.ContainsRune(allowed, r) {
		return "", 0
	}
	result = append(result, r)

	success := false
	for len(buf) > 0 {
		r = next()
		if strings.ContainsRune(allowed, r) {
			result = append(result, r)
			continue
		}
		if r == ':' {
			// Invariant: len(result) > 0. If len(result) == 1,
			// check that it is not just a '-'. If len(result) > 1, it is valid.
			if result[0] != '-' || len(result) > 1 {
				success = true
			}
		}
		break
	}
	if !success {
		return "", 0
	}
	return string(result), count
}

var fieldValuePattern = lazyregexp.New("(^-?[a-zA-Z0-9]+):(.*)")

// ScanSearchPatternHeuristic scans for a pattern using a heuristic that allows it to
// contain parentheses, if balanced, with appropriate lexical handling for
// traditional escape sequences, escaped parentheses, and escaped whitespace.
func ScanSearchPatternHeuristic(buf []byte) ([]string, int, bool) {
	var count, advance, balanced int
	var r rune
	var piece []rune
	var pieces []string

	next := func() rune {
		r, advance := utf8.DecodeRune(buf)
		count += advance
		buf = buf[advance:]
		return r
	}

loop:
	for len(buf) > 0 {
		r = next()
		switch {
		case unicode.IsSpace(r) && balanced == 0:
			// Stop scanning a potential pattern when we see
			// whitespace in a balanced state.
			break loop
		case r == '(':
			balanced++
			piece = append(piece, r)
		case r == ')':
			balanced--
			piece = append(piece, r)
		case unicode.IsSpace(r):
			// We see a space and the pattern is unbalanced, so assume this
			// terminates a piece of an incomplete search pattern.
			if len(piece) > 0 {
				pieces = append(pieces, string(piece))
			}
			piece = piece[:0]
		case r == '\\':
			// Handle escape sequence.
			if len(buf[advance:]) > 0 {
				r = next()
				if unicode.IsSpace(r) {
					// Interpret escaped whitespace.
					piece = append(piece, r)
					continue
				}
				switch r {
				case 'a', 'b', 'f', 'v', '(', ')':
					piece = append(piece, '\\', r)
				case ':', '\\', '"', '\'':
					piece = append(piece, r)
				case 'n':
					piece = append(piece, '\n')
				case 'r':
					piece = append(piece, '\r')
				case 't':
					piece = append(piece, '\t')
				default:
					// Heuristic is conservative: fail on
					// unrecognized escape sequence.
					// ScanValue will accept unrecognized
					// escape sequences, if applicable.
					return pieces, count, false
				}
			} else {
				// Unterminated escape sequence.
				return pieces, count, false
			}
		default:
			piece = append(piece, r)
		}

	}
	if len(piece) > 0 {
		pieces = append(pieces, string(piece))
	}
	return pieces, count, balanced == 0
}

// ParseSearchPatternHeuristic heuristically parses a search pattern containing
// parentheses at the current position. There are cases where we want to
// interpret parentheses as part of a search pattern, rather than an and/or
// expression group. For example, In the regex foo(a|b)bar, we want to preserve
// parentheses as part of the pattern.
func (p *parser) ParseSearchPatternHeuristic() (Node, bool) {
	if !p.heuristic.parensAsPatterns || p.heuristic.allowDanglingParens {
		return Parameter{Field: "", Value: ""}, false
	}
	if value, ok := p.TryParseDelimiter(); ok {
		return Parameter{Field: "", Value: value}, true
	}

	start := p.pos
	pieces, advance, ok := ScanSearchPatternHeuristic(p.buf[p.pos:])
	end := start + advance
	if !ok || len(p.buf[start:end]) == 0 || !isPureSearchPattern(p.buf[start:end]) {
		// We tried validating the pattern but it is either unbalanced
		// or malformed, empty, or an invalid and/or expression.
		return Parameter{Field: "", Value: ""}, false
	}
	// The heuristic succeeds: we can process the string as a pure search pattern.
	p.pos += advance
	if len(pieces) == 1 {
		return Parameter{Field: "", Value: pieces[0]}, true
	}
	parameters := []Node{}
	for _, piece := range pieces {
		parameters = append(parameters, Parameter{Field: "", Value: piece})
	}
	return Operator{Kind: Concat, Operands: parameters}, true
}

// ScanValue scans for a value (e.g., of a parameter, or a string corresponding
// to a search pattern). It's main function is to determine when to stop
// scanning a value (e.g., at a parentheses), and which escape sequences to
// interpret.
func ScanValue(buf []byte, allowDanglingParens bool) (string, int) {
	var count, advance int
	var r rune
	var result []rune

	next := func() rune {
		r, advance := utf8.DecodeRune(buf)
		count += advance
		buf = buf[advance:]
		return r
	}

	for len(buf) > 0 {
		start := count
		r = next()
		if unicode.IsSpace(r) {
			count = start // Backtrack.
			break
		}
		if r == '(' || r == ')' {
			if allowDanglingParens {
				result = append(result, r)
				continue
			}
			count = start // Backtrack.
			break
		}
		if r == '\\' {
			// Handle escape sequence.
			if len(buf[advance:]) > 0 {
				r = next()
				if unicode.IsSpace(r) {
					// Interpret escaped whitespace.
					result = append(result, r)
					continue
				}
				// Interpret escape sequences for:
				// (1) special syntax in our language :\"'/
				// (2) whitespace escape sequences \n\r\t
				switch r {
				case ':', '\\', '"', '\'':
					result = append(result, r)
				case 'n':
					result = append(result, '\n')
				case 'r':
					result = append(result, '\r')
				case 't':
					result = append(result, '\t')
				default:
					// Accept anything else literally.
					result = append(result, '\\', r)
				}
				continue
			}
		}
		result = append(result, r)
	}
	return string(result), count
}

// TryParseDelimiter tries to parse a delimited, returning whether it succeeded,
// and the interpreted (i.e., unquoted) value if it succeeds.
func (p *parser) TryParseDelimiter() (string, bool) {
	delimited := func(delimiter rune) (string, error) {
		value, advance, err := ScanDelimited(p.buf[p.pos:], delimiter)
		if err != nil {
			return "", err
		}
		p.pos += advance
		return value, nil
	}
	tryScanDelimiter := func() (string, error) {
		if p.match(SQUOTE) {
			return delimited('\'')
		}
		if p.match(DQUOTE) {
			return delimited('"')
		}
		return "", errors.New("failed to scan delimiter")
	}
	if value, err := tryScanDelimiter(); err == nil {
		return value, true
	}
	return "", false
}

// ParseValue parses a value at the current position and whether it was quoted.
// A value may belong to a parameter like repo:<value> or the <value> may be a
// search pattern. Values may be quoted. ParseValue cannot fail: it will first
// try to scan well-formed delimiters, like quoted strings. If that fails, it
// will accept unbalanced strings as patterns. Uses of value can then be
// validated along other concerns for different search types (literal, regexp,
// etc.).
func (p *parser) ParseValue() (string, bool) {
	if value, ok := p.TryParseDelimiter(); ok {
		return value, true
	}
	value, advance := ScanValue(p.buf[p.pos:], p.heuristic.allowDanglingParens)
	p.pos += advance
	return value, false
}

// ParseParameter returns a leaf node usable by _any_ kind of search (e.g.,
// literal or regexp, or...) and always succeeds.
//
// A parameter is a contiguous sequence of characters, where the following two forms are distinguished:
// (1) a string of syntax field:<string> where : matches the first encountered colon, and field must match ^-?[a-zA-Z]+
// (2) <string>
//
// When a parameter is of form (1), the <string> corresponds to Parameter.Value,
// field corresponds to Parameter.Field and Parameter.Negated is set if Field
// starts with '-'. When form (1) does not match, Value corresponds to <string>
// and Field is the empty string.
func (p *parser) ParseParameter() Parameter {
	field, advance := ScanField(p.buf[p.pos:])
	p.pos += advance
	value, quoted := p.ParseValue()
	negated := len(field) > 0 && field[0] == '-'
	if negated {
		field = field[1:]
	}
	return Parameter{Field: field, Value: value, Negated: negated, Quoted: quoted}
}

// containsPattern returns true if any descendent of nodes is a search pattern
// (i.e., a parameter where the field is the empty string).
func containsPattern(node Node) bool {
	var result bool
	VisitField([]Node{node}, "", func(_ string, _, _ bool) {
		result = true
	})
	return result
}

// returns true if descendent of node contains and/or expressions.
func containsAndOrExpression(nodes []Node) bool {
	var result bool
	VisitOperator(nodes, func(kind operatorKind, _ []Node) {
		if kind == And || kind == Or {
			result = true
		}
	})
	return result
}

// partitionParameters constructs a parse tree to distinguish terms where
// ordering is insignificant (e.g., "repo:foo file:bar") versus terms where
// ordering may be significant (e.g., search patterns like "foo bar"). Search
// patterns are parameters whose field is the empty string.
//
// The resulting tree defines an ordering relation on nodes in the following cases:
// (1) When more than one search patterns exist at the same operator level, they
// are concatenated in order.
// (2) Any nonterminal node is concatenated (ordered in the tree) if its
// descendents contain one or more search patterns.
func partitionParameters(nodes []Node) []Node {
	var patterns, unorderedParams []Node
	for _, n := range nodes {
		switch v := n.(type) {
		case Parameter:
			if v.Field == "" {
				patterns = append(patterns, n)
			} else {
				unorderedParams = append(unorderedParams, n)
			}
		case Operator:
			if containsPattern(n) {
				patterns = append(patterns, n)
			} else {
				unorderedParams = append(unorderedParams, n)
			}
		}
	}
	if len(patterns) > 1 {
		orderedPatterns := newOperator(patterns, Concat)
		return newOperator(append(unorderedParams, orderedPatterns...), And)
	}
	return newOperator(append(unorderedParams, patterns...), And)
}

// parseParameterParameterList scans for consecutive leaf nodes.
func (p *parser) parseParameterList() ([]Node, error) {
	var nodes []Node
loop:
	for {
		if err := p.skipSpaces(); err != nil {
			return nil, err
		}
		if p.done() {
			break loop
		}
		switch {
		case p.match(LPAREN) && !p.heuristic.allowDanglingParens:
			// First try parse a parameter as a search pattern containing parens.
			if patterns, ok := p.ParseSearchPatternHeuristic(); ok {
				nodes = append(nodes, patterns)
			} else {
				// If the above failed, we treat this paren
				// group as part of an and/or expression.
				_ = p.expect(LPAREN) // Guaranteed to succeed.
				p.balanced++
				p.unambiguated = true
				result, err := p.parseOr()
				if err != nil {
					return nil, err
				}
				nodes = append(nodes, result...)
			}
		case p.expect(RPAREN) && !p.heuristic.allowDanglingParens:
			p.balanced--
			p.unambiguated = true
			if len(nodes) == 0 {
				// We parsed "()".
				if p.heuristic.parensAsPatterns {
					// Interpret literally.
					nodes = []Node{Parameter{Value: "()"}}
				} else {
					// Interpret as a group: return an empty non-nil node.
					nodes = []Node{Parameter{Value: ""}}
				}
			}
			break loop
		case p.matchKeyword(AND), p.matchKeyword(OR):
			// Caller advances.
			break loop
		default:
			// First try parse a parameter as a search pattern containing parens.
			if parameter, ok := p.ParseSearchPatternHeuristic(); ok {
				nodes = append(nodes, parameter)
			} else {
				parameter := p.ParseParameter()
				nodes = append(nodes, parameter)
			}
		}
	}
	return partitionParameters(nodes), nil
}

// reduce takes lists of left and right nodes and reduces them if possible. For example,
// (and a (b and c))       => (and a b c)
// (((a and b) or c) or d) => (or (and a b) c d)
func reduce(left, right []Node, kind operatorKind) ([]Node, bool) {
	if param, ok := left[0].(Parameter); ok && param.Value == "" {
		// Remove empty string parameter.
		return right, true
	}

	switch term := right[0].(type) {
	case Operator:
		if kind == term.Kind {
			// Reduce right node.
			left = append(left, term.Operands...)
			if len(right) > 1 {
				left = append(left, right[1:]...)
			}
			return left, true
		}
	case Parameter:
		if term.Value == "" {
			// Remove empty string parameter.
			if len(right) > 1 {
				return append(left, right[1:]...), true
			}
			return left, true
		}
		if operator, ok := left[0].(Operator); ok && operator.Kind == kind {
			// Reduce left node.
			return append(operator.Operands, right...), true

		}
	}
	if len(right) > 1 {
		// Reduce right list.
		reduced, changed := reduce(append(left, right[0]), right[1:], kind)
		if changed {
			return reduced, true
		}
	}
	return append(left, right...), false
}

// newOperator constructs a new node of kind operatorKind with operands nodes,
// reducing nodes as needed.
func newOperator(nodes []Node, kind operatorKind) []Node {
	if len(nodes) == 0 {
		return nil
	} else if len(nodes) == 1 {
		return nodes
	}

	reduced, changed := reduce([]Node{nodes[0]}, nodes[1:], kind)
	if changed {
		return newOperator(reduced, kind)
	}
	return []Node{Operator{Kind: kind, Operands: reduced}}
}

// parseAnd parses and-expressions.
func (p *parser) parseAnd() ([]Node, error) {
	left, err := p.parseParameterList()
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, &UnsupportedError{Msg: fmt.Sprintf("expected operand at %d", p.pos)}
	}
	if !p.expect(AND) {
		return left, nil
	}
	right, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	return newOperator(append(left, right...), And), nil
}

// parseOr parses or-expressions. Or operators have lower precedence than And
// operators, therefore this function calls parseAnd.
func (p *parser) parseOr() ([]Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, &UnsupportedError{Msg: fmt.Sprintf("expected operand at %d", p.pos)}
	}
	if !p.expect(OR) {
		return left, nil
	}
	right, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	return newOperator(append(left, right...), Or), nil
}

func tryFallbackParser(in string) ([]Node, error) {
	parser := &parser{
		buf:       []byte(in),
		heuristic: heuristic{allowDanglingParens: true},
	}
	nodes, err := parser.parseOr()
	if err != nil {
		return nil, err
	}
	if hoistedNodes, err := Hoist(nodes); err == nil {
		return newOperator(hoistedNodes, And), nil
	}
	return newOperator(nodes, And), nil
}

// ParseAndOr a raw input string into a parse tree comprising Nodes.
func ParseAndOr(in string) ([]Node, error) {
	if strings.TrimSpace(in) == "" {
		return nil, nil
	}
	parser := &parser{
		buf:       []byte(in),
		heuristic: heuristic{parensAsPatterns: true},
	}

	nodes, err := parser.parseOr()
	if err != nil {
		if nodes, err := tryFallbackParser(in); err == nil {
			return nodes, nil
		}
		return nil, err
	}
	if parser.balanced != 0 {
		if nodes, err := tryFallbackParser(in); err == nil {
			return nodes, nil
		}
		return nil, errors.New("unbalanced expression")
	}
	if !parser.unambiguated {
		// Hoist or expressions if this query is potential ambiguous.
		if hoistedNodes, err := Hoist(nodes); err == nil {
			nodes = hoistedNodes
		}
	}
	return newOperator(nodes, And), nil
}

// ProcessAndOr query parses and validates an and/or query for a given search type.
func ProcessAndOr(in string) (QueryInfo, error) {
	query, err := ParseAndOr(in)
	if err != nil {
		return nil, err
	}
	query = LowercaseFieldNames(query)
	err = validate(query)
	if err != nil {
		return nil, err
	}
	query = SubstituteAliases(query)
	return &AndOrQuery{Query: query}, nil
}

package query

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/inconshreveable/log15"
)

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
func (Pattern) node()   {}
func (Parameter) node() {}
func (Operator) node()  {}

// An annotation stores information associated with a node.
type Annotation struct {
	Labels labels `json:"labels"`
	Range  Range  `json:"range"`
}

// Pattern is a leaf node of expressions representing a search pattern fragment.
type Pattern struct {
	Value      string     `json:"value"`   // The pattern value.
	Negated    bool       `json:"negated"` // True if this pattern is negated.
	Annotation Annotation `json:"-"`       // An annotation attached to this pattern.
}

// Parameter is a leaf node of expressions representing a parameter of format "repo:foo".
type Parameter struct {
	Field      string     `json:"field"`   // The repo part in repo:sourcegraph.
	Value      string     `json:"value"`   // The sourcegraph part in repo:sourcegraph.
	Negated    bool       `json:"negated"` // True if the - prefix exists, as in -repo:sourcegraph.
	Annotation Annotation `json:"-"`
}

type operatorKind int

const (
	Or operatorKind = iota
	And
	Concat
)

// Operator is a nonterminal node of kind Kind with child nodes Operands.
type Operator struct {
	Kind       operatorKind
	Operands   []Node
	Annotation Annotation
}

func (node Pattern) String() string {
	var v string
	if node.Negated {
		v = fmt.Sprintf("NOT %s", node.Value)
	} else {
		v = node.Value
	}
	return strconv.Quote(v)
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
	NOT    keyword = "not"
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

type heuristics uint8

const (
	// If set, balanced parentheses, which would normally be treated as
	// delimiting expression groups, may in select cases be parsed as
	// literal search patterns instead.
	parensAsPatterns heuristics = 1 << iota
	// If set, all parentheses, whether balanced or unbalanced, are parsed
	// as literal search patterns (i.e., interpreting parentheses as
	// expression groups is completely disabled).
	allowDanglingParens
	// If set, implies that at least one expression was disambiguated by
	// explicit parentheses.
	disambiguated
)

func isSet(h, heuristic heuristics) bool { return h&heuristic != 0 }

type parser struct {
	buf        []byte
	heuristics heuristics
	pos        int
	balanced   int
	leafParser SearchType
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
	if after >= len(p.buf) || !isSpace(p.buf[after:after+1]) {
		return false
	}
	return strings.EqualFold(v, string(keyword))
}

// matchUnaryKeyword is like match but expects the keyword to be followed by whitespace.
func (p *parser) matchUnaryKeyword(keyword keyword) bool {
	if p.pos != 0 && !isSpace(p.buf[p.pos-1:p.pos]) {
		return false
	}
	v, err := p.peek(len(string(keyword)))
	if err != nil {
		return false
	}
	after := p.pos + len(string(keyword))
	if after >= len(p.buf) || !isSpace(p.buf[after:after+1]) {
		return false
	}
	return strings.EqualFold(v, string(keyword))
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

// parseNegatedLeafNode parses `NOT field:value` or `NOT pattern` and
// translates it to `-field:value` or as negated pattern respectively.
func (p *parser) parseNegatedLeafNode() (Node, error) {
	start := p.pos
	_ = p.expect(NOT)

	err := p.skipSpaces()
	if err != nil {
		return Parameter{}, err
	}

	// try parsing as parameter. If it doesn't work we treat NOT's operand
	// as pattern.
	parameter, ok, err := p.ParseParameter()
	if err != nil {
		return nil, err
	}
	if !ok {
		pattern := p.ParsePattern()
		pattern.Negated = true
		pattern.Annotation.Range = newRange(start, p.pos)
		return pattern, nil
	}
	// we don't support NOT -field:value
	if parameter.Negated {
		return nil, fmt.Errorf("Unexpected NOT before \"-%s:%s\". Remove NOT and try again.",
			parameter.Field, parameter.Value)
	}
	parameter.Negated = true
	parameter.Annotation.Range = newRange(start, p.pos)
	return parameter, nil
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
				case 'a', 'b', 'f', 'v':
					result = append(result, '\\', r)
				case 'n':
					result = append(result, '\n')
				case 'r':
					result = append(result, '\r')
				case 't':
					result = append(result, '\t')
				case '\\', delimiter:
					result = append(result, r)
				default:
					return "", count, errors.New("unrecognized escape sequence")
				}
				if len(buf) == 0 {
					return "", count, errors.New("unterminated literal: expected " + string(delimiter))
				}
			} else {
				return "", count, errors.New("unterminated escape sequence")
			}
		default:
			result = append(result, r)
		}
	}

	if r != delimiter || (r == delimiter && count == 1) {
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

// ScanValue scans for a value (e.g., of a parameter, or a string corresponding
// to a search pattern). Its main function is to determine when to stop scanning
// a value (e.g., at a parentheses), and which escape sequences to interpret. It
// returns the scanned value, how much was advanced, and whether the
// allowDanglingParenthesis heuristic was applied
func ScanValue(buf []byte, allowDanglingParens bool) (string, int, bool) {
	var count, advance, balanced int
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
			if r == '(' {
				balanced++
			}
			if r == ')' {
				balanced--
			}
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
				// Interpret escape sequences for whitespace escape
				// sequences \n\r\t. Other escape sequences fall
				// through to a regexp interpretation.
				switch r {
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
	return string(result), count, balanced != 0
}

// TryParseDelimiter tries to parse a delimited string, returning whether it
// succeeded, and the interpreted (i.e., unquoted) value if it succeeds.
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

// ParseFieldValue parses a value after a field like "repo:". If the value
// starts with a recognized quoting delimiter but does not close it, an error is
// returned.
func (p *parser) ParseFieldValue() (string, error) {
	delimited := func(delimiter rune) (string, error) {
		value, advance, err := ScanDelimited(p.buf[p.pos:], delimiter)
		if err != nil {
			return "", err
		}
		p.pos += advance
		return value, nil
	}
	if p.match(SQUOTE) {
		return delimited('\'')
	}
	if p.match(DQUOTE) {
		return delimited('"')
	}
	// First try scan a field value for cases like (a b repo:foo), where a
	// trailing ) may be closing a group, and not part of the value.
	value, advance, ok := ScanBalancedPatternLiteral(p.buf[p.pos:])
	if !ok {
		// The above failed, so attempt a best effort.
		value, advance, _ = ScanValue(p.buf[p.pos:], false)
	}
	p.pos += advance
	return value, nil
}

// ParsePattern parses a leaf node Pattern that corresponds to a search pattern.
// Note that ParsePattern may be called multiple times (a query can have
// multiple Patterns concatenated together).
func (p *parser) ParsePattern() Pattern {
	start := p.pos
	// If we can parse a well-delimited value, that takes precedence, and we
	// denote it with Quoted set to true.
	if value, ok := p.TryParseDelimiter(); ok {
		return Pattern{
			Value:   value,
			Negated: false,
			Annotation: Annotation{
				Labels: Literal | Quoted,
				Range:  newRange(start, p.pos),
			},
		}
	}

	value, advance, sawDanglingParen := ScanValue(p.buf[p.pos:], isSet(p.heuristics, allowDanglingParens))
	var labels labels
	if sawDanglingParen {
		// If we saw a dangling parenthesis, this is not a well-formed
		// regular expression and we will interpret it as a literal.
		// TODO(rvantonder): Try to still support a trailing parentheses
		// combined with regex, like "foo.*bar(".
		labels = HeuristicDanglingParens | Literal
	} else {
		labels = Regexp
	}
	p.pos += advance
	// Invariant: the pattern can't be quoted since we checked for that.
	return Pattern{
		Value:   value,
		Negated: false,
		Annotation: Annotation{
			Labels: labels,
			Range:  newRange(start, p.pos),
		},
	}
}

// ParseParameter returns a leaf node corresponding to the syntax
// (-?)field:<string> where : matches the first encountered colon, and field
// must match ^[a-zA-Z]+ and be allowed by allFields. Field may optionally
// be preceded by '-' which means the parameter is negated.
func (p *parser) ParseParameter() (Parameter, bool, error) {
	start := p.pos
	field, advance := ScanField(p.buf[p.pos:])
	if field == "" {
		return Parameter{}, false, nil
	}

	negated := field[0] == '-'
	if negated {
		field = field[1:]
	}

	if _, exists := allFields[strings.ToLower(field)]; !exists {
		// Not a recognized parameter field.
		return Parameter{}, false, nil
	}

	p.pos += advance
	value, err := p.ParseFieldValue()
	if err != nil {
		return Parameter{}, false, err
	}
	return Parameter{
		Field:      field,
		Value:      value,
		Negated:    negated,
		Annotation: Annotation{Range: newRange(start, p.pos)},
	}, true, nil
}

// partitionParameters constructs a parse tree to distinguish terms where
// ordering is insignificant (e.g., "repo:foo file:bar") versus terms where
// ordering may be significant (e.g., search patterns like "foo bar").
//
// The resulting tree defines an ordering relation on nodes in the following cases:
// (1) When more than one search patterns exist at the same operator level, they
// are concatenated in order.
// (2) Any nonterminal node is concatenated (ordered in the tree) if its
// descendents contain one or more search patterns.
func partitionParameters(nodes []Node) []Node {
	var patterns, unorderedParams []Node
	for _, n := range nodes {
		switch n.(type) {
		case Pattern:
			patterns = append(patterns, n)
		case Parameter:
			unorderedParams = append(unorderedParams, n)
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

// concatPatterns extends the left pattern with the right pattern, appropriately
// updating the ranges and annotations.
func concatPatterns(left, right Pattern) (Pattern, error) {
	if left.Annotation.Range.End.Column != right.Annotation.Range.Start.Column {
		log15.Warn("parser can't process concatPatterns",
			"left", left, "right", right,
			"leftEnd", left.Annotation.Range.End.Column,
			"rightBegin", right.Annotation.Range.Start.Column)
		return Pattern{}, &UnsupportedError{Msg: "invalid query syntax"}
	}
	left.Value += right.Value
	left.Annotation.Labels |= right.Annotation.Labels
	left.Annotation.Range.End.Column += len(right.Value)
	return left, nil
}

// parseLeavesRegexp scans for consecutive leaf nodes when interpreting the
// query as containing regexp patterns.
func (p *parser) parseLeavesRegexp() ([]Node, error) {
	var nodes []Node
	start := p.pos
loop:
	for {
		if err := p.skipSpaces(); err != nil {
			return nil, err
		}
		if p.done() {
			break loop
		}
		switch {
		case p.match(LPAREN) && !isSet(p.heuristics, allowDanglingParens):
			if isSet(p.heuristics, parensAsPatterns) {
				if value, advance, ok := ScanBalancedPatternLiteral(p.buf[p.pos:]); ok {
					pattern := Pattern{
						Value:   value,
						Negated: false,
						Annotation: Annotation{
							Labels: Regexp,
							Range:  newRange(p.pos, p.pos+advance),
						},
					}

					// Concat when a pattern like foo()bar is parsed as (concat "foo" "()bar")
					// if these are not space-separated.
					if p.pos > 0 {
						if r, _ := utf8.DecodeRune([]byte{p.buf[p.pos-1]}); !unicode.IsSpace(r) {
							if len(nodes) > 0 {
								if previous, ok := nodes[len(nodes)-1].(Pattern); ok {
									previous.Annotation.Labels = Regexp | HeuristicParensAsPatterns
									result, err := concatPatterns(previous, pattern)
									if err != nil {
										return nil, err
									}
									nodes[len(nodes)-1] = result
									p.pos += advance
									continue
								}
							}
						}
					}

					p.pos += advance
					nodes = append(nodes, pattern)
					continue
				}
			}
			// If the above failed, we treat this paren
			// group as part of an and/or expression.
			_ = p.expect(LPAREN) // Guaranteed to succeed.
			p.balanced++
			p.heuristics |= disambiguated
			result, err := p.parseOr()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, result...)
		case p.expect(RPAREN) && !isSet(p.heuristics, allowDanglingParens):
			p.balanced--
			p.heuristics |= disambiguated
			if len(nodes) == 0 {
				// We parsed "()".
				if isSet(p.heuristics, parensAsPatterns) {
					// Interpret literally.
					nodes = []Node{
						Pattern{
							Value: "()",
							Annotation: Annotation{
								Labels: Literal | HeuristicParensAsPatterns,
								Range:  newRange(start, p.pos),
							},
						},
					}
				} else {
					// Interpret as a group: return an empty non-nil node.
					nodes = []Node{Parameter{}}
				}
			}
			break loop
		case p.matchKeyword(AND), p.matchKeyword(OR):
			// Caller advances.
			break loop
		case p.matchUnaryKeyword(NOT):
			parameter, err := p.parseNegatedLeafNode()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, parameter)
		default:
			parameter, ok, err := p.ParseParameter()
			if err != nil {
				return nil, err
			}
			if ok {
				nodes = append(nodes, parameter)
			} else {
				pattern := p.ParsePattern()
				nodes = append(nodes, pattern)
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
	case Pattern:
		if term.Value == "" {
			// Remove empty string pattern.
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
	var left []Node
	var err error
	if p.leafParser == SearchTypeRegex {
		left, err = p.parseLeavesRegexp()
	} else {
		left, err = p.parseLeavesLiteral()
	}
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, &ExpectedOperand{Msg: fmt.Sprintf("expected operand at %d", p.pos)}
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
		return nil, &ExpectedOperand{Msg: fmt.Sprintf("expected operand at %d", p.pos)}
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

func (p *parser) tryFallbackParser(in string) ([]Node, error) {
	newParser := &parser{
		buf:        []byte(in),
		heuristics: allowDanglingParens,
		leafParser: p.leafParser,
	}
	nodes, err := newParser.parseOr()
	if err != nil {
		return nil, err
	}
	if hoistedNodes, err := Hoist(nodes); err == nil {
		return newOperator(hoistedNodes, And), nil
	}
	return newOperator(nodes, And), nil
}

// ParseAndOr a raw input string into a parse tree comprising Nodes.
func ParseAndOr(in string, searchType SearchType) ([]Node, error) {
	if strings.TrimSpace(in) == "" {
		return nil, nil
	}

	parser := &parser{
		buf:        []byte(in),
		heuristics: parensAsPatterns,
		leafParser: searchType,
	}

	nodes, err := parser.parseOr()
	if err != nil {
		switch err.(type) {
		case *ExpectedOperand:
			// The query may be unbalanced or malformed as in "(" or
			// "x or" and expects an operand. Try harder to parse it.
			if nodes, err := parser.tryFallbackParser(in); err == nil {
				return nodes, nil
			}
		}
		// Another kind of error, like a malformed parameter.
		return nil, err
	}
	if parser.balanced != 0 {
		// The query is unbalanced and might be something like "(x" or
		// "x or (x" where patterns start with a leading open
		// parenthesis. Try harder to parse it.
		if nodes, err := parser.tryFallbackParser(in); err == nil {
			return nodes, nil
		}
		return nil, errors.New("unbalanced expression")
	}
	if !isSet(parser.heuristics, disambiguated) {
		// Hoist or expressions if this query is potential ambiguous.
		if hoistedNodes, err := Hoist(nodes); err == nil {
			nodes = hoistedNodes
		}
	}
	if searchType == SearchTypeLiteral {
		err = validatePureLiteralPattern(nodes, parser.balanced == 0)
		if err != nil {
			return nil, err
		}
	}
	return newOperator(nodes, And), nil
}

type ParserOptions struct {
	SearchType SearchType

	// treat repo, file, or repohasfile values as glob syntax if true.
	Globbing bool
}

// ProcessAndOr query parses and validates an and/or query for a given search type.
func ProcessAndOr(in string, options ParserOptions) (QueryInfo, error) {
	var query []Node
	var err error

	query, err = ParseAndOr(in, options.SearchType)
	if err != nil {
		return nil, err
	}

	query = Map(query, LowercaseFieldNames, SubstituteAliases)

	switch options.SearchType {
	case SearchTypeLiteral:
		query = substituteConcat(query, " ")
	case SearchTypeStructural:
		if containsNegatedPattern(query) {
			return nil, errors.New("The query contains a negated search pattern. Structural search does not support negated search patterns at the moment.")
		}
		query = substituteConcat(query, " ")
	case SearchTypeRegex:
		query = EmptyGroupsToLiteral(query)
	}

	if options.Globbing {
		query, err = mapGlobToRegex(query)
		if err != nil {
			return nil, err
		}
	}

	err = validate(query)
	if err != nil {
		return nil, err
	}
	return &AndOrQuery{Query: query}, nil
}

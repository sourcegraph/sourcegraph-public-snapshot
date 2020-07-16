package query

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/pkg/errors"
)

// ScanAnyPatternLiteral consumes all characters up to a whitespace character
// and returns the string and how much it consumed.
func ScanAnyPatternLiteral(buf []byte) (scanned string, count int) {
	var advance int
	var r rune
	var result []rune

	next := func() rune {
		r, advance = utf8.DecodeRune(buf)
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
		result = append(result, r)
	}
	scanned = string(result)
	return scanned, count
}

// isParameter returns whether the token is a parameter.
func isParameter(token []byte) bool {
	parser := &parser{buf: token}
	_, ok, _ := parser.ParseParameter()
	return ok
}

// ScanBalancedPatternLiteral attempts to scan parentheses as literal patterns.
// It returns the scanned string, how much to advance, and whether it succeeded.
// Basically it scans any literal string, including whitespace, but ensures that
// a resulting string does not contain 'and' or 'or keywords, and is balanced.
func ScanBalancedPatternLiteral(buf []byte) (scanned string, count int, ok bool) {
	var advance, balanced int
	var r rune
	var result []rune

	next := func() rune {
		r, advance = utf8.DecodeRune(buf)
		count += advance
		buf = buf[advance:]
		return r
	}

	var token []byte

loop:
	for len(buf) > 0 {
		start := count
		r = next()
		switch {
		case unicode.IsSpace(r) && balanced == 0:
			// Stop scanning a potential pattern when we see
			// whitespace in a balanced state.
			break loop
		case r == '(':
			balanced++
			result = append(result, r)
		case r == ')':
			balanced--
			if balanced < 0 {
				// This paren is an unmatched closing paren, so
				// we stop treating it as a potential pattern
				// here--it might be closing a group.
				count = start // Backtrack.
				balanced = 0  // Pattern is balanced up to this point.
				break loop
			}
			result = append(result, r)
		case unicode.IsSpace(r):
			if isParameter(token) {
				// This is not a pattern, one of the tokens is parameter.
				return "", 0, false
			}
			token = []byte{}

			// We see a space and the pattern is unbalanced, so assume this
			// this space is still part of the pattern.
			result = append(result, r)
		default:
			token = append(token, []byte(string(r))...)
			result = append(result, r)
		}
	}

	if isParameter(token) {
		// This is not a pattern, one of the tokens is parameter.
		return "", 0, false
	}

	scanned = string(result)
	if ContainsAndOrKeyword(scanned) {
		// Reject if we scanned 'and' or 'or'. Preceding parentheses
		// likely refer to a group, not a pattern.
		return "", 0, false
	}
	return scanned, count, balanced == 0
}

func (p *parser) ParsePatternLiteral() Pattern {
	start := p.pos
	if value, advance, ok := ScanBalancedPatternLiteral(p.buf[p.pos:]); ok && value != "" {
		p.pos += advance
		return Pattern{
			Value:   value,
			Negated: false,
			Annotation: Annotation{
				Labels: Literal,
				Range:  newRange(start, p.pos),
			},
		}
	}
	value, advance := ScanAnyPatternLiteral(p.buf[p.pos:])
	p.pos += advance
	return Pattern{
		Value:   value,
		Negated: false,
		Annotation: Annotation{
			Labels: Literal,
			Range:  newRange(start, p.pos),
		},
	}
}

// parseParameterParameterList scans for consecutive leaf nodes.
func (p *parser) parseParameterListLiteral() ([]Node, error) {
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
		case p.match(LPAREN):
			if value, advance, ok := ScanBalancedPatternLiteral(p.buf[p.pos:]); ok {
				p.pos += advance
				pattern := Pattern{
					Value:   value,
					Negated: false,
					Annotation: Annotation{
						Labels: Literal | HeuristicParensAsPatterns,
						Range:  newRange(start, p.pos),
					},
				}
				nodes = append(nodes, pattern)
				continue
			}
			if isSet(p.heuristics, allowDanglingParens) {
				// Consume strings containing unbalanced
				// parentheses up to whitespace.
				pattern := p.ParsePatternLiteral()
				pattern.Annotation.Labels |= HeuristicDanglingParens
				nodes = append(nodes, pattern)
				continue
			}
			_ = p.expect(LPAREN) // Guaranteed to succeed.
			p.balanced++
			p.heuristics |= disambiguated
			result, err := p.parseOrLiteral()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, result...)
		case p.match(RPAREN):
			if p.balanced <= 0 {
				// This is a dangling right paren. It can't possibly help
				// us parse a well-formed query, so try treat it as a pattern.
				pattern := p.ParsePatternLiteral()
				pattern.Annotation.Labels |= HeuristicDanglingParens

				// Heuristic: This right paren may be one we should associate with a previous pattern, and not
				// just a dangling one. Check if a pattern occurred before it and append it if so.
				if p.pos > 0 {
					// Heuristic is imprecise and that's OK: It will only look for a 1-byte whitespace
					// character (not any unicode whitespace) before this paren.
					if r, _ := utf8.DecodeRune([]byte{p.buf[p.pos-1]}); !unicode.IsSpace(r) {
						if len(nodes) > 0 {
							if previous, ok := nodes[len(nodes)-1].(Pattern); ok {
								nodes[len(nodes)-1] = concatPatterns(previous, pattern)
								continue
							}
						}
					}
				}

				nodes = append(nodes, pattern)
				continue
			}
			_ = p.expect(RPAREN) // Guaranteed to succeed.
			p.balanced--
			p.heuristics |= disambiguated
			if len(nodes) == 0 {
				// We parsed "()", interpret it literally.
				nodes = []Node{
					Pattern{
						Value: "()",
						Annotation: Annotation{
							Labels: Literal | HeuristicParensAsPatterns,
							Range:  newRange(start, p.pos),
						},
					},
				}
			}
			break loop
		case p.matchKeyword(AND), p.matchKeyword(OR):
			// Caller advances.
			break loop
		default:
			parameter, ok, err := p.ParseParameter()
			if err != nil {
				return nil, err
			}
			if ok {
				nodes = append(nodes, parameter)
			} else {
				pattern := p.ParsePatternLiteral()
				nodes = append(nodes, pattern)
			}
		}
	}
	return partitionParameters(nodes), nil
}

// parseAnd parses and-expressions.
func (p *parser) parseAndLiteral() ([]Node, error) {
	left, err := p.parseParameterListLiteral()
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, &ExpectedOperand{Msg: fmt.Sprintf("expected operand at %d", p.pos)}
	}
	if !p.expect(AND) {
		return left, nil
	}
	right, err := p.parseAndLiteral()
	if err != nil {
		return nil, err
	}
	return newOperator(append(left, right...), And), nil
}

// parseOr parses or-expressions. Or operators have lower precedence than And
// operators, therefore this function calls parseAnd.
func (p *parser) parseOrLiteral() ([]Node, error) {
	left, err := p.parseAndLiteral()
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, &ExpectedOperand{Msg: fmt.Sprintf("expected operand at %d", p.pos)}
	}
	if !p.expect(OR) {
		return left, nil
	}
	right, err := p.parseOrLiteral()
	if err != nil {
		return nil, err
	}
	return newOperator(append(left, right...), Or), nil
}

func literalFallbackParser(in string) ([]Node, error) {
	parser := &parser{
		buf:        []byte(in),
		heuristics: allowDanglingParens,
	}
	nodes, err := parser.parseOrLiteral()
	if err != nil {
		return nil, err
	}
	if hoistedNodes, err := Hoist(nodes); err == nil {
		return newOperator(hoistedNodes, And), nil
	}
	return newOperator(nodes, And), nil
}

// validatePureLiteralPattern checks that no pattern expression contains and/or
// operators nested inside concat. It may happen that we interpret a query this
// way due to ambiguity. If this happens, return an error message.
func validatePureLiteralPattern(nodes []Node, balanced bool) error {
	impure := exists(nodes, func(node Node) bool {
		if operator, ok := node.(Operator); ok && operator.Kind == Concat {
			for _, node := range operator.Operands {
				if op, ok := node.(Operator); ok && (op.Kind == Or || op.Kind == And) {
					return true
				}
			}
		}
		return false
	})
	if impure {
		if !balanced {
			return errors.New("this literal search query contains unbalanced parentheses. I tried to guess what you meant, but wasn't able to. Maybe you missed a parenthesis? Otherwise, try using the content: filter if the pattern is unbalanced")
		}
		return errors.New("i'm having trouble understanding that query. The combination of parentheses is the problem. Try using the content: filter to quote patterns that contain parentheses")
	}
	return nil
}

func ParseAndOrLiteral(in string) ([]Node, error) {
	if strings.TrimSpace(in) == "" {
		return nil, nil
	}
	parser := &parser{buf: []byte(in)}
	nodes, err := parser.parseOrLiteral()
	if err != nil {
		switch err.(type) {
		case *ExpectedOperand:
			// The query may be unbalanced or malformed as in "(" or
			// "x or" and expects an operand. Try harder to parse it.
			if nodes, err := literalFallbackParser(in); err == nil {
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
		if nodes, err := literalFallbackParser(in); err == nil {
			return nodes, nil
		}
		return nil, errors.Wrap(err, "unbalanced expression")
	}
	if !isSet(parser.heuristics, disambiguated) {
		// Hoist or expressions if this query is potential ambiguous.
		if hoistedNodes, err := Hoist(nodes); err == nil {
			nodes = hoistedNodes
		}
	}

	err = validatePureLiteralPattern(nodes, parser.balanced == 0)
	if err != nil {
		return nil, err
	}
	return newOperator(nodes, And), nil
}

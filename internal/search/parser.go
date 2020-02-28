package search

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Node interface {
	String() string
	node()
}

func (Parameter) node() {}
func (Operator) node()  {}

type Parameter struct {
	Value string
}

type Operator struct {
	Kind     string
	Operands []Node
}

func (node Parameter) String() string {
	return node.Value
}

func (node Operator) String() string {
	var result []string
	for _, child := range node.Operands {
		result = append(result, child.String())
	}
	return fmt.Sprintf("(%s %s)", strings.ToLower(node.Kind), strings.Join(result, " "))
}

func isSpace(c byte) bool {
	return (c == ' ') || (c == '\n') || (c == '\r') || (c == '\t')
}

func skipSpace(buf []byte) int {
	for i, c := range buf {
		if !isSpace(c) {
			return i
		}
	}
	return len(buf)
}

type parser struct {
	buf      []byte
	pos      int
	balanced int
}

func (p *parser) done() bool {
	return p.pos >= len(p.buf)
}

func (p *parser) match(keyword string) bool {
	v, err := p.peek(len(keyword))
	if err != nil {
		return false
	}
	return strings.ToLower(v) == keyword
}

func (p *parser) expect(keyword string) bool {
	if !p.match(keyword) {
		return false
	}
	p.pos += len(keyword)
	return true
}

func (p *parser) isKeyword() bool {
	return p.match("and") || p.match("or") || p.match("(") || p.match(")")
}

func (p *parser) peek(n int) (string, error) {
	if p.pos+n > len(p.buf) {
		return "", io.ErrShortBuffer
	}
	return string(p.buf[p.pos : p.pos+n]), nil
}

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

func (p *parser) scanParameter() (string, error) {
	start := p.pos
	for {
		if p.isKeyword() {
			break
		}
		if p.done() {
			break
		}
		if isSpace(p.buf[p.pos]) {
			break
		}
		p.pos++
	}
	return string(p.buf[start:p.pos]), nil
}

func (p *parser) parseParameterList() ([]Node, error) {
	var nodes []Node
	for {
		if err := p.skipSpaces(); err != nil {
			return nil, err
		}
		if p.done() {
			break
		}
		switch {
		case p.expect("("):
			p.balanced++
			result, err := p.parseOr()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, result...)
		case p.expect(")"):
			p.balanced--
			if len(nodes) == 0 {
				// Return a non-nil node if we parsed "()".
				return []Node{Parameter{Value: ""}}, nil
			}
			return nodes, nil
		case p.match("and"), p.match("or"):
			// Caller advances.
			return nodes, nil
		default:
			value, err := p.scanParameter()
			if err != nil {
				return nil, err
			}
			if value != "" {
				nodes = append(nodes, Parameter{Value: value})
			}
		}
	}
	return nodes, nil
}

func reduce(left, right []Node, kind string) ([]Node, bool) {
	if param, ok := left[0].(Parameter); ok && param.Value == "" {
		// Remove empty string parameter.
		return right, true
	}

	switch right[0].(type) {
	case Operator:
		if kind == right[0].(Operator).Kind {
			// Reduce right node.
			left = append(left, right[0].(Operator).Operands...)
			if len(right) > 1 {
				left = append(left, right[1:]...)
			}
			return left, true
		}
	case Parameter:
		if right[0].(Parameter).Value == "" {
			// Remove empty string parameter.
			if len(right) > 1 {
				return append(left, right[1:]...), true
			}
			return left, true
		}
		if operator, ok := left[0].(Operator); ok && operator.Kind == kind {
			// Reduce left node.
			return append(left[0].(Operator).Operands, right...), true

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

func newOperator(nodes []Node, kind string) []Node {
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

func (p *parser) parseAnd() ([]Node, error) {
	left, err := p.parseParameterList()
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, fmt.Errorf("expected operand at %d", p.pos)
	}
	if !p.expect("and") {
		return left, nil
	}
	right, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	return newOperator(append(left, right...), "and"), nil
}

func (p *parser) parseOr() ([]Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, fmt.Errorf("expected operand at %d", p.pos)
	}
	if !p.expect("or") {
		return left, nil
	}
	right, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	return newOperator(append(left, right...), "or"), nil
}

func Parse(in string) ([]Node, error) {
	if in == "" {
		return nil, nil
	}
	parser := &parser{buf: []byte(in)}
	nodes, err := parser.parseOr()
	if err != nil {
		return nil, err
	}
	if parser.balanced != 0 {
		return nil, errors.New("unbalanced expression")
	}
	return newOperator(nodes, "and"), nil
}

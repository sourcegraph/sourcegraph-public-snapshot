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

type state struct {
	buf      []byte
	pos      int
	balanced int
}

func (s *state) done() bool {
	return s.pos >= len(s.buf)
}

func (s *state) advance(n int) {
	s.pos += n
}

func (s *state) peek(n int) (string, error) {
	if s.pos+n > len(s.buf) {
		return "", io.ErrShortBuffer
	}
	return string(s.buf[s.pos : s.pos+n]), nil
}

func (s *state) skipSpaces() error {
	if s.pos > len(s.buf) {
		return io.ErrShortBuffer
	}

	s.pos += skipSpace(s.buf[s.pos:])
	if s.pos > len(s.buf) {
		return io.ErrShortBuffer
	}
	return nil
}

// reserved returns a reserved string (token) and its value at the current
// position. If no such reserved string exists, it returns the empty string.
// This lets the parser observe syntactic cues and decide to, e.g., keep lexing
// or return control to parsing a different term.
func (s *state) reserved() string {
	if v, err := s.peek(3); err == nil && (v == "AND" || v == "and") {
		return "and"
	}
	if v, err := s.peek(2); err == nil && (v == "OR" || v == "or") {
		return "or"
	}
	if v, err := s.peek(1); err == nil && (v == "(" || v == ")") {
		return v
	}
	return ""
}

func (s *state) scanParameter() (string, error) {
	start := s.pos
	for {
		if s.reserved() != "" {
			break
		}
		if s.done() {
			break
		}
		if isSpace(s.buf[s.pos]) {
			break
		}
		s.pos++
	}
	return string(s.buf[start:s.pos]), nil
}

func (s *state) parseParameterList() ([]Node, error) {
	var nodes []Node
	for {
		if err := s.skipSpaces(); err != nil {
			return nil, err
		}
		if s.done() {
			break
		}
		switch s.reserved() {
		case "(":
			s.balanced++
			s.advance(1)
			result, err := s.parseOr()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, result...)
		case ")":
			s.balanced--
			s.advance(1)
			if len(nodes) == 0 {
				// Return a non-nil node if we parsed "()".
				return []Node{Parameter{Value: ""}}, nil
			}
			return nodes, nil
		case "and", "or":
			// Caller advances.
			return nodes, nil
		default:
			value, err := s.scanParameter()
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

func newAnd(nodes []Node) []Node {
	return newOperator(nodes, "and")
}

func newOr(nodes []Node) []Node {
	return newOperator(nodes, "or")
}

func (s *state) parseAnd() ([]Node, error) {
	left, err := s.parseParameterList()
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, fmt.Errorf("expected operand at %d", s.pos)
	}
	if s.done() || s.reserved() != "and" {
		return newAnd(left), nil
	}
	s.advance(len("and"))
	right, err := s.parseAnd()
	if err != nil {
		return nil, err
	}
	return newAnd(append(left, right...)), nil
}

func (s *state) parseOr() ([]Node, error) {
	left, err := s.parseAnd()
	if err != nil {
		return nil, err
	}
	if left == nil {
		return nil, fmt.Errorf("expected operand at %d", s.pos)
	}
	if s.done() || s.reserved() != "or" {
		return newOr(left), nil
	}

	s.advance(len("or"))
	right, err := s.parseOr()
	if err != nil {
		return nil, err
	}
	return newOr(append(left, right...)), nil
}

func Parse(in string) ([]Node, error) {
	if in == "" {
		return nil, nil
	}
	state := &state{buf: []byte(in)}
	nodes, err := state.parseOr()
	if err != nil {
		return nil, err
	}
	if state.balanced != 0 {
		return nil, errors.New("unbalanced expression")
	}
	return nodes, nil
}

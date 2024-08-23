package ast

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/op"
)

// Parser represents a general purpose AST parser.
type Parser struct {
	internal *parser.Parser

	converter func(interface{}) interface{}
	operator  func(interface{}) (*Node, error)
}

// New creates a new Parser.
func New(input []byte) (*Parser, error) {
	internal, err := parser.New(input)
	if err != nil {
		return nil, err
	}
	return NewFromParser(internal)
}

// SetConverter allows you to add additional (prioritized) converters to the
// parser. e.g. convert aliases to other types or overwrite defaults.
func (ap *Parser) SetConverter(c func(i interface{}) interface{}) {
	ap.converter = c
}

// SetOperator allows you to support additional (prioritized) operators.
// Should return an UnsupportedType error if the given value is not supported.
func (ap *Parser) SetOperator(o func(i interface{}) (*Node, error)) {
	ap.operator = o
}

// NewFromParser creates a new Parser from a parser.Parser. This allows you to
// customize the internal parser. If no customization is needed, use New.
func NewFromParser(p *parser.Parser) (*Parser, error) {
	return &Parser{
		internal: p,
	}, nil
}

// Expect checks whether the buffer contains the given value.
func (ap *Parser) Expect(i interface{}) (*Node, error) {
	i = ConvertAliases(i)
	if ap.converter != nil {
		i = ap.converter(i)
	}

	p := ap.internal
	start := p.Mark()
	if ap.operator != nil {
		// Takes priority over default values. If an unsupported error is
		// returned we can check if one of the predefined types match.
		node, err := ap.operator(i)
		if err == nil {
			return node, nil
		}

		if _, ok := err.(*parser.UnsupportedType); !ok {
			p.Jump(start)
			return node, err
		}
	}
	switch v := i.(type) {
	case rune, string, parser.AnonymousClass:
		// Just check if it matches.
		if _, err := p.Expect(v); err != nil {
			return nil, err
		}

	case ParseNode:
		node, err := v(ap)
		if err != nil {
			p.Jump(start)
			return nil, err
		}
		return node, nil

	case Capture:
		node, err := ap.Expect(v.Value)
		if err != nil {
			p.Jump(start)
			return nil, err
		}
		if node != nil {
			// Return the node.
			if node.Type == -1 {
				node.Type = v.Type
			}
			if len(node.TypeStrings) == 0 {
				node.TypeStrings = v.TypeStrings
			}
			return node, nil
		}

		return &Node{
			Type:        v.Type,
			TypeStrings: v.TypeStrings,
			Value:       p.Slice(start, p.LookBack()),
		}, nil

	case LoopUp:
		i, err := v.Get()
		if err != nil {
			return nil, err
		}
		return ap.Expect(i)

	case op.Not:
		defer p.Jump(start)
		if _, err := ap.Expect(v.Value); err == nil {
			// Return error if match is found.
			return nil, p.ExpectedParseError(v, start, p.LookBack())
		}
	case op.Ensure:
		if n, err := ap.Expect(v.Value); err != nil {
			return n, err
		}
		p.Jump(start)
	case op.And:
		node := &Node{Type: -1}
		for _, i := range v {
			n, err := ap.Expect(i)
			if err != nil {
				p.Jump(start)
				return nil, err
			}
			if n != nil {
				if n.Type == -1 {
					node.Adopt(n)
				} else {
					node.SetLast(n)
				}
			}
		}

		if node.IsParent() {
			// Only return node if it has children.
			return node, nil
		}
	case op.Or:
		// To keep track whether we encountered a valid value, node or not.
		var hit bool
		for _, i := range v {
			node, err := ap.Expect(i)
			if err == nil {
				hit = true
				if node != nil {
					// Return node if found.
					return node, nil
				}
				break
			}
			p.Jump(start)
		}
		if !hit {
			return nil, p.ExpectedParseError(v, start, p.Peek())
		}
	case op.XOr:
		var (
			node *Node
			last *parser.Cursor
		)
		for _, i := range v {
			n, err := ap.Expect(i)
			if err != nil {
				p.Jump(start)
				continue
			}
			if last != nil {
				// We already got a match.
				return nil, p.ExpectedParseError(v, start, last)
			}
			last = p.Mark()
			node = n
		}
		if last == nil {
			return nil, p.ExpectedParseError(v, start, start)
		}
		if node != nil {
			return node, nil
		}

	case op.Range:
		var (
			count int
			last  *parser.Cursor
			node  = &Node{Type: -1}
		)
		for {
			n, err := ap.Expect(v.Value)
			if err != nil {
				break
			}
			if n != nil {
				if n.Type == -1 {
					node.Adopt(n)
				} else {
					node.SetLast(n)
				}
			}
			last = p.LookBack()
			count++

			if v.Max != -1 && count == v.Max {
				// Break if you have parsed the maximum amount of values.
				// This way count will never be larger than v.Max.
				break
			}
		}
		if count < v.Min {
			if last == nil {
				last = start
			}
			return nil, p.ExpectedParseError(v, start, p.Jump(last).Peek())
		}

		if node.IsParent() {
			// Only return node if it has children.
			return node, nil
		}

	default:
		return nil, &parser.UnsupportedType{
			Value: i,
		}
	}

	return nil, nil
}

// ConvertAliases converts various default primitive types to aliases for type
// matching.
func ConvertAliases(i interface{}) interface{} {
	switch v := i.(type) {
	case func(p *Parser) (*Node, error):
		return ParseNode(v)

	default:
		return parser.ConvertAliases(i)
	}
}

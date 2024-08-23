package filter

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/scim2/filter-parser/v2/internal/grammar"
	"github.com/scim2/filter-parser/v2/internal/types"
)

// ParseFilter parses the given raw data as an Expression.
func ParseFilter(raw []byte) (Expression, error) {
	return parseFilter(raw, config{})
}

// ParseFilterNumber parses the given raw data as an Expression with json.Number.
func ParseFilterNumber(raw []byte) (Expression, error) {
	return parseFilter(raw, config{useNumber: true})
}

func parseFilter(raw []byte, c config) (Expression, error) {
	p, err := ast.New(raw)
	if err != nil {
		return nil, err
	}
	node, err := grammar.Filter(p)
	if err != nil {
		return nil, err
	}
	if _, err := p.Expect(parser.EOD); err != nil {
		return nil, err
	}
	return c.parseFilterOr(node)
}

func (p config) parseFilterAnd(node *ast.Node) (Expression, error) {
	if node.Type != typ.FilterAnd {
		return nil, invalidTypeError(typ.FilterAnd, node.Type)
	}

	children := node.Children()
	if len(children) == 0 {
		return nil, invalidLengthError(typ.FilterAnd, 1, 0)
	}

	if len(children) == 1 {
		return p.parseFilterValue(children[0])
	}

	var and LogicalExpression
	for _, node := range children {
		exp, err := p.parseFilterValue(node)
		if err != nil {
			return nil, err
		}
		switch {
		case and.Left == nil:
			and.Left = exp
		case and.Right == nil:
			and.Right = exp
			and.Operator = AND
		default:
			and = LogicalExpression{
				Left: &LogicalExpression{
					Left:     and.Left,
					Right:    and.Right,
					Operator: AND,
				},
				Right:    exp,
				Operator: AND,
			}
		}
	}
	return &and, nil
}

func (p config) parseFilterOr(node *ast.Node) (Expression, error) {
	if node.Type != typ.FilterOr {
		return nil, invalidTypeError(typ.FilterOr, node.Type)
	}

	children := node.Children()
	if len(children) == 0 {
		return nil, invalidLengthError(typ.FilterOr, 1, 0)
	}

	if len(children) == 1 {
		return p.parseFilterAnd(children[0])
	}

	var or LogicalExpression
	for _, node := range children {
		exp, err := p.parseFilterAnd(node)
		if err != nil {
			return nil, err
		}
		switch {
		case or.Left == nil:
			or.Left = exp
		case or.Right == nil:
			or.Right = exp
			or.Operator = OR
		default:
			or = LogicalExpression{
				Left: &LogicalExpression{
					Left:     or.Left,
					Right:    or.Right,
					Operator: OR,
				},
				Right:    exp,
				Operator: OR,
			}
		}
	}
	return &or, nil
}

func (p config) parseFilterValue(node *ast.Node) (Expression, error) {
	switch t := node.Type; t {
	case typ.ValuePath:
		valuePath, err := p.parseValuePath(node)
		if err != nil {
			return nil, err
		}
		return &valuePath, nil
	case typ.AttrExp:
		attrExp, err := p.parseAttrExp(node)
		if err != nil {
			return nil, err
		}
		return &attrExp, nil
	case typ.FilterNot:
		children := node.Children()
		if l := len(children); l != 1 {
			return nil, invalidLengthError(typ.FilterNot, 1, l)
		}

		exp, err := p.parseFilterOr(children[0])
		if err != nil {
			return nil, err
		}
		return &NotExpression{
			Expression: exp,
		}, nil
	case typ.FilterOr:
		return p.parseFilterOr(node)
	default:
		return nil, invalidChildTypeError(typ.FilterAnd, t)
	}
}

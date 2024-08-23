package filter

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/scim2/filter-parser/v2/internal/grammar"
	"github.com/scim2/filter-parser/v2/internal/types"
)

// ParseValuePath parses the given raw data as an ValuePath.
func ParseValuePath(raw []byte) (ValuePath, error) {
	return parseValuePath(raw, config{})
}

// ParseValuePathNumber parses the given raw data as an ValuePath with json.Number.
func ParseValuePathNumber(raw []byte) (ValuePath, error) {
	return parseValuePath(raw, config{useNumber: true})
}

func parseValuePath(raw []byte, c config) (ValuePath, error) {
	p, err := ast.New(raw)
	if err != nil {
		return ValuePath{}, err
	}
	node, err := grammar.ValuePath(p)
	if err != nil {
		return ValuePath{}, err
	}
	if _, err := p.Expect(parser.EOD); err != nil {
		return ValuePath{}, err
	}
	return c.parseValuePath(node)
}

func (p config) parseValueFilter(node *ast.Node) (Expression, error) {
	switch t := node.Type; t {
	case typ.ValueLogExpOr, typ.ValueLogExpAnd:
		children := node.Children()
		if l := len(children); l != 2 {
			return nil, invalidLengthError(node.Type, 2, l)
		}

		left, err := p.parseAttrExp(children[0])
		if err != nil {
			return nil, err
		}
		right, err := p.parseAttrExp(children[1])
		if err != nil {
			return nil, err
		}

		var operator LogicalOperator
		if node.Type == typ.ValueLogExpOr {
			operator = OR
		} else {
			operator = AND
		}

		return &LogicalExpression{
			Left:     &left,
			Right:    &right,
			Operator: operator,
		}, nil
	case typ.AttrExp:
		attrExp, err := p.parseAttrExp(node)
		if err != nil {
			return nil, err
		}
		return &attrExp, nil
	case typ.ValueFilterNot:
		children := node.Children()
		if l := len(children); l != 1 {
			return nil, invalidLengthError(typ.ValueFilterNot, 1, l)
		}

		valueFilter, err := p.parseValueFilter(children[0])
		if err != nil {
			return nil, err
		}
		return &NotExpression{
			Expression: valueFilter,
		}, nil
	default:
		return nil, invalidChildTypeError(typ.ValuePath, t)
	}
}

func (p config) parseValuePath(node *ast.Node) (ValuePath, error) {
	if node.Type != typ.ValuePath {
		return ValuePath{}, invalidTypeError(typ.ValuePath, node.Type)
	}

	children := node.Children()
	if l := len(children); l != 2 {
		return ValuePath{}, invalidLengthError(typ.ValuePath, 2, l)
	}

	attrPath, err := parseAttrPath(children[0])
	if err != nil {
		return ValuePath{}, err
	}

	valueFilter, err := p.parseValueFilter(children[1])
	if err != nil {
		return ValuePath{}, err
	}

	return ValuePath{
		AttributePath: attrPath,
		ValueFilter:   valueFilter,
	}, nil
}

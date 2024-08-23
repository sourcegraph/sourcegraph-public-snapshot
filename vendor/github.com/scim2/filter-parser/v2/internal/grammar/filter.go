package grammar

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/di-wu/parser/op"
	"github.com/scim2/filter-parser/v2/internal/types"
)

func Filter(p *ast.Parser) (*ast.Node, error) {
	return FilterOr(p)
}

func FilterAnd(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(ast.Capture{
		Type:        typ.FilterAnd,
		TypeStrings: typ.Stringer,
		Value: op.And{
			FilterValue,
			op.MinZero(op.And{
				op.MinOne(SP),
				parser.CheckStringCI("and"),
				op.MinOne(SP),
				FilterValue,
			}),
		},
	})
}

func FilterNot(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(ast.Capture{
		Type:        typ.FilterNot,
		TypeStrings: typ.Stringer,
		Value: op.And{
			parser.CheckStringCI("not"),
			op.MinZero(SP),
			FilterParentheses,
		},
	})
}

func FilterOr(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(ast.Capture{
		Type:        typ.FilterOr,
		TypeStrings: typ.Stringer,
		Value: op.And{
			FilterAnd,
			op.MinZero(op.And{
				op.MinOne(SP),
				parser.CheckStringCI("or"),
				op.MinOne(SP),
				FilterAnd,
			}),
		},
	})
}

func FilterParentheses(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(op.And{
		'(',
		op.MinZero(SP),
		FilterOr,
		op.MinZero(SP),
		')',
	})
}

func FilterValue(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(op.Or{
		ValuePath,
		AttrExp,
		FilterNot,
		FilterParentheses,
	})
}

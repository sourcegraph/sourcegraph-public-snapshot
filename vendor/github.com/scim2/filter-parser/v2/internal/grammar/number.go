package grammar

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/di-wu/parser/op"
	"github.com/scim2/filter-parser/v2/internal/types"
)

func Digits(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        typ.Digits,
			TypeStrings: typ.Stringer,
			Value: op.MinOne(
				parser.CheckRuneRange('0', '9'),
			),
		},
	)
}

func Exp(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        typ.Exp,
			TypeStrings: typ.Stringer,
			Value: op.And{
				op.Or{
					"e",
					"E",
				},
				op.Optional(
					Sign,
				),
				Digits,
			},
		},
	)
}

func Frac(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        typ.Frac,
			TypeStrings: typ.Stringer,
			Value: op.And{
				".",
				Digits,
			},
		},
	)
}

func Int(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        typ.Int,
			TypeStrings: typ.Stringer,
			Value: op.Or{
				"0",
				op.And{
					parser.CheckRuneRange('1', '9'),
					op.MinZero(
						parser.CheckRuneRange('0', '9'),
					),
				},
			},
		},
	)
}

func Minus(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        typ.Minus,
			TypeStrings: typ.Stringer,
			Value:       "-",
		},
	)
}

func Number(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        typ.Number,
			TypeStrings: typ.Stringer,
			Value: op.And{
				op.Optional(
					Minus,
				),
				Int,
				op.Optional(
					Frac,
				),
				op.Optional(
					Exp,
				),
			},
		},
	)
}

func Sign(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        typ.Sign,
			TypeStrings: typ.Stringer,
			Value: op.Or{
				"-",
				"+",
			},
		},
	)
}

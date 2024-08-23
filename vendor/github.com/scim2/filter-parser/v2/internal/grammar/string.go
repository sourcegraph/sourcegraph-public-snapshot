package grammar

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/di-wu/parser/op"
	"github.com/scim2/filter-parser/v2/internal/types"
)

func Character(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			Unescaped,
			op.And{
				"\\",
				Escaped,
			},
		},
	)
}

func Escaped(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			"\"",
			"\\",
			"/",
			0x0062,
			0x0066,
			0x006E,
			0x0072,
			0x0074,
			op.And{
				"u",
				op.Repeat(4,
					op.Or{
						parser.CheckRuneRange('0', '9'),
						parser.CheckRuneRange('A', 'F'),
					},
				),
			},
		},
	)
}

func String(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        typ.String,
			TypeStrings: typ.Stringer,
			Value: op.And{
				"\"",
				op.MinZero(
					Character,
				),
				"\"",
			},
		},
	)
}

func Unescaped(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		op.Or{
			parser.CheckRuneRange(0x0020, 0x0021),
			parser.CheckRuneRange(0x0023, 0x005B),
			parser.CheckRuneRange(0x005D, 0x0010FFFF),
		},
	)
}

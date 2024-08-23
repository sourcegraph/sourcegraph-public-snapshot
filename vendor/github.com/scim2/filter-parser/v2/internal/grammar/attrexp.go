package grammar

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/di-wu/parser/op"
	"github.com/scim2/filter-parser/v2/internal/types"
)

func AttrExp(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(ast.Capture{
		Type:        typ.AttrExp,
		TypeStrings: typ.Stringer,
		Value: op.And{
			AttrPath,
			op.MinOne(SP),
			op.Or{
				parser.CheckStringCI("pr"),
				op.And{
					CompareOp,
					op.MinOne(SP),
					CompareValue,
				},
			},
		},
	})
}

func AttrName(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(ast.Capture{
		Type:        typ.AttrName,
		TypeStrings: typ.Stringer,
		Value: op.And{
			op.Optional('$'),
			Alpha,
			op.MinZero(NameChar),
		},
	})
}

func AttrPath(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(ast.Capture{
		Type:        typ.AttrPath,
		TypeStrings: typ.Stringer,
		Value: op.And{
			op.Optional(URI),
			AttrName,
			op.Optional(SubAttr),
		},
	})
}

func CompareOp(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(ast.Capture{
		Type:        typ.CompareOp,
		TypeStrings: typ.Stringer,
		Value: op.Or{
			parser.CheckStringCI("eq"),
			parser.CheckStringCI("ne"),
			parser.CheckStringCI("co"),
			parser.CheckStringCI("sw"),
			parser.CheckStringCI("ew"),
			parser.CheckStringCI("gt"),
			parser.CheckStringCI("lt"),
			parser.CheckStringCI("ge"),
			parser.CheckStringCI("le"),
		},
	})
}

func CompareValue(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(op.Or{False, Null, True, Number, String})
}

func NameChar(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(op.Or{'-', '_', Digit, Alpha})
}

func SubAttr(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(op.And{'.', AttrName})
}

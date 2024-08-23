package grammar

import (
	"github.com/di-wu/parser"
	"github.com/di-wu/parser/ast"
	"github.com/scim2/filter-parser/v2/internal/types"
)

// A boolean has no case sensitivity or uniqueness.
// More info: https://tools.ietf.org/html/rfc7643#section-2.3.2

func False(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        typ.False,
			TypeStrings: typ.Stringer,
			Value:       parser.CheckStringCI("false"),
		},
	)
}

func Null(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        typ.Null,
			TypeStrings: typ.Stringer,
			Value:       parser.CheckStringCI("null"),
		},
	)
}

func True(p *ast.Parser) (*ast.Node, error) {
	return p.Expect(
		ast.Capture{
			Type:        typ.True,
			TypeStrings: typ.Stringer,
			Value:       parser.CheckStringCI("true"),
		},
	)
}
